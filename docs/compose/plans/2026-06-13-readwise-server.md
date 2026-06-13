# Readwise-Compatible Highlight Server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go + Gin HTTP server that implements Readwise API v3 for receiving highlights from Moon Reader, stores them in SQLite, and provides export endpoints.

**Architecture:** Single Go binary with Gin HTTP framework, SQLite storage via pure-Go driver, token-based auth from config file, and JSON/CSV export capabilities.

**Tech Stack:** Go, Gin, SQLite (modernc.org/sqlite), YAML config

---

### Task 1: Project Scaffolding

**Covers:** [S10]
**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `config.yaml`

- [ ] **Step 1: Initialize Go module**

```bash
cd /home/shixl/workdir/moon
go mod init github.com/moon-readwise-server
```

- [ ] **Step 2: Create config.yaml**

```yaml
server:
  host: "0.0.0.0"
  port: 8080

auth:
  token: "your-secret-token-here"

database:
  path: "./data/highlights.db"
```

- [ ] **Step 3: Create minimal main.go**

```go
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Auth struct {
		Token string `yaml:"token"`
	} `yaml:"auth"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Server starting on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
```

- [ ] **Step 4: Install dependencies**

```bash
go mod tidy
```

- [ ] **Step 5: Test config loading**

```bash
go run cmd/server/main.go
```
Expected: `Server starting on 0.0.0.0:8080`

- [ ] **Step 6: Commit**

```bash
git init
git add go.mod go.sum cmd/server/main.go config.yaml
git commit -m "feat: project scaffolding with config loading"
```

---

### Task 2: SQLite Storage Layer

**Covers:** [S5]
**Files:**
- Create: `internal/storage/models.go`
- Create: `internal/storage/sqlite.go`
- Test: `internal/storage/sqlite_test.go`

- [ ] **Step 1: Write failing test for storage**

```go
package storage

import (
	"os"
	"testing"
)

func TestCreateAndGetHighlight(t *testing.T) {
	dbPath := "/tmp/test_highlights.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	h := &Highlight{
		Text:  "test highlight",
		Title: "Test Book",
	}

	id, err := store.CreateHighlight(h)
	if err != nil {
		t.Fatalf("Failed to create highlight: %v", err)
	}

	got, err := store.GetHighlight(id)
	if err != nil {
		t.Fatalf("Failed to get highlight: %v", err)
	}

	if got.Text != "test highlight" {
		t.Errorf("Expected text 'test highlight', got '%s'", got.Text)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/shixl/workdir/moon
go test ./internal/storage/...
```
Expected: FAIL (package doesn't exist)

- [ ] **Step 3: Create models.go**

```go
package storage

import "time"

type Highlight struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	Note      string    `json:"note,omitempty"`
	Title     string    `json:"title"`
	Author    string    `json:"author,omitempty"`
	Source    string    `json:"source,omitempty"`
	Color     string    `json:"color,omitempty"`
	BookType  string    `json:"book_type,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Book struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author,omitempty"`
	NumHighlights int    `json:"num_highlights"`
}
```

- [ ] **Step 4: Create sqlite.go**

```go
package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := initDB(db); err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func initDB(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS highlights (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			readwise_id TEXT UNIQUE,
			text TEXT NOT NULL,
			note TEXT,
			title TEXT NOT NULL,
			author TEXT,
			source TEXT,
			color TEXT,
			book_type TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			author TEXT,
			num_highlights INTEGER DEFAULT 0
		)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) CreateHighlight(h *Highlight) (int64, error) {
	now := time.Now()
	h.CreatedAt = now
	h.UpdatedAt = now

	result, err := s.db.Exec(
		`INSERT INTO highlights (text, note, title, author, source, color, book_type, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		h.Text, h.Note, h.Title, h.Author, h.Source, h.Color, h.BookType, h.CreatedAt, h.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	h.ID = id
	return id, nil
}

func (s *SQLiteStore) GetHighlight(id int64) (*Highlight, error) {
	var h Highlight
	err := s.db.QueryRow(
		`SELECT id, text, note, title, author, source, color, book_type, created_at, updated_at
		 FROM highlights WHERE id = ?`, id,
	).Scan(&h.ID, &h.Text, &h.Note, &h.Title, &h.Author, &h.Source, &h.Color, &h.BookType, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (s *SQLiteStore) ListHighlights() ([]Highlight, error) {
	rows, err := s.db.Query(
		`SELECT id, text, note, title, author, source, color, book_type, created_at, updated_at
		 FROM highlights ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var highlights []Highlight
	for rows.Next() {
		var h Highlight
		if err := rows.Scan(&h.ID, &h.Text, &h.Note, &h.Title, &h.Author, &h.Source, &h.Color, &h.BookType, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		highlights = append(highlights, h)
	}
	return highlights, nil
}

func (s *SQLiteStore) DeleteHighlight(id int64) error {
	_, err := s.db.Exec(`DELETE FROM highlights WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) ListBooks() ([]Book, error) {
	rows, err := s.db.Query(
		`SELECT id, title, author, num_highlights FROM books ORDER BY title`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.NumHighlights); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (s *SQLiteStore) GetBook(id int64) (*Book, error) {
	var b Book
	err := s.db.QueryRow(
		`SELECT id, title, author, num_highlights FROM books WHERE id = ?`, id,
	).Scan(&b.ID, &b.Title, &b.Author, &b.NumHighlights)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd /home/shixl/workdir/moon
go test ./internal/storage/... -v
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/storage/
git commit -m "feat: SQLite storage layer with highlight CRUD"
```

---

### Task 3: HTTP Router and Middleware

**Covers:** [S6]
**Files:**
- Create: `internal/api/router.go`
- Create: `internal/api/middleware.go`

- [ ] **Step 1: Create middleware.go**

```go
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		if parts[1] != token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 2: Create router.go**

```go
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
		auth.GET("/api/v3/highlights/", ListHighlights(store))
		auth.POST("/api/v3/highlights/", CreateHighlights(store))
		auth.GET("/api/v3/highlights/:id", GetHighlight(store))
		auth.DELETE("/api/v3/highlights/:id", DeleteHighlight(store))

		auth.GET("/api/v3/books/", ListBooks(store))
		auth.GET("/api/v3/books/:id", GetBook(store))

		auth.GET("/api/export/json", ExportJSON(store))
		auth.GET("/api/export/csv", ExportCSV(store))
	}

	return r
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/api/
git commit -m "feat: HTTP router with auth middleware"
```

---

### Task 4: Highlight API Handlers

**Covers:** [S6]
**Files:**
- Create: `internal/api/highlights.go`
- Test: `internal/api/highlights_test.go`

- [ ] **Step 1: Write failing test for highlights API**

```go
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
	defer os.Remove(dbPath)

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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/shixl/workdir/moon
go test ./internal/api/... -v
```
Expected: FAIL (CreateHighlights not defined)

- [ ] **Step 3: Create highlights.go**

```go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/storage"
)

type HighlightRequest struct {
	Highlights []struct {
		Text     string `json:"text"`
		Note     string `json:"note"`
		Title    string `json:"title"`
		Author   string `json:"author"`
		Source   string `json:"source"`
		Color    string `json:"color"`
		BookType string `json:"book_type"`
	} `json:"highlights"`
}

func CreateHighlights(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req HighlightRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var created []storage.Highlight
		for _, h := range req.Highlights {
			highlight := &storage.Highlight{
				Text:     h.Text,
				Note:     h.Note,
				Title:    h.Title,
				Author:   h.Author,
				Source:   h.Source,
				Color:    h.Color,
				BookType: h.BookType,
			}

			id, err := store.CreateHighlight(highlight)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			highlight.ID = id
			created = append(created, *highlight)
		}

		c.JSON(http.StatusOK, gin.H{"highlights": created})
	}
}

func ListHighlights(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		highlights, err := store.ListHighlights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"highlights": highlights})
	}
}

func GetHighlight(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		// Parse id...
		var highlightID int64
		if _, err := fmt.Sscanf(id, "%d", &highlightID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		h, err := store.GetHighlight(highlightID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "highlight not found"})
			return
		}
		c.JSON(http.StatusOK, h)
	}
}

func DeleteHighlight(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var highlightID int64
		if _, err := fmt.Sscanf(id, "%d", &highlightID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := store.DeleteHighlight(highlightID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}
```

- [ ] **Step 4: Fix import in highlights.go**

Add `"fmt"` to imports.

- [ ] **Step 5: Run test to verify it passes**

```bash
cd /home/shixl/workdir/moon
go test ./internal/api/... -v
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/api/highlights.go internal/api/highlights_test.go
git commit -m "feat: highlight API handlers with tests"
```

---

### Task 5: Book API Handlers

**Covers:** [S7]
**Files:**
- Create: `internal/api/books.go`

- [ ] **Step 1: Create books.go**

```go
package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/storage"
)

func ListBooks(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		books, err := store.ListBooks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"books": books})
	}
}

func GetBook(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var bookID int64
		if _, err := fmt.Sscanf(id, "%d", &bookID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		b, err := store.GetBook(bookID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}
		c.JSON(http.StatusOK, b)
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/api/books.go
git commit -m "feat: book API handlers"
```

---

### Task 6: Export Handlers

**Covers:** [S8]
**Files:**
- Create: `internal/api/export.go`

- [ ] **Step 1: Create export.go**

```go
package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/storage"
)

func ExportJSON(store *storage.SQLiteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		highlights, err := store.ListHighlights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
```

- [ ] **Step 2: Commit**

```bash
git add internal/api/export.go
git commit -m "feat: JSON and CSV export endpoints"
```

---

### Task 7: Wire Everything Together

**Covers:** [S3, S4, S6]
**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Update main.go**

```go
package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/moon-readwise-server/internal/api"
	"github.com/moon-readwise-server/internal/storage"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Auth struct {
		Token string `yaml:"token"`
	} `yaml:"auth"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	store, err := storage.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	gin.SetMode(gin.ReleaseMode)
	r := api.NewRouter(store, cfg.Auth.Token)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Install Gin dependency**

```bash
cd /home/shixl/workdir/moon
go get -u github.com/gin-gonic/gin
go mod tidy
```

- [ ] **Step 3: Build and test manually**

```bash
go build -o moon-readwise cmd/server/main.go
./moon-readwise
```

In another terminal:
```bash
curl -H "Authorization: Token your-secret-token-here" http://localhost:8080/api/v3/highlights/
```

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go go.mod go.sum
git commit -m "feat: wire server with all components"
```

---

### Task 8: systemd Service File

**Covers:** [S9]
**Files:**
- Create: `deploy/moon-readwise.service`

- [ ] **Step 1: Create service file**

```ini
[Unit]
Description=Moon Readwise Server
After=network.target

[Service]
Type=simple
User=shixl
WorkingDirectory=/home/shixl/workdir/moon
ExecStart=/home/shixl/workdir/moon/moon-readwise
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

- [ ] **Step 2: Commit**

```bash
git add deploy/
git commit -m "feat: systemd service file for deployment"
```

---

### Task 9: Run All Tests

**Files:**
- Test: `internal/storage/sqlite_test.go`
- Test: `internal/api/highlights_test.go`

- [ ] **Step 1: Run all tests**

```bash
cd /home/shixl/workdir/moon
go test ./... -v
```
Expected: All tests PASS

- [ ] **Step 2: Build final binary**

```bash
go build -o moon-readwise cmd/server/main.go
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: ensure all tests pass, final build"
```

---

## Spec Coverage Summary

| Task | Covers |
|------|--------|
| 1 | [S10] Project Structure |
| 2 | [S5] Data Model |
| 3 | [S6] API Endpoints |
| 4 | [S6] Highlight Endpoints |
| 5 | [S7] Book Endpoints |
| 6 | [S8] Export Format |
| 7 | [S3, S4] Architecture & API |
| 8 | [S9] Deployment |
| 9 | [S12] Success Criteria |
