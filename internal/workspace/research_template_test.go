package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/markdown"
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
	for _, tc := range []struct{ from, to string }{
		{"reading", "unread"},
		{"annotated", "reading"},
		{"summarized", "annotated"},
	} {
		if err := workflow.ValidateTransition(w, tc.from, tc.to); err != nil {
			t.Fatalf("expected backward transition %s -> %s: %v", tc.from, tc.to, err)
		}
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
	delete(fm, "workflow")
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for missing workflow")
	}
	fm["workflow"] = "reading"
	fm["venue"] = "Test Conference"
	fm["authors"] = []any{"Author One"}
	fm["state"] = "done"
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for invalid state enum")
	}
	fm["state"] = "unread"
	fm["workflow"] = "tasks"
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for wrong workflow const")
	}
}

func TestResearchTemplateLintClean(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "research-lint")
	if err := Init(root, "research"); err != nil {
		t.Fatal(err)
	}
	res, err := schema.Lint(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, is := range res.Issues {
		if is.Kind == "broken-link" || is.Kind == "orphan" || is.Kind == "empty-file" {
			t.Fatalf("lint issue: %+v", is)
		}
	}

	sv := schema.NewValidator(root)
	for _, rel := range []string{
		"papers/example-paper.md",
		"papers/transformer-survey.md",
	} {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		fm, err := markdown.Frontmatter(data)
		if err != nil {
			t.Fatalf("%s frontmatter: %v", rel, err)
		}
		if verr := sv.Validate(fm); verr != nil {
			t.Fatalf("%s schema validation: %v", rel, verr)
		}
	}

	cfg, err := os.ReadFile(filepath.Join(root, ".kiwi/config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), "typed_fields") || !strings.Contains(string(cfg), "cites") {
		t.Fatal("research config should enable cites typed_fields")
	}
}

func TestInitResearchTemplateMetadata(t *testing.T) {
	t.Parallel()
	list, err := ListInitTemplates()
	if err != nil {
		t.Fatal(err)
	}
	var found *InitTemplate
	for i := range list {
		if list[i].ID == "research" {
			found = &list[i]
			break
		}
	}
	if found == nil {
		t.Fatal("research template not listed")
	}
	if found.Label != "Research" {
		t.Fatalf("label = %q, want %q", found.Label, "Research")
	}
	if !strings.Contains(found.Description, "reading") {
		t.Fatalf("description should mention reading workflow: %q", found.Description)
	}
}
