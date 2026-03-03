// cmd/client 实现一个简单的 SIP UAC（用户代理客户端）。
//
// 功能演示（按顺序执行）：
//  1. 发送 OPTIONS 请求 → 探测服务器能力
//  2. 发送 REGISTER 请求 → 注册到服务器
//  3. 发送 INVITE 请求  → 发起呼叫
//  4. 等待 200 OK
//  5. 发送 ACK          → 确认会话建立
//  6. 等待 3 秒（模拟通话）
//  7. 发送 BYE          → 挂断
//
// 运行方式（先启动 server）：
//
//	go run ./cmd/server
//	go run ./cmd/client -server 127.0.0.1:5060
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lccxxo/go_/mini_sip/internal/dialog"
	"github.com/lccxxo/go_/mini_sip/internal/message"
	"github.com/lccxxo/go_/mini_sip/internal/stack"
	"go.uber.org/zap"
)

var (
	serverAddr = flag.String("server", "127.0.0.1:5060", "SIP server address")
	listenAddr = flag.String("addr", "0.0.0.0:5070", "local SIP listen address (client uses 5070)")
	fromURI    = flag.String("from", "sip:alice@127.0.0.1:5070", "caller URI (From)")
	toURI      = flag.String("to", "sip:bob@127.0.0.1:5060", "callee URI (To)")
)

func main() {
	flag.Parse()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	uac := &UAC{
		logger:     logger,
		serverAddr: *serverAddr,
		responseCh: make(chan *message.Response, 10),
	}
	s, err := stack.NewStack(*listenAddr, uac, logger)
	if err != nil {
		logger.Fatal("start SIP stack", zap.Error(err))
	}
	uac.stack = s
	defer s.Stop()

	logger.Info("SIP UAC started",
		zap.String("local", *listenAddr),
		zap.String("server", *serverAddr),
	)

	// ── 步骤 1：OPTIONS ────────────────────────────────────────────
	fmt.Println("\n[Step 1] Sending OPTIONS to probe server capabilities...")
	if err := uac.sendOptions(); err != nil {
		logger.Error("OPTIONS failed", zap.Error(err))
		os.Exit(1)
	}
	resp := uac.waitResponse(5 * time.Second)
	if resp != nil {
		fmt.Printf("  <- %d %s\n", resp.StatusCode, resp.Reason)
		if allow := resp.Headers.Get(message.HeaderAllow); allow != "" {
			fmt.Printf("     Allow: %s\n", allow)
		}
	}

	time.Sleep(500 * time.Millisecond)

	// ── 步骤 2：REGISTER ───────────────────────────────────────────
	fmt.Println("\n[Step 2] Sending REGISTER...")
	if err := uac.sendRegister(*fromURI, "sip:"+*serverAddr, 3600); err != nil {
		logger.Error("REGISTER failed", zap.Error(err))
		os.Exit(1)
	}
	resp = uac.waitResponse(5 * time.Second)
	if resp != nil {
		fmt.Printf("  <- %d %s\n", resp.StatusCode, resp.Reason)
	}

	time.Sleep(500 * time.Millisecond)

	// ── 步骤 3：INVITE ─────────────────────────────────────────────
	fmt.Println("\n[Step 3] Sending INVITE (initiating call)...")
	inviteReq, err := uac.stack.BuildInviteRequest(*fromURI, *toURI)
	if err != nil {
		logger.Fatal("build INVITE", zap.Error(err))
	}
	if err := uac.stack.SendRequest(inviteReq, *serverAddr); err != nil {
		logger.Fatal("send INVITE", zap.Error(err))
	}
	fmt.Println("  -> INVITE sent")

	// 等待最终响应（2xx 或 4xx+）
	var finalResp *message.Response
	for {
		resp = uac.waitResponse(10 * time.Second)
		if resp == nil {
			fmt.Println("  [timeout] no response")
			os.Exit(1)
		}
		fmt.Printf("  <- %d %s\n", resp.StatusCode, resp.Reason)
		if resp.StatusCode >= 200 {
			finalResp = resp
			break
		}
	}

	if finalResp.StatusCode != message.StatusOK {
		fmt.Printf("  Call rejected: %d %s\n", finalResp.StatusCode, finalResp.Reason)
		os.Exit(0)
	}

	// ── 步骤 4：ACK ────────────────────────────────────────────────
	fmt.Println("\n[Step 4] Sending ACK (confirming dialog)...")
	ack := uac.buildACK(inviteReq, finalResp)
	if err := uac.stack.SendRequest(ack, *serverAddr); err != nil {
		logger.Error("send ACK", zap.Error(err))
	}
	fmt.Println("  -> ACK sent, dialog established!")

	// ── 步骤 5：模拟通话 ───────────────────────────────────────────
	fmt.Println("\n[Step 5] Call in progress (3 seconds)...")
	time.Sleep(3 * time.Second)

	// ── 步骤 6：BYE ────────────────────────────────────────────────
	fmt.Println("\n[Step 6] Sending BYE (hanging up)...")
	bye := uac.buildBYE(inviteReq, finalResp)
	if err := uac.stack.SendRequest(bye, *serverAddr); err != nil {
		logger.Error("send BYE", zap.Error(err))
	}
	resp = uac.waitResponse(5 * time.Second)
	if resp != nil {
		fmt.Printf("  <- %d %s\n", resp.StatusCode, resp.Reason)
	}
	fmt.Println("\nCall completed. Bye!")
}

// UAC 实现 stack.Handler
type UAC struct {
	stack      *stack.Stack
	logger     *zap.Logger
	serverAddr string
	responseCh chan *message.Response
}

func (u *UAC) OnRequest(req *message.Request, tx *dialog.Transaction) {
	// UAC 通常不处理来自服务器的请求（除非是 re-INVITE 等）
	u.logger.Info("unexpected request from server", zap.String("method", string(req.Method)))
}

func (u *UAC) OnResponse(resp *message.Response, req *message.Request) {
	select {
	case u.responseCh <- resp:
	default:
		u.logger.Warn("response channel full, dropping response")
	}
}

func (u *UAC) waitResponse(timeout time.Duration) *message.Response {
	select {
	case resp := <-u.responseCh:
		return resp
	case <-time.After(timeout):
		return nil
	}
}

func (u *UAC) sendOptions() error {
	toURI, err := message.ParseURI("sip:" + u.serverAddr)
	if err != nil {
		return err
	}
	req := message.NewRequest(message.MethodOPTIONS, toURI)
	req.Headers.Set(message.HeaderVia, fmt.Sprintf("SIP/2.0/UDP %s;branch=%s",
		u.stack.LocalAddr(), stack.NewBranch()))
	req.Headers.Set(message.HeaderMaxForwards, "70")
	req.Headers.Set(message.HeaderFrom, fmt.Sprintf("<sip:%s>;tag=%s", u.stack.LocalAddr(), stack.NewTag()))
	req.Headers.Set(message.HeaderTo, fmt.Sprintf("<sip:%s>", u.serverAddr))
	req.Headers.Set(message.HeaderCallID, stack.NewCallID(u.stack.LocalAddr()))
	req.Headers.Set(message.HeaderCSeq, "1 OPTIONS")
	req.Headers.Set(message.HeaderAccept, "application/sdp")
	req.Headers.Set(message.HeaderContentLen, "0")
	return u.stack.SendRequest(req, u.serverAddr)
}

func (u *UAC) sendRegister(aor, registrar string, expires int) error {
	req, err := u.stack.BuildRegisterRequest(aor, registrar, expires)
	if err != nil {
		return err
	}
	return u.stack.SendRequest(req, u.serverAddr)
}

// buildACK 根据 INVITE 请求和 200 OK 响应构造 ACK。
//
// ACK 在 INVITE 事务中的特殊性（RFC 3261 §17.1.1.3）：
//   - ACK 只用于对 INVITE 的最终响应进行确认
//   - 对 2xx 的 ACK：由 TU（Transaction User）直接生成，不经过事务层，
//     使用 To tag（dialog 的 tag），CSeq 方法改为 ACK，序号与 INVITE 相同
func (u *UAC) buildACK(invite *message.Request, resp *message.Response) *message.Request {
	toURI := invite.RequestURI
	ack := message.NewRequest(message.MethodACK, toURI)

	// 复用 INVITE 的 branch（对 2xx 的 ACK 需要新 branch）
	ack.Headers.Set(message.HeaderVia, fmt.Sprintf("SIP/2.0/UDP %s;branch=%s",
		u.stack.LocalAddr(), stack.NewBranch()))
	ack.Headers.Set(message.HeaderMaxForwards, "70")
	ack.Headers.Set(message.HeaderFrom, invite.Headers.Get(message.HeaderFrom))
	// To 使用响应中带 tag 的 To 头域
	ack.Headers.Set(message.HeaderTo, resp.Headers.Get(message.HeaderTo))
	ack.Headers.Set(message.HeaderCallID, invite.Headers.Get(message.HeaderCallID))

	// CSeq：序号与 INVITE 相同，方法改为 ACK
	cseq, _ := message.ParseCSeq(invite.Headers.Get(message.HeaderCSeq))
	if cseq != nil {
		ack.Headers.Set(message.HeaderCSeq, fmt.Sprintf("%d ACK", cseq.Seq))
	}
	ack.Headers.Set(message.HeaderContentLen, "0")
	return ack
}

// buildBYE 构造 BYE 请求终止会话。
//
// BYE 特点：
//   - 使用 dialog 内的 Route set（如果有）
//   - CSeq 序号需要递增（不是从 INVITE 的序号继续，而是 dialog 内的序号）
//   - Request-URI 使用对端的 Contact URI（Remote Target）
func (u *UAC) buildBYE(invite *message.Request, resp *message.Response) *message.Request {
	// 使用对端 Contact 作为 Request-URI（如果有）
	reqURI := invite.RequestURI
	if contact := resp.Headers.Get(message.HeaderContact); contact != "" {
		addr, err := message.ParseAddress(contact)
		if err == nil && addr.URI != nil {
			reqURI = addr.URI
		}
	}

	bye := message.NewRequest(message.MethodBYE, reqURI)
	bye.Headers.Set(message.HeaderVia, fmt.Sprintf("SIP/2.0/UDP %s;branch=%s",
		u.stack.LocalAddr(), stack.NewBranch()))
	bye.Headers.Set(message.HeaderMaxForwards, "70")
	bye.Headers.Set(message.HeaderFrom, invite.Headers.Get(message.HeaderFrom))
	bye.Headers.Set(message.HeaderTo, resp.Headers.Get(message.HeaderTo))
	bye.Headers.Set(message.HeaderCallID, invite.Headers.Get(message.HeaderCallID))

	// CSeq 递增
	cseq, _ := message.ParseCSeq(invite.Headers.Get(message.HeaderCSeq))
	seq := uint32(2)
	if cseq != nil {
		seq = cseq.Seq + 1
	}
	bye.Headers.Set(message.HeaderCSeq, fmt.Sprintf("%d BYE", seq))
	bye.Headers.Set(message.HeaderContentLen, "0")
	return bye
}
