package api

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/storage"
)

const maxBodySize = 1 << 20 // 1MB

func ListHighlights(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		highlights, err := store.ListHighlights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"highlights": highlights})
	}
}

func CreateHighlights(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}

		var highlights []storage.Highlight

		var batch struct {
			Highlights []storage.Highlight `json:"highlights"`
		}
		if err := json.Unmarshal(body, &batch); err == nil && len(batch.Highlights) > 0 {
			highlights = batch.Highlights
		} else {
			var single storage.Highlight
			if err := json.Unmarshal(body, &single); err == nil && single.Title != "" {
				highlights = append(highlights, single)
			}
		}

		if len(highlights) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no highlights found in request"})
			return
		}

		var created []storage.Highlight
		for _, h := range highlights {
			if _, err := store.CreateHighlight(&h); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
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

func UpdateHighlight(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		// 获取现有高亮
		h, err := store.GetHighlight(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "highlight not found"})
			return
		}

		// 读取请求体
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}

		// 解析更新字段
		var updates map[string]interface{}
		if err := json.Unmarshal(body, &updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		// 应用更新
		if v, ok := updates["text"].(string); ok {
			h.Text = v
		}
		if v, ok := updates["note"].(string); ok {
			h.Note = v
		}
		if v, ok := updates["title"].(string); ok {
			h.Title = v
		}
		if v, ok := updates["author"].(string); ok {
			h.Author = v
		}
		if v, ok := updates["source"].(string); ok {
			h.Source = v
		}
		if v, ok := updates["source_url"].(string); ok {
			h.Source = v // source_url 映射到 source
		}
		if v, ok := updates["color"].(string); ok {
			h.Color = v
		}
		if v, ok := updates["book_type"].(string); ok {
			h.BookType = v
		}

		// 更新到数据库
		if err := store.UpdateHighlight(h); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		c.JSON(http.StatusOK, h)
	}
}

func DeleteAllHighlights(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := store.DeleteAllHighlights(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "all highlights deleted"})
	}
}

func ListBooks(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		books, err := store.ListBooks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"highlights":  highlights,
			"total":       len(highlights),
			"exported_at": time.Now().UTC(),
		})
	}
}

func ExportCSV(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		highlights, err := store.ListHighlights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=highlights.csv")

		w := csv.NewWriter(c.Writer)
		defer w.Flush()

		w.Write([]string{"id", "text", "note", "title", "author", "source", "color", "created_at"})

		for _, h := range highlights {
			w.Write([]string{
				strconv.FormatInt(h.ID, 10),
				h.Text,
				h.Note,
				h.Title,
				h.Author,
				h.Source,
				h.Color,
				h.CreatedAt.Format(time.RFC3339),
			})
		}
	}
}
