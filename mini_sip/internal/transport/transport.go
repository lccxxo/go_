// Package transport 实现 SIP 传输层（RFC 3261 §18）。
//
// SIP 支持三种传输协议：
//   - UDP（无连接，最常用）
//   - TCP（有连接，大消息时使用）
//   - TLS（加密，对应 sips:）
//
// 本实现聚焦 UDP，便于学习 SIP 基础。
// UDP 特点：
//   - 无连接，每条消息独立发送
//   - SIP 消息必须 < 1300 bytes（避免 IP 分片），超出应切换 TCP
//   - 需要在应用层处理重传（事务层负责）
package transport

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

const (
	maxUDPPacket = 65507          // UDP 单包最大字节数
	readTimeout  = 5 * time.Second
)

// Message 封装从网络收到的原始 SIP 消息。
type Message struct {
	Data   []byte      // 原始字节
	Source *net.UDPAddr // 来源地址（用于发送响应）
}

// UDPTransport 监听 UDP 端口并收发 SIP 消息。
type UDPTransport struct {
	conn    *net.UDPConn
	addr    *net.UDPAddr
	logger  *zap.Logger
	recvCh  chan *Message
	stopCh  chan struct{}
}

// NewUDPTransport 创建并绑定 UDP 传输层。
// addr 格式："0.0.0.0:5060"（SIP 默认端口 5060）
func NewUDPTransport(addr string, logger *zap.Logger) (*UDPTransport, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve UDP addr %q: %w", addr, err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("listen UDP %q: %w", addr, err)
	}
	logger.Info("UDP transport listening", zap.String("addr", addr))
	return &UDPTransport{
		conn:   conn,
		addr:   udpAddr,
		logger: logger,
		recvCh: make(chan *Message, 64),
		stopCh: make(chan struct{}),
	}, nil
}

// Start 开始后台接收循环，收到的消息可从 Recv() 读取。
func (t *UDPTransport) Start() {
	go t.readLoop()
}

// Recv 返回接收通道（只读）。
func (t *UDPTransport) Recv() <-chan *Message {
	return t.recvCh
}

// Send 将数据发送到指定 UDP 地址。
func (t *UDPTransport) Send(data []byte, dst *net.UDPAddr) error {
	if len(data) > maxUDPPacket {
		return fmt.Errorf("message too large for UDP: %d bytes", len(data))
	}
	_, err := t.conn.WriteToUDP(data, dst)
	if err != nil {
		return fmt.Errorf("send to %s: %w", dst, err)
	}
	t.logger.Debug("sent SIP message", zap.String("dst", dst.String()), zap.Int("bytes", len(data)))
	return nil
}

// SendTo 将数据发送到 host:port 字符串指定的目标。
func (t *UDPTransport) SendTo(data []byte, hostPort string) error {
	dst, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return fmt.Errorf("resolve dst %q: %w", hostPort, err)
	}
	return t.Send(data, dst)
}

// LocalAddr 返回本地监听地址。
func (t *UDPTransport) LocalAddr() *net.UDPAddr {
	return t.addr
}

// Stop 关闭传输层。
func (t *UDPTransport) Stop() {
	close(t.stopCh)
	t.conn.Close()
}

func (t *UDPTransport) readLoop() {
	buf := make([]byte, maxUDPPacket)
	for {
		// 设置读超时，以便能定期检查 stopCh
		t.conn.SetReadDeadline(time.Now().Add(readTimeout))
		n, src, err := t.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-t.stopCh:
				t.logger.Info("UDP transport stopped")
				close(t.recvCh)
				return
			default:
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				t.logger.Error("UDP read error", zap.Error(err))
				continue
			}
		}
		// 拷贝数据（buf 会被下次读取覆盖）
		data := make([]byte, n)
		copy(data, buf[:n])
		msg := &Message{Data: data, Source: src}
		select {
		case t.recvCh <- msg:
		default:
			t.logger.Warn("receive buffer full, dropping message", zap.String("src", src.String()))
		}
		t.logger.Debug("received SIP message",
			zap.String("src", src.String()),
			zap.Int("bytes", n),
		)
	}
}
