package workspace

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/schema"
	"github.com/kiwifs/kiwifs/internal/workflow"
)

func TestResearchTemplateReadingWorkflowValid(t *testing.T) {
	t.Parallel()
	data, err := templates.ReadFile("templates/research/.kiwi/workflows/reading.json")
	if err != nil {
		t.Fatal(err)
	}
	var w workflow.Workflow
	if err := json.Unmarshal(data, &w); err != nil {
		t.Fatal(err)
	}
	if err := workflow.Validate(w); err != nil {
		t.Fatalf("reading workflow invalid: %v", err)
	}
	for _, tc := range []struct{ from, to string }{
		{"unread", "reading"},
		{"reading", "annotated"},
		{"annotated", "summarized"},
		{"summarized", "incorporated"},
	} {
		if err := workflow.ValidateTransition(w, tc.from, tc.to); err != nil {
			t.Fatalf("expected transition %s -> %s: %v", tc.from, tc.to, err)
		}
	}
	if err := workflow.ValidateTransition(w, "unread", "incorporated"); err == nil {
		t.Fatal("expected error for unread -> incorporated skip")
	}
}

func TestResearchTemplatePaperSchemaValidatesExample(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "research-schema")
	if err := Init(root, "research"); err != nil {
		t.Fatal(err)
	}
	v := schema.NewValidator(root)
	fm := map[string]any{
		"type":     "paper",
		"title":    "Test Paper",
		"authors":  []any{"Author One"},
		"year":     2024,
		"venue":    "Test Conference",
		"workflow": "reading",
		"state":    "unread",
	}
	if err := v.Validate(fm); err != nil {
		t.Fatalf("valid paper frontmatter rejected: %v", err)
	}
	fm["authors"] = []any{}
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for empty authors")
	}
	delete(fm, "venue")
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for missing venue")
	}
}
