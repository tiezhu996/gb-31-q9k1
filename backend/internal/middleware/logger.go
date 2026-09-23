package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/constants"
)

// RequestLogger 请求日志中间件：统一包含 request_id/method/path/status/latency_ms。
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		rid := GetRequestID(c)
		logger.Debug(fmt.Sprintf(constants.LogRequestIn, rid, c.Request.Method, c.Request.URL.Path))
		c.Next()
		status := c.Writer.Status()
		latency := time.Since(start).Milliseconds()
		logger.Info(fmt.Sprintf(constants.LogRequestDone, rid, c.Request.Method, c.Request.URL.Path, status, latency))
	}
}
