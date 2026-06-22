package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/workspace"
	"github.com/spf13/cobra"
)

func TestKnowledgeTemplateEmbedded(t *testing.T) {
	t.Parallel()
	embedded := workspace.EmbeddedTemplates()
	paths := []string{
		"templates/knowledge/SCHEMA.md",
		"templates/knowledge/index.md",
		"templates/knowledge/log.md",
		"templates/knowledge/episodes/example-episode.md",
		"templates/knowledge/pages/getting-started.md",
		"templates/knowledge/playbook.md",
	}
	for _, p := range paths {
		if _, err := fs.Stat(embedded, p); err != nil {
			t.Fatalf("embedded template missing %s: %v", p, err)
		}
	}

	absent := []string{
		"templates/knowledge/concepts",
		"templates/knowledge/entities",
		"templates/knowledge/reports",
		"templates/knowledge/decisions",
		"templates/knowledge/welcome.md",
	}
	for _, p := range absent {
		if _, err := fs.Stat(embedded, p); err == nil {
			t.Fatalf("expected %s to be removed, but it still exists", p)
		}
	}
}

func TestMemoryTemplateRemoved(t *testing.T) {
	t.Parallel()
	if _, err := fs.Stat(workspace.EmbeddedTemplates(), "templates/memory/SCHEMA.md"); err == nil {
		t.Fatal("memory template should be removed from embedded files")
	}
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "init",
		RunE: runInit,
	}
	cmd.Flags().StringP("root", "r", "./knowledge", "directory to initialize")
	cmd.Flags().String("template", "knowledge", "template")
	return cmd
}

func TestKnowledgeTemplateInit(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "kb")

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--root", root, "--template", "knowledge"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	mustExist := []string{
		"pages/getting-started.md",
		"episodes/example-episode.md",
		"SCHEMA.md",
		"index.md",
		"log.md",
		".kiwi/playbook.md",
		".kiwi/config.toml",
	}
	for _, p := range mustExist {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}

	mustNotExist := []string{
		"concepts",
		"entities",
		"reports",
		"decisions",
		"welcome.md",
	}
	for _, p := range mustNotExist {
		if _, err := os.Stat(filepath.Join(root, p)); err == nil {
			t.Errorf("expected %s to NOT exist", p)
		}
	}
}

func TestWikiTemplateEmbedded(t *testing.T) {
	t.Parallel()
	embedded := workspace.EmbeddedTemplates()
	paths := []string{
		"templates/wiki/SCHEMA.md",
		"templates/wiki/index.md",
		"templates/wiki/welcome.md",
		"templates/wiki/how-we-work.md",
		"templates/wiki/architecture.md",
		"templates/wiki/playbook.md",
		"templates/wiki/onboarding/index.md",
		"templates/wiki/decisions/index.md",
		"templates/wiki/decisions/adr-001-example.md",
		"templates/wiki/processes/index.md",
		"templates/wiki/processes/deployment.md",
		"templates/wiki/processes/dev-setup.md",
		"templates/wiki/processes/incident-response.md",
		"templates/wiki/reference/index.md",
		"templates/wiki/reference/glossary.md",
		"templates/wiki/reference/faq.md",
	}
	for _, p := range paths {
		if _, err := fs.Stat(embedded, p); err != nil {
			t.Fatalf("embedded template missing %s: %v", p, err)
		}
	}

	absent := []string{
		"templates/wiki/getting-started.md",
		"templates/wiki/engineering",
		"templates/wiki/product",
	}
	for _, p := range absent {
		if _, err := fs.Stat(embedded, p); err == nil {
			t.Fatalf("expected old wiki file %s to be removed, but it still exists", p)
		}
	}
}

func TestWikiTemplateInit(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "wiki")

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--root", root, "--template", "wiki"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	mustExist := []string{
		"SCHEMA.md",
		"index.md",
		"welcome.md",
		"how-we-work.md",
		"architecture.md",
		"onboarding/index.md",
		"decisions/index.md",
		"decisions/adr-001-example.md",
		"processes/index.md",
		"processes/deployment.md",
		"processes/dev-setup.md",
		"processes/incident-response.md",
		"reference/index.md",
		"reference/glossary.md",
		"reference/faq.md",
		".kiwi/playbook.md",
		".kiwi/config.toml",
		".kiwi/templates/decision.md",
		".kiwi/templates/sop.md",
		".kiwi/templates/meeting-notes.md",
		".gitignore",
	}
	for _, p := range mustExist {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}

	welcome, err := os.ReadFile(filepath.Join(root, "welcome.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(welcome), "title: Welcome") {
		t.Error("welcome.md missing frontmatter")
	}
	if !strings.Contains(string(welcome), "owner:") {
		t.Error("welcome.md missing owner field")
	}

	adr, err := os.ReadFile(filepath.Join(root, "decisions/adr-001-example.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adr), "type: decision") {
		t.Error("adr-001-example.md missing type: decision")
	}
}

func TestMemoryTemplateMigrationError(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "kb")

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--root", root, "--template", "memory"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for memory template, got nil")
	}
	if got := err.Error(); got != "the 'memory' template has been merged into 'knowledge' — use --template knowledge instead" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitCmdDocumentsRunbookTemplate(t *testing.T) {
	t.Parallel()
	usage := initCmd.Flags().Lookup("template").Usage
	if !strings.Contains(usage, "runbook") {
		t.Fatalf("template flag usage missing runbook: %q", usage)
	}
	if !strings.Contains(initCmd.Example, "--template runbook") {
		t.Fatalf("init example missing runbook template:\n%s", initCmd.Example)
	}
}

func TestRunbookTemplateInitBlankRoot(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "empty-parent", "runbooks")

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--root", root, "--template", "runbook"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	cfg, err := os.ReadFile(filepath.Join(root, ".kiwi/config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(cfg)
	for _, want := range []string{"127.0.0.1", "[auth]", "apikey", "perspace", "execution_staleness"} {
		if !strings.Contains(content, want) {
			t.Errorf("config.toml missing %q", want)
		}
	}
}

func TestRunbookTemplateEmbedded(t *testing.T) {
	t.Parallel()
	embedded := workspace.EmbeddedTemplates()
	paths := []string{
		"templates/runbook/SCHEMA.md",
		"templates/runbook/index.md",
		"templates/runbook/playbook.md",
		"templates/runbook/example-high-cpu.md",
		"templates/runbook/.kiwi/schemas/runbook.json",
		"templates/runbook/.kiwi/config.toml",
		"templates/runbook/.kiwi/templates/runbook.md",
	}
	for _, p := range paths {
		if _, err := fs.Stat(embedded, p); err != nil {
			t.Fatalf("embedded template missing %s: %v", p, err)
		}
	}
}

func TestRunbookTemplateInit(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "runbook")

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--root", root, "--template", "runbook"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	mustExist := []string{
		"SCHEMA.md",
		"index.md",
		"example-high-cpu.md",
		".kiwi/schemas/runbook.json",
		".kiwi/templates/runbook.md",
		".kiwi/config.toml",
		".kiwi/playbook.md",
	}
	for _, p := range mustExist {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}

	example, err := os.ReadFile(filepath.Join(root, "example-high-cpu.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"type: runbook",
		"trigger:",
		"severity: P2",
		"owner:",
		"services:",
		"Trigger / When to Use",
		"## 2. Diagnosis",
		"## 3. Mitigation",
		"## 4. Verification",
		"## 5. Rollback",
		"## 6. RTO and Data Loss Expectations",
		"## 7. Escalation Path",
		"```bash",
	} {
		if !strings.Contains(string(example), want) {
			t.Errorf("example runbook missing %q", want)
		}
	}
}
