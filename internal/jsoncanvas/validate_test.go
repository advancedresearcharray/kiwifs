package jsoncanvas

import (
	"testing"
)

func TestValidate_acceptsEmptyCanvas(t *testing.T) {
	if err := Validate(EmptyDocument()); err != nil {
		t.Fatalf("empty document: %v", err)
	}
}

func TestValidate_preservesSpecFields(t *testing.T) {
	raw := []byte(`{
  "nodes":[{"id":"a","type":"text","x":0,"y":0,"width":100,"height":50,"text":"hi"}],
  "edges":[{"id":"e1","fromNode":"a","toNode":"b","fromSide":"right","toSide":"left","label":"link"}]
}`)
	if err := Validate(raw); err != nil {
		t.Fatalf("valid spec document: %v", err)
	}
}

func TestValidate_rejectsMissingNodes(t *testing.T) {
	if err := Validate([]byte(`{"edges":[]}`)); err == nil {
		t.Fatal("expected error for missing nodes")
	}
}

func TestValidate_rejectsMissingEdges(t *testing.T) {
	if err := Validate([]byte(`{"nodes":[]}`)); err == nil {
		t.Fatal("expected error for missing edges")
	}
}

func TestValidate_rejectsInvalidJSON(t *testing.T) {
	if err := Validate([]byte(`not json`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
