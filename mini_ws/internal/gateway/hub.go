package gateway

import (
	"sync"

	"go.uber.org/zap"
)

// 管理所有客户端连接
type Hub struct {
	// 所有注册上来的客户端
	clients map[string]*Client

	// 用户到客户端的映射
	userClients map[string]map[string]*Client

	mu     sync.RWMutex
	logger *zap.Logger
}

// 创建Hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		userClients: make(map[string]map[string]*Client),
		logger:      logger,
	}
}

// 客户端注册
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 新增到clients中
	h.clients[client.ID] = client

	// 如果用户映射中不存在 则新增
	if _, ok := h.userClients[client.UserID]; !ok {
		h.userClients[client.UserID] = make(map[string]*Client)
	}
	h.userClients[client.UserID][client.DeviceID] = client

	h.logger.Debug("client registered",
		zap.String("client_id", client.ID),
		zap.Int("total_clients", len(h.clients)))
}

// 注销客户端
func (h *Hub) UnRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 从clients中删除
	delete(h.clients, client.ID)

	// 从用户映射中删除
	if devices, ok := h.userClients[client.UserID]; ok {
		delete(devices, client.DeviceID)
		if len(devices) == 0 {
			delete(h.userClients, client.UserID)
		}
	}

	h.logger.Debug("client unregistered",
		zap.String("client_id", client.ID),
		zap.Int("total_clients", len(h.clients)))
}

// 获取客户端
func (h *Hub) GetClient(clientID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.clients[clientID]
	return client, ok
}

// 获取用户的所有客户端
func (h *Hub) GetUserClients(userID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if devices, ok := h.userClients[userID]; ok {
		for _, client := range devices {
			clients = append(clients, client)
		}
	}

	return clients
}

// 广播消息给用户的所有设备
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	clients := h.GetUserClients(userID)
	for _, client := range clients {
		client.Send(message)
	}
}

// 获取统计信息
func (h *Hub) Stats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"totalClients": len(h.clients),
		"totalUsers":   len(h.userClients),
	}
}
