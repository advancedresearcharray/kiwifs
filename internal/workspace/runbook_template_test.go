package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/schema"
)

func TestRunbookSchemaValidatesExample(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "runbook-schema")
	if err := Init(root, "runbook"); err != nil {
		t.Fatal(err)
	}
	v := schema.NewValidator(root)
	fm := map[string]any{
		"type":     "runbook",
		"title":    "Test Runbook",
		"trigger":  "CPU > 90%",
		"severity": "P2",
		"owner":    "platform-oncall",
		"services": []any{"[[api-service]]"},
	}
	if err := v.Validate(fm); err != nil {
		t.Fatalf("valid runbook frontmatter rejected: %v", err)
	}
	delete(fm, "trigger")
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for missing trigger")
	}
	fm["trigger"] = "CPU > 90%"
	fm["severity"] = "P5"
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for invalid severity")
	}
	fm["severity"] = "P2"
	fm["services"] = []any{}
	if err := v.Validate(fm); err == nil {
		t.Fatal("expected validation error for empty services")
	}
}

func TestInitRunbookTemplateScaffold(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "runbook-ws")
	if err := Init(root, "runbook"); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		".kiwi/schemas/runbook.json",
		".kiwi/templates/runbook.md",
		".kiwi/config.toml",
		".kiwi/playbook.md",
		"example-high-cpu.md",
		"index.md",
		"SCHEMA.md",
	} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Fatalf("missing %s: %v", p, err)
		}
	}
}

func TestRunbookTemplateLintClean(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "runbook-lint")
	if err := Init(root, "runbook"); err != nil {
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
	data, err := os.ReadFile(filepath.Join(root, "example-high-cpu.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, err := markdown.Frontmatter(data)
	if err != nil {
		t.Fatal(err)
	}
	if verr := sv.Validate(fm); verr != nil {
		t.Fatalf("example runbook schema validation: %v", verr)
	}

	cfg, err := os.ReadFile(filepath.Join(root, ".kiwi/config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(cfg)
	for _, want := range []string{"execution_staleness", "last_executed", "services", "[auth]"} {
		if !strings.Contains(content, want) {
			t.Fatalf("config.toml missing %q:\n%s", want, content)
		}
	}
}

func TestInitRunbookTemplateMetadata(t *testing.T) {
	t.Parallel()
	list, err := ListInitTemplates()
	if err != nil {
		t.Fatal(err)
	}
	var found *InitTemplate
	for i := range list {
		if list[i].ID == "runbook" {
			found = &list[i]
			break
		}
	}
	if found == nil {
		t.Fatal("runbook template not listed")
	}
	if found.Label != "Runbook" {
		t.Fatalf("label = %q, want %q", found.Label, "Runbook")
	}
	if !strings.Contains(found.Description, "runbook") {
		t.Fatalf("description should mention runbooks: %q", found.Description)
	}
}

func TestExampleRunbookHasSevenSections(t *testing.T) {
	t.Parallel()
	data, err := templates.ReadFile("templates/runbook/example-high-cpu.md")
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, section := range []string{
		"Trigger / When to Use",
		"## 2. Diagnosis",
		"## 3. Mitigation",
		"## 4. Verification",
		"## 5. Rollback",
		"## 6. RTO and Data Loss Expectations",
		"## 7. Escalation Path",
		"```bash",
		"Expected output",
	} {
		if !strings.Contains(body, section) {
			t.Fatalf("example runbook missing %q", section)
		}
	}
}

func TestRunbookSchemaRejectsInvalidFrontmatter(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "runbook-schema-reject")
	if err := Init(root, "runbook"); err != nil {
		t.Fatal(err)
	}
	sv := schema.NewValidator(root)

	cases := []struct {
		name string
		fm   map[string]any
	}{
		{
			name: "missing trigger",
			fm: map[string]any{
				"type": "runbook", "title": "RB", "severity": "P2", "owner": "oncall",
				"services": []any{"[[svc]]"},
			},
		},
		{
			name: "missing owner",
			fm: map[string]any{
				"type": "runbook", "title": "RB", "trigger": "alert", "severity": "P2",
				"services": []any{"[[svc]]"},
			},
		},
		{
			name: "missing services",
			fm: map[string]any{
				"type": "runbook", "title": "RB", "trigger": "alert", "severity": "P2",
				"owner": "oncall",
			},
		},
		{
			name: "invalid severity",
			fm: map[string]any{
				"type": "runbook", "title": "RB", "trigger": "alert", "severity": "high",
				"owner": "oncall", "services": []any{"[[svc]]"},
			},
		},
		{
			name: "empty services",
			fm: map[string]any{
				"type": "runbook", "title": "RB", "trigger": "alert", "severity": "P2",
				"owner": "oncall", "services": []any{},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if verr := sv.Validate(tc.fm); verr == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestRunbookConfigHasAuthGuidance(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "runbook-config")
	if err := Init(root, "runbook"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".kiwi/config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"[auth]", "127.0.0.1", "apikey", "perspace"} {
		if !strings.Contains(content, want) {
			t.Fatalf("config.toml missing %q:\n%s", want, content)
		}
	}
}

func TestInitRunbookIntoEmptyParent(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	root := filepath.Join(parent, "nested", "runbooks")
	if err := Init(root, "runbook"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "example-high-cpu.md")); err != nil {
		t.Fatalf("expected scaffold in empty nested dir: %v", err)
	}
}

func TestInitRunbookDoesNotOverwriteExisting(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "runbook")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	custom := []byte("# Custom index\n")
	if err := os.WriteFile(filepath.Join(root, "index.md"), custom, 0644); err != nil {
		t.Fatal(err)
	}
	if err := Init(root, "runbook"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(custom) {
		t.Fatalf("Init overwrote existing index.md:\n%s", data)
	}
	if _, err := os.Stat(filepath.Join(root, "example-high-cpu.md")); err != nil {
		t.Fatal("expected example-high-cpu.md to be created alongside existing index.md")
	}
}
