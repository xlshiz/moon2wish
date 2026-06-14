package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DebugSanitizeMiddleware 在调试模式下清除敏感信息防止日志泄露
func DebugSanitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这个中间件可以用来在调试模式下清除敏感信息
		// 但Gin的默认日志中间件不会记录Authorization header
		// 所以我们暂时不需要做任何事情
		c.Next()
	}
}

// SecurityHeadersMiddleware 添加安全HTTP头
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

// ValidateContentTypeMiddleware 验证Content-Type
func ValidateContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对需要JSON的POST/PUT/PATCH请求进行验证
		if c.Request.Method == http.MethodPost || 
		   c.Request.Method == http.MethodPut || 
		   c.Request.Method == http.MethodPatch {
		   
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type must be application/json"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}