package graphutil

import (
	"testing"

	"github.com/kiwifs/kiwifs/internal/links"
)

func TestDetectCommunities(t *testing.T) {
	// Two clusters with a bridge:
	// Cluster 1: a-b-c (densely connected)
	// Cluster 2: d-e-f (densely connected)
	// Bridge: c -> d
	linkMap := map[string][]string{
		"a.md": {"b.md", "c.md"},
		"b.md": {"a.md", "c.md"},
		"c.md": {"a.md", "b.md", "d.md"},
		"d.md": {"e.md", "f.md"},
		"e.md": {"d.md", "f.md"},
		"f.md": {"d.md", "e.md"},
	}

	partition := DetectCommunities(linkMap)
	if len(partition) != 6 {
		t.Fatalf("expected 6 nodes, got %d", len(partition))
	}

	// All nodes should have a community assignment.
	for _, n := range []string{"a.md", "b.md", "c.md", "d.md", "e.md", "f.md"} {
		if _, ok := partition[n]; !ok {
			t.Errorf("node %s missing from partition", n)
		}
	}

	// Nodes in the same dense cluster should be in the same community.
	if partition["a.md"] != partition["b.md"] {
		t.Errorf("expected a.md and b.md in same community, got %d vs %d",
			partition["a.md"], partition["b.md"])
	}
	if partition["d.md"] != partition["e.md"] {
		t.Errorf("expected d.md and e.md in same community, got %d vs %d",
			partition["d.md"], partition["e.md"])
	}
}

func TestDetectCommunitiesEmpty(t *testing.T) {
	partition := DetectCommunities(map[string][]string{})
	if len(partition) != 0 {
		t.Fatalf("expected empty partition, got %d", len(partition))
	}
}

func TestCommunitiesFromEdges(t *testing.T) {
	edges := []links.Edge{
		{Source: "a.md", Target: "b.md"},
		{Source: "b.md", Target: "a.md"},
		{Source: "c.md", Target: "d.md"},
		{Source: "d.md", Target: "c.md"},
	}

	communities := CommunitiesFromEdges(edges)
	if len(communities) == 0 {
		t.Fatal("expected at least one community")
	}

	// All pages should be covered.
	totalPages := 0
	for _, c := range communities {
		totalPages += len(c.Pages)
	}
	if totalPages != 4 {
		t.Errorf("expected 4 total pages, got %d", totalPages)
	}

	// Communities should be sorted by ID.
	for i := 1; i < len(communities); i++ {
		if communities[i].ID < communities[i-1].ID {
			t.Errorf("communities not sorted by ID at index %d", i)
		}
	}
}

func TestCommunitiesFromEdgesEmpty(t *testing.T) {
	communities := CommunitiesFromEdges(nil)
	if len(communities) != 0 {
		t.Fatalf("expected empty communities, got %d", len(communities))
	}
}
