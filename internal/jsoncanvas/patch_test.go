package jsoncanvas

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestApplyPatch_AddNode(t *testing.T) {
	doc := EmptyDocument()
	ops := []PatchOp{{
		Op:   "add_node",
		Node: json.RawMessage(`{"id":"n1","type":"text","x":0,"y":0,"width":100,"height":50,"text":"hello"}`),
	}}
	out, res, err := ApplyPatch(doc, ops)
	if err != nil {
		t.Fatalf("add_node: %v", err)
	}
	if res.NodesAdded != 1 {
		t.Fatalf("NodesAdded = %d, want 1", res.NodesAdded)
	}
	if !strings.Contains(string(out), `"n1"`) {
		t.Fatalf("output missing n1: %s", out)
	}
}

func TestApplyPatch_UpdateNode(t *testing.T) {
	doc := []byte(`{"nodes":[{"id":"n1","type":"text","x":0,"y":0,"text":"old"}],"edges":[]}`)
	ops := []PatchOp{{
		Op:   "update_node",
		ID:   "n1",
		Node: json.RawMessage(`{"text":"new","color":"#ff0000"}`),
	}}
	out, res, err := ApplyPatch(doc, ops)
	if err != nil {
		t.Fatalf("update_node: %v", err)
	}
	if res.NodesUpdated != 1 {
		t.Fatalf("NodesUpdated = %d", res.NodesUpdated)
	}
	if !strings.Contains(string(out), `"new"`) {
		t.Fatalf("text not updated: %s", out)
	}
	if !strings.Contains(string(out), `"#ff0000"`) {
		t.Fatalf("color not merged: %s", out)
	}
}

func TestApplyPatch_RemoveNodeCascadesEdges(t *testing.T) {
	doc := []byte(`{
		"nodes":[{"id":"a"},{"id":"b"},{"id":"c"}],
		"edges":[
			{"id":"e1","fromNode":"a","toNode":"b"},
			{"id":"e2","fromNode":"b","toNode":"c"},
			{"id":"e3","fromNode":"a","toNode":"c"}
		]
	}`)
	ops := []PatchOp{{Op: "remove_node", ID: "b"}}
	out, res, err := ApplyPatch(doc, ops)
	if err != nil {
		t.Fatalf("remove_node: %v", err)
	}
	if res.NodesRemoved != 1 {
		t.Fatalf("NodesRemoved = %d", res.NodesRemoved)
	}
	if res.EdgesRemoved != 2 {
		t.Fatalf("EdgesRemoved = %d, want 2 (e1+e2)", res.EdgesRemoved)
	}
	if strings.Contains(string(out), `"b"`) {
		t.Fatalf("node b still present: %s", out)
	}
	if !strings.Contains(string(out), `"e3"`) {
		t.Fatalf("edge e3 should survive: %s", out)
	}
}

func TestApplyPatch_AddAndRemoveEdge(t *testing.T) {
	doc := []byte(`{"nodes":[{"id":"a"},{"id":"b"}],"edges":[]}`)
	ops := []PatchOp{
		{Op: "add_edge", Edge: json.RawMessage(`{"id":"e1","fromNode":"a","toNode":"b","label":"link"}`)},
	}
	out, res, err := ApplyPatch(doc, ops)
	if err != nil {
		t.Fatalf("add_edge: %v", err)
	}
	if res.EdgesAdded != 1 {
		t.Fatalf("EdgesAdded = %d", res.EdgesAdded)
	}

	ops2 := []PatchOp{{Op: "remove_edge", ID: "e1"}}
	out2, res2, err := ApplyPatch(out, ops2)
	if err != nil {
		t.Fatalf("remove_edge: %v", err)
	}
	if res2.EdgesRemoved != 1 {
		t.Fatalf("EdgesRemoved = %d", res2.EdgesRemoved)
	}
	if strings.Contains(string(out2), `"e1"`) {
		t.Fatalf("edge e1 still present: %s", out2)
	}
}

func TestApplyPatch_DuplicateNodeRejected(t *testing.T) {
	doc := []byte(`{"nodes":[{"id":"a"}],"edges":[]}`)
	ops := []PatchOp{{
		Op:   "add_node",
		Node: json.RawMessage(`{"id":"a","type":"text"}`),
	}}
	_, _, err := ApplyPatch(doc, ops)
	if err == nil {
		t.Fatal("expected error for duplicate node")
	}
}

func TestApplyPatch_UnknownOpRejected(t *testing.T) {
	doc := EmptyDocument()
	ops := []PatchOp{{Op: "nuke_everything"}}
	_, _, err := ApplyPatch(doc, ops)
	if err == nil {
		t.Fatal("expected error for unknown op")
	}
}

func TestApplyPatch_MultipleBatch(t *testing.T) {
	doc := EmptyDocument()
	ops := []PatchOp{
		{Op: "add_node", Node: json.RawMessage(`{"id":"a","type":"text","x":0,"y":0,"width":100,"height":50,"text":"A"}`)},
		{Op: "add_node", Node: json.RawMessage(`{"id":"b","type":"file","x":200,"y":0,"width":100,"height":50,"file":"notes.md"}`)},
		{Op: "add_edge", Edge: json.RawMessage(`{"id":"e1","fromNode":"a","toNode":"b"}`)},
		{Op: "update_node", ID: "a", Node: json.RawMessage(`{"text":"Updated A"}`)},
	}
	_, res, err := ApplyPatch(doc, ops)
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if res.NodesAdded != 2 || res.EdgesAdded != 1 || res.NodesUpdated != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
}
