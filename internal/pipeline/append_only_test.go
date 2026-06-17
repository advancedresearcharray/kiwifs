package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/events"
	"github.com/kiwifs/kiwifs/internal/search"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/kiwifs/kiwifs/internal/versioning"
)

func newAppendOnlyPipeline(t *testing.T) (*Pipeline, context.Context) {
	t.Helper()
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	p := New(store, versioning.NewNoop(), search.NewGrep(dir), nil, events.NewHub(), nil, dir)
	return p, context.Background()
}

func TestIsAppendOnly(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"no frontmatter", "# body\n", false},
		{"append_only true", "---\nappend_only: true\n---\n# log\n", true},
		{"append_only false", "---\nappend_only: false\n---\n# log\n", false},
		{"append_only string", "---\nappend_only: \"true\"\n---\n# log\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAppendOnly([]byte(tt.content)); got != tt.want {
				t.Fatalf("IsAppendOnly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrite_RejectsAppendOnlyOverwrite(t *testing.T) {
	p, ctx := newAppendOnlyPipeline(t)
	initial := "---\nappend_only: true\n---\nentry one\n"
	if _, err := p.Write(ctx, "events/log.md", []byte(initial), "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := p.Write(ctx, "events/log.md", []byte("replaced\n"), "test")
	if !errors.Is(err, ErrAppendOnlyDenied) {
		t.Fatalf("overwrite: got %v, want ErrAppendOnlyDenied", err)
	}
}

func TestWrite_AllowsCreateWithAppendOnly(t *testing.T) {
	p, ctx := newAppendOnlyPipeline(t)
	content := "---\nappend_only: true\n---\nfirst\n"
	if _, err := p.Write(ctx, "events/new-log.md", []byte(content), "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
}

func TestWrite_UnaffectedWithoutAppendOnly(t *testing.T) {
	p, ctx := newAppendOnlyPipeline(t)
	if _, err := p.Write(ctx, "note.md", []byte("v1\n"), "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := p.Write(ctx, "note.md", []byte("v2\n"), "test"); err != nil {
		t.Fatalf("overwrite: %v", err)
	}
}

func TestAppend_AllowsAppendOnlyFile(t *testing.T) {
	p, ctx := newAppendOnlyPipeline(t)
	initial := "---\nappend_only: true\n---\nentry one\n"
	if _, err := p.Write(ctx, "events/log.md", []byte(initial), "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := p.Append(ctx, "events/log.md", "entry two\n", "\n", "test"); err != nil {
		t.Fatalf("append: %v", err)
	}
	got, err := p.Store.Read(ctx, "events/log.md")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(got), "entry one") || !strings.Contains(string(got), "entry two") {
		t.Fatalf("content = %q", string(got))
	}
}

func TestBulkWrite_RejectsAppendOnlyOverwrite(t *testing.T) {
	p, ctx := newAppendOnlyPipeline(t)
	initial := "---\nappend_only: true\n---\nentry\n"
	if _, err := p.Write(ctx, "events/log.md", []byte(initial), "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := p.BulkWrite(ctx, []struct {
		Path    string
		Content []byte
	}{{Path: "events/log.md", Content: []byte("nope\n")}}, "test", "")
	if !errors.Is(err, ErrAppendOnlyDenied) {
		t.Fatalf("bulk overwrite: got %v, want ErrAppendOnlyDenied", err)
	}
}
