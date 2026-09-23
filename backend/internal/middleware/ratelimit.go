package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/util"
)

// RateLimit 基于 Redis 固定窗口限流。
func RateLimit(rdb *redis.Client, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil || limit <= 0 {
			c.Next()
			return
		}
		ip := c.ClientIP()
		key := "ratelimit:" + c.Request.URL.Path + ":" + ip
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			_ = rdb.Expire(ctx, key, time.Minute).Err()
		}
		if count > int64(limit) {
			util.Fail(c, http.StatusTooManyRequests, constants.CodeRateLimited, constants.MsgRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}
