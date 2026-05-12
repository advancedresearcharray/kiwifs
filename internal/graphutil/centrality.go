package graphutil

import (
	"sort"

	"github.com/kiwifs/kiwifs/internal/links"
)

// CentralityEntry holds PageRank and betweenness scores for a single page.
type CentralityEntry struct {
	Path        string  `json:"path"`
	PageRank    float64 `json:"pagerank"`
	Betweenness float64 `json:"betweenness"`
	InDegree    int     `json:"in_degree"`
	OutDegree   int     `json:"out_degree"`
}

// PageRankFromLinks computes PageRank for a link map using the iterative
// power-iteration algorithm (damping=0.85, 30 iterations). This wraps the
// existing analytics.go pageRank implementation but accepts the
// map[string][]string format used by higher-level callers.
func PageRankFromLinks(linkMap map[string][]string) map[string]float64 {
	edges := linksFromMap(linkMap)
	if len(edges) == 0 {
		return make(map[string]float64)
	}

	nodeSet := make(map[string]struct{})
	for _, e := range edges {
		nodeSet[e.Source] = struct{}{}
		nodeSet[e.Target] = struct{}{}
	}
	nodes := make([]string, 0, len(nodeSet))
	nodeIdx := make(map[string]int, len(nodeSet))
	for n := range nodeSet {
		nodeIdx[n] = len(nodes)
		nodes = append(nodes, n)
	}

	rank := pageRank(nodes, edges, nodeIdx)
	result := make(map[string]float64, len(nodes))
	for i, n := range nodes {
		result[n] = rank[i]
	}
	return result
}

// BetweennessFromLinks computes betweenness centrality for a link map.
// It delegates to the existing ComputeBetweenness which uses Brandes'
// algorithm with sampling.
func BetweennessFromLinks(linkMap map[string][]string) map[string]float64 {
	edges := linksFromMap(linkMap)
	return ComputeBetweenness(edges)
}

// Centrality computes both PageRank and betweenness centrality for all
// nodes in the link graph and returns a sorted slice (highest PageRank first).
func Centrality(edges []links.Edge) []CentralityEntry {
	if len(edges) == 0 {
		return []CentralityEntry{}
	}

	// Build node set and degree maps.
	nodeSet := make(map[string]struct{})
	inDegree := make(map[string]int)
	outDegree := make(map[string]int)
	for _, e := range edges {
		nodeSet[e.Source] = struct{}{}
		nodeSet[e.Target] = struct{}{}
		outDegree[e.Source]++
		inDegree[e.Target]++
	}

	nodes := make([]string, 0, len(nodeSet))
	nodeIdx := make(map[string]int, len(nodeSet))
	for n := range nodeSet {
		nodeIdx[n] = len(nodes)
		nodes = append(nodes, n)
	}

	// Compute PageRank.
	rank := pageRank(nodes, edges, nodeIdx)

	// Compute betweenness.
	betweenness := ComputeBetweenness(edges)

	// Assemble results.
	entries := make([]CentralityEntry, len(nodes))
	for i, n := range nodes {
		entries[i] = CentralityEntry{
			Path:        n,
			PageRank:    rank[i],
			Betweenness: betweenness[n],
			InDegree:    inDegree[n],
			OutDegree:   outDegree[n],
		}
	}

	// Sort by PageRank descending.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PageRank > entries[j].PageRank
	})

	return entries
}

// linksFromMap converts a map[string][]string to []links.Edge.
func linksFromMap(m map[string][]string) []links.Edge {
	var edges []links.Edge
	for src, targets := range m {
		for _, tgt := range targets {
			edges = append(edges, links.Edge{Source: src, Target: tgt})
		}
	}
	return edges
}
