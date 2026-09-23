package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件。
func CORS(origins string) gin.HandlerFunc {
	cfg := cors.Config{
		AllowAllOrigins:  origins == "*",
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	if origins != "*" {
		cfg.AllowOrigins = splitComma(origins)
	}
	return cors.New(cfg)
}

func splitComma(s string) []string {
	var out []string
	for _, p := range split(s) {
		out = append(out, p)
	}
	return out
}

func split(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ',' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}
