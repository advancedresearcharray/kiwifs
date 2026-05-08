package draft

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound    = errors.New("draft not found")
	ErrConflict    = errors.New("merge conflict: main has diverged — rebase or resolve manually")
	ErrMaxActive   = errors.New("maximum active drafts reached")
	ErrEmptyRepo   = errors.New("repository has no commits — write at least one file before creating a draft")
)

const gitTimeout = 30 * time.Second

type Draft struct {
	ID        string    `json:"id"`
	Branch    string    `json:"branch"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
	WorkDir   string    `json:"work_dir"`
}

type Manager struct {
	root      string
	draftsDir string
	maxActive int
	mu        sync.RWMutex
	active    map[string]*Draft
}

func NewManager(root string, maxActive int) (*Manager, error) {
	draftsDir := filepath.Join(root, ".kiwi", "drafts")
	if err := os.MkdirAll(draftsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create drafts dir: %w", err)
	}
	if maxActive <= 0 {
		maxActive = 10
	}
	m := &Manager{
		root:      root,
		draftsDir: draftsDir,
		maxActive: maxActive,
		active:    make(map[string]*Draft),
	}
	if err := m.loadManifest(); err != nil {
		log.Printf("draft: load manifest: %v (starting fresh)", err)
	}
	m.Cleanup()
	return m, nil
}

func (m *Manager) Create(ctx context.Context, actor string) (*Draft, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.active) >= m.maxActive {
		return nil, ErrMaxActive
	}

	if !m.hasCommits(ctx) {
		return nil, ErrEmptyRepo
	}

	id := shortID()
	branch := "draft/" + id
	workDir := filepath.Join(m.draftsDir, id)

	if err := m.git(ctx, "branch", branch, "HEAD"); err != nil {
		return nil, fmt.Errorf("create branch: %w", err)
	}

	if err := m.git(ctx, "worktree", "add", workDir, branch); err != nil {
		_ = m.git(ctx, "branch", "-D", branch)
		return nil, fmt.Errorf("create worktree: %w", err)
	}

	d := &Draft{
		ID:        id,
		Branch:    branch,
		Actor:     actor,
		CreatedAt: time.Now().UTC(),
		WorkDir:   workDir,
	}
	m.active[id] = d
	m.saveManifestLocked()
	return d, nil
}

func (m *Manager) Get(id string) (*Draft, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.active[id]
	if !ok {
		return nil, ErrNotFound
	}
	return d, nil
}

func (m *Manager) List() []*Draft {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Draft, 0, len(m.active))
	for _, d := range m.active {
		out = append(out, d)
	}
	return out
}

func (m *Manager) Diff(ctx context.Context, id string) (string, error) {
	m.mu.RLock()
	d, ok := m.active[id]
	m.mu.RUnlock()
	if !ok {
		return "", ErrNotFound
	}
	base := m.defaultBranch(ctx)
	return m.gitOutput(ctx, "diff", base+"..."+d.Branch)
}

func (m *Manager) Merge(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, ok := m.active[id]
	if !ok {
		return ErrNotFound
	}

	// Ensure any uncommitted changes in the draft worktree are committed
	// before attempting the merge. External tools (imports, direct writes)
	// may leave unstaged/uncommitted files in the worktree.
	if err := m.commitWorktreeChanges(ctx, d); err != nil {
		log.Printf("draft: auto-commit in %s: %v", d.ID, err)
	}

	if err := m.git(ctx, "merge", "--ff-only", d.Branch); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "Not possible to fast-forward") ||
			strings.Contains(errMsg, "not something we can merge") ||
			strings.Contains(errMsg, "fatal: refusing to merge unrelated histories") {
			return ErrConflict
		}
		return fmt.Errorf("merge: %w", err)
	}

	m.removeDraftLocked(ctx, d)
	return nil
}

func (m *Manager) Discard(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, ok := m.active[id]
	if !ok {
		return ErrNotFound
	}
	m.removeDraftLocked(ctx, d)
	return nil
}

func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for id, d := range m.active {
		if _, err := os.Stat(d.WorkDir); os.IsNotExist(err) {
			log.Printf("draft: cleaning orphaned draft %s (worktree missing)", id)
			_ = m.git(ctx, "branch", "-D", d.Branch)
			delete(m.active, id)
		}
	}

	_ = m.git(ctx, "worktree", "prune")
	m.saveManifestLocked()
}

func (m *Manager) removeDraftLocked(ctx context.Context, d *Draft) {
	if err := m.git(ctx, "worktree", "remove", "--force", d.WorkDir); err != nil {
		log.Printf("draft: remove worktree %s: %v", d.ID, err)
		os.RemoveAll(d.WorkDir)
	}
	if err := m.git(ctx, "branch", "-D", d.Branch); err != nil {
		log.Printf("draft: delete branch %s: %v", d.Branch, err)
	}
	delete(m.active, d.ID)
	m.saveManifestLocked()
}

// Root returns the main repo root.
func (m *Manager) Root() string { return m.root }

func (m *Manager) git(ctx context.Context, args ...string) error {
	_, err := m.gitOutputErr(ctx, args...)
	return err
}

func (m *Manager) gitOutput(ctx context.Context, args ...string) (string, error) {
	return m.gitOutputErr(ctx, args...)
}

func (m *Manager) gitOutputErr(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = m.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

type manifestEntry struct {
	ID        string    `json:"id"`
	Branch    string    `json:"branch"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *Manager) manifestPath() string {
	return filepath.Join(m.draftsDir, "manifest.json")
}

func (m *Manager) loadManifest() error {
	data, err := os.ReadFile(m.manifestPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var entries []manifestEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	for _, e := range entries {
		m.active[e.ID] = &Draft{
			ID:        e.ID,
			Branch:    e.Branch,
			Actor:     e.Actor,
			CreatedAt: e.CreatedAt,
			WorkDir:   filepath.Join(m.draftsDir, e.ID),
		}
	}
	return nil
}

func (m *Manager) saveManifestLocked() {
	entries := make([]manifestEntry, 0, len(m.active))
	for _, d := range m.active {
		entries = append(entries, manifestEntry{
			ID:        d.ID,
			Branch:    d.Branch,
			Actor:     d.Actor,
			CreatedAt: d.CreatedAt,
		})
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		log.Printf("draft: marshal manifest: %v", err)
		return
	}
	if err := os.WriteFile(m.manifestPath(), data, 0o644); err != nil {
		log.Printf("draft: write manifest: %v", err)
	}
}

// hasCommits returns true if the repository has at least one commit.
// An empty repo (freshly git-init'd, no commits) has an invalid HEAD;
// branching from HEAD in that state fails. We detect this up front and
// return a clear error rather than leaking a git message to the user.
func (m *Manager) hasCommits(ctx context.Context) bool {
	_, err := m.gitOutputErr(ctx, "rev-parse", "--verify", "HEAD")
	return err == nil
}

// defaultBranch returns the name of the HEAD branch (e.g. "main", "master").
// Falls back to "main" if detection fails.
func (m *Manager) defaultBranch(ctx context.Context) string {
	out, err := m.gitOutputErr(ctx, "symbolic-ref", "--short", "HEAD")
	if err != nil || out == "" {
		return "main"
	}
	return out
}

// commitWorktreeChanges stages and commits any uncommitted changes in the
// draft worktree. This handles the case where external tools (like the
// import command) or the pipeline's async commit haven't flushed before
// merge is called.
func (m *Manager) commitWorktreeChanges(ctx context.Context, d *Draft) error {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	// Check for uncommitted changes in the worktree
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = d.WorkDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return nil
	}

	// Stage all changes
	cmd = exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = d.WorkDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("add: %s: %w", strings.TrimSpace(string(out)), err)
	}

	// Commit
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", fmt.Sprintf("draft/%s: auto-commit before merge", d.ID))
	cmd.Dir = d.WorkDir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+d.Actor,
		"GIT_AUTHOR_EMAIL=kiwifs@internal",
		"GIT_COMMITTER_NAME="+d.Actor,
		"GIT_COMMITTER_EMAIL=kiwifs@internal",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("commit: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func shortID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
