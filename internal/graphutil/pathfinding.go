package graphutil

import (
	"errors"

	"github.com/kiwifs/kiwifs/internal/links"
)

// ErrNoPath is returned when no path exists between the source and target.
var ErrNoPath = errors.New("no path exists between the given pages")

// ErrNodeNotFound is returned when a start or end node is not in the graph.
var ErrNodeNotFound = errors.New("node not found in graph")

// ShortestPath finds the shortest path between two pages in the link graph
// using BFS. The link graph is treated as directed: an edge from A to B
// means A contains [[B]].
//
// Returns the path as a slice of file paths from `from` to `to` inclusive.
func ShortestPath(linkMap map[string][]string, from, to string) ([]string, error) {
	edges := linksFromMap(linkMap)
	return ShortestPathFromEdges(edges, from, to)
}

// ShortestPathFromEdges finds the shortest path between two pages using
// BFS on the directed edge list.
func ShortestPathFromEdges(edges []links.Edge, from, to string) ([]string, error) {
	if from == to {
		return []string{from}, nil
	}

	// Build directed adjacency list.
	adj := make(map[string][]string)
	nodeSet := make(map[string]struct{})
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
		nodeSet[e.Source] = struct{}{}
		nodeSet[e.Target] = struct{}{}
	}

	// Validate that both nodes exist.
	if _, ok := nodeSet[from]; !ok {
		return nil, ErrNodeNotFound
	}
	if _, ok := nodeSet[to]; !ok {
		return nil, ErrNodeNotFound
	}

	// BFS.
	visited := map[string]bool{from: true}
	parent := make(map[string]string)
	queue := []string{from}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, neighbor := range adj[curr] {
			if visited[neighbor] {
				continue
			}
			visited[neighbor] = true
			parent[neighbor] = curr

			if neighbor == to {
				// Reconstruct path.
				return reconstructPath(parent, from, to), nil
			}

			queue = append(queue, neighbor)
		}
	}

	return nil, ErrNoPath
}

// reconstructPath walks the parent map backwards from target to source
// and returns the path in source-to-target order.
func reconstructPath(parent map[string]string, from, to string) []string {
	var path []string
	for curr := to; curr != from; curr = parent[curr] {
		path = append(path, curr)
	}
	path = append(path, from)

	// Reverse to get from → to order.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
