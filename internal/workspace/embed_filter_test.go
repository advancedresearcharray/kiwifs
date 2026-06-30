package workspace

import (
	"io/fs"
	"testing"
)

func TestIsExcludedEmbedPath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path string
		want bool
	}{
		{"templates/runbook/incidents", true},
		{"templates/runbook/incidents/template.md", true},
		{"templates/runbook/postmortems/template.md", true},
		{"templates/runbook/procedures/scale-up.md", true},
		{"templates/research/experiments/exp-001-baseline.md", true},
		{"templates/research/literature/example-paper.md", true},
		{"templates/knowledge/index.md", true},
		{"templates/runbook/example-high-cpu.md", false},
		{"templates/runbook/SCHEMA.md", false},
		{"templates/runbook/.kiwi/schemas/runbook.json", false},
		{"templates/runbook/services/api-service.md", false},
		{"templates/research/papers/example-paper.md", false},
		{"templates/wiki/index.md", false},
	}
	for _, tc := range cases {
		if got := isExcludedEmbedPath(tc.path); got != tc.want {
			t.Errorf("isExcludedEmbedPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestFilteredTemplatesFSHidesLegacyRunbookPaths(t *testing.T) {
	t.Parallel()
	legacy := []string{
		"templates/runbook/incidents/template.md",
		"templates/runbook/postmortems/template.md",
		"templates/runbook/procedures/deploy-rollback.md",
		"templates/knowledge/index.md",
		"templates/research/experiments/exp-001-baseline.md",
	}
	for _, p := range legacy {
		if _, err := templates.ReadFile(p); err == nil {
			t.Fatalf("filtered FS should hide %q", p)
		}
	}

	entries, err := fs.ReadDir(templates, "templates/runbook")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		switch e.Name() {
		case "incidents", "postmortems", "procedures":
			t.Fatalf("legacy dir %q listed in templates/runbook", e.Name())
		}
	}
}
