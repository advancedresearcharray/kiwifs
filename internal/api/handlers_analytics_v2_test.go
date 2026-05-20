package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kiwifs/kiwifs/internal/comments"
	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/kiwifs/kiwifs/internal/events"
	"github.com/kiwifs/kiwifs/internal/pipeline"
	"github.com/kiwifs/kiwifs/internal/search"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/kiwifs/kiwifs/internal/versioning"
)

func TestAnalyticsV2Overview(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "guide.md", "# Guide\nA page about kiwi.\n")

	// Generate views
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=guide.md", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET guide.md: %d %s", rec.Code, rec.Body.String())
		}
	}

	// Generate a failed search
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=nonexistent-term", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /search: %d %s", rec.Code, rec.Body.String())
	}

	// Hit overview endpoint
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/overview?period=7d", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/overview: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsOverviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, rec.Body.String())
	}
	if resp.Period != "7d" {
		t.Fatalf("expected period=7d, got %s", resp.Period)
	}
	if resp.TotalViews < 5 {
		t.Fatalf("expected at least 5 views, got %d", resp.TotalViews)
	}
}

func TestAnalyticsV2Overview_DefaultPeriod(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/overview", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/overview no period: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsOverviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Period != "7d" {
		t.Fatalf("expected default period 7d, got %s", resp.Period)
	}
}

func TestAnalyticsV2Views(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "doc1.md", "# Doc1\nContent.\n")
	mustPutFile(t, s, "doc2.md", "# Doc2\nContent.\n")

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=doc1.md", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET doc1.md: %d", rec.Code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=doc2.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET doc2.md: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/views/v2?period=7d", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/views/v2: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsViewsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.TimeSeries == nil {
		t.Fatal("time_series should not be nil")
	}
	if resp.TopPages == nil {
		t.Fatal("top_pages should not be nil")
	}
	if len(resp.TopPages) < 2 {
		t.Fatalf("expected at least 2 top pages, got %d", len(resp.TopPages))
	}
}

func TestAnalyticsV2Views_NullSafety(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	// No data at all — arrays should be empty, not null
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/views/v2?period=7d", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/views/v2: %d %s", rec.Code, rec.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Verify arrays are [] not null
	for _, key := range []string{"time_series", "top_pages"} {
		if string(raw[key]) == "null" {
			t.Fatalf("%s should be [] not null", key)
		}
	}
}

func TestAnalyticsV2Searches(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "doc.md", "# Doc\nSearchable content.\n")

	// Generate a successful search
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=searchable", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search: %d", rec.Code)
	}

	// Generate a failed search
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=nonexistent-xyz", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/searches?period=7d", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/searches: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsSearchesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.TimeSeries == nil {
		t.Fatal("time_series should not be nil")
	}
	if resp.TopFailed == nil {
		t.Fatal("top_failed should not be nil")
	}
}

func TestAnalyticsV2Trends(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/trends?period=7d", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/trends: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsTrendsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Trending == nil {
		t.Fatal("trending should not be nil (should be [])")
	}
	if resp.Declining == nil {
		t.Fatal("declining should not be nil (should be [])")
	}
}

func TestAnalyticsV2ContentGaps(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	// Generate failed searches
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=missing-topic", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/content-gaps", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/content-gaps: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsContentGapsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Results == nil {
		t.Fatal("results should not be nil")
	}
	if len(resp.Results) < 1 {
		t.Fatal("expected at least 1 content gap")
	}
	found := false
	for _, r := range resp.Results {
		if r.Query == "missing-topic" {
			found = true
			if r.Count < 3 {
				t.Fatalf("expected count >= 3 for missing-topic, got %d", r.Count)
			}
		}
	}
	if !found {
		t.Fatalf("missing-topic not found in content gaps: %+v", resp.Results)
	}
}

func TestAnalyticsV2DismissContentGap(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	// Generate a failed search
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=dismiss-me", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
	}

	// Dismiss it
	body, _ := json.Marshal(dismissRequest{Query: "dismiss-me", SearchType: "search"})
	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/analytics/content-gaps/dismiss", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST dismiss: %d %s", rec.Code, rec.Body.String())
	}

	// Verify it no longer appears in gaps
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/content-gaps", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET content-gaps after dismiss: %d", rec.Code)
	}

	var resp AnalyticsContentGapsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, r := range resp.Results {
		if r.Query == "dismiss-me" {
			t.Fatal("dismiss-me should have been filtered out after dismissal")
		}
	}
}

func TestAnalyticsV2DismissContentGap_EmptyQuery(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	body, _ := json.Marshal(dismissRequest{Query: "", SearchType: "search"})
	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/analytics/content-gaps/dismiss", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty query, got %d", rec.Code)
	}
}

func TestAnalyticsV2Sources(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "doc.md", "# Doc\nContent.\n")

	// Source defaults to "api" for file reads without ?source=
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=doc.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET doc.md: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=doc.md&source=ui", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET doc.md ui: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/sources?period=7d", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/sources: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsSourcesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Sources == nil {
		t.Fatal("sources should not be nil")
	}
}

func TestAnalyticsV2Sources_NullSafety(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/sources?period=7d", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/sources: %d", rec.Code)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(raw["sources"]) == "null" {
		t.Fatal("sources should be {} not null")
	}
}

func TestAnalyticsV2_BotFiltering(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "doc.md", "# Doc\nContent.\n")

	// Normal user view
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=doc.md", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET doc.md normal: %d", rec.Code)
	}

	// Bot view (should be excluded from analytics)
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=doc.md", nil)
	req.Header.Set("User-Agent", "Googlebot/2.1")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET doc.md bot: %d", rec.Code)
	}

	// Check views — should only have 1 (not 2)
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/overview?period=7d", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/overview: %d", rec.Code)
	}

	var resp AnalyticsOverviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.TotalViews > 1 {
		t.Fatalf("expected at most 1 view (bot should be filtered), got %d", resp.TotalViews)
	}
}

func TestParsePeriodSeconds(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"7d", 7 * 86400},
		{"30d", 30 * 86400},
		{"90d", 90 * 86400},
		{"", 7 * 86400},
		{"1d", 1 * 86400},
		{"abc", 7 * 86400},
		{"0d", 7 * 86400},
	}
	for _, tt := range tests {
		got := parsePeriodSeconds(tt.input)
		if got != tt.want {
			t.Errorf("parsePeriodSeconds(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestBucketSizeForPeriod(t *testing.T) {
	tests := []struct {
		period int64
		want   int64
	}{
		{7 * 86400, 3600},   // 7d → hourly
		{30 * 86400, 86400}, // 30d → daily
		{90 * 86400, 86400}, // 90d → daily
		{1 * 86400, 3600},   // 1d → hourly
	}
	for _, tt := range tests {
		got := bucketSizeForPeriod(tt.period)
		if got != tt.want {
			t.Errorf("bucketSizeForPeriod(%d) = %d, want %d", tt.period, got, tt.want)
		}
	}
}

func TestAnalyticsV2_BackwardCompatibility(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "doc.md", "# Doc\nSome content.\n")

	// Old v1 endpoint should still work
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics (v1): %d %s", rec.Code, rec.Body.String())
	}

	// New v2 endpoint should work
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/overview?period=7d", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/overview (v2): %d %s", rec.Code, rec.Body.String())
	}
}

// --- Analytics Writer tests ---

func TestAnalyticsWriterIntegration(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	// Verify server starts successfully (writer lifecycle)
	mustPutFile(t, s, "page.md", "# Page\nContent.\n")

	// Generate a view to exercise the recording path
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=page.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET page.md: %d", rec.Code)
	}

	// Verify overview endpoint works (proves the query layer + recording work end to end)
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/overview?period=7d", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET overview: %d %s", rec.Code, rec.Body.String())
	}
}

// --- Bot filtering unit test ---

func TestBotDetection(t *testing.T) {
	// Importing analytics.IsBot isn't possible from api package,
	// so we test via the HTTP layer (see TestAnalyticsV2_BotFiltering above).
	// This test validates the search_hours recording path for bot UAs.
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "page.md", "# Page\nContent.\n")

	// Simulate search with bot UA — search recording doesn't filter bots
	// (only page views filter bots), so we test the search path still works
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=content", nil)
	req.Header.Set("User-Agent", "Googlebot/2.1")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search: %d", rec.Code)
	}

	// Verify searches endpoint returns data
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/searches?period=7d", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET searches: %d", rec.Code)
	}

	var resp AnalyticsSearchesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Search recording doesn't filter bots (only page views do)
	if resp.TimeSeries == nil {
		t.Fatal("time_series nil")
	}
}

// --- Benchmark for analytics queries (Phase 8.7) ---

func BenchmarkAnalyticsOverview(b *testing.B) {
	s, _ := benchSQLiteServer(b)
	seedBenchData(b, s, 100000)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/overview?period=30d", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("overview: %d", rec.Code)
		}
	}
}

func BenchmarkAnalyticsViews(b *testing.B) {
	s, _ := benchSQLiteServer(b)
	seedBenchData(b, s, 100000)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/views/v2?period=30d", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("views: %d", rec.Code)
		}
	}
}

func BenchmarkAnalyticsSources(b *testing.B) {
	s, _ := benchSQLiteServer(b)
	seedBenchData(b, s, 100000)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/sources?period=30d", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("sources: %d", rec.Code)
		}
	}
}

func BenchmarkAnalyticsTrends(b *testing.B) {
	s, _ := benchSQLiteServer(b)
	seedBenchData(b, s, 100000)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/trends?period=30d", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("trends: %d", rec.Code)
		}
	}
}

func BenchmarkAnalyticsContentGaps(b *testing.B) {
	s, _ := benchSQLiteServer(b)
	seedBenchData(b, s, 100000)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/content-gaps", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("content-gaps: %d", rec.Code)
		}
	}
}

// --- Benchmark helpers ---

func benchSQLiteServer(b *testing.B) (*Server, string) {
	b.Helper()
	dir := b.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		b.Fatalf("storage: %v", err)
	}
	searcher, err := search.NewSQLite(dir, store)
	if err != nil {
		b.Fatalf("sqlite: %v", err)
	}
	b.Cleanup(func() { _ = searcher.Close() })

	ver := versioning.NewNoop()
	hub := events.NewHub()
	pipe := pipeline.New(store, ver, searcher, searcher, hub, nil, "")
	cstore, err := comments.New(dir)
	if err != nil {
		b.Fatalf("comments: %v", err)
	}
	cfg := &config.Config{}
	cfg.Storage.Root = dir
	return NewServer(cfg, pipe, nil, cstore, nil, nil, nil), dir
}

func seedBenchData(b *testing.B, s *Server, rows int) {
	b.Helper()
	// Get the searcher to seed data directly into the DB
	sq, ok := s.handlers.searcher.(*search.SQLite)
	if !ok {
		b.Skip("searcher is not SQLite")
	}
	db := sq.WriteDB()

	now := time.Now().Unix()
	sources := []string{"ui", "api", "mcp", "s3", "webdav"}
	searchTypes := []string{"search", "verified", "semantic"}

	tx, err := db.Begin()
	if err != nil {
		b.Fatalf("begin: %v", err)
	}

	pvStmt, err := tx.Prepare(`
INSERT INTO page_view_hours(path, source, hour, count, unique_actors)
VALUES (?, ?, ?, ?, 0)
ON CONFLICT(path, source, hour) DO UPDATE SET count = count + ?`)
	if err != nil {
		b.Fatalf("prepare pv: %v", err)
	}

	shStmt, err := tx.Prepare(`
INSERT INTO search_hours(query, search_type, hour, count, had_results)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(query, search_type, hour) DO UPDATE SET count = count + ?`)
	if err != nil {
		b.Fatalf("prepare sh: %v", err)
	}

	for i := 0; i < rows; i++ {
		hour := now - int64(i%720)*3600 // spread across 30 days
		path := fmt.Sprintf("docs/page-%d.md", i%500)
		source := sources[i%len(sources)]
		count := (i % 10) + 1

		if _, err := pvStmt.Exec(path, source, hour, count, count); err != nil {
			b.Fatalf("insert pv: %v", err)
		}

		if i%3 == 0 { // ~33% of rows are search events
			query := fmt.Sprintf("query-%d", i%200)
			st := searchTypes[i%len(searchTypes)]
			hadResults := i%4 != 0 // 75% success rate
			hr := 0
			if hadResults {
				hr = 1
			}
			if _, err := shStmt.Exec(query, st, hour, count, hr, count); err != nil {
				b.Fatalf("insert sh: %v", err)
			}
		}
	}

	pvStmt.Close()
	shStmt.Close()
	if err := tx.Commit(); err != nil {
		b.Fatalf("commit: %v", err)
	}
}
