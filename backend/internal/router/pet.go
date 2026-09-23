package router

import (
	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/handler"
)

// RegisterPetPublic 宠物公开路由。
func RegisterPetPublic(r *gin.RouterGroup, h *handler.PetHandler) {
	r.GET("/pets", h.List)
	r.GET("/pets/:id", h.Get)
}

// RegisterPetAuth 宠物登录路由。
func RegisterPetAuth(r *gin.RouterGroup, h *handler.PetHandler) {
	r.POST("/pets", h.Create)
	r.PUT("/pets/:id", h.Update)
	r.DELETE("/pets/:id", h.Delete)
	r.GET("/pets/me", h.ListMy)
	r.POST("/pets/:id/follow", h.Follow)
	r.DELETE("/pets/:id/follow", h.Unfollow)
}
