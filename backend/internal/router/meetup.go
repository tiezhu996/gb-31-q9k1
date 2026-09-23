package router

import (
	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/handler"
)

// RegisterMeetupPublic 约伴公开路由。
func RegisterMeetupPublic(r *gin.RouterGroup, h *handler.MeetupHandler) {
	r.GET("/meetups", h.List)
	r.GET("/meetups/:id", h.Get)
}

// RegisterMeetupAuth 约伴登录路由。
func RegisterMeetupAuth(r *gin.RouterGroup, h *handler.MeetupHandler) {
	r.POST("/meetups", h.Create)
	r.GET("/meetups/me", h.ListMy)
	r.POST("/meetups/:id/join", h.Join)
	r.DELETE("/meetups/:id/join", h.CancelJoin)
	r.PATCH("/meetups/:id/status", h.UpdateStatus)
}
