// Package stack 实现 SIP 用户代理核心（RFC 3261 §8）。
//
// SIP 用户代理分为：
//   - UAC (User Agent Client)：发起请求的一方
//   - UAS (User Agent Server)：接收并响应请求的一方
//
// 同一个 UA 在不同呼叫中可同时扮演两种角色。
//
// # 消息流示意
//
// REGISTER 流程（UA 向注册服务器注册）：
//
//	UAC                      Registrar
//	 |--REGISTER------------->|
//	 |<-200 OK----------------|
//
// INVITE 呼叫流程（两个 UA 建立通话）：
//
//	Alice(UAC)               Bob(UAS)
//	 |--INVITE--------------->|
//	 |<-100 Trying------------|
//	 |<-180 Ringing-----------|
//	 |<-200 OK (SDP answer)---|
//	 |--ACK------------------>|
//	 |     (RTP media stream) |
//	 |--BYE------------------>|
//	 |<-200 OK----------------|
package stack

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/lccxxo/go_/mini_sip/internal/dialog"
	"github.com/lccxxo/go_/mini_sip/internal/message"
	"github.com/lccxxo/go_/mini_sip/internal/transport"
	"go.uber.org/zap"
)

// Handler 是上层应用处理 SIP 消息的回调接口。
type Handler interface {
	// OnRequest 处理收到的请求（UAS 角色）。
	OnRequest(req *message.Request, tx *dialog.Transaction)
	// OnResponse 处理收到的响应（UAC 角色）。
	OnResponse(resp *message.Response, req *message.Request)
}

// Stack 是 SIP 协议栈，整合传输层、事务层。
type Stack struct {
	mu        sync.Mutex
	transport *transport.UDPTransport
	handler   Handler
	logger    *zap.Logger

	// 本地信息
	localHost string
	localPort int

	// 事务表：txID -> Transaction
	txMu sync.RWMutex
	txs  map[string]*dialog.Transaction

	// 待匹配的客户端请求：txID -> 原始请求
	pendingMu  sync.RWMutex
	pendingReq map[string]*message.Request

	stopCh chan struct{}
}

// NewStack 创建并启动协议栈。
// listenAddr 格式："0.0.0.0:5060"
func NewStack(listenAddr string, handler Handler, logger *zap.Logger) (*Stack, error) {
	tp, err := transport.NewUDPTransport(listenAddr, logger)
	if err != nil {
		return nil, err
	}

	host, portStr, _ := net.SplitHostPort(listenAddr)
	port, _ := strconv.Atoi(portStr)
	if host == "0.0.0.0" || host == "" {
		// 尝试获取本机 IP
		host = localIP()
	}

	s := &Stack{
		transport:  tp,
		handler:    handler,
		logger:     logger,
		localHost:  host,
		localPort:  port,
		txs:        make(map[string]*dialog.Transaction),
		pendingReq: make(map[string]*message.Request),
		stopCh:     make(chan struct{}),
	}
	tp.Start()
	go s.dispatchLoop()
	return s, nil
}

// Stop 关闭协议栈。
func (s *Stack) Stop() {
	close(s.stopCh)
	s.transport.Stop()
}

// LocalAddr 返回本地监听地址字符串。
func (s *Stack) LocalAddr() string {
	return fmt.Sprintf("%s:%d", s.localHost, s.localPort)
}

// ---- UAC 方法 ----

// SendRequest 发送请求到目标地址，并注册事务。
func (s *Stack) SendRequest(req *message.Request, dst string) error {
	tx, err := dialog.NewTransaction(req, s.logger)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	s.txMu.Lock()
	s.txs[tx.ID] = tx
	s.txMu.Unlock()
	s.pendingMu.Lock()
	s.pendingReq[tx.ID] = req
	s.pendingMu.Unlock()

	data := []byte(req.String())
	if err := s.transport.SendTo(data, dst); err != nil {
		return err
	}
	s.logger.Info("sent request",
		zap.String("method", string(req.Method)),
		zap.String("dst", dst),
		zap.String("txID", tx.ID),
	)
	return nil
}

// SendResponse 发送响应到指定来源地址。
// src 通常是 Via 头域中解析出的地址。
func (s *Stack) SendResponse(resp *message.Response, dst *net.UDPAddr) error {
	data := []byte(resp.String())
	return s.transport.Send(data, dst)
}

// ---- 消息构造工具 ----

// NewBranch 生成符合 RFC 3261 要求的 branch 参数。
// magic cookie "z9hG4bK" + 随机字符串
func NewBranch() string {
	return fmt.Sprintf("z9hG4bK%x", rand.Int63())
}

// NewTag 生成随机 tag。
func NewTag() string {
	return fmt.Sprintf("%x", rand.Int63())
}

// NewCallID 生成全局唯一的 Call-ID。
func NewCallID(host string) string {
	return fmt.Sprintf("%x@%s", rand.Int63(), host)
}

// BuildRegisterRequest 构造 REGISTER 请求。
//
// REGISTER 用于 UA 向注册服务器声明自己的联系地址：
//   - To: 注册的逻辑地址（AOR, Address of Record）
//   - From: 与 To 相同（自注册）
//   - Contact: UA 实际可达的 URI（含 IP:Port）
//   - Expires: 注册有效期（秒），0 表示注销
func (s *Stack) BuildRegisterRequest(aor, registrar string, expires int) (*message.Request, error) {
	reqURI, err := message.ParseURI(registrar)
	if err != nil {
		return nil, fmt.Errorf("parse registrar URI: %w", err)
	}
	req := message.NewRequest(message.MethodREGISTER, reqURI)

	via := fmt.Sprintf("SIP/2.0/UDP %s;branch=%s", s.LocalAddr(), NewBranch())
	req.Headers.Set(message.HeaderVia, via)
	req.Headers.Set(message.HeaderMaxForwards, "70")

	fromTag := NewTag()
	req.Headers.Set(message.HeaderFrom, fmt.Sprintf("<%s>;tag=%s", aor, fromTag))
	req.Headers.Set(message.HeaderTo, fmt.Sprintf("<%s>", aor))
	req.Headers.Set(message.HeaderCallID, NewCallID(s.localHost))
	req.Headers.Set(message.HeaderCSeq, "1 REGISTER")
	req.Headers.Set(message.HeaderContact, fmt.Sprintf("<sip:%s>", s.LocalAddr()))
	req.Headers.Set(message.HeaderExpires, strconv.Itoa(expires))
	req.Headers.Set(message.HeaderUserAgent, "mini_sip/1.0")
	req.Headers.Set(message.HeaderContentLen, "0")
	return req, nil
}

// BuildInviteRequest 构造 INVITE 请求（不含 SDP，仅演示信令）。
//
// INVITE 用于发起会话邀请：
//   - Request-URI: 被叫方 URI
//   - From: 主叫方 URI + tag（UAC 生成）
//   - To: 被叫方 URI（无 tag，dialog 建立后由 UAS 填充）
func (s *Stack) BuildInviteRequest(from, to string) (*message.Request, error) {
	toURI, err := message.ParseURI(to)
	if err != nil {
		return nil, fmt.Errorf("parse To URI: %w", err)
	}
	req := message.NewRequest(message.MethodINVITE, toURI)

	via := fmt.Sprintf("SIP/2.0/UDP %s;branch=%s", s.LocalAddr(), NewBranch())
	req.Headers.Set(message.HeaderVia, via)
	req.Headers.Set(message.HeaderMaxForwards, "70")

	fromTag := NewTag()
	req.Headers.Set(message.HeaderFrom, fmt.Sprintf("<%s>;tag=%s", from, fromTag))
	req.Headers.Set(message.HeaderTo, fmt.Sprintf("<%s>", to))
	req.Headers.Set(message.HeaderCallID, NewCallID(s.localHost))
	req.Headers.Set(message.HeaderCSeq, "1 INVITE")
	req.Headers.Set(message.HeaderContact, fmt.Sprintf("<sip:%s>", s.LocalAddr()))
	req.Headers.Set(message.HeaderUserAgent, "mini_sip/1.0")
	req.Headers.Set(message.HeaderAllow, "INVITE, ACK, BYE, CANCEL, OPTIONS")
	req.Headers.Set(message.HeaderContentLen, "0")
	return req, nil
}

// BuildResponse 从请求构造响应（复制必要头域）。
//
// RFC 3261 §8.2.6: 响应必须复制请求的 Via、From、To、Call-ID、CSeq 头域。
func BuildResponse(req *message.Request, code int, localTag string) *message.Response {
	resp := message.NewResponse(code)

	// 复制 Via（所有值，用于逐跳路由响应）
	for _, v := range req.Headers.GetAll(message.HeaderVia) {
		resp.Headers.Add(message.HeaderVia, v)
	}
	resp.Headers.Set(message.HeaderFrom, req.Headers.Get(message.HeaderFrom))
	// To 头域：若是 dialog 建立响应，追加 tag
	to := req.Headers.Get(message.HeaderTo)
	if localTag != "" && !containsTag(to) {
		to = to + ";tag=" + localTag
	}
	resp.Headers.Set(message.HeaderTo, to)
	resp.Headers.Set(message.HeaderCallID, req.Headers.Get(message.HeaderCallID))
	resp.Headers.Set(message.HeaderCSeq, req.Headers.Get(message.HeaderCSeq))
	return resp
}

func containsTag(s string) bool {
	for _, seg := range splitParams(s) {
		if len(seg) >= 3 && seg[:4] == "tag=" {
			return true
		}
	}
	return false
}

func splitParams(s string) []string {
	var out []string
	for _, p := range splitAfterSemi(s) {
		out = append(out, p)
	}
	return out
}

func splitAfterSemi(s string) []string {
	parts := []string{}
	for _, p := range s[indexOf(s, ';')+1:] {
		_ = p
	}
	// 简化实现：直接 split
	for i, p := range (func() []string {
		result := []string{}
		start := 0
		for j := 0; j < len(s); j++ {
			if s[j] == ';' {
				result = append(result, s[start:j])
				start = j + 1
			}
		}
		result = append(result, s[start:])
		return result
	})() {
		if i > 0 {
			parts = append(parts, p)
		}
	}
	return parts
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// ---- 内部分发 ----

func (s *Stack) dispatchLoop() {
	for {
		select {
		case <-s.stopCh:
			return
		case raw, ok := <-s.transport.Recv():
			if !ok {
				return
			}
			go s.handleRaw(raw)
		}
	}
}

func (s *Stack) handleRaw(raw *transport.Message) {
	msg, err := message.Parse(raw.Data)
	if err != nil {
		s.logger.Warn("failed to parse SIP message", zap.Error(err),
			zap.String("src", raw.Source.String()))
		return
	}

	switch m := msg.(type) {
	case *message.Request:
		s.handleRequest(m, raw.Source)
	case *message.Response:
		s.handleResponse(m)
	}
}

func (s *Stack) handleRequest(req *message.Request, src *net.UDPAddr) {
	s.logger.Info("received request",
		zap.String("method", string(req.Method)),
		zap.String("src", src.String()),
	)
	// 创建服务端事务
	tx, err := dialog.NewTransaction(req, s.logger)
	if err != nil {
		s.logger.Error("create server transaction", zap.Error(err))
		return
	}
	s.txMu.Lock()
	s.txs[tx.ID] = tx
	s.txMu.Unlock()

	if s.handler != nil {
		s.handler.OnRequest(req, tx)
	}
}

func (s *Stack) handleResponse(resp *message.Response) {
	s.logger.Info("received response",
		zap.Int("code", resp.StatusCode),
		zap.String("reason", resp.Reason),
	)
	// 匹配事务
	via := resp.Headers.Get(message.HeaderVia)
	if via == "" {
		s.logger.Warn("response missing Via header")
		return
	}
	parsed, err := message.ParseVia(via)
	if err != nil {
		s.logger.Warn("parse Via in response", zap.Error(err))
		return
	}
	branch := parsed.Params["branch"]
	cseq, _ := message.ParseCSeq(resp.Headers.Get(message.HeaderCSeq))
	txID := ""
	if cseq != nil {
		txID = branch + ":" + string(cseq.Method)
	}

	s.txMu.RLock()
	tx := s.txs[txID]
	s.txMu.RUnlock()

	if tx != nil {
		tx.HandleResponse(resp)
	}

	// 通知上层
	s.pendingMu.RLock()
	req := s.pendingReq[txID]
	s.pendingMu.RUnlock()

	if s.handler != nil {
		s.handler.OnResponse(resp, req)
	}
}

// localIP 获取本机非回环 IP。
func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// init 初始化随机种子
func init() {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck
}
