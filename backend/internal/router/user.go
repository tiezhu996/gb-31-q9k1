package router

import (
	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/handler"
)

// RegisterUserPublic 用户公开路由。
func RegisterUserPublic(r *gin.RouterGroup, h *handler.UserHandler) {
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.GET("/users/:id", h.GetUserByID)
}

// RegisterUserAuth 用户登录路由。
func RegisterUserAuth(r *gin.RouterGroup, h *handler.UserHandler) {
	r.GET("/users/me", h.GetProfile)
	r.PUT("/users/me", h.UpdateProfile)
	r.POST("/users/me/follow", h.Follow)
	r.DELETE("/users/me/follow/:id", h.Unfollow)
	r.GET("/users/me/following", h.ListFollowing)
	r.GET("/users/me/followers", h.ListFollowers)
}
