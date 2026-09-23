package router

import (
	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/handler"
)

// RegisterChatAuth 私信路由（含 websocket）。
func RegisterChatAuth(r *gin.RouterGroup, h *handler.ChatHandler) {
	r.GET("/chat/ws", h.WS)
	r.POST("/chat/messages", h.Send)
	r.GET("/chat/conversations", h.Conversations)
	r.GET("/chat/conversations/:peer_id", h.Conversation)
}
