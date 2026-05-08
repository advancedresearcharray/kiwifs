package graphutil

import "github.com/kiwifs/kiwifs/internal/links"

// FindComponents uses union-find to identify connected components.
// Returns a map of component ID → member paths.
func FindComponents(edges []links.Edge) map[int][]string {
	parent := make(map[string]string)
	rank := make(map[string]int)

	var find func(string) string
	find = func(x string) string {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}

	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra == rb {
			return
		}
		if rank[ra] < rank[rb] {
			ra, rb = rb, ra
		}
		parent[rb] = ra
		if rank[ra] == rank[rb] {
			rank[ra]++
		}
	}

	for _, e := range edges {
		if _, ok := parent[e.Source]; !ok {
			parent[e.Source] = e.Source
		}
		if _, ok := parent[e.Target]; !ok {
			parent[e.Target] = e.Target
		}
		union(e.Source, e.Target)
	}

	groups := make(map[string][]string)
	for node := range parent {
		root := find(node)
		groups[root] = append(groups[root], node)
	}

	result := make(map[int][]string)
	id := 0
	for _, members := range groups {
		result[id] = members
		id++
	}
	return result
}

// ComputeBetweenness approximates betweenness centrality using BFS
// from a sample of source nodes (Brandes' algorithm).
func ComputeBetweenness(edges []links.Edge) map[string]float64 {
	adj := make(map[string][]string)
	nodes := make(map[string]struct{})
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
		nodes[e.Source] = struct{}{}
		nodes[e.Target] = struct{}{}
	}

	betweenness := make(map[string]float64)
	n := len(nodes)
	if n < 2 {
		return betweenness
	}

	maxSources := 100
	if n < maxSources {
		maxSources = n
	}
	sourceList := make([]string, 0, maxSources)
	for node := range nodes {
		sourceList = append(sourceList, node)
		if len(sourceList) >= maxSources {
			break
		}
	}

	for _, s := range sourceList {
		stack := []string{}
		pred := make(map[string][]string)
		sigma := make(map[string]float64)
		dist := make(map[string]int)
		for node := range nodes {
			dist[node] = -1
		}
		sigma[s] = 1
		dist[s] = 0
		queue := []string{s}

		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			stack = append(stack, v)
			for _, w := range adj[v] {
				if dist[w] < 0 {
					dist[w] = dist[v] + 1
					queue = append(queue, w)
				}
				if dist[w] == dist[v]+1 {
					sigma[w] += sigma[v]
					pred[w] = append(pred[w], v)
				}
			}
		}

		delta := make(map[string]float64)
		for i := len(stack) - 1; i >= 0; i-- {
			w := stack[i]
			for _, v := range pred[w] {
				if sigma[w] > 0 {
					delta[v] += (sigma[v] / sigma[w]) * (1 + delta[w])
				}
			}
			if w != s {
				betweenness[w] += delta[w]
			}
		}
	}

	if n > 2 {
		norm := float64((n - 1) * (n - 2))
		for k := range betweenness {
			betweenness[k] /= norm
		}
	}

	return betweenness
}

// FindTopInCluster returns the member with the highest in-degree.
func FindTopInCluster(members []string, edges []links.Edge) string {
	inDegree := make(map[string]int)
	memberSet := make(map[string]bool)
	for _, m := range members {
		memberSet[m] = true
	}
	for _, e := range edges {
		if memberSet[e.Target] {
			inDegree[e.Target]++
		}
	}
	top := members[0]
	for _, m := range members {
		if inDegree[m] > inDegree[top] {
			top = m
		}
	}
	return top
}
