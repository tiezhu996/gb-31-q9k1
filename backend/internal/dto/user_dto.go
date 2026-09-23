package dto

import (
	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RegisterRequest 注册入参。
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" binding:"required,min=1,max=32"`
	City     string `json:"city" binding:"max=32"`
}

// LoginRequest 登录入参。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新资料入参。
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"max=32"`
	Avatar   string `json:"avatar" binding:"max=512"`
	Bio      string `json:"bio" binding:"max=255"`
	City     string `json:"city" binding:"max=32"`
}

// UserResponse 用户出参（隐藏密码）。
type UserResponse struct {
	ID            string               `json:"id"`
	Username      string               `json:"username"`
	Nickname      string               `json:"nickname"`
	Avatar        string               `json:"avatar"`
	Bio           string               `json:"bio"`
	City          string               `json:"city"`
	Role          constants.Role       `json:"role"`
	Status        constants.UserStatus `json:"status"`
	FollowCount   int64                `json:"follow_count"`
	FollowerCount int64                `json:"follower_count"`
	CreatedAt     string               `json:"created_at"`
}

// TokenResponse 登录/注册返回 token。
type TokenResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// FollowRequest 关注/取消关注入参。
type FollowRequest struct {
	FolloweeID string `json:"followee_id" binding:"required"`
}

// PageResponse 统一分页出参。
type PageResponse struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ToUserResponse 模型转出参。
func ToUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:            u.ID.Hex(),
		Username:      u.Username,
		Nickname:      u.Nickname,
		Avatar:        u.Avatar,
		Bio:           u.Bio,
		City:          u.City,
		Role:          u.Role,
		Status:        u.Status,
		FollowCount:   u.FollowCount,
		FollowerCount: u.FollowerCount,
		CreatedAt:     u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ParseObjectID 解析 ObjectID，失败时返回 ErrInvalidID。
func ParseObjectID(s string) (primitive.ObjectID, error) {
	if !primitive.IsValidObjectID(s) {
		return primitive.NilObjectID, constants.ErrInvalidID
	}
	return primitive.ObjectIDFromHex(s)
}
