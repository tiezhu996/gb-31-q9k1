package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OK 统一成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

// Fail 统一失败响应。
func Fail(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, gin.H{"code": code, "message": message, "data": nil})
}

// Page 统一分页成功响应。
func Page(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	OK(c, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}
