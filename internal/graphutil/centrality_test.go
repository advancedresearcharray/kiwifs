package graphutil

import (
	"testing"

	"github.com/kiwifs/kiwifs/internal/links"
)

func TestPageRankFromLinks(t *testing.T) {
	linkMap := map[string][]string{
		"a.md": {"b.md", "c.md"},
		"b.md": {"c.md"},
		"c.md": {"a.md"},
	}

	ranks := PageRankFromLinks(linkMap)
	if len(ranks) != 3 {
		t.Fatalf("expected 3 ranks, got %d", len(ranks))
	}
	for _, path := range []string{"a.md", "b.md", "c.md"} {
		if r, ok := ranks[path]; !ok || r <= 0 {
			t.Errorf("expected positive rank for %s, got %v", path, r)
		}
	}
}

func TestPageRankFromLinksEmpty(t *testing.T) {
	ranks := PageRankFromLinks(map[string][]string{})
	if len(ranks) != 0 {
		t.Fatalf("expected empty ranks, got %d", len(ranks))
	}
}

func TestBetweennessFromLinks(t *testing.T) {
	// Linear chain: a -> b -> c -> d
	linkMap := map[string][]string{
		"a.md": {"b.md"},
		"b.md": {"c.md"},
		"c.md": {"d.md"},
	}

	betweenness := BetweennessFromLinks(linkMap)
	// b and c should have higher betweenness than endpoints.
	if betweenness["b.md"] <= betweenness["a.md"] {
		t.Errorf("expected b.md to have higher betweenness than a.md")
	}
	if betweenness["c.md"] <= betweenness["d.md"] {
		t.Errorf("expected c.md to have higher betweenness than d.md")
	}
}

func TestCentrality(t *testing.T) {
	edges := []links.Edge{
		{Source: "a.md", Target: "b.md"},
		{Source: "b.md", Target: "c.md"},
		{Source: "c.md", Target: "a.md"},
	}

	entries := Centrality(edges)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	// All entries should have non-zero PageRank.
	for _, e := range entries {
		if e.PageRank <= 0 {
			t.Errorf("expected positive PageRank for %s", e.Path)
		}
	}
	// Should be sorted by PageRank descending.
	for i := 1; i < len(entries); i++ {
		if entries[i].PageRank > entries[i-1].PageRank {
			t.Errorf("entries not sorted by PageRank descending at index %d", i)
		}
	}
}

func TestCentralityEmpty(t *testing.T) {
	entries := Centrality(nil)
	if len(entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(entries))
	}
}
