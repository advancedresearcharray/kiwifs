package cmd

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/memory"
	"github.com/kiwifs/kiwifs/internal/storage"
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

func TestKnowledgeTemplateMemorySchema(t *testing.T) {
	t.Parallel()
	embedded := workspace.EmbeddedTemplates()

	schema, err := fs.ReadFile(embedded, "templates/knowledge/SCHEMA.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## Memory fields",
		"`memory_status`",
		"`valid_from`",
		"`valid_until`",
		"`confidence`",
		"`expires_at`",
		"`ttl`",
		"`scope`",
		"`contradicts`",
	} {
		if !strings.Contains(string(schema), want) {
			t.Errorf("embedded SCHEMA.md missing %q", want)
		}
	}

	episode, err := fs.ReadFile(embedded, "templates/knowledge/episodes/example-episode.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"memory_kind: episodic",
		"scope: user:demo",
		"confidence: 0.9",
		"expires_at: 2026-12-31T00:00:00Z",
	} {
		if !strings.Contains(string(episode), want) {
			t.Errorf("embedded example-episode.md missing %q", want)
		}
	}

	gettingStarted, err := fs.ReadFile(embedded, "templates/knowledge/pages/getting-started.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gettingStarted), "## Memory lifecycle") {
		t.Error("embedded getting-started.md missing Memory lifecycle section")
	}
	if !strings.Contains(string(gettingStarted), "merged-from") {
		t.Error("embedded getting-started.md should mention merged-from in lifecycle")
	}

	root := filepath.Join(t.TempDir(), "kb")
	cmd := newInitCmd()
	cmd.SetArgs([]string{"--root", root, "--template", "knowledge"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	initEpisode, err := os.ReadFile(filepath.Join(root, "episodes/example-episode.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"episode_id: example-001",
		"memory_kind: episodic",
		"scope: user:demo",
		"confidence: 0.9",
		"expires_at: 2026-12-31T00:00:00Z",
	} {
		if !strings.Contains(string(initEpisode), want) {
			t.Errorf("initialized example episode missing %q", want)
		}
	}

	store, err := storage.NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := memory.Scan(context.Background(), store, memory.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.EpisodicCount != 1 {
		t.Fatalf("memory report episodic count = %d, want 1", rep.EpisodicCount)
	}
	if len(rep.Unmerged) != 1 || rep.Unmerged[0].EpisodeID != "example-001" {
		t.Fatalf("memory report unmerged = %+v, want example-001 unmerged", rep.Unmerged)
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
