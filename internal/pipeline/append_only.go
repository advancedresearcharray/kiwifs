package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/kiwifs/kiwifs/internal/markdown"
)

// ErrAppendOnly is returned when a PUT overwrite is attempted on a file whose
// frontmatter has append_only: true.
var ErrAppendOnly = fmt.Errorf("file is append-only; use append")

func isAppendOnly(content []byte) bool {
	fm, err := markdown.Frontmatter(content)
	if err != nil || fm == nil {
		return false
	}
	v, ok := fm["append_only"]
	if !ok {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(b, "true") || b == "1"
	}
	return false
}

func (p *Pipeline) rejectAppendOnlyOverwrite(ctx context.Context, path string) error {
	existing, err := p.Store.Read(ctx, path)
	if err != nil {
		return nil
	}
	if isAppendOnly(existing) {
		return fmt.Errorf("%w: PUT not allowed on %q", ErrAppendOnly, path)
	}
	return nil
}

// rejectAppendOnlyBulkOverwrite checks append_only for a bulk batch under
// writeMu. It rejects overwrites of on-disk append-only files and duplicate
// paths where an earlier batch entry is append-only.
func (p *Pipeline) rejectAppendOnlyBulkOverwrite(ctx context.Context, files []struct {
	Path    string
	Content []byte
}) error {
	seen := make(map[string][]byte, len(files))
	for i, f := range files {
		if err := p.rejectAppendOnlyOverwrite(ctx, f.Path); err != nil {
			return fmt.Errorf("files[%d] (%s): %w", i, f.Path, err)
		}
		if prev, ok := seen[f.Path]; ok && isAppendOnly(prev) {
			return fmt.Errorf("files[%d] (%s): %w: PUT not allowed on %q", i, f.Path, ErrAppendOnly, f.Path)
		}
		seen[f.Path] = f.Content
	}
	return nil
}
