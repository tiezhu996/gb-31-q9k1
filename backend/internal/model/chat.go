package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
)

// ChatMessage 私信消息实体：支持图片、表情包、宠物贴纸。
type ChatMessage struct {
	ID        primitive.ObjectID        `bson:"_id,omitempty" json:"id"`
	FromID    primitive.ObjectID        `bson:"from_id" json:"from_id"`
	ToID      primitive.ObjectID        `bson:"to_id" json:"to_id"`
	Type      constants.ChatMessageType `bson:"type" json:"type"`
	Content   string                    `bson:"content" json:"content"`
	Read      bool                      `bson:"read" json:"read"`
	CreatedAt time.Time                 `bson:"created_at" json:"created_at"`
}
