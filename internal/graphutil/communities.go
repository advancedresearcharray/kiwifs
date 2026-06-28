package graphutil

import (
	"sort"

	loomgraph "github.com/bluuewhale/loom/graph"
	"github.com/kiwifs/kiwifs/internal/links"
)

// Community represents a group of pages detected by community detection.
type Community struct {
	ID    int      `json:"id"`
	Pages []string `json:"pages"`
}

// DetectCommunities runs the Louvain algorithm on the wiki-link graph and
// returns a mapping of file path to community ID.
//
// The link graph is treated as undirected for community detection: if A links
// to B, that implies a topical relationship in both directions.
func DetectCommunities(linkMap map[string][]string) map[string]int {
	edges := linksFromMap(linkMap)
	return DetectCommunitiesFromEdges(edges)
}

// DetectCommunitiesFromEdges runs the Louvain algorithm on the given edge
// list and returns community ID per file path.
func DetectCommunitiesFromEdges(edges []links.Edge) map[string]int {
	if len(edges) == 0 {
		return make(map[string]int)
	}

	// Map string paths to integer node IDs for loom.
	nodeSet := make(map[string]struct{})
	for _, e := range edges {
		nodeSet[e.Source] = struct{}{}
		nodeSet[e.Target] = struct{}{}
	}

	// Sorted for deterministic ID assignment.
	nodes := make([]string, 0, len(nodeSet))
	for n := range nodeSet {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)

	nodeToID := make(map[string]loomgraph.NodeID, len(nodes))
	idToNode := make(map[loomgraph.NodeID]string, len(nodes))
	for i, n := range nodes {
		nid := loomgraph.NodeID(i)
		nodeToID[n] = nid
		idToNode[nid] = n
	}

	// Build an undirected loom graph.
	g := loomgraph.NewGraph(false)
	for _, n := range nodes {
		g.AddNode(nodeToID[n], 1.0)
	}

	// Deduplicate edges: for an undirected graph, only add each pair once.
	type edgePair struct{ a, b loomgraph.NodeID }
	seen := make(map[edgePair]bool)
	for _, e := range edges {
		src := nodeToID[e.Source]
		tgt := nodeToID[e.Target]
		if src == tgt {
			continue // skip self-loops
		}
		lo, hi := src, tgt
		if lo > hi {
			lo, hi = hi, lo
		}
		p := edgePair{lo, hi}
		if seen[p] {
			continue
		}
		seen[p] = true
		g.AddEdge(src, tgt, 1.0)
	}

	// Run Louvain community detection.
	detector := loomgraph.NewLouvain(loomgraph.LouvainOptions{
		Resolution: 1.0,
	})
	result, err := detector.Detect(g)
	if err != nil {
		// Fallback: assign all nodes to community 0 on error (e.g. directed graph).
		fallback := make(map[string]int, len(nodes))
		for _, n := range nodes {
			fallback[n] = 0
		}
		return fallback
	}

	// Convert loom partition back to string keys.
	partition := make(map[string]int, len(nodes))
	for nid, comm := range result.Partition {
		partition[idToNode[nid]] = comm
	}
	return partition
}

// CommunitiesFromEdges runs community detection and returns structured
// community groups sorted by community ID.
func CommunitiesFromEdges(edges []links.Edge) []Community {
	partition := DetectCommunitiesFromEdges(edges)
	if len(partition) == 0 {
		return []Community{}
	}

	// Group by community ID.
	groups := make(map[int][]string)
	for path, comm := range partition {
		groups[comm] = append(groups[comm], path)
	}

	// Build sorted communities.
	communities := make([]Community, 0, len(groups))
	for id, pages := range groups {
		sort.Strings(pages)
		communities = append(communities, Community{
			ID:    id,
			Pages: pages,
		})
	}
	sort.Slice(communities, func(i, j int) bool {
		return communities[i].ID < communities[j].ID
	})

	return communities
}
