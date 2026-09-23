package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/petsocial/petsocial/internal/model"
)

// ChatRepository 私信数据访问层。
type ChatRepository struct {
	coll *mongo.Collection
}

// NewChatRepository 构造注入。
func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{coll: db.Collection("chat_messages")}
}

func (r *ChatRepository) Create(ctx context.Context, m *model.ChatMessage) error {
	if m.ID.IsZero() {
		m.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.InsertOne(ctx, m); err != nil {
		return fmt.Errorf("insert chat message: %w", err)
	}
	return nil
}

func (r *ChatRepository) ListConversation(ctx context.Context, userID, peerID primitive.ObjectID, page, pageSize int) ([]model.ChatMessage, int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_id": userID, "to_id": peerID},
			{"from_id": peerID, "to_id": userID},
		},
	}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count conversation: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list conversation: %w", err)
	}
	defer cursor.Close(ctx)
	msgs := []model.ChatMessage{}
	if err := cursor.All(ctx, &msgs); err != nil {
		return nil, 0, fmt.Errorf("decode conversation: %w", err)
	}
	return msgs, total, nil
}

func (r *ChatRepository) MarkRead(ctx context.Context, toID, fromID primitive.ObjectID) error {
	_, err := r.coll.UpdateMany(ctx, bson.M{"to_id": toID, "from_id": fromID, "read": false}, bson.M{"$set": bson.M{"read": true}})
	if err != nil {
		return fmt.Errorf("mark chat read: %w", err)
	}
	return nil
}

func (r *ChatRepository) ListConversations(ctx context.Context, userID primitive.ObjectID) ([]model.ChatMessage, error) {
	// 取当前用户相关的最新消息作为会话摘要
	filter := bson.M{"$or": []bson.M{{"from_id": userID}, {"to_id": userID}}}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(100)
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer cursor.Close(ctx)
	msgs := []model.ChatMessage{}
	if err := cursor.All(ctx, &msgs); err != nil {
		return nil, fmt.Errorf("decode conversations: %w", err)
	}
	return msgs, nil
}

func (r *ChatRepository) CountUnread(ctx context.Context, userID, peerID primitive.ObjectID) (int64, error) {
	filter := bson.M{"to_id": userID, "from_id": peerID, "read": false}
	n, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return n, nil
}
