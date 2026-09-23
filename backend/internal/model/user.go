package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
)

// User 用户实体：角色字段贯穿 JWT/RBAC/前端按钮显隐。
type User struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username      string               `bson:"username" json:"username"`
	PasswordHash  string               `bson:"password_hash" json:"-"`
	Nickname      string               `bson:"nickname" json:"nickname"`
	Avatar        string               `bson:"avatar" json:"avatar"`
	Bio           string               `bson:"bio" json:"bio"`
	City          string               `bson:"city" json:"city"`
	Role          constants.Role       `bson:"role" json:"role"`
	Status        constants.UserStatus `bson:"status" json:"status"`
	FollowCount   int64                `bson:"follow_count" json:"follow_count"`
	FollowerCount int64                `bson:"follower_count" json:"follower_count"`
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`
}

// Follow 关注关系：follow 动作的并发安全依赖唯一索引 (follower_id, followee_id)。
type Follow struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FollowerID primitive.ObjectID `bson:"follower_id" json:"follower_id"`
	FolloweeID primitive.ObjectID `bson:"followee_id" json:"followee_id"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}
