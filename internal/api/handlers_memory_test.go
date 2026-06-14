package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/memory"
)

func TestMemoryReportEndpoint(t *testing.T) {
	s, dir := buildSQLiteTestServer(t)

	epDir := filepath.Join(dir, "episodes")
	if err := os.MkdirAll(epDir, 0755); err != nil {
		t.Fatal(err)
	}
	epBody := `---
memory_kind: episodic
episode_id: ep-api-1
---
# Episode
`
	if err := os.WriteFile(filepath.Join(epDir, "e.md"), []byte(epBody), 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/memory/report", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /memory/report: %d %s", rec.Code, rec.Body.String())
	}

	var rep memory.Report
	if err := json.Unmarshal(rec.Body.Bytes(), &rep); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, rec.Body.String())
	}
	if rep.EpisodicCount != 1 {
		t.Fatalf("episodic_count want 1 got %d", rep.EpisodicCount)
	}
	if len(rep.Unmerged) != 1 || rep.Unmerged[0].EpisodeID != "ep-api-1" {
		t.Fatalf("unmerged: %+v", rep.Unmerged)
	}
	if rep.CoveragePct != 0 {
		t.Fatalf("coverage_pct want 0 got %v", rep.CoveragePct)
	}

	// Semantic page cites the episode
	mustPutFile(t, s, "concepts/c.md", `---
memory_kind: semantic
merged-from:
  - type: episode
    id: ep-api-1
---
# Concept
`)

	req2 := httptest.NewRequest(http.MethodGet, "/api/kiwi/memory/report", nil)
	rec2 := httptest.NewRecorder()
	s.echo.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET /memory/report 2: %d %s", rec2.Code, rec2.Body.String())
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if len(rep.Unmerged) != 0 {
		t.Fatalf("want 0 unmerged after merge ref, got %+v", rep.Unmerged)
	}
	if rep.CoveragePct != 100 {
		t.Fatalf("coverage_pct want 100 got %v", rep.CoveragePct)
	}
}

func TestMemoryReportEpisodesPrefixQuery(t *testing.T) {
	s, dir := buildSQLiteTestServer(t)

	// Under raw/ — not under default episodes/ prefix
	rawDir := filepath.Join(dir, "raw")
	if err := os.MkdirAll(rawDir, 0755); err != nil {
		t.Fatal(err)
	}
	body := `---
memory_kind: episodic
episode_id: z1
---
# z
`
	if err := os.WriteFile(filepath.Join(rawDir, "z.md"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/memory/report?episodes_prefix=raw/", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var rep memory.Report
	if err := json.Unmarshal(rec.Body.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.EpisodicCount != 1 {
		t.Fatalf("want 1 episodic under raw/, got %d", rep.EpisodicCount)
	}
}

func TestMemoryReportPagination(t *testing.T) {
	s, dir := buildSQLiteTestServer(t)

	epDir := filepath.Join(dir, "episodes")
	if err := os.MkdirAll(epDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"ep-page-1", "ep-page-2", "ep-page-3"} {
		body := `---
memory_kind: episodic
episode_id: ` + id + `
---
# Episode
`
		if err := os.WriteFile(filepath.Join(epDir, id+".md"), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/memory/report?limit=2&offset=1", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var rep memory.Report
	if err := json.Unmarshal(rec.Body.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.TotalEpisodic != 3 || rep.TotalUnmerged != 3 {
		t.Fatalf("totals = episodic:%d unmerged:%d", rep.TotalEpisodic, rep.TotalUnmerged)
	}
	if len(rep.Episodes) != 2 || len(rep.Unmerged) != 2 {
		t.Fatalf("paginated lengths = episodes:%d unmerged:%d", len(rep.Episodes), len(rep.Unmerged))
	}
}
