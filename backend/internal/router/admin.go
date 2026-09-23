package router

import (
	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/handler"
)

// RegisterAdmin 管理员路由（RBAC：仅 admin）。
func RegisterAdmin(r *gin.RouterGroup, postH *handler.PostHandler, auditH *handler.AuditHandler) {
	r.PATCH("/admin/posts/:id/status", postH.UpdateStatus)
	r.POST("/admin/topics", postH.CreateTopic)
	r.GET("/admin/audit-logs", auditH.List)
}
