package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/storage"
)

func setupTestRouter() (*gin.Engine, *storage.SQLiteStore) {
	gin.SetMode(gin.TestMode)
	dbPath := "/tmp/test_api_highlights.db"
	os.Remove(dbPath)

	store, _ := storage.NewSQLiteStore(dbPath)
	token := "test-token"
	r := NewRouter(store, token)
	return r, store
}

func TestCreateHighlights(t *testing.T) {
	r, store := setupTestRouter()
	defer store.Close()

	body := `{
		"highlights": [{
			"text": "test highlight",
			"title": "Test Book",
			"author": "Test Author"
		}]
	}`

	req, _ := http.NewRequest("POST", "/api/v3/highlights/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token test-token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestListHighlights(t *testing.T) {
	r, store := setupTestRouter()
	defer store.Close()

	req, _ := http.NewRequest("GET", "/api/v3/highlights/", nil)
	req.Header.Set("Authorization", "Token test-token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["highlights"]; !ok {
		t.Error("Expected 'highlights' key in response")
	}
}
