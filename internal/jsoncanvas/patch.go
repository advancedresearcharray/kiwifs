package jsoncanvas

import (
	"encoding/json"
	"fmt"
)

// PatchRequest is a batch of atomic canvas operations.
// Agents POST an array of operations; they execute in order against the
// current document. If any operation fails, the whole batch is rejected.
type PatchRequest struct {
	Operations []PatchOp `json:"operations"`
}

// PatchOp is a single canvas mutation.
//
// Supported op values:
//   - add_node:    insert a new node (Node in payload)
//   - update_node: merge fields into an existing node by ID
//   - remove_node: delete a node and its connected edges by ID
//   - add_edge:    insert a new edge (Edge in payload)
//   - update_edge: merge fields into an existing edge by ID
//   - remove_edge: delete an edge by ID
type PatchOp struct {
	Op   string          `json:"op"`
	ID   string          `json:"id,omitempty"`
	Node json.RawMessage `json:"node,omitempty"`
	Edge json.RawMessage `json:"edge,omitempty"`
}

// PatchResult summarizes what a patch changed.
type PatchResult struct {
	NodesAdded   int `json:"nodes_added"`
	NodesUpdated int `json:"nodes_updated"`
	NodesRemoved int `json:"nodes_removed"`
	EdgesAdded   int `json:"edges_added"`
	EdgesUpdated int `json:"edges_updated"`
	EdgesRemoved int `json:"edges_removed"`
}

// ApplyPatch applies a batch of operations to a raw canvas document.
// Returns the updated document bytes and a summary, or an error on the
// first invalid operation (no partial application).
func ApplyPatch(doc []byte, ops []PatchOp) ([]byte, *PatchResult, error) {
	var canvas struct {
		Nodes []json.RawMessage `json:"nodes"`
		Edges []json.RawMessage `json:"edges"`
	}
	if err := json.Unmarshal(doc, &canvas); err != nil {
		return nil, nil, fmt.Errorf("parse canvas: %w", err)
	}

	nodeIdx := buildIDIndex(canvas.Nodes)
	edgeIdx := buildIDIndex(canvas.Edges)
	res := &PatchResult{}

	for i, op := range ops {
		switch op.Op {
		case "add_node":
			id, err := extractID(op.Node)
			if err != nil {
				return nil, nil, fmt.Errorf("op %d (add_node): %w", i, err)
			}
			if _, exists := nodeIdx[id]; exists {
				return nil, nil, fmt.Errorf("op %d (add_node): node %q already exists", i, id)
			}
			canvas.Nodes = append(canvas.Nodes, op.Node)
			nodeIdx[id] = len(canvas.Nodes) - 1
			res.NodesAdded++

		case "update_node":
			id := op.ID
			if id == "" {
				id, _ = extractID(op.Node)
			}
			idx, exists := nodeIdx[id]
			if !exists {
				return nil, nil, fmt.Errorf("op %d (update_node): node %q not found", i, id)
			}
			merged, err := mergeJSON(canvas.Nodes[idx], op.Node)
			if err != nil {
				return nil, nil, fmt.Errorf("op %d (update_node): %w", i, err)
			}
			canvas.Nodes[idx] = merged
			res.NodesUpdated++

		case "remove_node":
			id := op.ID
			idx, exists := nodeIdx[id]
			if !exists {
				return nil, nil, fmt.Errorf("op %d (remove_node): node %q not found", i, id)
			}
			canvas.Nodes = append(canvas.Nodes[:idx], canvas.Nodes[idx+1:]...)
			nodeIdx = buildIDIndex(canvas.Nodes)
			removed := removeEdgesForNode(id, &canvas.Edges)
			edgeIdx = buildIDIndex(canvas.Edges)
			res.NodesRemoved++
			res.EdgesRemoved += removed

		case "add_edge":
			id, err := extractID(op.Edge)
			if err != nil {
				return nil, nil, fmt.Errorf("op %d (add_edge): %w", i, err)
			}
			if _, exists := edgeIdx[id]; exists {
				return nil, nil, fmt.Errorf("op %d (add_edge): edge %q already exists", i, id)
			}
			canvas.Edges = append(canvas.Edges, op.Edge)
			edgeIdx[id] = len(canvas.Edges) - 1
			res.EdgesAdded++

		case "update_edge":
			id := op.ID
			if id == "" {
				id, _ = extractID(op.Edge)
			}
			idx, exists := edgeIdx[id]
			if !exists {
				return nil, nil, fmt.Errorf("op %d (update_edge): edge %q not found", i, id)
			}
			merged, err := mergeJSON(canvas.Edges[idx], op.Edge)
			if err != nil {
				return nil, nil, fmt.Errorf("op %d (update_edge): %w", i, err)
			}
			canvas.Edges[idx] = merged
			res.EdgesUpdated++

		case "remove_edge":
			id := op.ID
			idx, exists := edgeIdx[id]
			if !exists {
				return nil, nil, fmt.Errorf("op %d (remove_edge): edge %q not found", i, id)
			}
			canvas.Edges = append(canvas.Edges[:idx], canvas.Edges[idx+1:]...)
			edgeIdx = buildIDIndex(canvas.Edges)
			res.EdgesRemoved++

		default:
			return nil, nil, fmt.Errorf("op %d: unknown op %q", i, op.Op)
		}
	}

	out, err := json.MarshalIndent(canvas, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}
	return out, res, nil
}

func buildIDIndex(items []json.RawMessage) map[string]int {
	idx := make(map[string]int, len(items))
	for i, raw := range items {
		id, _ := extractID(raw)
		if id != "" {
			idx[id] = i
		}
	}
	return idx
}

func extractID(raw json.RawMessage) (string, error) {
	if raw == nil {
		return "", fmt.Errorf("payload is nil")
	}
	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", fmt.Errorf("cannot read id: %w", err)
	}
	if obj.ID == "" {
		return "", fmt.Errorf("id is required")
	}
	return obj.ID, nil
}

// mergeJSON does a shallow merge of patch fields into base (patch wins).
func mergeJSON(base, patch json.RawMessage) (json.RawMessage, error) {
	var baseMap map[string]json.RawMessage
	if err := json.Unmarshal(base, &baseMap); err != nil {
		return nil, err
	}
	var patchMap map[string]json.RawMessage
	if err := json.Unmarshal(patch, &patchMap); err != nil {
		return nil, err
	}
	for k, v := range patchMap {
		baseMap[k] = v
	}
	return json.Marshal(baseMap)
}

// removeEdgesForNode removes edges that reference the given node ID
// (as fromNode or toNode). Returns the count of removed edges.
func removeEdgesForNode(nodeID string, edges *[]json.RawMessage) int {
	removed := 0
	kept := (*edges)[:0]
	for _, raw := range *edges {
		var e struct {
			FromNode string `json:"fromNode"`
			ToNode   string `json:"toNode"`
		}
		_ = json.Unmarshal(raw, &e)
		if e.FromNode == nodeID || e.ToNode == nodeID {
			removed++
		} else {
			kept = append(kept, raw)
		}
	}
	*edges = kept
	return removed
}
