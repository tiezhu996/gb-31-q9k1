package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
)

// MediaItem 动态媒体资源。
type MediaItem struct {
	Type string `bson:"type" json:"type"`
	URL  string `bson:"url" json:"url"`
}

// Post 图文/短视频动态实体：话题、宠物标签、地点、审核状态机。
type Post struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	AuthorID      primitive.ObjectID   `bson:"author_id" json:"author_id"`
	PetIDs        []primitive.ObjectID `bson:"pet_ids" json:"pet_ids"`
	Content       string               `bson:"content" json:"content"`
	Type          constants.PostType   `bson:"type" json:"type"`
	Media         []MediaItem          `bson:"media" json:"media"`
	Topics        []string             `bson:"topics" json:"topics"`
	Location      string               `bson:"location" json:"location"`
	City          string               `bson:"city" json:"city"`
	Status        constants.PostStatus `bson:"status" json:"status"`
	RejectReason  string               `bson:"reject_reason" json:"reject_reason"`
	LikeCount     int64                `bson:"like_count" json:"like_count"`
	CommentCount  int64                `bson:"comment_count" json:"comment_count"`
	FavoriteCount int64                `bson:"favorite_count" json:"favorite_count"`
	ForwardCount  int64                `bson:"forward_count" json:"forward_count"`
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`
}

// Comment 评论实体。
type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID    primitive.ObjectID `bson:"post_id" json:"post_id"`
	AuthorID  primitive.ObjectID `bson:"author_id" json:"author_id"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Interaction 点赞/收藏/转发互动实体：唯一索引 (post_id, user_id, type) 保证幂等。
type Interaction struct {
	ID        primitive.ObjectID        `bson:"_id,omitempty" json:"id"`
	PostID    primitive.ObjectID        `bson:"post_id" json:"post_id"`
	UserID    primitive.ObjectID        `bson:"user_id" json:"user_id"`
	Type      constants.InteractionType `bson:"type" json:"type"`
	CreatedAt time.Time                 `bson:"created_at" json:"created_at"`
}

// Topic 话题实体：官方话题运营与参与活动榜。
type Topic struct {
	ID          primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	Name        string                `bson:"name" json:"name"`
	Description string                `bson:"description" json:"description"`
	PostCount   int64                 `bson:"post_count" json:"post_count"`
	Status      constants.TopicStatus `bson:"status" json:"status"`
	CreatedAt   time.Time             `bson:"created_at" json:"created_at"`
}
