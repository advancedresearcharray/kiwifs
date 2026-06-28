package graphutil

import (
	"errors"
	"testing"

	"github.com/kiwifs/kiwifs/internal/links"
)

func TestShortestPathDirect(t *testing.T) {
	linkMap := map[string][]string{
		"a.md": {"b.md"},
		"b.md": {"c.md"},
	}

	path, err := ShortestPath(linkMap, "a.md", "c.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"a.md", "b.md", "c.md"}
	if len(path) != len(expected) {
		t.Fatalf("expected path length %d, got %d", len(expected), len(path))
	}
	for i, p := range path {
		if p != expected[i] {
			t.Errorf("path[%d] = %s, want %s", i, p, expected[i])
		}
	}
}

func TestShortestPathSameNode(t *testing.T) {
	linkMap := map[string][]string{
		"a.md": {"b.md"},
	}

	path, err := ShortestPath(linkMap, "a.md", "a.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 1 || path[0] != "a.md" {
		t.Errorf("expected [a.md], got %v", path)
	}
}

func TestShortestPathNoPath(t *testing.T) {
	linkMap := map[string][]string{
		"a.md": {"b.md"},
		"c.md": {"d.md"},
	}

	_, err := ShortestPath(linkMap, "a.md", "d.md")
	if !errors.Is(err, ErrNoPath) {
		t.Errorf("expected ErrNoPath, got %v", err)
	}
}

func TestShortestPathNodeNotFound(t *testing.T) {
	linkMap := map[string][]string{
		"a.md": {"b.md"},
	}

	_, err := ShortestPath(linkMap, "a.md", "missing.md")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestShortestPathFromEdges(t *testing.T) {
	edges := []links.Edge{
		{Source: "a.md", Target: "b.md"},
		{Source: "b.md", Target: "c.md"},
		{Source: "a.md", Target: "c.md"},
	}

	// Should find the direct a -> c path (length 2) over a -> b -> c.
	path, err := ShortestPathFromEdges(edges, "a.md", "c.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 2 {
		t.Errorf("expected direct path of length 2, got %d: %v", len(path), path)
	}
}

func TestShortestPathFromEdgesEmpty(t *testing.T) {
	_, err := ShortestPathFromEdges(nil, "a.md", "b.md")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("expected ErrNodeNotFound for empty edges, got %v", err)
	}
}
