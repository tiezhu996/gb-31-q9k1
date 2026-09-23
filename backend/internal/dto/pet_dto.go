package dto

import (
	"time"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
)

// CreatePetRequest 创建宠物档案入参。
type CreatePetRequest struct {
	Name        string               `json:"name" binding:"required,min=1,max=32"`
	Species     constants.PetSpecies `json:"species" binding:"required,oneof=dog cat other"`
	Breed       string               `json:"breed" binding:"max=32"`
	Birthday    string               `json:"birthday" binding:"required"`
	Gender      constants.PetGender  `json:"gender" binding:"required,oneof=male female"`
	Personality string               `json:"personality" binding:"max=255"`
	Bio         string               `json:"bio" binding:"max=255"`
	Avatar      string               `json:"avatar" binding:"max=512"`
	Album       []string             `json:"album"`
	City        string               `json:"city" binding:"max=32"`
}

// UpdatePetRequest 更新宠物档案入参。
type UpdatePetRequest struct {
	Name        string              `json:"name" binding:"max=32"`
	Breed       string              `json:"breed" binding:"max=32"`
	Birthday    string              `json:"birthday"`
	Gender      constants.PetGender `json:"gender" binding:"omitempty,oneof=male female"`
	Personality string              `json:"personality" binding:"max=255"`
	Bio         string              `json:"bio" binding:"max=255"`
	Avatar      string              `json:"avatar" binding:"max=512"`
	Album       []string            `json:"album"`
	City        string              `json:"city" binding:"max=32"`
	Status      constants.PetStatus `json:"status" binding:"omitempty,oneof=active hidden"`
}

// PetResponse 宠物出参。
type PetResponse struct {
	ID            string               `json:"id"`
	OwnerID       string               `json:"owner_id"`
	Name          string               `json:"name"`
	Species       constants.PetSpecies `json:"species"`
	Breed         string               `json:"breed"`
	Birthday      string               `json:"birthday"`
	Gender        constants.PetGender  `json:"gender"`
	Personality   string               `json:"personality"`
	Bio           string               `json:"bio"`
	Avatar        string               `json:"avatar"`
	Album         []string             `json:"album"`
	City          string               `json:"city"`
	Status        constants.PetStatus  `json:"status"`
	FollowerCount int64                `json:"follower_count"`
	CreatedAt     string               `json:"created_at"`
}

// ToPetResponse 模型转出参。
func ToPetResponse(p *model.Pet) PetResponse {
	return PetResponse{
		ID:            p.ID.Hex(),
		OwnerID:       p.OwnerID.Hex(),
		Name:          p.Name,
		Species:       p.Species,
		Breed:         p.Breed,
		Birthday:      p.Birthday.Format("2006-01-02"),
		Gender:        p.Gender,
		Personality:   p.Personality,
		Bio:           p.Bio,
		Avatar:        p.Avatar,
		Album:         p.Album,
		City:          p.City,
		Status:        p.Status,
		FollowerCount: p.FollowerCount,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ParseBirthday 解析生日字符串。
func ParseBirthday(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
