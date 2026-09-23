package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"encoding/json"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/repository"
	"github.com/petsocial/petsocial/internal/util"
)

// Client websocket 客户端连接。
type Client struct {
	UserID primitive.ObjectID
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub 实时聊天中心：按用户维护连接集合。
type Hub struct {
	mu      sync.RWMutex
	clients map[primitive.ObjectID]map[*Client]bool
}

// NewHub 构造 Hub。
func NewHub() *Hub {
	return &Hub{clients: make(map[primitive.ObjectID]map[*Client]bool)}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.UserID] == nil {
		h.clients[c.UserID] = make(map[*Client]bool)
	}
	h.clients[c.UserID][c] = true
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[c.UserID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.clients, c.UserID)
		}
	}
}

func (h *Hub) sendTo(userID primitive.ObjectID, payload []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns := h.clients[userID]
	if len(conns) == 0 {
		return false
	}
	for c := range conns {
		select {
		case c.Send <- payload:
		default:
		}
	}
	return true
}

// ChatService 私信业务 + 实时推送。
type ChatService struct {
	repo   *repository.ChatRepository
	users  *repository.UserRepository
	audit  *AuditService
	logger *slog.Logger
	hub    *Hub
}

// NewChatService 构造注入。
func NewChatService(repo *repository.ChatRepository, users *repository.UserRepository, audit *AuditService, logger *slog.Logger) *ChatService {
	return &ChatService{repo: repo, users: users, audit: audit, logger: logger, hub: NewHub()}
}

// Hub 暴露 Hub 供 handler 注册连接。
func (s *ChatService) Hub() *Hub { return s.hub }

// SendMessage 发送私信并实时推送。
func (s *ChatService) SendMessage(ctx context.Context, fromID primitive.ObjectID, req dto.SendMessageRequest, ip string) (*model.ChatMessage, error) {
	toID, err := dto.ParseObjectID(req.ToID)
	if err != nil {
		return nil, util.Validation("接收方 ID 不合法", err)
	}
	if fromID == toID {
		return nil, util.BadRequest("不能给自己发私信", errors.New("cannot message self"))
	}
	if _, err := s.users.FindByID(ctx, toID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound("接收方不存在", err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	msg := &model.ChatMessage{
		FromID:    fromID,
		ToID:      toID,
		Type:      req.Type,
		Content:   req.Content,
		Read:      false,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogChatSent, msg.ID.Hex(), fromID.Hex(), toID.Hex(), msg.Type), "ip", ip)
	s.audit.Record(ctx, fromID, "", "chat.send", "chat", msg.ID.Hex(), "发送私信", ip)
	// 实时推送
	payload, _ := jsonMarshal(map[string]interface{}{
		"event": "message",
		"data":  dto.ToChatMessageResponse(msg),
	})
	s.hub.sendTo(toID, payload)
	return msg, nil
}

// ListConversation 会话消息列表。
func (s *ChatService) ListConversation(ctx context.Context, userID, peerID primitive.ObjectID, page, pageSize int) ([]model.ChatMessage, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	msgs, total, err := s.repo.ListConversation(ctx, userID, peerID, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.MarkRead(ctx, userID, peerID); err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	return msgs, total, nil
}

// ListConversations 会话列表。
func (s *ChatService) ListConversations(ctx context.Context, userID primitive.ObjectID) ([]dto.ConversationResponse, error) {
	msgs, err := s.repo.ListConversations(ctx, userID)
	if err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	seen := map[string]bool{}
	var result []dto.ConversationResponse
	for _, m := range msgs {
		peerID := m.ToID
		if peerID == userID {
			peerID = m.FromID
		}
		key := peerID.Hex()
		if seen[key] {
			continue
		}
		seen[key] = true
		peer, err := s.users.FindByID(ctx, peerID)
		if err != nil {
			continue
		}
		unread, _ := s.repo.CountUnread(ctx, userID, peerID)
		result = append(result, dto.ConversationResponse{
			PeerID:      peerID.Hex(),
			PeerName:    peer.Nickname,
			PeerAvatar:  peer.Avatar,
			LastMessage: m.Content,
			LastTime:    m.CreatedAt.Format("2006-01-02 15:04:05"),
			UnreadCount: unread,
		})
	}
	return result, nil
}

// jsonMarshal 序列化 JSON（封装标准库）。
func jsonMarshal(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return b, nil
}
