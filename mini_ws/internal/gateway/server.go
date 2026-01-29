package gateway

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

// ws升级
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息发送接口
type MessageHandler interface {
	HandleClientMessage(ctx context.Context, client *http.Client, msg []byte) error
}

type Server struct {
	// hub *Hub

}
