// Package jsoncanvas validates documents in the JSON Canvas 1.0 format.
package jsoncanvas

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrInvalidDocument is returned when canvas JSON fails structural validation.
var ErrInvalidDocument = errors.New("invalid json canvas document")

// Document is the top-level JSON Canvas structure (nodes + edges).
type Document struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node is a positioned canvas element (text, file, link, or group).
type Node struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Text   string  `json:"text,omitempty"`
	File   string  `json:"file,omitempty"`
	URL    string  `json:"url,omitempty"`
	Color  string  `json:"color,omitempty"`
}

// Edge connects two nodes using JSON Canvas field names.
type Edge struct {
	ID       string `json:"id"`
	FromNode string `json:"fromNode"`
	ToNode   string `json:"toNode"`
	FromSide string `json:"fromSide,omitempty"`
	ToSide   string `json:"toSide,omitempty"`
	Label    string `json:"label,omitempty"`
	Color    string `json:"color,omitempty"`
}

// Validate checks that content is a JSON Canvas document with nodes and edges arrays.
// Extra fields are preserved when callers write the original bytes back to storage.
func Validate(content []byte) error {
	var raw struct {
		Nodes *json.RawMessage `json:"nodes"`
		Edges *json.RawMessage `json:"edges"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}
	if raw.Nodes == nil {
		return fmt.Errorf("%w: missing nodes array", ErrInvalidDocument)
	}
	if raw.Edges == nil {
		return fmt.Errorf("%w: missing edges array", ErrInvalidDocument)
	}
	var nodes []json.RawMessage
	if err := json.Unmarshal(*raw.Nodes, &nodes); err != nil {
		return fmt.Errorf("%w: nodes must be an array", ErrInvalidDocument)
	}
	var edges []json.RawMessage
	if err := json.Unmarshal(*raw.Edges, &edges); err != nil {
		return fmt.Errorf("%w: edges must be an array", ErrInvalidDocument)
	}
	return nil
}

// EmptyDocument returns minimal valid JSON for a new canvas file.
func EmptyDocument() []byte {
	return []byte("{\n  \"nodes\": [],\n  \"edges\": []\n}\n")
}
