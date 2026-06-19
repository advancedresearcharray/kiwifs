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

func TestADRWorkflowValid(t *testing.T) {
	t.Parallel()
	data, err := templates.ReadFile("templates/adr/.kiwi/workflows/adr.json")
	if err != nil {
		t.Fatal(err)
	}
	var w workflow.Workflow
	if err := json.Unmarshal(data, &w); err != nil {
		t.Fatal(err)
	}
	if err := workflow.Validate(w); err != nil {
		t.Fatalf("adr workflow invalid: %v", err)
	}
	if w.Name != "adr" {
		t.Fatalf("workflow name = %q, want %q", w.Name, "adr")
	}
	for _, tc := range []struct{ from, to string }{
		{"proposed", "accepted"},
		{"proposed", "deprecated"},
		{"accepted", "deprecated"},
		{"accepted", "superseded"},
		{"deprecated", "superseded"},
	} {
		if err := workflow.ValidateTransition(w, tc.from, tc.to); err != nil {
			t.Fatalf("expected transition %s -> %s: %v", tc.from, tc.to, err)
		}
	}
	if err := workflow.ValidateTransition(w, "proposed", "superseded"); err == nil {
		t.Fatal("expected error for proposed -> superseded skip")
	}
	if err := workflow.ValidateTransition(w, "accepted", "proposed"); err == nil {
		t.Fatal("expected error for backward accepted -> proposed")
	}
	for _, to := range []string{"proposed", "accepted", "deprecated"} {
		if err := workflow.ValidateTransition(w, "superseded", to); err == nil {
			t.Fatalf("expected error for terminal superseded -> %s", to)
		}
	}
}

func TestADRSchemaValidatesExample(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "adr-schema")
	if err := Init(root, "adr"); err != nil {
		t.Fatal(err)
	}
	v := schema.NewValidator(root)
	fm := map[string]any{
		"type":     "adr",
		"title":    "Test ADR",
		"status":   "proposed",
		"date":     "2026-06-19",
		"deciders": []any{"team-a"},
		"workflow": "adr",
		"state":    "proposed",
	}
	if err := v.Validate(fm); err != nil {
		t.Fatalf("valid adr frontmatter rejected: %v", err)
	}
	fm["deciders"] = []any{}
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for empty deciders")
	}
	delete(fm, "date")
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for missing date")
	}
	fm["date"] = "2026-06-19"
	fm["deciders"] = []any{"team-a"}
	fm["status"] = "draft"
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for invalid status enum")
	}
	fm["status"] = "proposed"
	fm["state"] = "draft"
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for invalid state enum")
	}
}

func TestInitADRTemplateScaffold(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "adr-ws")
	if err := Init(root, "adr"); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		".kiwi/workflows/adr.json",
		".kiwi/schemas/adr.json",
		".kiwi/templates/adr.md",
		".kiwi/config.toml",
		".kiwi/playbook.md",
		"decisions/ADR-001-use-markdown-for-adrs.md",
		"index.md",
		"SCHEMA.md",
	} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Fatalf("missing %s: %v", p, err)
		}
	}
}

func TestADRTemplateLintClean(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "adr-lint")
	if err := Init(root, "adr"); err != nil {
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
	data, err := os.ReadFile(filepath.Join(root, "decisions/ADR-001-use-markdown-for-adrs.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, err := markdown.Frontmatter(data)
	if err != nil {
		t.Fatal(err)
	}
	if verr := sv.Validate(fm); verr != nil {
		t.Fatalf("example ADR schema validation: %v", verr)
	}
	if fm["status"] != "accepted" || fm["state"] != "accepted" {
		t.Fatalf("example ADR status/state mismatch: %+v", fm)
	}

	cfg, err := os.ReadFile(filepath.Join(root, ".kiwi/config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(cfg)
	for _, want := range []string{"auto_sequence", "decisions/", "adr_number", "supersedes"} {
		if !strings.Contains(content, want) {
			t.Fatalf("config.toml missing %q:\n%s", want, content)
		}
	}
}

func TestInitADRTemplateMetadata(t *testing.T) {
	t.Parallel()
	list, err := ListInitTemplates()
	if err != nil {
		t.Fatal(err)
	}
	var found *InitTemplate
	for i := range list {
		if list[i].ID == "adr" {
			found = &list[i]
			break
		}
	}
	if found == nil {
		t.Fatal("adr template not listed")
	}
	if found.Label != "Architecture Decision Records" {
		t.Fatalf("label = %q, want %q", found.Label, "Architecture Decision Records")
	}
	if !strings.Contains(found.Description, "MADR") {
		t.Fatalf("description should mention MADR: %q", found.Description)
	}
}

func TestExampleADRHasMADRSections(t *testing.T) {
	t.Parallel()
	data, err := templates.ReadFile("templates/adr/decisions/ADR-001-use-markdown-for-adrs.md")
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, section := range []string{
		"Context and Problem Statement",
		"Decision Drivers",
		"Considered Options",
		"Decision Outcome",
		"### Positive",
		"### Negative",
		"### Neutral",
	} {
		if !strings.Contains(body, section) {
			t.Fatalf("example ADR missing section %q", section)
		}
	}
}
