package graphutil

import (
	"sort"

	"github.com/kiwifs/kiwifs/internal/links"
)

type PageRankEntry struct {
	Path      string  `json:"path"`
	PageRank  float64 `json:"pagerank"`
	InDegree  int     `json:"in_degree"`
	OutDegree int     `json:"out_degree"`
}

type Result struct {
	TotalNodes           int             `json:"total_nodes"`
	TotalEdges           int             `json:"total_edges"`
	Components           int             `json:"components"`
	TopPages             []PageRankEntry `json:"top_pages"`
	Orphans              []string        `json:"orphans"`
	LargestComponentSize int             `json:"largest_component_size"`
}

func Analyze(edges []links.Edge, limit int) *Result {
	if limit <= 0 {
		limit = 20
	}

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
	nodeIdx := make(map[string]int)
	for n := range nodeSet {
		nodeIdx[n] = len(nodes)
		nodes = append(nodes, n)
	}

	n := len(nodes)
	if n == 0 {
		return &Result{Orphans: []string{}, TopPages: []PageRankEntry{}}
	}

	rank := pageRank(nodes, edges, nodeIdx)

	type rankedNode struct {
		path string
		rank float64
	}
	ranked := make([]rankedNode, n)
	for i, node := range nodes {
		ranked[i] = rankedNode{path: node, rank: rank[i]}
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].rank > ranked[j].rank })

	topN := limit
	if topN > len(ranked) {
		topN = len(ranked)
	}
	topPages := make([]PageRankEntry, topN)
	for i := 0; i < topN; i++ {
		topPages[i] = PageRankEntry{
			Path:      ranked[i].path,
			PageRank:  ranked[i].rank,
			InDegree:  inDegree[ranked[i].path],
			OutDegree: outDegree[ranked[i].path],
		}
	}

	var orphans []string
	for _, node := range nodes {
		if inDegree[node] == 0 {
			orphans = append(orphans, node)
		}
	}
	if orphans == nil {
		orphans = []string{}
	}

	components, largestComponent := connectedComponents(n, edges, nodeIdx)

	return &Result{
		TotalNodes:           n,
		TotalEdges:           len(edges),
		Components:           components,
		TopPages:             topPages,
		Orphans:              orphans,
		LargestComponentSize: largestComponent,
	}
}

func pageRank(nodes []string, edges []links.Edge, nodeIdx map[string]int) []float64 {
	n := len(nodes)
	damping := 0.85
	iterations := 30
	rank := make([]float64, n)
	newRank := make([]float64, n)
	for i := range rank {
		rank[i] = 1.0 / float64(n)
	}

	adjOut := make([][]int, n)
	for _, e := range edges {
		si := nodeIdx[e.Source]
		ti := nodeIdx[e.Target]
		adjOut[si] = append(adjOut[si], ti)
	}

	for iter := 0; iter < iterations; iter++ {
		base := (1.0 - damping) / float64(n)
		for i := range newRank {
			newRank[i] = base
		}
		for i := range nodes {
			if len(adjOut[i]) == 0 {
				share := damping * rank[i] / float64(n)
				for j := range newRank {
					newRank[j] += share
				}
			} else {
				share := damping * rank[i] / float64(len(adjOut[i]))
				for _, j := range adjOut[i] {
					newRank[j] += share
				}
			}
		}
		rank, newRank = newRank, rank
	}
	return rank
}

func connectedComponents(n int, edges []links.Edge, nodeIdx map[string]int) (components, largest int) {
	adj := make([][]int, n)
	for _, e := range edges {
		si := nodeIdx[e.Source]
		ti := nodeIdx[e.Target]
		adj[si] = append(adj[si], ti)
		adj[ti] = append(adj[ti], si)
	}

	visited := make([]bool, n)
	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}
		components++
		size := 0
		queue := []int{i}
		visited[i] = true
		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			size++
			for _, nb := range adj[curr] {
				if !visited[nb] {
					visited[nb] = true
					queue = append(queue, nb)
				}
			}
		}
		if size > largest {
			largest = size
		}
	}
	return
}
