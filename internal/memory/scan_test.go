package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kiwifs/kiwifs/internal/storage"
)

func TestScan_unmerged(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Episodic file
	ep := filepath.Join(root, "episodes", "a.md")
	if err := os.MkdirAll(filepath.Dir(ep), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ep, []byte(`---
memory_kind: episodic
episode_id: run-1
---
# run
`), 0644); err != nil {
		t.Fatal(err)
	}
	// Semantic with merged-from
	sem := filepath.Join(root, "concepts", "x.md")
	if err := os.MkdirAll(filepath.Dir(sem), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sem, []byte(`---
memory_kind: semantic
title: x
merged-from:
  - type: episode
    id: run-1
---
# x
`), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := storage.NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Scan(context.Background(), s, Options{EpisodesPathPrefix: "episodes/"})
	if err != nil {
		t.Fatal(err)
	}
	if rep.EpisodicCount != 1 {
		t.Fatalf("episodic: %d", rep.EpisodicCount)
	}
	if len(rep.Unmerged) != 0 {
		t.Fatalf("unmerged: %+v", rep.Unmerged)
	}
	if rep.CoveragePct != 100 {
		t.Fatalf("coverage_pct: got %v want 100", rep.CoveragePct)
	}
}

func TestScan_pathOnlyMerge(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ep := filepath.Join(root, "episodes", "b.md")
	if err := os.MkdirAll(filepath.Dir(ep), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ep, []byte(`---
memory_kind: episodic
---
# no id
`), 0644); err != nil {
		t.Fatal(err)
	}
	sem := filepath.Join(root, "c.md")
	if err := os.WriteFile(sem, []byte(`---
memory_kind: semantic
merged-from:
  - type: episode
    path: episodes/b.md
---
# c
`), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := storage.NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Scan(context.Background(), s, Options{EpisodesPathPrefix: "episodes/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Unmerged) != 0 {
		t.Fatalf("expected path merge, unmerged: %+v", rep.Unmerged)
	}
}

func TestScan_healthMetrics(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	write := func(rel, body string, modTime time.Time) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatal(err)
		}
	}

	now := time.Now()
	write("episodes/merged.md", `---
memory_kind: episodic
episode_id: ep-merged
---
# merged
`, now.Add(-48*time.Hour))
	write("episodes/open.md", `---
memory_kind: episodic
episode_id: ep-open
---
# open
`, now.Add(-24*time.Hour))
	write("pages/active-a.md", `---
memory_status: active
scope: user:alice
---
# active a
`, now.Add(-72*time.Hour))
	write("pages/active-b.md", `---
scope: team:core
---
# active b (default status)
`, now.Add(-24*time.Hour))
	write("pages/contested.md", `---
memory_status: contested
scope: user:bob
---
# contested
`, now)
	write("pages/expired.md", `---
memory_status: active
expires_at: 2020-01-01
scope: user:alice
---
# expired
`, now)
	write("pages/superseded.md", `---
memory_status: superseded
---
# superseded
`, now.Add(-96*time.Hour))
	write("pages/future.md", `---
memory_status: active
expires_at: 2099-01-01
---
# future expiry
`, now)
	write("concepts/summary.md", `---
memory_kind: semantic
merged-from:
  - type: episode
    id: ep-merged
---
# summary
`, now)

	s, err := storage.NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Scan(context.Background(), s, Options{EpisodesPathPrefix: "episodes/"})
	if err != nil {
		t.Fatal(err)
	}

	if rep.TotalEpisodic != 2 {
		t.Fatalf("total_episodic = %d, want 2", rep.TotalEpisodic)
	}
	if rep.TotalUnmerged != 1 {
		t.Fatalf("total_unmerged = %d, want 1", rep.TotalUnmerged)
	}
	if rep.CoveragePct != 50 {
		t.Fatalf("coverage_pct = %v, want 50", rep.CoveragePct)
	}
	if rep.ContestedCount != 1 {
		t.Fatalf("contested_count = %d, want 1", rep.ContestedCount)
	}
	if rep.ExpiredCount != 1 {
		t.Fatalf("expired_count = %d, want 1", rep.ExpiredCount)
	}
	wantScopes := map[string]int{
		"user:alice": 2,
		"team:core":  1,
		"user:bob":   1,
	}
	if len(rep.ScopeCounts) != len(wantScopes) {
		t.Fatalf("scope_counts = %+v, want %+v", rep.ScopeCounts, wantScopes)
	}
	for k, want := range wantScopes {
		if got := rep.ScopeCounts[k]; got != want {
			t.Fatalf("scope_counts[%q] = %d, want %d", k, got, want)
		}
	}
	// Active pages include default-status episodic/semantic files (7 total):
	// 72h, 48h, 24h, 24h, 0, 0, 0 -> mean 1 day.
	wantAvg := 1.0
	if rep.AvgAgeDays < wantAvg-0.5 || rep.AvgAgeDays > wantAvg+0.5 {
		t.Fatalf("avg_age_days = %v, want ~%.1f", rep.AvgAgeDays, wantAvg)
	}
}

func TestScan_contradictions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("pages/a.md", `---
memory_kind: semantic
contradicts: pages/b.md
---
# a
`)
	write("pages/b.md", `---
memory_kind: semantic
memory_status: contested
---
# b
`)
	write("pages/c.md", `---
memory_kind: semantic
contradicts:
  - pages/d.md
  - pages/e.md
---
# c
`)
	write("pages/d.md", `---
memory_kind: semantic
---
# d
`)

	s, err := storage.NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Scan(context.Background(), s, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Contradictions != 3 {
		t.Fatalf("contradictions: got %d want 3 (a, b contested, c)", rep.Contradictions)
	}
}

func TestCoveragePercent(t *testing.T) {
	t.Parallel()
	if got := coveragePercent(0, 0); got != 0 {
		t.Fatalf("zero episodic: got %v", got)
	}
	if got := coveragePercent(4, 1); got != 75 {
		t.Fatalf("75%% coverage: got %v", got)
	}
}
