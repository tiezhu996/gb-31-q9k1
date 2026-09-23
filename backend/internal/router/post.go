package router

import (
	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/handler"
)

// RegisterPostPublic 动态公开路由。
func RegisterPostPublic(r *gin.RouterGroup, h *handler.PostHandler) {
	r.GET("/posts", h.List)
	r.GET("/posts/:id", h.Get)
	r.GET("/posts/:id/comments", h.ListComments)
	r.GET("/topics", h.ListTopics)
	r.GET("/topics/:name/posts", h.ListByTopic)
}

// RegisterPostAuth 动态登录路由。
func RegisterPostAuth(r *gin.RouterGroup, h *handler.PostHandler) {
	r.POST("/posts", h.Create)
	r.POST("/posts/:id/comments", h.AddComment)
	r.DELETE("/posts/:id/comments/:comment_id", h.DeleteComment)
	r.POST("/posts/:id/interact", h.Interact)
	r.GET("/users/:id/posts", h.ListByAuthor)
}
