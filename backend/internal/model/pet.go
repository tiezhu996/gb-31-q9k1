package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
)

// Pet 宠物主页实体：品种/生日/性格/相册，宠物有独立主页与粉丝。
type Pet struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	OwnerID       primitive.ObjectID   `bson:"owner_id" json:"owner_id"`
	Name          string               `bson:"name" json:"name"`
	Species       constants.PetSpecies `bson:"species" json:"species"`
	Breed         string               `bson:"breed" json:"breed"`
	Birthday      time.Time            `bson:"birthday" json:"birthday"`
	Gender        constants.PetGender  `bson:"gender" json:"gender"`
	Personality   string               `bson:"personality" json:"personality"`
	Bio           string               `bson:"bio" json:"bio"`
	Avatar        string               `bson:"avatar" json:"avatar"`
	Album         []string             `bson:"album" json:"album"`
	City          string               `bson:"city" json:"city"`
	Status        constants.PetStatus  `bson:"status" json:"status"`
	FollowerCount int64                `bson:"follower_count" json:"follower_count"`
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`
}

// PetFollow 宠物粉丝关系。
type PetFollow struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PetID     primitive.ObjectID `bson:"pet_id" json:"pet_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
