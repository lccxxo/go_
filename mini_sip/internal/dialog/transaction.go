// Package dialog 实现 SIP 事务层和对话层（RFC 3261 §17 & §12）。
//
// # 事务 (Transaction)
//
// 事务是一次请求和所有响应的集合，分为客户端事务（CTX）和服务端事务（STX）。
// 事务由 branch 参数唯一标识（z9hG4bK... magic cookie）。
//
// 状态机（以 INVITE 为例）：
//
// 客户端 INVITE 事务：
//
//	         INVITE sent
//	            │
//	        [Calling]
//	            │ 1xx
//	        [Proceeding]
//	            │ 2xx (交给 TU 直接处理，不经过事务层)
//	            │ 3xx-6xx
//	        [Completed]──ACK sent──>[Terminated]
//
// 服务端 INVITE 事务：
//
//	         INVITE rcvd
//	            │
//	        [Proceeding] ──1xx sent──>
//	            │ 2xx / 3xx-6xx sent
//	        [Completed] ──ACK rcvd──>[Confirmed]──>[Terminated]
//
// # 对话 (Dialog)
//
// 对话是两个 UA 之间的点对点 SIP 关系，由三元组唯一标识：
//   - Call-ID
//   - 本地 tag（From tag）
//   - 远端 tag（To tag）
//
// 对话在收到 2xx 或可靠 1xx 响应后建立，BYE 后终止。
package dialog

import (
	"fmt"
	"sync"
	"time"

	"github.com/lccxxo/go_/mini_sip/internal/message"
	"go.uber.org/zap"
)

// TxState 事务状态
type TxState int

const (
	TxStateCalling     TxState = iota // 客户端：已发送请求，等待响应
	TxStateProceeding                 // 收到 1xx
	TxStateCompleted                  // 收到最终响应（3xx-6xx），等待 ACK / Timer D
	TxStateConfirmed                  // INVITE 服务端：已收到 ACK
	TxStateTerminated                 // 事务结束
)

func (s TxState) String() string {
	switch s {
	case TxStateCalling:
		return "Calling"
	case TxStateProceeding:
		return "Proceeding"
	case TxStateCompleted:
		return "Completed"
	case TxStateConfirmed:
		return "Confirmed"
	case TxStateTerminated:
		return "Terminated"
	}
	return "Unknown"
}

// Transaction 表示一个 SIP 事务。
type Transaction struct {
	mu        sync.RWMutex
	ID        string          // branch 参数作为事务 ID
	Method    message.Method  // 原始请求方法
	State     TxState
	Request   *message.Request
	Responses []*message.Response
	logger    *zap.Logger
	done      chan struct{}
}

// NewTransaction 创建新事务。
func NewTransaction(req *message.Request, logger *zap.Logger) (*Transaction, error) {
	via := req.Headers.Get(message.HeaderVia)
	if via == "" {
		return nil, fmt.Errorf("request missing Via header")
	}
	parsed, err := message.ParseVia(via)
	if err != nil {
		return nil, fmt.Errorf("parse Via: %w", err)
	}
	branch, ok := parsed.Params["branch"]
	if !ok || branch == "" {
		return nil, fmt.Errorf("Via missing branch parameter")
	}
	// 事务 ID = branch + method（非 INVITE 的 ACK 使用 INVITE 的 branch）
	txID := branch + ":" + string(req.Method)

	return &Transaction{
		ID:      txID,
		Method:  req.Method,
		State:   TxStateCalling,
		Request: req,
		logger:  logger,
		done:    make(chan struct{}),
	}, nil
}

// HandleResponse 将响应送入事务状态机。
func (tx *Transaction) HandleResponse(resp *message.Response) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	tx.Responses = append(tx.Responses, resp)
	code := resp.StatusCode

	switch tx.State {
	case TxStateCalling, TxStateProceeding:
		if code >= 100 && code < 200 {
			tx.State = TxStateProceeding
			tx.logger.Info("tx provisional response",
				zap.String("id", tx.ID), zap.Int("code", code))
		} else if code >= 200 {
			tx.State = TxStateCompleted
			tx.logger.Info("tx final response",
				zap.String("id", tx.ID), zap.Int("code", code))
			// Timer D: 32s (UDP) 后进入 Terminated
			go tx.timerD(32 * time.Second)
		}
	}
}

// HandleACK 服务端 INVITE 事务收到 ACK 后调用。
func (tx *Transaction) HandleACK() {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.State == TxStateCompleted {
		tx.State = TxStateConfirmed
		tx.logger.Info("tx confirmed (ACK received)", zap.String("id", tx.ID))
		go tx.timerI(0) // Timer I: UDP 下等待重传的 ACK
	}
}

// Done 返回事务结束信号通道。
func (tx *Transaction) Done() <-chan struct{} {
	return tx.done
}

func (tx *Transaction) timerD(d time.Duration) {
	time.Sleep(d)
	tx.mu.Lock()
	tx.State = TxStateTerminated
	tx.mu.Unlock()
	select {
	case <-tx.done:
	default:
		close(tx.done)
	}
	tx.logger.Info("tx terminated (Timer D)", zap.String("id", tx.ID))
}

func (tx *Transaction) timerI(d time.Duration) {
	time.Sleep(d)
	tx.mu.Lock()
	tx.State = TxStateTerminated
	tx.mu.Unlock()
	select {
	case <-tx.done:
	default:
		close(tx.done)
	}
	tx.logger.Info("tx terminated (Timer I)", zap.String("id", tx.ID))
}

// ---- Dialog ----

// DialogState 对话状态
type DialogState int

const (
	DialogStateEarly      DialogState = iota // 收到 1xx，对话尚未确认
	DialogStateConfirmed                     // 收到 2xx，对话已建立
	DialogStateTerminated                    // BYE 后结束
)

func (s DialogState) String() string {
	switch s {
	case DialogStateEarly:
		return "Early"
	case DialogStateConfirmed:
		return "Confirmed"
	case DialogStateTerminated:
		return "Terminated"
	}
	return "Unknown"
}

// DialogID 对话标识三元组
type DialogID struct {
	CallID   string
	LocalTag string
	RemoteTag string
}

func (d DialogID) String() string {
	return fmt.Sprintf("%s;local=%s;remote=%s", d.CallID, d.LocalTag, d.RemoteTag)
}

// Dialog 表示一个 SIP 对话（两个 UA 之间的逻辑关系）。
type Dialog struct {
	mu         sync.RWMutex
	ID         DialogID
	State      DialogState
	LocalURI   *message.URI
	RemoteURI  *message.URI
	RemoteTarget *message.URI // Contact 中的 URI，下一跳目标
	RouteSet   []string      // Record-Route 构建的路由集
	LocalCSeq  uint32
	RemoteCSeq uint32
	logger     *zap.Logger
}

// NewDialogFromRequest 从收到的 INVITE 创建对话（服务端视角）。
func NewDialogFromRequest(req *message.Request, localTag string, logger *zap.Logger) (*Dialog, error) {
	callID := req.Headers.Get(message.HeaderCallID)
	fromAddr, err := message.ParseAddress(req.Headers.Get(message.HeaderFrom))
	if err != nil {
		return nil, fmt.Errorf("parse From: %w", err)
	}
	cseq, err := message.ParseCSeq(req.Headers.Get(message.HeaderCSeq))
	if err != nil {
		return nil, fmt.Errorf("parse CSeq: %w", err)
	}

	d := &Dialog{
		ID: DialogID{
			CallID:    callID,
			LocalTag:  localTag,
			RemoteTag: fromAddr.Tag,
		},
		State:      DialogStateEarly,
		RemoteCSeq: cseq.Seq,
		logger:     logger,
	}
	if req.RequestURI != nil {
		d.LocalURI = req.RequestURI.Clone()
	}
	if fromAddr.URI != nil {
		d.RemoteURI = fromAddr.URI.Clone()
	}
	return d, nil
}

// Confirm 对话确认（收到 2xx 后调用）。
func (d *Dialog) Confirm(resp *message.Response) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	toAddr, err := message.ParseAddress(resp.Headers.Get(message.HeaderTo))
	if err != nil {
		return fmt.Errorf("parse To in 2xx: %w", err)
	}
	if d.ID.RemoteTag == "" {
		d.ID.RemoteTag = toAddr.Tag
	}
	// 更新 Remote-Target 为 Contact URI
	if contact := resp.Headers.Get(message.HeaderContact); contact != "" {
		addr, err := message.ParseAddress(contact)
		if err == nil && addr.URI != nil {
			d.RemoteTarget = addr.URI.Clone()
		}
	}
	d.State = DialogStateConfirmed
	d.logger.Info("dialog confirmed", zap.String("id", d.ID.String()))
	return nil
}

// Terminate 终止对话。
func (d *Dialog) Terminate() {
	d.mu.Lock()
	d.State = DialogStateTerminated
	d.mu.Unlock()
	d.logger.Info("dialog terminated", zap.String("id", d.ID.String()))
}

// NextLocalCSeq 获取并递增本地 CSeq。
func (d *Dialog) NextLocalCSeq() uint32 {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.LocalCSeq++
	return d.LocalCSeq
}
