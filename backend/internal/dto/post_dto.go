package dto

import (
	"time"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreatePostRequest 发布动态入参。
type CreatePostRequest struct {
	Content  string             `json:"content" binding:"required,min=1,max=2000"`
	Type     constants.PostType `json:"type" binding:"required,oneof=image video"`
	Media    []MediaDTO         `json:"media"`
	PetIDs   []string           `json:"pet_ids"`
	Topics   []string           `json:"topics"`
	Location string             `json:"location" binding:"max=128"`
	City     string             `json:"city" binding:"max=32"`
}

// MediaDTO 媒体资源入参。
type MediaDTO struct {
	Type string `json:"type" binding:"required,oneof=image video"`
	URL  string `json:"url" binding:"required,max=1024"`
}

// UpdatePostStatusRequest 审核状态流转入参。
type UpdatePostStatusRequest struct {
	Status       constants.PostStatus `json:"status" binding:"required,oneof=approved rejected"`
	RejectReason string               `json:"reject_reason" binding:"max=255"`
}

// CommentRequest 评论入参。
type CommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=500"`
}

// InteractionRequest 互动入参。
type InteractionRequest struct {
	Type constants.InteractionType `json:"type" binding:"required,oneof=like favorite forward"`
}

// PostResponse 动态出参。
type PostResponse struct {
	ID            string               `json:"id"`
	AuthorID      string               `json:"author_id"`
	AuthorName    string               `json:"author_name"`
	AuthorAvatar  string               `json:"author_avatar"`
	PetIDs        []string             `json:"pet_ids"`
	Content       string               `json:"content"`
	Type          constants.PostType   `json:"type"`
	Media         []MediaDTO           `json:"media"`
	Topics        []string             `json:"topics"`
	Location      string               `json:"location"`
	City          string               `json:"city"`
	Status        constants.PostStatus `json:"status"`
	RejectReason  string               `json:"reject_reason"`
	LikeCount     int64                `json:"like_count"`
	CommentCount  int64                `json:"comment_count"`
	FavoriteCount int64                `json:"favorite_count"`
	ForwardCount  int64                `json:"forward_count"`
	CreatedAt     string               `json:"created_at"`
	Liked         bool                 `json:"liked"`
	Favorited     bool                 `json:"favorited"`
}

// CommentResponse 评论出参。
type CommentResponse struct {
	ID         string `json:"id"`
	PostID     string `json:"post_id"`
	AuthorID   string `json:"author_id"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

// ToPostResponse 模型转出参。
func ToPostResponse(p *model.Post) PostResponse {
	media := make([]MediaDTO, 0, len(p.Media))
	for _, m := range p.Media {
		media = append(media, MediaDTO{Type: m.Type, URL: m.URL})
	}
	petIDs := make([]string, 0, len(p.PetIDs))
	for _, id := range p.PetIDs {
		petIDs = append(petIDs, id.Hex())
	}
	return PostResponse{
		ID:            p.ID.Hex(),
		AuthorID:      p.AuthorID.Hex(),
		PetIDs:        petIDs,
		Content:       p.Content,
		Type:          p.Type,
		Media:         media,
		Topics:        p.Topics,
		Location:      p.Location,
		City:          p.City,
		Status:        p.Status,
		RejectReason:  p.RejectReason,
		LikeCount:     p.LikeCount,
		CommentCount:  p.CommentCount,
		FavoriteCount: p.FavoriteCount,
		ForwardCount:  p.ForwardCount,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToCommentResponse 评论转出参。
func ToCommentResponse(c *model.Comment, authorName string) CommentResponse {
	return CommentResponse{
		ID:         c.ID.Hex(),
		PostID:     c.PostID.Hex(),
		AuthorID:   c.AuthorID.Hex(),
		AuthorName: authorName,
		Content:    c.Content,
		CreatedAt:  c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ParsePostIDs 解析宠物 ID 列表。
func ParsePostIDs(ids []string) ([]primitive.ObjectID, error) {
	out := make([]primitive.ObjectID, 0, len(ids))
	for _, s := range ids {
		id, err := ParseObjectID(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

// ParseTimePtr 解析可选时间。
func ParseTimePtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
