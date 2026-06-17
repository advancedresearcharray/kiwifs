package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/storage"
)

// ErrAppendOnlyDenied is returned when a full overwrite (PUT) is attempted on a
// file whose existing frontmatter has append_only: true. Mapped to HTTP 409.
var ErrAppendOnlyDenied = fmt.Errorf("append-only file: use POST /api/kiwi/file/append")

// IsAppendOnly reports whether content carries append_only: true in YAML frontmatter.
func IsAppendOnly(content []byte) bool {
	fm, err := markdown.Frontmatter(content)
	if err != nil || len(fm) == 0 {
		return false
	}
	return frontmatterBool(fm["append_only"])
}

func frontmatterBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case int:
		return val != 0
	case string:
		s := strings.ToLower(strings.TrimSpace(val))
		return s == "true" || s == "1" || s == "yes"
	default:
		return false
	}
}

// rejectAppendOnlyOverwrite returns ErrAppendOnlyDenied when path already exists
// and its current content is marked append-only. Creating a new file is allowed
// even when the incoming payload sets append_only: true.
func rejectAppendOnlyOverwrite(ctx context.Context, store storage.Storage, path string) error {
	existing, err := store.Read(ctx, path)
	if err != nil {
		return nil
	}
	if IsAppendOnly(existing) {
		return ErrAppendOnlyDenied
	}
	return nil
}
