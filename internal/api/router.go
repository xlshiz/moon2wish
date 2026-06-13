package api

import (
	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/storage"
)

func NewRouter(store *storage.SQLiteStore, token string) *gin.Engine {
	r := gin.Default()

	auth := r.Group("/")
	auth.Use(AuthMiddleware(token))
	{
		auth.POST("/api/v2/highlights/", CreateHighlights(store))
		auth.GET("/api/v2/highlights/", ListHighlights(store))

		auth.POST("/api/v3/highlights/", CreateHighlights(store))
		auth.GET("/api/v3/highlights/", ListHighlights(store))
		auth.GET("/api/v3/highlights/:id", GetHighlight(store))
		auth.DELETE("/api/v3/highlights/:id", DeleteHighlight(store))

		auth.GET("/api/v2/export/", ExportJSON(store))
		auth.GET("/api/v2/books/", ListBooks(store))
		auth.GET("/api/v2/books/:id", GetBook(store))

		auth.GET("/api/v3/books/", ListBooks(store))
		auth.GET("/api/v3/books/:id", GetBook(store))

		auth.GET("/api/export/json", ExportJSON(store))
		auth.GET("/api/export/csv", ExportCSV(store))
	}

	return r
}
