package dto

import (
	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
)

// SendMessageRequest 发送私信入参。
type SendMessageRequest struct {
	ToID    string                    `json:"to_id" binding:"required"`
	Type    constants.ChatMessageType `json:"type" binding:"required,oneof=text image sticker"`
	Content string                    `json:"content" binding:"required,min=1,max=2000"`
}

// ChatMessageResponse 私信出参。
type ChatMessageResponse struct {
	ID        string                    `json:"id"`
	FromID    string                    `json:"from_id"`
	ToID      string                    `json:"to_id"`
	Type      constants.ChatMessageType `json:"type"`
	Content   string                    `json:"content"`
	Read      bool                      `json:"read"`
	CreatedAt string                    `json:"created_at"`
}

// ConversationResponse 会话摘要出参。
type ConversationResponse struct {
	PeerID      string `json:"peer_id"`
	PeerName    string `json:"peer_name"`
	PeerAvatar  string `json:"peer_avatar"`
	LastMessage string `json:"last_message"`
	LastTime    string `json:"last_time"`
	UnreadCount int64  `json:"unread_count"`
}

// ToChatMessageResponse 模型转出参。
func ToChatMessageResponse(m *model.ChatMessage) ChatMessageResponse {
	return ChatMessageResponse{
		ID:        m.ID.Hex(),
		FromID:    m.FromID.Hex(),
		ToID:      m.ToID.Hex(),
		Type:      m.Type,
		Content:   m.Content,
		Read:      m.Read,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// WSMessage websocket 消息协议。
type WSMessage struct {
	Type    string `json:"type"`
	ToID    string `json:"to_id"`
	MsgType string `json:"msg_type"`
	Content string `json:"content"`
}
