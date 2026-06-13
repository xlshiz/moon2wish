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
