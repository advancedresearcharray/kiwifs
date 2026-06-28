package rbac

import (
	"os"
	"testing"
)

func TestPublishMetricsStore(t *testing.T) {
	tmp := t.TempDir()

	store, err := NewPublishMetricsStore(tmp)
	if err != nil {
		t.Fatal(err)
	}

	// Initially empty.
	if m := store.Get("test.md"); m != nil {
		t.Errorf("expected nil for unknown path, got %+v", m)
	}

	// Increment creates entry.
	store.Increment("test.md")
	m := store.Get("test.md")
	if m == nil {
		t.Fatal("expected non-nil after increment")
	}
	if m.Views != 1 {
		t.Errorf("expected 1 view, got %d", m.Views)
	}
	if m.FirstViewed.IsZero() {
		t.Error("expected non-zero FirstViewed")
	}
	if m.LastViewed.IsZero() {
		t.Error("expected non-zero LastViewed")
	}

	// Multiple increments.
	store.Increment("test.md")
	store.Increment("test.md")
	m = store.Get("test.md")
	if m.Views != 3 {
		t.Errorf("expected 3 views, got %d", m.Views)
	}

	// Different paths are independent.
	store.Increment("other.md")
	if om := store.Get("other.md"); om == nil || om.Views != 1 {
		t.Errorf("expected 1 view for other.md, got %+v", om)
	}

	// List returns all.
	all := store.List()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}

	// Persistence: reload from disk.
	store2, err := NewPublishMetricsStore(tmp)
	if err != nil {
		t.Fatal(err)
	}
	m2 := store2.Get("test.md")
	if m2 == nil || m2.Views != 3 {
		t.Errorf("expected 3 views after reload, got %+v", m2)
	}
}

func TestPublishMetricsStore_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	// Remove the temp dir to test mkdir behavior.
	os.RemoveAll(tmp)

	store, err := NewPublishMetricsStore(tmp)
	if err != nil {
		t.Fatal(err)
	}
	store.Increment("x.md")
	if m := store.Get("x.md"); m == nil || m.Views != 1 {
		t.Errorf("expected 1 view, got %+v", m)
	}
}
