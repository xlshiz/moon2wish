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
