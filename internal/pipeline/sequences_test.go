package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendSequenceNumbers(t *testing.T) {
	p, store, dir := newTestPipeline(t)
	p.SequenceDirs = []string{"events/"}
	ctx := context.Background()
	path := "events/log.md"

	if _, err := p.Append(ctx, path, "first entry", "\n", "test"); err != nil {
		t.Fatalf("append 1: %v", err)
	}
	if _, err := p.Append(ctx, path, "second entry", "\n", "test"); err != nil {
		t.Fatalf("append 2: %v", err)
	}

	body, err := store.Read(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "<!-- seq:1 -->") || !strings.Contains(string(body), "<!-- seq:2 -->") {
		t.Fatalf("missing seq markers: %q", string(body))
	}

	issues, err := CheckSequenceGaps(dir, []string{"events/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no gaps, got %v", issues)
	}
}

func TestAppendSequenceNumbersLeadingSlashPath(t *testing.T) {
	p, store, _ := newTestPipeline(t)
	p.SequenceDirs = []string{"events/"}
	ctx := context.Background()
	path := "/events/log.md"

	if _, err := p.Append(ctx, path, "entry", "\n", "test"); err != nil {
		t.Fatalf("append: %v", err)
	}
	body, err := store.Read(ctx, "events/log.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "<!-- seq:1 -->") {
		t.Fatalf("missing seq marker for API-style path: %q", string(body))
	}
}

func TestAppendSequenceNumbersConcurrent(t *testing.T) {
	p, _, _ := newTestPipeline(t)
	p.SequenceDirs = []string{"events/"}
	ctx := context.Background()
	path := "events/concurrent.md"

	done := make(chan error, 4)
	for i := 0; i < 4; i++ {
		go func(n int) {
			_, err := p.Append(ctx, path, fmtLine(n), "\n", "test")
			done <- err
		}(i)
	}
	for i := 0; i < 4; i++ {
		if err := <-done; err != nil {
			t.Fatalf("append: %v", err)
		}
	}
}

func fmtLine(n int) string {
	return fmt.Sprintf("line-%d", n)
}

func TestCheckSequenceGapsDetectsMissing(t *testing.T) {
	dir := t.TempDir()
	eventsDir := filepath.Join(dir, "events")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(dir, ".kiwi", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "sequences.json"), []byte(`{"counters":{"events":3}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eventsDir, "log.md"), []byte("<!-- seq:1 -->\n<!-- seq:3 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	issues, err := CheckSequenceGaps(dir, []string{"events/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || !strings.Contains(issues[0], "seq:2") {
		t.Fatalf("issues: %v", issues)
	}
}
