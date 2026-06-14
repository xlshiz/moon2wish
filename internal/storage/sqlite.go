package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
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

	if h.ReadwiseID == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return 0, fmt.Errorf("failed to generate readwise_id: %w", err)
		}
		h.ReadwiseID = hex.EncodeToString(b)
	}

	result, err := s.db.Exec(
		`INSERT INTO highlights (readwise_id, text, note, title, author, source, color, book_type, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		h.ReadwiseID, h.Text, h.Note, h.Title, h.Author, h.Source, h.Color, h.BookType, h.CreatedAt, h.UpdatedAt,
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
		`SELECT id, readwise_id, text, note, title, author, source, color, book_type, created_at, updated_at
		 FROM highlights WHERE id = ?`, id,
	).Scan(&h.ID, &h.ReadwiseID, &h.Text, &h.Note, &h.Title, &h.Author, &h.Source, &h.Color, &h.BookType, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (s *SQLiteStore) ListHighlights() ([]Highlight, error) {
	rows, err := s.db.Query(
		`SELECT id, readwise_id, text, note, title, author, source, color, book_type, created_at, updated_at
		 FROM highlights ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var highlights []Highlight
	for rows.Next() {
		var h Highlight
		if err := rows.Scan(&h.ID, &h.ReadwiseID, &h.Text, &h.Note, &h.Title, &h.Author, &h.Source, &h.Color, &h.BookType, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		highlights = append(highlights, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return highlights, nil
}

func (s *SQLiteStore) DeleteHighlight(id int64) error {
	_, err := s.db.Exec(`DELETE FROM highlights WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) UpdateHighlight(h *Highlight) error {
	h.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE highlights SET text = ?, note = ?, title = ?, author = ?, source = ?, color = ?, book_type = ?, updated_at = ?
		 WHERE id = ?`,
		h.Text, h.Note, h.Title, h.Author, h.Source, h.Color, h.BookType, h.UpdatedAt, h.ID,
	)
	return err
}

func (s *SQLiteStore) DeleteAllHighlights() error {
	_, err := s.db.Exec(`DELETE FROM highlights`)
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
	if err := rows.Err(); err != nil {
		return nil, err
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
