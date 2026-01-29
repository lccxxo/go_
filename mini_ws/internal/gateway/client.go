package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	// 心跳间隔
	heartbeatInterval = 30 * time.Second
	// 写超时
	writeTimeout = 10 * time.Second
	// 读超时
	readTimeout = 60 * time.Second
	// 最大消息大小
	maxMessageSize = 1024 * 1024
	// 发送通道缓冲
	sendChannelSize = 256
)

// 一个Client表示一个websocket链接
type Client struct {
	ID       string
	UserID   string
	DeviceID string
	Conn     *websocket.Conn
	Hub      *Hub
	Presence PresenceManager
	// 发送通道
	sendChan chan []byte

	// 限流器
	rateLimiter *rate.Limiter

	// 状态
	mu       sync.RWMutex
	isAlive  bool
	lastPing time.Time

	// 日志
	logger *zap.Logger
}

// 状态管理接口
type PresenceManager interface {
	Register(ctx context.Context, userID, deviceID, clientID string) error
	Unregister(ctx context.Context, userID, deviceID string) error
	GetUserClients(ctx context.Context, userID string) ([]string, error)
}

func NewClient(conn *websocket.Conn, userID, deviceID string, hub *Hub, presence PresenceManager, logger *zap.Logger) *Client {
	return &Client{
		ID:          fmt.Sprintf("%s:%s", userID, deviceID),
		UserID:      userID,
		DeviceID:    deviceID,
		Conn:        conn,
		Hub:         hub,
		Presence:    presence,
		sendChan:    make(chan []byte, sendChannelSize),
		rateLimiter: rate.NewLimiter(rate.Limit(100), 10),
		isAlive:     true,
		logger:      logger,
	}
}

// 处理读取消息
func (c *Client) ReadPump() {}

// 处理写入消息
func (c *Client) WritePump() {}

// 发送消息到客户端
func (c *Client) Send(message []byte) error {
	return nil
}

// 关闭连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isAlive {
		c.isAlive = false
		c.Conn.Close()
		close(c.sendChan)
	}
}

func (c *Client) handleMessage(data []byte) {
	// todo 对不同类型的消息进行不同的处理
}
