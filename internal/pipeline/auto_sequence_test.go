package pipeline

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/search"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/kiwifs/kiwifs/internal/versioning"
)

type stubMetaMax struct {
	max int
	mu  sync.Mutex
}

func (s *stubMetaMax) MaxFrontmatterIntInDirectory(_ context.Context, _, _ string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.max, nil
}

func (s *stubMetaMax) bumpAssigned() {
	s.mu.Lock()
	s.max++
	s.mu.Unlock()
}

func TestAutoSequencerAssignsNextNumber(t *testing.T) {
	meta := &stubMetaMax{max: 3}
	seq := NewAutoSequencer(config.AutoSequenceConfig{
		Directory: "decisions/",
		Field:     "adr_number",
	}, meta)

	content := []byte("---\ntitle: New ADR\n---\n# Context\n")
	got := seq.FormatWrite("decisions/new-adr.md", content)
	fm, err := markdown.Frontmatter(got)
	if err != nil {
		t.Fatalf("frontmatter: %v", err)
	}
	if fm["adr_number"] != 4 {
		t.Fatalf("adr_number = %v, want 4", fm["adr_number"])
	}
}

func TestAutoSequencerSkipsExistingNumber(t *testing.T) {
	meta := &stubMetaMax{max: 10}
	seq := NewAutoSequencer(config.AutoSequenceConfig{
		Directory: "decisions/",
		Field:     "adr_number",
	}, meta)

	input := []byte("---\nadr_number: 7\ntitle: Existing\n---\n# Context\n")
	got := seq.FormatWrite("decisions/existing.md", input)
	if string(got) != string(input) {
		t.Fatalf("content changed:\n%s", got)
	}
}

func TestAutoSequencerSkipsOtherDirectories(t *testing.T) {
	meta := &stubMetaMax{max: 5}
	seq := NewAutoSequencer(config.AutoSequenceConfig{
		Directory: "decisions/",
		Field:     "adr_number",
	}, meta)

	input := []byte("---\ntitle: Note\n---\n# Hello\n")
	got := seq.FormatWrite("notes/hello.md", input)
	if string(got) != string(input) {
		t.Fatalf("unexpected change outside directory:\n%s", got)
	}
}

func TestAutoSequencerConcurrentWritesUnique(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	sqliteSearcher, err := search.NewSQLite(dir, store)
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	defer sqliteSearcher.Close()

	ctx := context.Background()
	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf("---\nadr_number: %d\ntitle: ADR %d\n---\n# ADR %d\n", i, i, i)
		if err := sqliteSearcher.IndexMeta(ctx, fmt.Sprintf("decisions/adr-%d.md", i), []byte(body)); err != nil {
			t.Fatalf("IndexMeta(%d): %v", i, err)
		}
	}

	seq := NewAutoSequencer(config.AutoSequenceConfig{
		Directory: "decisions/",
		Field:     "adr_number",
	}, sqliteSearcher)

	const n = 8
	var wg sync.WaitGroup
	numbers := make([]int, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := fmt.Sprintf("decisions/concurrent-%d.md", idx)
			content := []byte("---\ntitle: Concurrent\n---\n# Body\n")
			out := seq.FormatWrite(path, content)
			fm, err := markdown.Frontmatter(out)
			if err != nil {
				errs[idx] = err
				return
			}
			switch v := fm["adr_number"].(type) {
			case int:
				numbers[idx] = v
			default:
				errs[idx] = fmt.Errorf("unexpected adr_number type %T", fm["adr_number"])
			}
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}
	seen := make(map[int]struct{}, n)
	for _, num := range numbers {
		if num < 4 || num > 11 {
			t.Fatalf("number out of expected range 4..11: %d", num)
		}
		if _, dup := seen[num]; dup {
			t.Fatalf("duplicate adr_number assigned: %d", num)
		}
		seen[num] = struct{}{}
	}
}

func TestAutoSequencerBulkWriteSequential(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	sqliteSearcher, err := search.NewSQLite(dir, store)
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	defer sqliteSearcher.Close()

	ctx := context.Background()
	if err := sqliteSearcher.IndexMeta(ctx, "decisions/first.md", []byte("---\nadr_number: 2\n---\n# One\n")); err != nil {
		t.Fatalf("IndexMeta: %v", err)
	}

	seq := NewAutoSequencer(config.AutoSequenceConfig{
		Directory: "decisions/",
		Field:     "adr_number",
	}, sqliteSearcher)

	files := []struct {
		Path    string
		Content []byte
	}{
		{"decisions/a.md", []byte("---\ntitle: A\n---\n# A\n")},
		{"decisions/b.md", []byte("---\ntitle: B\n---\n# B\n")},
	}
	for i := range files {
		files[i].Content = seq.FormatWrite(files[i].Path, files[i].Content)
	}

	nums := make([]int, len(files))
	for i, f := range files {
		fm, err := markdown.Frontmatter(f.Content)
		if err != nil {
			t.Fatalf("frontmatter %s: %v", f.Path, err)
		}
		switch v := fm["adr_number"].(type) {
		case int:
			nums[i] = v
		default:
			t.Fatalf("unexpected type for %s: %T", f.Path, fm["adr_number"])
		}
	}
	if nums[0] != 3 || nums[1] != 4 {
		t.Fatalf("assigned numbers = %v, want [3 4]", nums)
	}
}

func TestPipelineAutoSequenceIntegration(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	sqliteSearcher, err := search.NewSQLite(dir, store)
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	defer sqliteSearcher.Close()

	seq := NewAutoSequencer(config.AutoSequenceConfig{
		Directory: "decisions/",
		Field:     "adr_number",
	}, sqliteSearcher)

	p := New(store, versioning.NewNoop(), sqliteSearcher, sqliteSearcher, nil, nil, "")
	p.FormatWrite = seq.FormatWrite

	ctx := context.Background()
	if _, err := p.Write(ctx, "decisions/seed.md", []byte("---\nadr_number: 5\n---\n# Seed\n"), "tester"); err != nil {
		t.Fatalf("seed write: %v", err)
	}
	res, err := p.Write(ctx, "decisions/next.md", []byte("---\ntitle: Next\n---\n# Next\n"), "tester")
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if res.Path != "decisions/next.md" {
		t.Fatalf("unexpected path: %s", res.Path)
	}
	onDisk, err := store.Read(ctx, "decisions/next.md")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	fm, err := markdown.Frontmatter(onDisk)
	if err != nil {
		t.Fatalf("frontmatter: %v", err)
	}
	if fm["adr_number"] != 6 {
		t.Fatalf("adr_number = %v, want 6", fm["adr_number"])
	}
}

func TestPathInDirectory(t *testing.T) {
	prefix := "decisions/"
	cases := map[string]bool{
		"decisions/a.md":       true,
		"decisions/nested/x.md": true,
		"notes/decisions/x.md": false,
		"decisions-backup/x.md": false,
	}
	for path, want := range cases {
		if got := pathInDirectory(path, prefix); got != want {
			t.Errorf("pathInDirectory(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestChainFormatWriteOrder(t *testing.T) {
	var log strings.Builder
	h1 := func(path string, content []byte) []byte {
		log.WriteString("1")
		return content
	}
	h2 := func(path string, content []byte) []byte {
		log.WriteString("2")
		return content
	}
	chain := ChainFormatWrite(h1, h2)
	chain("x.md", []byte("body"))
	if log.String() != "12" {
		t.Fatalf("hook order = %q, want 12", log.String())
	}
}
