package draft_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/draft"
)

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	// Create an initial commit so HEAD exists
	seedFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(seedFile, []byte("# Test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")
	return dir
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %s: %v", name, args, out, err)
	}
}

func TestCreateAndList(t *testing.T) {
	root := initGitRepo(t)
	mgr, err := draft.NewManager(root, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	d1, err := mgr.Create(ctx, "agent-1")
	if err != nil {
		t.Fatal(err)
	}
	d2, err := mgr.Create(ctx, "agent-2")
	if err != nil {
		t.Fatal(err)
	}
	d3, err := mgr.Create(ctx, "agent-3")
	if err != nil {
		t.Fatal(err)
	}

	list := mgr.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 drafts, got %d", len(list))
	}

	// Verify each draft is retrievable
	for _, d := range []*draft.Draft{d1, d2, d3} {
		got, err := mgr.Get(d.ID)
		if err != nil {
			t.Fatalf("Get(%s): %v", d.ID, err)
		}
		if got.Branch != d.Branch {
			t.Fatalf("branch mismatch: %s != %s", got.Branch, d.Branch)
		}
	}

	// Verify not found
	_, err = mgr.Get("nonexistent")
	if err != draft.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestWriteAndRead(t *testing.T) {
	root := initGitRepo(t)
	mgr, err := draft.NewManager(root, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	d, err := mgr.Create(ctx, "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	pipe, err := mgr.Pipeline(d.ID)
	if err != nil {
		t.Fatal(err)
	}

	content := []byte("# Draft page\nThis is draft content.\n")
	_, err = pipe.Write(ctx, "docs/new-page.md", content, "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	// Read back from draft
	got, err := pipe.Store.Read(ctx, "docs/new-page.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("content mismatch:\n  got:  %q\n  want: %q", got, content)
	}

	// Verify main doesn't have it
	mainFile := filepath.Join(root, "docs", "new-page.md")
	if _, err := os.Stat(mainFile); !os.IsNotExist(err) {
		t.Fatal("expected file to not exist in main worktree")
	}
}

func TestDiff(t *testing.T) {
	root := initGitRepo(t)
	mgr, err := draft.NewManager(root, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	d, err := mgr.Create(ctx, "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	pipe, err := mgr.Pipeline(d.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pipe.Write(ctx, "docs/diff-test.md", []byte("# Diff test\n"), "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	diff, err := mgr.Diff(ctx, d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff == "" {
		t.Fatal("expected non-empty diff")
	}
	if !containsString(diff, "diff-test.md") {
		t.Fatalf("diff doesn't mention the changed file:\n%s", diff)
	}
}

func TestMerge(t *testing.T) {
	root := initGitRepo(t)
	mgr, err := draft.NewManager(root, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	d, err := mgr.Create(ctx, "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	pipe, err := mgr.Pipeline(d.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pipe.Write(ctx, "docs/merged.md", []byte("# Merged\n"), "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	if err := mgr.Merge(ctx, d.ID); err != nil {
		t.Fatal(err)
	}

	// Verify main now has the file
	mainFile := filepath.Join(root, "docs", "merged.md")
	data, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("expected file in main after merge: %v", err)
	}
	if string(data) != "# Merged\n" {
		t.Fatalf("unexpected content: %q", data)
	}

	// Verify draft is gone
	if _, err := mgr.Get(d.ID); err != draft.ErrNotFound {
		t.Fatalf("expected draft to be removed after merge, got %v", err)
	}
}

func TestMergeConflict(t *testing.T) {
	root := initGitRepo(t)
	mgr, err := draft.NewManager(root, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	d, err := mgr.Create(ctx, "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	pipe, err := mgr.Pipeline(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pipe.Write(ctx, "docs/conflict.md", []byte("# From draft\n"), "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	// Advance main independently
	mainFile := filepath.Join(root, "docs")
	os.MkdirAll(mainFile, 0o755)
	os.WriteFile(filepath.Join(mainFile, "main-only.md"), []byte("# Main\n"), 0o644)
	run(t, root, "git", "add", ".")
	run(t, root, "git", "commit", "-m", "advance main")

	err = mgr.Merge(ctx, d.ID)
	if err != draft.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestDiscard(t *testing.T) {
	root := initGitRepo(t)
	mgr, err := draft.NewManager(root, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	d, err := mgr.Create(ctx, "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	pipe, err := mgr.Pipeline(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pipe.Write(ctx, "docs/discarded.md", []byte("# Discarded\n"), "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	if err := mgr.Discard(ctx, d.ID); err != nil {
		t.Fatal(err)
	}

	// Verify draft is gone
	if _, err := mgr.Get(d.ID); err != draft.ErrNotFound {
		t.Fatalf("expected ErrNotFound after discard")
	}

	// Verify worktree directory is removed
	if _, err := os.Stat(d.WorkDir); !os.IsNotExist(err) {
		t.Fatal("expected worktree directory to be removed")
	}

	// Verify branch is removed
	cmd := exec.Command("git", "branch", "--list", d.Branch)
	cmd.Dir = root
	out, _ := cmd.Output()
	if len(out) > 0 {
		t.Fatalf("expected branch to be deleted, but found: %s", out)
	}
}

func TestCleanupOrphans(t *testing.T) {
	root := initGitRepo(t)
	mgr, err := draft.NewManager(root, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	d, err := mgr.Create(ctx, "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	// Simulate crash: remove the worktree directory but leave the branch
	os.RemoveAll(d.WorkDir)

	// Create a new manager which runs Cleanup on init
	mgr2, err := draft.NewManager(root, 10)
	if err != nil {
		t.Fatal(err)
	}
	_ = ctx

	list := mgr2.List()
	if len(list) != 0 {
		t.Fatalf("expected 0 drafts after cleanup, got %d", len(list))
	}
}

func TestMaxActive(t *testing.T) {
	root := initGitRepo(t)
	mgr, err := draft.NewManager(root, 2)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = mgr.Create(ctx, "a1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = mgr.Create(ctx, "a2")
	if err != nil {
		t.Fatal(err)
	}
	_, err = mgr.Create(ctx, "a3")
	if err != draft.ErrMaxActive {
		t.Fatalf("expected ErrMaxActive, got %v", err)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
