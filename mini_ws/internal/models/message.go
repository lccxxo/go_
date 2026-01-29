package models

import (
	"time"

	"github.com/google/uuid"
)

// 消息类型
const (
	MessageTypeText    = "text"
	MessageTypeImage   = "image"
	MessageTypeCommand = "command"
)

// 客户端上行消息
type ClientMessage struct {
	MsgID    string                 `json:"msgID"`            // 客户端生成的ID
	Type     string                 `json:"type"`             // 消息类型
	ConvID   string                 `json:"convID"`           // 会话ID
	ToUserID string                 `json:"toUserID"`         // 接受用户ID
	Content  string                 `json:"content"`          // 消息内容
	Extras   map[string]interface{} `json:"extras,omitempty"` // 补充内容
}

// 服务端下行消息
type ServerMessage struct {
	MsgID     string                 `json:"msgID"`            // 服务端生成的ID
	Type      string                 `json:"type"`             // 消息类型
	ConvID    string                 `json:"conv_id"`          // 会话ID
	From      string                 `json:"from"`             // 发送者
	Content   string                 `json:"content"`          // 消息内容
	Seq       int64                  `json:"seq"`              // 会话内序号
	Timestamp int64                  `json:"timestamp"`        // 服务端时间戳
	Extras    map[string]interface{} `json:"extras,omitempty"` // 补充内容
}

// ACK确认消息
type AckMessage struct {
	AckMsgID string `json:"ackMsgID"`
	Type     string `json:"type"`    // 消息类型
	ConvID   string `json:"conv_id"` // 会话ID
}

// 心跳消息
type HeartBeatMessage struct {
	Timestamp int64 `json:"timestamp"` // 时间戳
}

// 连接注册消息
type RegisterMessage struct {
	UserID   string `json:"userID"`
	DeviceID string `json:"deviceID"`
	Token    string `json:"token"`
}

// 服务端消息ID生成
func GenerateMsgID() string {
	return uuid.New().String()
}

// 创建服务端消息
func NewServerMessage(from, convID, msgType, content string, seq int64) *ServerMessage {
	return &ServerMessage{
		MsgID:     GenerateMsgID(),
		Type:      msgType,
		ConvID:    convID,
		From:      from,
		Content:   content,
		Seq:       seq,
		Timestamp: time.Now().UnixMilli(),
		Extras:    make(map[string]interface{}),
	}
}
