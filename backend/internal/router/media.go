package router

import (
	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/handler"
)

// RegisterMediaAuth 媒体上传路由。
func RegisterMediaAuth(r *gin.RouterGroup, h *handler.MediaHandler) {
	r.POST("/media/upload", h.Upload)
}
