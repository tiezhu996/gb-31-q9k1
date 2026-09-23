package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/service"
)

// Audit 操作审计中间件：对 /api/v1 写操作自动记录审计日志。
func Audit(audit *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() >= 400 {
			return
		}
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
			return
		}
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/v1") {
			return
		}
		userID, _ := GetUserID(c)
		if userID == primitive.NilObjectID {
			return
		}
		action := fmt.Sprintf("%s %s", method, path)
		audit.Record(c.Request.Context(), userID, GetUsername(c), action, "http", path, "HTTP 写操作", c.ClientIP())
	}
}

var _ = constants.LogAuditCreated
