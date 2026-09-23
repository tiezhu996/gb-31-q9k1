package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/util"
)

// RequireRole RBAC 权限中间件：仅允许指定角色访问。
func RequireRole(roles ...constants.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		for _, r := range roles {
			if constants.Role(role) == r {
				c.Next()
				return
			}
		}
		util.Fail(c, http.StatusForbidden, constants.CodeRoleForbidden, constants.MsgForbidden)
		c.Abort()
	}
}
