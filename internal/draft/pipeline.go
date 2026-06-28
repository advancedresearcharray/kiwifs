package draft

import (
	"fmt"
	"log"

	"github.com/kiwifs/kiwifs/internal/pipeline"
	"github.com/kiwifs/kiwifs/internal/search"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/kiwifs/kiwifs/internal/versioning"
)

// Pipeline returns a lightweight Pipeline scoped to the draft's worktree.
// Writes go to the worktree directory and commits land on the draft branch.
// Search indexing and SSE broadcasting are disabled — drafts are short-lived
// write-review-merge workflows, not long-running indexed spaces.
func (m *Manager) Pipeline(id string) (*pipeline.Pipeline, error) {
	d, err := m.Get(id)
	if err != nil {
		return nil, err
	}

	store, err := storage.NewLocal(d.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("draft store: %w", err)
	}

	ver, err := versioning.NewGit(d.WorkDir)
	if err != nil {
		log.Printf("draft: git versioning unavailable for %s (%v) — using noop", id, err)
		return pipeline.New(store, versioning.NewNoop(), search.NewGrep(d.WorkDir), nil, nil, nil, d.WorkDir), nil
	}

	searcher := search.NewGrep(d.WorkDir)
	return pipeline.New(store, ver, searcher, nil, nil, nil, d.WorkDir), nil
}
