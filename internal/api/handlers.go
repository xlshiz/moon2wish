package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/storage"
)

func ListHighlights(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		highlights, err := store.ListHighlights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, highlights)
	}
}

func CreateHighlights(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Highlights []storage.Highlight `json:"highlights" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var created []storage.Highlight
		for _, h := range input.Highlights {
			if _, err := store.CreateHighlight(&h); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			created = append(created, h)
		}

		c.JSON(http.StatusOK, gin.H{"highlights": created})
	}
}

func GetHighlight(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		h, err := store.GetHighlight(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "highlight not found"})
			return
		}

		c.JSON(http.StatusOK, h)
	}
}

func DeleteHighlight(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := store.DeleteHighlight(id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "highlight not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func ListBooks(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		books, err := store.ListBooks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, books)
	}
}

func GetBook(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		b, err := store.GetBook(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}

		c.JSON(http.StatusOK, b)
	}
}

func ExportJSON(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		highlights, err := store.ListHighlights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, highlights)
	}
}

func ExportCSV(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		highlights, err := store.ListHighlights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=highlights.csv")
		c.String(http.StatusOK, "id,text,title,author,source,color\n")
		for _, h := range highlights {
			c.String(http.StatusOK, "%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
				h.ID, h.Text, h.Title, h.Author, h.Source, h.Color)
		}
	}
}
