package graphutil

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/kiwifs/kiwifs/internal/links"
)

// CanvasNode represents a positioned node for JSON Canvas output.
type CanvasNode struct {
	ID     string  `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Type   string  `json:"type"`
	Text   string  `json:"text,omitempty"`
	File   string  `json:"file,omitempty"`
	Color  string  `json:"color,omitempty"`
}

// CanvasEdge represents a connection between two canvas nodes.
type CanvasEdge struct {
	ID       string `json:"id"`
	FromNode string `json:"fromNode"`
	ToNode   string `json:"toNode"`
	Label    string `json:"label,omitempty"`
}

// CanvasLayout holds the computed layout result.
type CanvasLayout struct {
	Nodes []CanvasNode `json:"nodes"`
	Edges []CanvasEdge `json:"edges"`
}

// LayoutAlgorithm selects the Graphviz engine.
type LayoutAlgorithm string

const (
	LayoutHierarchical LayoutAlgorithm = "dot"
	LayoutRadial       LayoutAlgorithm = "neato"
	LayoutForce        LayoutAlgorithm = "fdp"
	LayoutCircular     LayoutAlgorithm = "circo"
)

var communityColors = []string{
	"#4A90D9", "#E67E22", "#2ECC71", "#9B59B6",
	"#E74C3C", "#1ABC9C", "#F39C12", "#3498DB",
	"#E91E63", "#00BCD4", "#8BC34A", "#FF9800",
}

// GenerateCanvasLayout computes positioned nodes and edges from a link graph
// using Graphviz layout algorithms. Each page becomes a node; wiki-links
// become edges. The layout engine places them spatially so the result can
// be written directly as a .canvas.json file.
func GenerateCanvasLayout(ctx context.Context, edges []links.Edge, pages []string, algo LayoutAlgorithm, communities map[string]int) (*CanvasLayout, error) {
	if len(pages) == 0 {
		return &CanvasLayout{Nodes: []CanvasNode{}, Edges: []CanvasEdge{}}, nil
	}

	if algo == "" {
		algo = LayoutHierarchical
	}

	positions, err := computeGraphvizPositions(ctx, pages, edges, algo)
	if err != nil || len(positions) == 0 {
		return fallbackGridLayout(pages, edges, communities), nil
	}

	return buildCanvasFromPositions(positions, pages, edges, communities), nil
}

type nodePosition struct {
	x, y float64
}

// computeGraphvizPositions builds a DOT graph, runs layout via go-graphviz,
// renders to DOT format (which embeds pos="x,y" attributes), then parses positions.
func computeGraphvizPositions(ctx context.Context, pages []string, edges []links.Edge, algo LayoutAlgorithm) (map[string]nodePosition, error) {
	g, err := graphviz.New(ctx)
	if err != nil {
		return nil, err
	}
	defer g.Close()

	graph, err := g.Graph()
	if err != nil {
		return nil, err
	}
	defer graph.Close()

	nodeMap := make(map[string]*graphviz.Node, len(pages))
	for _, p := range pages {
		n, nerr := graph.CreateNodeByName(p)
		if nerr != nil {
			continue
		}
		n.SetLabel(pageLabel(p))
		nodeMap[p] = n
	}

	edgeSet := make(map[string]bool)
	for _, e := range edges {
		if e.Source == e.Target {
			continue
		}
		key := e.Source + "->" + e.Target
		if edgeSet[key] {
			continue
		}
		edgeSet[key] = true
		srcNode, ok1 := nodeMap[e.Source]
		tgtNode, ok2 := nodeMap[e.Target]
		if !ok1 || !ok2 {
			continue
		}
		graph.CreateEdgeByName("", srcNode, tgtNode)
	}

	g.SetLayout(graphviz.Layout(algo))

	var buf bytes.Buffer
	if err := g.Render(ctx, graph, "dot", &buf); err != nil {
		return nil, err
	}

	return parseDOTPositions(buf.String(), pages), nil
}

// parseDOTPositions extracts node positions from rendered DOT output.
// After layout, Graphviz annotates nodes with pos="x,y" in the output.
func parseDOTPositions(dot string, pages []string) map[string]nodePosition {
	positions := make(map[string]nodePosition)

	for _, line := range strings.Split(dot, "\n") {
		line = strings.TrimSpace(line)

		if !strings.Contains(line, "pos=") {
			continue
		}
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}
		// Skip edge lines (contain "->")
		if strings.Contains(line, "->") {
			continue
		}

		for _, page := range pages {
			quoted := fmt.Sprintf(`"%s"`, page)
			if !strings.HasPrefix(line, quoted) && !strings.HasPrefix(line, page+" ") {
				continue
			}

			posIdx := strings.Index(line, `pos="`)
			if posIdx < 0 {
				continue
			}
			rest := line[posIdx+5:]
			endIdx := strings.Index(rest, `"`)
			if endIdx < 0 {
				continue
			}
			coords := strings.Split(rest[:endIdx], ",")
			if len(coords) >= 2 {
				x, xerr := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
				y, yerr := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)
				if xerr == nil && yerr == nil {
					positions[page] = nodePosition{x, y}
				}
			}
			break
		}
	}
	return positions
}

func buildCanvasFromPositions(positions map[string]nodePosition, pages []string, edges []links.Edge, communities map[string]int) *CanvasLayout {
	const nodeWidth = 250.0
	const nodeHeight = 60.0
	const scale = 1.5

	nodes := make([]CanvasNode, 0, len(pages))
	for _, p := range pages {
		pos, ok := positions[p]
		if !ok {
			continue
		}
		color := ""
		if communities != nil {
			if cid, ok := communities[p]; ok && cid < len(communityColors) {
				color = communityColors[cid]
			}
		}
		nodes = append(nodes, CanvasNode{
			ID:     slugID(p),
			X:      pos.x * scale,
			Y:      pos.y * scale,
			Width:  nodeWidth,
			Height: nodeHeight,
			Type:   "file",
			File:   p,
			Color:  color,
		})
	}

	canvasEdges := buildCanvasEdges(edges)

	return &CanvasLayout{Nodes: nodes, Edges: canvasEdges}
}

func buildCanvasEdges(edges []links.Edge) []CanvasEdge {
	canvasEdges := make([]CanvasEdge, 0)
	seen := make(map[string]bool)
	for i, e := range edges {
		key := e.Source + "->" + e.Target
		if seen[key] || e.Source == e.Target {
			continue
		}
		seen[key] = true
		canvasEdges = append(canvasEdges, CanvasEdge{
			ID:       fmt.Sprintf("edge-%d", i),
			FromNode: slugID(e.Source),
			ToNode:   slugID(e.Target),
		})
	}
	return canvasEdges
}

// fallbackGridLayout provides a simple grid when Graphviz rendering fails.
func fallbackGridLayout(pages []string, edges []links.Edge, communities map[string]int) *CanvasLayout {
	sort.Strings(pages)

	const nodeWidth = 250.0
	const nodeHeight = 60.0
	const paddingX = 40.0
	const paddingY = 30.0

	cols := int(math.Ceil(math.Sqrt(float64(len(pages)))))

	nodes := make([]CanvasNode, len(pages))
	for i, p := range pages {
		row := i / cols
		col := i % cols
		color := ""
		if communities != nil {
			if cid, ok := communities[p]; ok && cid < len(communityColors) {
				color = communityColors[cid]
			}
		}
		nodes[i] = CanvasNode{
			ID:     slugID(p),
			X:      float64(col) * (nodeWidth + paddingX),
			Y:      float64(row) * (nodeHeight + paddingY),
			Width:  nodeWidth,
			Height: nodeHeight,
			Type:   "file",
			File:   p,
			Color:  color,
		}
	}

	return &CanvasLayout{Nodes: nodes, Edges: buildCanvasEdges(edges)}
}

func pageLabel(path string) string {
	name := path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		name = path[idx+1:]
	}
	name = strings.TrimSuffix(name, ".md")
	name = strings.TrimSuffix(name, ".mdx")
	return strings.ReplaceAll(name, "-", " ")
}

func slugID(path string) string {
	s := strings.ReplaceAll(path, "/", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
