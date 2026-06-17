package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/kiwifs/kiwifs/internal/search"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/kiwifs/kiwifs/internal/versioning"
)

func TestSequenceSequentialAppends(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	seqStore, err := NewSequenceStore(dir, []string{"events/"})
	if err != nil {
		t.Fatalf("NewSequenceStore: %v", err)
	}
	p := New(store, versioning.NewNoop(), search.NewGrep(dir), nil, nil, nil, dir)
	p.Sequences = seqStore

	ctx := context.Background()
	for i := 1; i <= 3; i++ {
		_, err := p.Append(ctx, "events/log.md", fmt.Sprintf("entry %d", i), "\n", "tester")
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	data, err := store.Read(ctx, "events/log.md")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	markers := seqMarkerRe.FindAllString(string(data), -1)
	if len(markers) != 3 {
		t.Fatalf("want 3 markers, got %d: %v", len(markers), markers)
	}
	for i, want := range []string{"<!-- seq:1 -->", "<!-- seq:2 -->", "<!-- seq:3 -->"} {
		if markers[i] != want {
			t.Fatalf("marker[%d]=%q, want %q", i, markers[i], want)
		}
	}
	if seqStore.Counter() != 3 {
		t.Fatalf("counter=%d, want 3", seqStore.Counter())
	}
}

func TestSequenceNotAppliedOutsideDirs(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	seqStore, err := NewSequenceStore(dir, []string{"events/"})
	if err != nil {
		t.Fatalf("NewSequenceStore: %v", err)
	}
	p := New(store, versioning.NewNoop(), search.NewGrep(dir), nil, nil, nil, dir)
	p.Sequences = seqStore

	ctx := context.Background()
	_, err = p.Append(ctx, "notes/other.md", "plain entry", "\n", "tester")
	if err != nil {
		t.Fatalf("append: %v", err)
	}

	data, err := store.Read(ctx, "notes/other.md")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if seqMarkerRe.Match(data) {
		t.Fatalf("unexpected sequence marker outside configured dir: %s", data)
	}
	if seqStore.Counter() != 0 {
		t.Fatalf("counter=%d, want 0", seqStore.Counter())
	}
}

func TestSequenceConcurrentAppends(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	seqStore, err := NewSequenceStore(dir, []string{"audit/"})
	if err != nil {
		t.Fatalf("NewSequenceStore: %v", err)
	}
	p := New(store, versioning.NewNoop(), search.NewGrep(dir), nil, nil, nil, dir)
	p.Sequences = seqStore

	const n = 20
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := p.Append(context.Background(), "audit/trail.md",
				fmt.Sprintf("event %d", i), "\n", "worker")
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("append: %v", err)
		}
	}

	data, err := store.Read(context.Background(), "audit/trail.md")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	nums := extractSeqNumbers(string(data))
	if len(nums) != n {
		t.Fatalf("want %d markers, got %d", n, len(nums))
	}
	seen := make(map[int]bool, n)
	for _, num := range nums {
		if seen[num] {
			t.Fatalf("duplicate sequence %d", num)
		}
		seen[num] = true
	}
	for i := 1; i <= n; i++ {
		if !seen[i] {
			t.Fatalf("missing sequence %d", i)
		}
	}
}

func TestCheckSequencesGapDetection(t *testing.T) {
	dir := t.TempDir()
	eventsDir := filepath.Join(dir, "events")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := "<!-- seq:1 -->\nfirst\n<!-- seq:3 -->\nthird\n"
	if err := os.WriteFile(filepath.Join(eventsDir, "log.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	stateDir := filepath.Join(dir, ".kiwi", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "sequences.json"),
		[]byte("{\n  \"counter\": 3\n}\n"), 0o644); err != nil {
		t.Fatalf("write sequences.json: %v", err)
	}

	result, err := CheckSequences(dir, []string{"events/"})
	if err != nil {
		t.Fatalf("CheckSequences: %v", err)
	}
	if !result.HasIssues() {
		t.Fatal("expected gap issue, got clean result")
	}
	foundGap := false
	for _, iss := range result.Issues {
		if iss.Kind == "gap" && iss.Missing == 2 {
			foundGap = true
		}
	}
	if !foundGap {
		t.Fatalf("expected gap at seq 2, issues: %+v", result.Issues)
	}
}

func TestCheckSequencesClean(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	seqStore, err := NewSequenceStore(dir, []string{"events/"})
	if err != nil {
		t.Fatalf("NewSequenceStore: %v", err)
	}
	p := New(store, versioning.NewNoop(), search.NewGrep(dir), nil, nil, nil, dir)
	p.Sequences = seqStore

	ctx := context.Background()
	for i := 1; i <= 2; i++ {
		if _, err := p.Append(ctx, "events/log.md", fmt.Sprintf("e%d", i), "\n", "t"); err != nil {
			t.Fatalf("append: %v", err)
		}
	}

	result, err := CheckSequences(dir, []string{"events/"})
	if err != nil {
		t.Fatalf("CheckSequences: %v", err)
	}
	if result.HasIssues() {
		t.Fatalf("expected clean check, got: %+v", result.Issues)
	}
}

func extractSeqNumbers(content string) []int {
	re := regexp.MustCompile(`<!-- seq:(\d+) -->`)
	var out []int
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		var n int
		fmt.Sscanf(m[1], "%d", &n)
		out = append(out, n)
	}
	return out
}

func TestNormalizeSequenceDirs(t *testing.T) {
	got := normalizeSequenceDirs([]string{"events", "audit/", " events/ "})
	want := []string{"events/", "audit/"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestSequenceAppliesTo(t *testing.T) {
	s := &SequenceStore{directories: normalizeSequenceDirs([]string{"events/", "audit/"})}
	cases := map[string]bool{
		"events/log.md":  true,
		"audit/trail.md": true,
		"notes/x.md":     false,
		"events":         true,
	}
	for path, want := range cases {
		if got := s.AppliesTo(path); got != want {
			t.Fatalf("AppliesTo(%q)=%v, want %v", path, got, want)
		}
	}
}

func TestSequenceStorePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	s1, err := NewSequenceStore(dir, []string{"events/"})
	if err != nil {
		t.Fatalf("NewSequenceStore: %v", err)
	}
	seq, err := s1.Next()
	if err != nil || seq != 1 {
		t.Fatalf("Next()=%d err=%v, want 1", seq, err)
	}

	s2, err := NewSequenceStore(dir, []string{"events/"})
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	seq, err = s2.Next()
	if err != nil || seq != 2 {
		t.Fatalf("Next() after reload=%d err=%v, want 2", seq, err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".kiwi", "state", "sequences.json"))
	if err != nil {
		t.Fatalf("read sequences.json: %v", err)
	}
	if !strings.Contains(string(raw), `"counter": 2`) {
		t.Fatalf("sequences.json=%q", raw)
	}
}
