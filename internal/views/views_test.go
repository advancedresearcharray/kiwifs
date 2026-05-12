package views

import (
	"os"
	"path/filepath"
	"testing"
)

func TestViewsListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	views, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(views) != 0 {
		t.Errorf("expected 0 views, got %d", len(views))
	}
}

func TestViewSaveGetDelete(t *testing.T) {
	tmpDir := t.TempDir()

	view := View{
		Name:   "test-view",
		Query:  `TABLE title, status WHERE type = "task"`,
		Layout: "table",
		Columns: []Column{
			{Property: "title", Label: "Title"},
			{Property: "status", Label: "Status"},
		},
	}

	// Save
	if err := Save(tmpDir, view); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, ".kiwi", "views", "test-view.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("view file not created")
	}

	// Get
	retrieved, err := Get(tmpDir, "test-view")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != "test-view" {
		t.Errorf("expected name test-view, got %s", retrieved.Name)
	}
	if retrieved.Query != view.Query {
		t.Errorf("query mismatch")
	}
	if retrieved.Layout != "table" {
		t.Errorf("expected layout table, got %s", retrieved.Layout)
	}

	// List
	views, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}

	// Delete
	if err := Delete(tmpDir, "test-view"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	views, err = List(tmpDir)
	if err != nil {
		t.Fatalf("List after delete failed: %v", err)
	}
	if len(views) != 0 {
		t.Errorf("expected 0 views after delete, got %d", len(views))
	}
}

func TestViewWithFiltersAndSort(t *testing.T) {
	tmpDir := t.TempDir()

	view := View{
		Name:   "filtered-view",
		Query:  `TABLE * WHERE status = "done"`,
		Layout: "list",
		Filters: []Filter{
			{Field: "priority", Operator: "=", Value: "high"},
		},
		Sort: []SortField{
			{Field: "updated", Order: "desc"},
		},
		GroupBy: "assignee",
	}

	if err := Save(tmpDir, view); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := Get(tmpDir, "filtered-view")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(retrieved.Filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(retrieved.Filters))
	}
	if len(retrieved.Sort) != 1 {
		t.Errorf("expected 1 sort field, got %d", len(retrieved.Sort))
	}
	if retrieved.GroupBy != "assignee" {
		t.Errorf("expected groupBy assignee, got %s", retrieved.GroupBy)
	}
}
