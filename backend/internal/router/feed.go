package router

import (
	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/handler"
)

// RegisterFeedPublic Feed 公开路由。
func RegisterFeedPublic(r *gin.RouterGroup, h *handler.FeedHandler) {
	r.GET("/feed/discover", h.Discover)
	r.GET("/feed/nearby", h.Nearby)
	r.GET("/feed/following", h.Following)
}
