// cmd/server 实现一个简单的 SIP UAS（用户代理服务器）。
//
// 功能：
//   - 监听 UDP:5060
//   - 响应 OPTIONS 请求（返回 200 OK + 支持的方法列表）
//   - 响应 REGISTER 请求（返回 200 OK，模拟注册成功）
//   - 响应 INVITE 请求（依次返回 100 Trying -> 180 Ringing -> 200 OK）
//   - 响应 BYE 请求（返回 200 OK，终止会话）
//
// 运行方式：
//
//	go run ./cmd/server
//
// 用 sngrep 或 Wireshark 抓包观察 SIP 消息格式。
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lccxxo/go_/mini_sip/internal/dialog"
	"github.com/lccxxo/go_/mini_sip/internal/message"
	"github.com/lccxxo/go_/mini_sip/internal/stack"
	"go.uber.org/zap"
)

var listenAddr = flag.String("addr", "0.0.0.0:5060", "SIP UDP listen address")

func main() {
	flag.Parse()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	uas := &UAS{logger: logger}
	srv, err := stack.NewStack(*listenAddr, uas, logger)
	if err != nil {
		logger.Fatal("start SIP stack", zap.Error(err))
	}
	uas.stack = srv

	logger.Info("SIP UAS started", zap.String("addr", *listenAddr))
	logger.Info("waiting for SIP messages... (Ctrl+C to stop)")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	srv.Stop()
	logger.Info("SIP UAS stopped")
}

// UAS 实现 stack.Handler 接口，处理各类请求。
type UAS struct {
	stack  *stack.Stack
	logger *zap.Logger
}

func (u *UAS) OnRequest(req *message.Request, tx *dialog.Transaction) {
	u.logger.Info("handling request",
		zap.String("method", string(req.Method)),
		zap.String("from", req.Headers.Get(message.HeaderFrom)),
	)

	// 从 Via 解析来源地址，用于发送响应
	src := u.extractResponseDst(req)
	if src == nil {
		u.logger.Error("cannot determine response destination")
		return
	}

	switch req.Method {
	case message.MethodOPTIONS:
		u.handleOptions(req, src)
	case message.MethodREGISTER:
		u.handleRegister(req, src)
	case message.MethodINVITE:
		u.handleInvite(req, src)
	case message.MethodBYE:
		u.handleBye(req, src)
	case message.MethodACK:
		// ACK 不需要响应，记录日志即可
		u.logger.Info("ACK received, dialog confirmed")
	case message.MethodCANCEL:
		u.handleCancel(req, src)
	default:
		u.send(stack.BuildResponse(req, message.StatusMethodNotAllowed, ""), src)
	}
}

func (u *UAS) OnResponse(resp *message.Response, req *message.Request) {
	// UAS 一般不发起请求，此处忽略
}

// handleOptions 响应 OPTIONS：返回 200 OK 和支持的方法列表。
//
// OPTIONS 用于探测对端能力，SIP 代理服务器也常用它做心跳检测。
func (u *UAS) handleOptions(req *message.Request, src *net.UDPAddr) {
	resp := stack.BuildResponse(req, message.StatusOK, "server")
	resp.Headers.Set(message.HeaderAllow, "INVITE, ACK, BYE, CANCEL, OPTIONS, REGISTER")
	resp.Headers.Set("Accept", "application/sdp")
	resp.Headers.Set("Accept-Encoding", "identity")
	resp.Headers.Set("Accept-Language", "en")
	u.send(resp, src)
	u.logger.Info("OPTIONS handled: 200 OK")
}

// handleRegister 响应 REGISTER：模拟注册成功。
//
// 真实场景中，注册服务器需要维护一个位置数据库（Location Database），
// 将 AOR（sip:alice@example.com）映射到联系地址（sip:alice@192.168.1.5:5060）。
func (u *UAS) handleRegister(req *message.Request, src *net.UDPAddr) {
	expires := req.Headers.Get(message.HeaderExpires)
	if expires == "" {
		expires = "3600"
	}
	resp := stack.BuildResponse(req, message.StatusOK, "server")
	resp.Headers.Set(message.HeaderContact, req.Headers.Get(message.HeaderContact))
	resp.Headers.Set(message.HeaderExpires, expires)
	resp.Headers.Set("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	u.send(resp, src)
	u.logger.Info("REGISTER handled: 200 OK",
		zap.String("contact", req.Headers.Get(message.HeaderContact)),
		zap.String("expires", expires),
	)
}

// handleInvite 处理 INVITE：模拟振铃后接听。
//
// 三步响应流程：
//  1. 100 Trying   - 已收到请求，正在处理（抑制 UAC 重传）
//  2. 180 Ringing  - 被叫正在振铃（UI 可播放回铃音）
//  3. 200 OK       - 接听（应含 SDP answer，此处省略）
func (u *UAS) handleInvite(req *message.Request, src *net.UDPAddr) {
	localTag := stack.NewTag()

	// 1. 100 Trying（不含 To tag，因为 dialog 尚未建立）
	trying := stack.BuildResponse(req, message.StatusTrying, "")
	u.send(trying, src)
	u.logger.Info("INVITE -> 100 Trying")

	// 模拟处理延迟
	time.Sleep(200 * time.Millisecond)

	// 2. 180 Ringing（Early Dialog：含 To tag）
	ringing := stack.BuildResponse(req, message.StatusRinging, localTag)
	ringing.Headers.Set(message.HeaderContact, fmt.Sprintf("<sip:%s>", u.stack.LocalAddr()))
	u.send(ringing, src)
	u.logger.Info("INVITE -> 180 Ringing")

	// 模拟振铃 1s
	time.Sleep(1 * time.Second)

	// 3. 200 OK（含 To tag，Dialog 建立；应含 SDP，此处演示信令只回空 body）
	ok := stack.BuildResponse(req, message.StatusOK, localTag)
	ok.Headers.Set(message.HeaderContact, fmt.Sprintf("<sip:%s>", u.stack.LocalAddr()))
	ok.Headers.Set(message.HeaderAllow, "INVITE, ACK, BYE, CANCEL, OPTIONS")
	// 在真实场景中这里需要填写 SDP（Session Description Protocol）body：
	// ok.Headers.Set(message.HeaderContentType, "application/sdp")
	// ok.Body = []byte("v=0\r\no=...\r\n...")
	u.send(ok, src)
	u.logger.Info("INVITE -> 200 OK (dialog established)", zap.String("tag", localTag))
}

// handleBye 响应 BYE：终止会话。
func (u *UAS) handleBye(req *message.Request, src *net.UDPAddr) {
	resp := stack.BuildResponse(req, message.StatusOK, "")
	u.send(resp, src)
	u.logger.Info("BYE handled: 200 OK (session terminated)")
}

// handleCancel 响应 CANCEL：取消未完成的 INVITE。
func (u *UAS) handleCancel(req *message.Request, src *net.UDPAddr) {
	resp := stack.BuildResponse(req, message.StatusOK, "")
	u.send(resp, src)
	u.logger.Info("CANCEL handled: 200 OK")
}

// extractResponseDst 从 Via 头域提取响应目标地址。
//
// RFC 3261 §18.2.2: 响应发送规则：
//   - 若 Via 含 maddr 参数，发往 maddr
//   - 若 Via 含 received 参数，发往 received（NAT 穿透）
//   - 否则发往 sent-by（Via 中的 host:port）
func (u *UAS) extractResponseDst(req *message.Request) *net.UDPAddr {
	via := req.Headers.Get(message.HeaderVia)
	if via == "" {
		return nil
	}
	parsed, err := message.ParseVia(via)
	if err != nil {
		u.logger.Error("parse Via", zap.Error(err))
		return nil
	}
	target := parsed.SentBy
	if received, ok := parsed.Params["received"]; ok && received != "" {
		// 使用 received 参数覆盖 host 部分
		host, _, _ := net.SplitHostPort(target)
		if host == "" {
			target = received
		} else {
			_, port, _ := net.SplitHostPort(target)
			if port != "" {
				target = received + ":" + port
			} else {
				target = received
			}
		}
	}
	// 如果没有端口，默认 5060
	if _, _, err := net.SplitHostPort(target); err != nil {
		target = target + ":5060"
	}
	addr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		u.logger.Error("resolve response dst", zap.Error(err))
		return nil
	}
	return addr
}

func (u *UAS) send(resp *message.Response, dst *net.UDPAddr) {
	if err := u.stack.SendResponse(resp, dst); err != nil {
		u.logger.Error("send response", zap.Error(err))
	}
}
