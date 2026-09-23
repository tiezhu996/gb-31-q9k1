package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/util"
)

// Recovery panic 恢复中间件。
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				rid := GetRequestID(c)
				logger.Error(fmt.Sprintf(constants.LogRecoveryPanic, rid, r), "stack", string(debug.Stack()))
				util.Fail(c, http.StatusInternalServerError, constants.CodeInternal, constants.MsgInternalError)
				c.Abort()
			}
		}()
		c.Next()
	}
}
