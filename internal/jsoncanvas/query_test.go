package jsoncanvas

import (
	"encoding/json"
	"testing"
)

const testCanvas = `{
  "nodes": [
    {"id":"n1","type":"text","x":0,"y":0,"width":100,"height":50,"text":"Auth flow"},
    {"id":"n2","type":"file","x":200,"y":0,"width":100,"height":50,"file":"auth.md"},
    {"id":"n3","type":"link","x":400,"y":0,"width":100,"height":50,"url":"https://example.com"},
    {"id":"n4","type":"text","x":0,"y":200,"width":100,"height":50,"text":"Database schema"}
  ],
  "edges": [
    {"id":"e1","fromNode":"n1","toNode":"n2","label":"implements"},
    {"id":"e2","fromNode":"n2","toNode":"n3"},
    {"id":"e3","fromNode":"n1","toNode":"n4","label":"related"}
  ]
}`

func TestQuery_FilterByType(t *testing.T) {
	res, err := Query([]byte(testCanvas), QueryParams{NodeType: "file"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(res.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(res.Nodes))
	}
}

func TestQuery_SearchText(t *testing.T) {
	res, err := Query([]byte(testCanvas), QueryParams{Search: "auth"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(res.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2 (n1 text + n2 file)", len(res.Nodes))
	}
}

func TestQuery_Connected(t *testing.T) {
	res, err := Query([]byte(testCanvas), QueryParams{Connected: "n1"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(res.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1 (n1 itself)", len(res.Nodes))
	}
	if len(res.Edges) != 2 {
		t.Fatalf("edges = %d, want 2 (e1, e3)", len(res.Edges))
	}
}

func TestQuery_NodesOnly(t *testing.T) {
	res, err := Query([]byte(testCanvas), QueryParams{NodesOnly: true})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(res.Edges) != 0 {
		t.Fatalf("edges = %d, want 0", len(res.Edges))
	}
	if len(res.Nodes) != 4 {
		t.Fatalf("nodes = %d, want 4", len(res.Nodes))
	}
}

func TestQuery_EdgesOnly(t *testing.T) {
	res, err := Query([]byte(testCanvas), QueryParams{EdgesOnly: true})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(res.Nodes) != 0 {
		t.Fatalf("nodes = %d, want 0", len(res.Nodes))
	}
	if len(res.Edges) != 3 {
		t.Fatalf("edges = %d, want 3", len(res.Edges))
	}
}

func TestQuery_SearchEdgeLabel(t *testing.T) {
	res, err := Query([]byte(testCanvas), QueryParams{Search: "implements", EdgesOnly: true})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(res.Edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(res.Edges))
	}
	var e struct{ ID string }
	_ = json.Unmarshal(res.Edges[0], &e)
	if e.ID != "e1" {
		t.Fatalf("edge id = %s, want e1", e.ID)
	}
}
