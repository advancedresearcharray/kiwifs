package jsoncanvas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-graphviz"
)

// AutoLayout repositions nodes in a canvas document using Graphviz.
// It preserves all node/edge data — only x/y coordinates change.
// Returns the updated document, node count, edge count, and any error.
func AutoLayout(doc []byte, algo string) ([]byte, int, int, error) {
	if algo == "" {
		algo = "dot"
	}

	var canvas struct {
		Nodes []json.RawMessage `json:"nodes"`
		Edges []json.RawMessage `json:"edges"`
	}
	if err := json.Unmarshal(doc, &canvas); err != nil {
		return nil, 0, 0, fmt.Errorf("parse canvas: %w", err)
	}

	if len(canvas.Nodes) == 0 {
		out, _ := json.MarshalIndent(canvas, "", "  ")
		return out, 0, len(canvas.Edges), nil
	}

	nodeIDs := make([]string, 0, len(canvas.Nodes))
	for _, raw := range canvas.Nodes {
		id, _ := extractID(raw)
		nodeIDs = append(nodeIDs, id)
	}

	var edges []edgePair
	for _, raw := range canvas.Edges {
		var e edgePair
		_ = json.Unmarshal(raw, &e)
		edges = append(edges, e)
	}

	positions, err := computePositions(nodeIDs, edges, algo)
	if err != nil || len(positions) == 0 {
		positions = gridFallback(nodeIDs)
	}

	for i, raw := range canvas.Nodes {
		id, _ := extractID(raw)
		if pos, ok := positions[id]; ok {
			var obj map[string]json.RawMessage
			_ = json.Unmarshal(raw, &obj)
			obj["x"], _ = json.Marshal(pos.x)
			obj["y"], _ = json.Marshal(pos.y)
			canvas.Nodes[i], _ = json.Marshal(obj)
		}
	}

	out, err := json.MarshalIndent(canvas, "", "  ")
	if err != nil {
		return nil, 0, 0, fmt.Errorf("marshal: %w", err)
	}
	return out, len(canvas.Nodes), len(canvas.Edges), nil
}

type pos struct {
	x, y float64
}

type edgePair struct {
	FromNode string `json:"fromNode"`
	ToNode   string `json:"toNode"`
}

func computePositions(nodeIDs []string, edges []edgePair, algo string) (map[string]pos, error) {
	ctx := context.Background()
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

	nodeMap := make(map[string]*graphviz.Node, len(nodeIDs))
	for _, id := range nodeIDs {
		n, nerr := graph.CreateNodeByName(id)
		if nerr != nil {
			continue
		}
		nodeMap[id] = n
	}

	seen := make(map[string]bool)
	for _, e := range edges {
		if e.FromNode == e.ToNode {
			continue
		}
		key := e.FromNode + "->" + e.ToNode
		if seen[key] {
			continue
		}
		seen[key] = true
		src, ok1 := nodeMap[e.FromNode]
		tgt, ok2 := nodeMap[e.ToNode]
		if !ok1 || !ok2 {
			continue
		}
		graph.CreateEdgeByName("", src, tgt)
	}

	g.SetLayout(graphviz.Layout(algo))

	var buf bytes.Buffer
	if err := g.Render(ctx, graph, "dot", &buf); err != nil {
		return nil, err
	}

	return parseDOT(buf.String(), nodeIDs), nil
}

func parseDOT(dot string, nodeIDs []string) map[string]pos {
	positions := make(map[string]pos)
	for _, line := range strings.Split(dot, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "pos=") || strings.Contains(line, "->") {
			continue
		}
		for _, id := range nodeIDs {
			quoted := fmt.Sprintf(`"%s"`, id)
			if !strings.HasPrefix(line, quoted) && !strings.HasPrefix(line, id+" ") {
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
					positions[id] = pos{x * 1.5, y * 1.5}
				}
			}
			break
		}
	}
	return positions
}

func gridFallback(nodeIDs []string) map[string]pos {
	sort.Strings(nodeIDs)
	cols := int(math.Ceil(math.Sqrt(float64(len(nodeIDs)))))
	const padX, padY = 290.0, 90.0
	positions := make(map[string]pos, len(nodeIDs))
	for i, id := range nodeIDs {
		row := i / cols
		col := i % cols
		positions[id] = pos{float64(col) * padX, float64(row) * padY}
	}
	return positions
}
