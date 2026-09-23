package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/middleware"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
)

// ChatHandler 私信接口层 + websocket 网关。
type ChatHandler struct {
	svc      *service.ChatService
	logger   *slog.Logger
	upgrader websocket.Upgrader
}

// NewChatHandler 构造注入。
func NewChatHandler(svc *service.ChatService, logger *slog.Logger) *ChatHandler {
	return &ChatHandler{
		svc:    svc,
		logger: logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Send 发送私信。
func (h *ChatHandler) Send(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	msg, err := h.svc.SendMessage(c.Request.Context(), userID, req, c.ClientIP())
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToChatMessageResponse(msg))
}

// Conversation 会话消息列表。
func (h *ChatHandler) Conversation(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	peerID, err := dto.ParseObjectID(c.Param("peer_id"))
	if err != nil {
		c.Error(util.Validation("对方 ID 不合法", err))
		return
	}
	page, pageSize := pageParams(c)
	msgs, total, err := h.svc.ListConversation(c.Request.Context(), userID, peerID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	out := make([]dto.ChatMessageResponse, 0, len(msgs))
	for i := range msgs {
		out = append(out, dto.ToChatMessageResponse(&msgs[i]))
	}
	util.Page(c, out, total, page, pageSize)
}

// Conversations 会话列表。
func (h *ChatHandler) Conversations(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	convs, err := h.svc.ListConversations(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, convs)
}

// WS 实时聊天 websocket 网关（token 通过查询参数传入）。
func (h *ChatHandler) WS(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		util.Fail(c, http.StatusInternalServerError, constants.CodeWSUpgradeError, "websocket 升级失败")
		return
	}
	client := &service.Client{UserID: userID, Conn: conn, Send: make(chan []byte, 64)}
	hub := h.svc.Hub()
	hub.Register(client)
	go h.writePump(client)
	h.readPump(client, hub)
}

func contextWithTimeout(userID primitive.ObjectID) context.Context {
	return context.Background()
}

func (h *ChatHandler) readPump(client *service.Client, hub *service.Hub) {
	defer func() {
		hub.Unregister(client)
		_ = client.Conn.Close()
	}()
	client.Conn.SetReadLimit(4096)
	_ = client.Conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		_ = client.Conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		return nil
	})
	for {
		var msg dto.WSMessage
		if err := client.Conn.ReadJSON(&msg); err != nil {
			return
		}
		toID, err := primitive.ObjectIDFromHex(msg.ToID)
		if err != nil {
			continue
		}
		req := dto.SendMessageRequest{ToID: msg.ToID, Type: constants.ChatMessageType(msg.MsgType), Content: msg.Content}
		if _, err := h.svc.SendMessage(contextWithTimeout(client.UserID), client.UserID, req, ""); err != nil {
			continue
		}
		_ = toID
	}
}

func (h *ChatHandler) writePump(client *service.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()
	for {
		select {
		case payload, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
