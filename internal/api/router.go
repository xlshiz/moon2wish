package api

import (
	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/storage"
)

func NewRouter(store *storage.SQLiteStore, token string, debug bool) *gin.Engine {
	r := gin.Default()

	// 注意：Gin的默认日志中间件不会记录Authorization header
	// 但为了确保安全，我们可以在调试模式下添加安全头中间件
	if debug {
		r.Use(SecurityHeadersMiddleware())
	}

	auth := r.Group("/api/moon")
	auth.Use(AuthMiddleware(token))
	{
		// V2 API - Readwise 兼容接口
		auth.POST("/v2/highlights/", CreateHighlights(store))
		auth.GET("/v2/highlights/", ListHighlights(store))
		auth.GET("/v2/highlights/:id", GetHighlight(store))
		auth.PATCH("/v2/highlights/:id", UpdateHighlight(store))
		auth.DELETE("/v2/highlights/:id", DeleteHighlight(store))
		auth.DELETE("/v2/highlights/", DeleteAllHighlights(store))

		auth.GET("/v2/export/", ExportJSON(store))
		auth.GET("/v2/books/", ListBooks(store))
		auth.GET("/v2/books/:id", GetBook(store))

		// 通用导出接口
		auth.GET("/export/json", ExportJSON(store))
		auth.GET("/export/csv", ExportCSV(store))
	}

	return r
}
