package storage

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestCreateAndGetHighlight(t *testing.T) {
	store := newTestStore(t)

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

func TestListHighlightsEmpty(t *testing.T) {
	store := newTestStore(t)

	highlights, err := store.ListHighlights()
	if err != nil {
		t.Fatalf("Failed to list highlights: %v", err)
	}
	if len(highlights) != 0 {
		t.Errorf("Expected 0 highlights, got %d", len(highlights))
	}
}

func TestDeleteHighlight(t *testing.T) {
	store := newTestStore(t)

	h := &Highlight{Text: "to delete", Title: "Del Book"}
	id, err := store.CreateHighlight(h)
	if err != nil {
		t.Fatalf("Failed to create highlight: %v", err)
	}

	if err := store.DeleteHighlight(id); err != nil {
		t.Fatalf("Failed to delete highlight: %v", err)
	}

	_, err = store.GetHighlight(id)
	if err == nil {
		t.Error("Expected error getting deleted highlight, got nil")
	}
}

func TestGetHighlightNotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.GetHighlight(9999)
	if err == nil {
		t.Error("Expected error for non-existent highlight, got nil")
	}
}

func TestCreateHighlightDuplicateReadwiseID(t *testing.T) {
	store := newTestStore(t)

	h1 := &Highlight{ReadwiseID: "rw-1", Text: "first", Title: "Book"}
	_, err := store.CreateHighlight(h1)
	if err != nil {
		t.Fatalf("Failed to create first highlight: %v", err)
	}

	h2 := &Highlight{ReadwiseID: "rw-1", Text: "second", Title: "Book"}
	_, err = store.CreateHighlight(h2)
	if err == nil {
		t.Error("Expected error for duplicate readwise_id, got nil")
	}
}
