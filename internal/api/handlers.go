package api

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/moon-readwise-server/internal/storage"
)

const maxBodySize = 1 << 20 // 1MB

func ListHighlights(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		highlights, err := store.ListHighlights()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "internal error")
			return
		}
		jsonOK(w, map[string]any{"highlights": highlights})
	}
}

func CreateHighlights(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
		if err != nil {
			jsonError(w, http.StatusBadRequest, "failed to read body")
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
			jsonError(w, http.StatusBadRequest, "no highlights found in request")
			return
		}

		var created []storage.Highlight
		for _, h := range highlights {
			if _, err := store.CreateHighlight(&h); err != nil {
				jsonError(w, http.StatusInternalServerError, "internal error")
				return
			}
			created = append(created, h)
		}

		jsonOK(w, map[string]any{"highlights": created})
	}
}

func GetHighlight(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "invalid id")
			return
		}

		h, err := store.GetHighlight(id)
		if err != nil {
			jsonError(w, http.StatusNotFound, "highlight not found")
			return
		}

		jsonOK(w, h)
	}
}

func DeleteHighlight(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "invalid id")
			return
		}

		if err := store.DeleteHighlight(id); err != nil {
			jsonError(w, http.StatusNotFound, "highlight not found")
			return
		}

		jsonOK(w, map[string]string{"message": "deleted"})
	}
}

func UpdateHighlight(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "invalid id")
			return
		}

		// 获取现有高亮
		h, err := store.GetHighlight(id)
		if err != nil {
			jsonError(w, http.StatusNotFound, "highlight not found")
			return
		}

		// 读取请求体
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
		if err != nil {
			jsonError(w, http.StatusBadRequest, "failed to read body")
			return
		}

		// 解析更新字段
		var updates map[string]any
		if err := json.Unmarshal(body, &updates); err != nil {
			jsonError(w, http.StatusBadRequest, "invalid JSON")
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
			jsonError(w, http.StatusInternalServerError, "internal error")
			return
		}

		jsonOK(w, h)
	}
}

func DeleteAllHighlights(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.DeleteAllHighlights(); err != nil {
			jsonError(w, http.StatusInternalServerError, "internal error")
			return
		}
		jsonOK(w, map[string]string{"message": "all highlights deleted"})
	}
}

func ListBooks(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		books, err := store.ListBooks()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "internal error")
			return
		}
		jsonOK(w, books)
	}
}

func GetBook(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "invalid id")
			return
		}

		b, err := store.GetBook(id)
		if err != nil {
			jsonError(w, http.StatusNotFound, "book not found")
			return
		}

		jsonOK(w, b)
	}
}

func ExportJSON(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		highlights, err := store.ListHighlights()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "internal error")
			return
		}

		jsonOK(w, map[string]any{
			"highlights":  highlights,
			"total":       len(highlights),
			"exported_at": time.Now().UTC(),
		})
	}
}

func ExportCSV(store *storage.SQLiteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		highlights, err := store.ListHighlights()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "internal error")
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=highlights.csv")

		cw := csv.NewWriter(w)
		defer cw.Flush()

		cw.Write([]string{"id", "text", "note", "title", "author", "source", "color", "created_at"})

		for _, h := range highlights {
			cw.Write([]string{
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
