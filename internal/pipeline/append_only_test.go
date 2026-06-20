package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAppendOnly_RejectsOverwrite(t *testing.T) {
	p, _, _ := newTestPipeline(t)
	ctx := context.Background()
	path := "events/log.md"
	initial := []byte("---\ntitle: Log\nappend_only: true\n---\n# Log\nline 1\n")

	if _, err := p.Write(ctx, path, initial, "test"); err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err := p.Write(ctx, path, []byte("---\ntitle: Log\nappend_only: true\n---\n# Log\nreplaced\n"), "test")
	if !errors.Is(err, ErrAppendOnly) {
		t.Fatalf("overwrite: got %v, want ErrAppendOnly", err)
	}
}

func TestAppendOnly_AllowsAppend(t *testing.T) {
	p, store, _ := newTestPipeline(t)
	ctx := context.Background()
	path := "events/log.md"
	initial := []byte("---\ntitle: Log\nappend_only: true\n---\n# Log\nline 1\n")

	if _, err := p.Write(ctx, path, initial, "test"); err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := p.Append(ctx, path, "line 2", "\n", "test"); err != nil {
		t.Fatalf("append: %v", err)
	}

	got, err := store.Read(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "line 2") {
		t.Fatalf("content %q missing appended line", string(got))
	}
}

func TestAppendOnly_AllowsFirstWrite(t *testing.T) {
	p, _, _ := newTestPipeline(t)
	ctx := context.Background()
	path := "events/new.md"
	body := []byte("---\ntitle: New\nappend_only: true\n---\n# New\n")

	if _, err := p.Write(ctx, path, body, "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
}

func TestAppendOnly_IgnoresNormalFiles(t *testing.T) {
	p, _, _ := newTestPipeline(t)
	ctx := context.Background()
	path := "notes/a.md"

	if _, err := p.Write(ctx, path, []byte("# A\n"), "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := p.Write(ctx, path, []byte("# B\n"), "test"); err != nil {
		t.Fatalf("overwrite: %v", err)
	}
}

func TestAppendOnly_BulkWriteRejectsOverwrite(t *testing.T) {
	p, ctx := newAppendOnlyPipeline(t)
	initial := []byte("---\nappend_only: true\n---\nentry\n")
	if _, err := p.Write(ctx, "events/log.md", initial, "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := p.BulkWrite(ctx, []struct {
		Path    string
		Content []byte
	}{{Path: "events/log.md", Content: []byte("nope\n")}}, "test", "")
	if !errors.Is(err, ErrAppendOnly) {
		t.Fatalf("bulk overwrite: got %v, want ErrAppendOnly", err)
	}
}

func TestAppendOnly_WriteStreamRejectsOverwrite(t *testing.T) {
	p, ctx := newAppendOnlyPipeline(t)
	initial := []byte("---\nappend_only: true\n---\nentry\n")
	if _, err := p.Write(ctx, "events/log.md", initial, "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := p.WriteStream(ctx, "events/log.md", strings.NewReader("replaced\n"), 9, "test")
	if !errors.Is(err, ErrAppendOnly) {
		t.Fatalf("WriteStream overwrite: got %v, want ErrAppendOnly", err)
	}
}

func TestAppendOnly_StringTrueValue(t *testing.T) {
	p, _, _ := newTestPipeline(t)
	ctx := context.Background()
	path := "events/log.md"
	initial := []byte("---\nappend_only: \"true\"\n---\nentry\n")

	if _, err := p.Write(ctx, path, initial, "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := p.Write(ctx, path, []byte("replaced\n"), "test")
	if !errors.Is(err, ErrAppendOnly) {
		t.Fatalf("overwrite string true: got %v, want ErrAppendOnly", err)
	}
}

func TestAppendOnly_StringOneValue(t *testing.T) {
	p, _, _ := newTestPipeline(t)
	ctx := context.Background()
	path := "events/log.md"
	initial := []byte("---\nappend_only: \"1\"\n---\nentry\n")

	if _, err := p.Write(ctx, path, initial, "test"); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := p.Write(ctx, path, []byte("replaced\n"), "test")
	if !errors.Is(err, ErrAppendOnly) {
		t.Fatalf("overwrite string 1: got %v, want ErrAppendOnly", err)
	}
}

func TestAppendOnly_BulkWriteRejectsDuplicatePathOverwrite(t *testing.T) {
	p, ctx := newAppendOnlyPipeline(t)
	_, err := p.BulkWrite(ctx, []struct {
		Path    string
		Content []byte
	}{
		{Path: "events/log.md", Content: []byte("---\nappend_only: true\n---\nfirst\n")},
		{Path: "events/log.md", Content: []byte("replaced\n")},
	}, "test", "")
	if !errors.Is(err, ErrAppendOnly) {
		t.Fatalf("duplicate path bulk overwrite: got %v, want ErrAppendOnly", err)
	}
}

func newAppendOnlyPipeline(t *testing.T) (*Pipeline, context.Context) {
	t.Helper()
	p, _, _ := newTestPipeline(t)
	return p, context.Background()
}
