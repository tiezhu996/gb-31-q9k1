package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/util"
)

// UserIDKey context 中用户 ID 的键。
const UserIDKey = "user_id"

// UsernameKey context 中用户名的键。
const UsernameKey = "username"

// RoleKey context 中角色的键。
const RoleKey = "role"

// Auth 认证中间件：校验 Authorization Bearer 或 token 查询参数（websocket）。
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		claims, err := util.ParseToken(secret, tokenStr)
		if err != nil {
			util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		userID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		c.Set(UserIDKey, userID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleKey, string(claims.Role))
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if token := c.Query("token"); token != "" {
		return token
	}
	return ""
}

// GetUserID 从 context 获取当前用户 ID。
func GetUserID(c *gin.Context) (primitive.ObjectID, bool) {
	v, ok := c.Get(UserIDKey)
	if !ok {
		return primitive.NilObjectID, false
	}
	id, ok := v.(primitive.ObjectID)
	return id, ok
}

// GetUsername 从 context 获取当前用户名。
func GetUsername(c *gin.Context) string {
	if v, ok := c.Get(UsernameKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetRole 从 context 获取当前角色。
func GetRole(c *gin.Context) string {
	if v, ok := c.Get(RoleKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

var _ = errors.New

// OptionalAuth 可选认证中间件：公开接口在携带 token 时注入用户上下文，不强制登录。
func OptionalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.Next()
			return
		}
		claims, err := util.ParseToken(secret, tokenStr)
		if err != nil {
			c.Next()
			return
		}
		userID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			c.Next()
			return
		}
		c.Set(UserIDKey, userID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleKey, string(claims.Role))
		c.Next()
	}
}
