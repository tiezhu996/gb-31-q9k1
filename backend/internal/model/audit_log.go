package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLog 操作审计日志实体。
type AuditLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username   string             `bson:"username" json:"username"`
	Action     string             `bson:"action" json:"action"`
	Resource   string             `bson:"resource" json:"resource"`
	ResourceID string             `bson:"resource_id" json:"resource_id"`
	Detail     string             `bson:"detail" json:"detail"`
	IP         string             `bson:"ip" json:"ip"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}
