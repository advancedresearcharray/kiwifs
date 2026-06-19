package pipeline

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/kiwifs/kiwifs/internal/markdown"
)

// MetaMaxQuerier is implemented by sqlite search to read max frontmatter
// field values from file_meta for a directory prefix.
type MetaMaxQuerier interface {
	MaxFrontmatterIntInDirectory(ctx context.Context, pathPrefix, field string) (int, error)
}

// AutoSequencer assigns the next sequence number to markdown files written
// under a configured directory when the target frontmatter field is absent.
type AutoSequencer struct {
	cfg  config.AutoSequenceConfig
	meta MetaMaxQuerier

	mu   sync.Mutex
	next map[string]int // normalized directory prefix → next number to assign
}

// NewAutoSequencer builds a FormatWrite hook from config and a meta querier.
func NewAutoSequencer(cfg config.AutoSequenceConfig, meta MetaMaxQuerier) *AutoSequencer {
	return &AutoSequencer{
		cfg:  cfg,
		meta: meta,
		next: make(map[string]int),
	}
}

// FormatWrite injects the next sequence number when path is under the configured
// directory and the field is missing or zero.
func (s *AutoSequencer) FormatWrite(path string, content []byte) []byte {
	if s == nil || s.meta == nil || s.cfg.Directory == "" || s.cfg.Field == "" {
		return content
	}
	if !strings.HasSuffix(strings.ToLower(path), ".md") {
		return content
	}
	dirPrefix := normalizeDirPrefix(s.cfg.Directory)
	if !pathInDirectory(path, dirPrefix) {
		return content
	}
	fm, err := markdown.Frontmatter(content)
	if err == nil && frontmatterFieldSet(fm, s.cfg.Field) {
		return content
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	next, ok := s.next[dirPrefix]
	if !ok {
		max, err := s.meta.MaxFrontmatterIntInDirectory(context.Background(), dirPrefix, s.cfg.Field)
		if err != nil {
			return content
		}
		next = max + 1
	}
	s.next[dirPrefix] = next + 1

	updated, err := markdown.SetFrontmatterField(content, s.cfg.Field, next)
	if err != nil {
		return content
	}
	return updated
}

// ChainFormatWrite runs multiple FormatWrite hooks in order.
func ChainFormatWrite(hooks ...func(path string, content []byte) []byte) func(path string, content []byte) []byte {
	if len(hooks) == 0 {
		return nil
	}
	if len(hooks) == 1 {
		return hooks[0]
	}
	return func(path string, content []byte) []byte {
		for _, hook := range hooks {
			if hook != nil {
				content = hook(path, content)
			}
		}
		return content
	}
}

func normalizeDirPrefix(dir string) string {
	dir = filepath.ToSlash(strings.TrimSpace(dir))
	dir = strings.TrimPrefix(dir, "/")
	if dir != "" && !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	return dir
}

func pathInDirectory(path, dirPrefix string) bool {
	if dirPrefix == "" {
		return false
	}
	path = filepath.ToSlash(strings.TrimPrefix(path, "/"))
	return path == strings.TrimSuffix(dirPrefix, "/") || strings.HasPrefix(path, dirPrefix)
}

func frontmatterFieldSet(fm map[string]any, field string) bool {
	if len(fm) == 0 {
		return false
	}
	val, ok := fm[field]
	if !ok || val == nil {
		return false
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case bool:
		return true
	default:
		return true
	}
}
