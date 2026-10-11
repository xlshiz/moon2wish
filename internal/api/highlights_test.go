package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/moon-readwise-server/internal/storage"
)

func setupTestRouter(t *testing.T) (http.Handler, *storage.SQLiteStore) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	token := "test-token"
	r := NewRouter(store, token, false)
	return r, store
}

func TestCreateHighlights(t *testing.T) {
	r, store := setupTestRouter(t)
	defer store.Close()

	body := `{
		"highlights": [{
			"text": "test highlight",
			"title": "Test Book",
			"author": "Test Author"
		}]
	}`

	req, _ := http.NewRequest("POST", "/api/moon/v2/highlights/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token test-token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestListHighlights(t *testing.T) {
	r, store := setupTestRouter(t)
	defer store.Close()

	req, _ := http.NewRequest("GET", "/api/moon/v2/highlights/", nil)
	req.Header.Set("Authorization", "Token test-token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["highlights"]; !ok {
		t.Error("Expected 'highlights' key in response")
	}
}
