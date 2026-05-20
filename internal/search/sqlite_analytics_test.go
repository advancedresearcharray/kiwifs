package search

import (
	"context"
	"testing"
	"time"

	"github.com/kiwifs/kiwifs/internal/storage"
)

// newAnalyticsTestSQLite creates a temporary SQLite search instance for analytics testing.
func newAnalyticsTestSQLite(t *testing.T) *SQLite {
	t.Helper()
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	s, err := NewSQLite(dir, store)
	if err != nil {
		t.Fatalf("new sqlite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// seedPageViewHours inserts page view rows directly for deterministic testing.
func seedPageViewHours(t *testing.T, s *SQLite, path, source string, hour int64, count int) {
	t.Helper()
	_, err := s.writeDB.Exec(`
INSERT INTO page_view_hours(path, source, hour, count, unique_actors)
VALUES (?, ?, ?, ?, 0)
ON CONFLICT(path, source, hour) DO UPDATE SET count = count + ?`, path, source, hour, count, count)
	if err != nil {
		t.Fatalf("seed page_view_hours: %v", err)
	}
}

// seedSearchHours inserts search rows directly for deterministic testing.
func seedSearchHours(t *testing.T, s *SQLite, query, searchType string, hour int64, count int, hadResults bool) {
	t.Helper()
	hr := 0
	if hadResults {
		hr = 1
	}
	_, err := s.writeDB.Exec(`
INSERT INTO search_hours(query, search_type, hour, count, had_results)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(query, search_type, hour) DO UPDATE SET count = count + ?`, query, searchType, hour, count, hr, count)
	if err != nil {
		t.Fatalf("seed search_hours: %v", err)
	}
}

func TestPageViewsInRange(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	seedPageViewHours(t, s, "doc1.md", "ui", hour, 5)
	seedPageViewHours(t, s, "doc2.md", "api", hour, 3)
	seedPageViewHours(t, s, "doc1.md", "ui", hour-3600, 2)

	stats, err := s.PageViewsInRange(ctx, "", hour-7200, now)
	if err != nil {
		t.Fatalf("PageViewsInRange: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 pages, got %d: %+v", len(stats), stats)
	}
	// doc1 should be first (7 total views)
	if stats[0].Path != "doc1.md" || stats[0].Count != 7 {
		t.Fatalf("expected doc1.md with 7 views, got %+v", stats[0])
	}
	if stats[1].Path != "doc2.md" || stats[1].Count != 3 {
		t.Fatalf("expected doc2.md with 3 views, got %+v", stats[1])
	}
}

func TestPageViewsInRange_PathPrefix(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	seedPageViewHours(t, s, "guides/setup.md", "ui", hour, 3)
	seedPageViewHours(t, s, "guides/deploy.md", "ui", hour, 2)
	seedPageViewHours(t, s, "notes/idea.md", "ui", hour, 1)

	stats, err := s.PageViewsInRange(ctx, "guides/", hour-3600, now)
	if err != nil {
		t.Fatalf("PageViewsInRange prefix: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 pages under guides/, got %d", len(stats))
	}
}

func TestPageViewTimeSeries(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	// Seed 3 hours of data for doc1.md
	seedPageViewHours(t, s, "doc1.md", "ui", hour, 10)
	seedPageViewHours(t, s, "doc1.md", "ui", hour-3600, 5)
	seedPageViewHours(t, s, "doc1.md", "ui", hour-7200, 3)

	pts, err := s.PageViewTimeSeries(ctx, "doc1.md", hour-7200, now, 3600)
	if err != nil {
		t.Fatalf("PageViewTimeSeries: %v", err)
	}
	if len(pts) != 3 {
		t.Fatalf("expected 3 time points, got %d: %+v", len(pts), pts)
	}
	// Should be ordered ascending by timestamp
	if pts[0].Timestamp > pts[1].Timestamp {
		t.Fatal("time series not in ascending order")
	}
	// Oldest bucket should have 3 views
	if pts[0].Count != 3 {
		t.Fatalf("oldest bucket expected 3 views, got %d", pts[0].Count)
	}
	// Newest bucket should have 10 views
	if pts[2].Count != 10 {
		t.Fatalf("newest bucket expected 10 views, got %d", pts[2].Count)
	}
}

func TestPageViewTimeSeries_AllPaths(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	seedPageViewHours(t, s, "a.md", "ui", hour, 4)
	seedPageViewHours(t, s, "b.md", "ui", hour, 6)

	pts, err := s.PageViewTimeSeries(ctx, "", hour-3600, now, 3600)
	if err != nil {
		t.Fatalf("PageViewTimeSeries all: %v", err)
	}
	if len(pts) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(pts))
	}
	if pts[0].Count != 10 {
		t.Fatalf("expected total 10, got %d", pts[0].Count)
	}
}

func TestSearchSuccessRate(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	// 6 successful, 4 failed = 60% success rate
	seedSearchHours(t, s, "golang", "search", hour, 6, true)
	seedSearchHours(t, s, "notfound", "search", hour, 4, false)

	rate, err := s.SearchSuccessRate(ctx, hour-3600, now)
	if err != nil {
		t.Fatalf("SearchSuccessRate: %v", err)
	}
	if rate < 0.59 || rate > 0.61 {
		t.Fatalf("expected ~0.6 success rate, got %f", rate)
	}
}

func TestSearchSuccessRate_NoData(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	rate, err := s.SearchSuccessRate(ctx, 0, time.Now().Unix())
	if err != nil {
		t.Fatalf("SearchSuccessRate no data: %v", err)
	}
	if rate != 0 {
		t.Fatalf("expected 0 rate with no data, got %f", rate)
	}
}

func TestTrendingPages(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	// Current period: last 7 days
	seedPageViewHours(t, s, "popular.md", "ui", hour, 20)
	seedPageViewHours(t, s, "stable.md", "ui", hour, 5)

	// Previous period: 7-14 days ago
	prevHour := hour - 7*86400
	seedPageViewHours(t, s, "popular.md", "ui", prevHour, 5)
	seedPageViewHours(t, s, "stable.md", "ui", prevHour, 5)

	trending, err := s.TrendingPages(ctx, 7)
	if err != nil {
		t.Fatalf("TrendingPages: %v", err)
	}
	if len(trending) < 1 {
		t.Fatal("expected at least 1 trending page")
	}
	// popular.md should be trending (20 vs 5 = +300%)
	if trending[0].Path != "popular.md" {
		t.Fatalf("expected popular.md to be top trending, got %s", trending[0].Path)
	}
	if trending[0].DeltaPercent <= 0 {
		t.Fatalf("expected positive delta for trending page, got %f", trending[0].DeltaPercent)
	}
}

func TestDecliningPages(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	// Current period: low views
	seedPageViewHours(t, s, "fading.md", "ui", hour, 2)

	// Previous period: high views
	prevHour := hour - 7*86400
	seedPageViewHours(t, s, "fading.md", "ui", prevHour, 20)

	declining, err := s.DecliningPages(ctx, 7)
	if err != nil {
		t.Fatalf("DecliningPages: %v", err)
	}
	if len(declining) < 1 {
		t.Fatal("expected at least 1 declining page")
	}
	if declining[0].Path != "fading.md" {
		t.Fatalf("expected fading.md, got %s", declining[0].Path)
	}
	if declining[0].DeltaPercent >= 0 {
		t.Fatalf("expected negative delta for declining page, got %f", declining[0].DeltaPercent)
	}
}

func TestContentGaps(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	seedSearchHours(t, s, "missing-feature", "search", hour, 5, false)
	seedSearchHours(t, s, "another-gap", "search", hour, 2, false)
	seedSearchHours(t, s, "found-it", "search", hour, 10, true)

	gaps, err := s.ContentGaps(ctx, 10)
	if err != nil {
		t.Fatalf("ContentGaps: %v", err)
	}
	if len(gaps) != 2 {
		t.Fatalf("expected 2 content gaps (failed searches only), got %d: %+v", len(gaps), gaps)
	}
	// Should be ordered by count desc
	if gaps[0].Query != "missing-feature" || gaps[0].Count != 5 {
		t.Fatalf("expected missing-feature with count 5, got %+v", gaps[0])
	}
}

func TestContentGaps_Dismissed(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	seedSearchHours(t, s, "irrelevant", "search", hour, 5, false)
	seedSearchHours(t, s, "real-gap", "search", hour, 3, false)

	// Dismiss "irrelevant"
	if err := s.DismissContentGap(ctx, "irrelevant", "search"); err != nil {
		t.Fatalf("DismissContentGap: %v", err)
	}

	gaps, err := s.ContentGaps(ctx, 10)
	if err != nil {
		t.Fatalf("ContentGaps after dismiss: %v", err)
	}
	if len(gaps) != 1 {
		t.Fatalf("expected 1 gap after dismissal, got %d", len(gaps))
	}
	if gaps[0].Query != "real-gap" {
		t.Fatalf("expected real-gap, got %s", gaps[0].Query)
	}
}

func TestDismissContentGap_EmptyQuery(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	err := s.DismissContentGap(ctx, "", "search")
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestAnalyticsOverview(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	// Current period views
	seedPageViewHours(t, s, "doc1.md", "ui", hour, 10)
	seedPageViewHours(t, s, "doc2.md", "api", hour, 5)

	// Previous period views (7 days ago)
	prevHour := hour - 7*86400
	seedPageViewHours(t, s, "doc1.md", "ui", prevHour, 3)

	// Current period searches
	seedSearchHours(t, s, "query1", "search", hour, 8, true)
	seedSearchHours(t, s, "query2", "search", hour, 2, false)

	// Previous period searches
	seedSearchHours(t, s, "query1", "search", prevHour, 5, true)

	stats, err := s.AnalyticsOverview(ctx, 7*86400)
	if err != nil {
		t.Fatalf("AnalyticsOverview: %v", err)
	}
	if stats.TotalViews != 15 {
		t.Fatalf("expected 15 total views, got %d", stats.TotalViews)
	}
	if stats.UniquePages != 2 {
		t.Fatalf("expected 2 unique pages, got %d", stats.UniquePages)
	}
	if stats.TotalSearches != 10 {
		t.Fatalf("expected 10 total searches, got %d", stats.TotalSearches)
	}
	// 8 successful out of 10 = 0.8
	if stats.SearchSuccessRate < 0.79 || stats.SearchSuccessRate > 0.81 {
		t.Fatalf("expected ~0.8 success rate, got %f", stats.SearchSuccessRate)
	}
	// Views delta: 15 current vs 3 previous = +400%
	if stats.ViewsDelta < 390 || stats.ViewsDelta > 410 {
		t.Fatalf("expected views delta ~400%%, got %f", stats.ViewsDelta)
	}
}

func TestAnalyticsOverview_EmptyData(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	stats, err := s.AnalyticsOverview(ctx, 7*86400)
	if err != nil {
		t.Fatalf("AnalyticsOverview empty: %v", err)
	}
	if stats.TotalViews != 0 {
		t.Fatalf("expected 0 views, got %d", stats.TotalViews)
	}
	if stats.TotalSearches != 0 {
		t.Fatalf("expected 0 searches, got %d", stats.TotalSearches)
	}
	if stats.ViewsDelta != 0 {
		t.Fatalf("expected 0 views delta, got %f", stats.ViewsDelta)
	}
}

func TestSearchTimeSeries(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	seedSearchHours(t, s, "query1", "search", hour, 5, true)
	seedSearchHours(t, s, "query2", "search", hour, 3, false)
	seedSearchHours(t, s, "query3", "search", hour-3600, 2, true)

	pts, err := s.SearchTimeSeries(ctx, hour-7200, now, 3600)
	if err != nil {
		t.Fatalf("SearchTimeSeries: %v", err)
	}
	if len(pts) != 2 {
		t.Fatalf("expected 2 time points, got %d: %+v", len(pts), pts)
	}
	// Older bucket has 2, newer has 8
	if pts[0].Count != 2 {
		t.Fatalf("older bucket expected 2, got %d", pts[0].Count)
	}
	if pts[1].Count != 8 {
		t.Fatalf("newer bucket expected 8, got %d", pts[1].Count)
	}
}

func TestTopSearches(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	seedSearchHours(t, s, "popular-query", "search", hour, 10, true)
	seedSearchHours(t, s, "failing-query", "search", hour, 7, false)
	seedSearchHours(t, s, "rare-query", "search", hour, 1, true)

	// All searches
	all, err := s.TopSearches(ctx, 10, hour-3600, false)
	if err != nil {
		t.Fatalf("TopSearches all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 searches, got %d", len(all))
	}
	if all[0].Query != "popular-query" {
		t.Fatalf("expected popular-query first, got %s", all[0].Query)
	}

	// Failed only
	failed, err := s.TopSearches(ctx, 10, hour-3600, true)
	if err != nil {
		t.Fatalf("TopSearches failed: %v", err)
	}
	if len(failed) != 1 {
		t.Fatalf("expected 1 failed search, got %d", len(failed))
	}
	if failed[0].Query != "failing-query" {
		t.Fatalf("expected failing-query, got %s", failed[0].Query)
	}
}

func TestSourceBreakdown(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	seedPageViewHours(t, s, "doc1.md", "ui", hour, 10)
	seedPageViewHours(t, s, "doc1.md", "api", hour, 5)
	seedPageViewHours(t, s, "doc2.md", "mcp", hour, 3)

	sources, err := s.SourceBreakdown(ctx, hour-3600, now)
	if err != nil {
		t.Fatalf("SourceBreakdown: %v", err)
	}
	if len(sources) != 3 {
		t.Fatalf("expected 3 sources, got %d: %v", len(sources), sources)
	}
	if sources["ui"] != 10 {
		t.Fatalf("expected ui=10, got %d", sources["ui"])
	}
	if sources["api"] != 5 {
		t.Fatalf("expected api=5, got %d", sources["api"])
	}
	if sources["mcp"] != 3 {
		t.Fatalf("expected mcp=3, got %d", sources["mcp"])
	}
}

func TestSourceBreakdown_Empty(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	sources, err := s.SourceBreakdown(ctx, 0, time.Now().Unix())
	if err != nil {
		t.Fatalf("SourceBreakdown empty: %v", err)
	}
	if len(sources) != 0 {
		t.Fatalf("expected empty sources, got %v", sources)
	}
}

func TestRollupHourlyToDaily(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	// Seed old hourly data (8+ days ago)
	oldHour := hour - 8*86400
	seedPageViewHours(t, s, "old-doc.md", "ui", oldHour, 10)
	seedPageViewHours(t, s, "old-doc.md", "ui", oldHour+3600, 5)
	seedPageViewHours(t, s, "old-doc.md", "api", oldHour, 3)

	// Seed recent hourly data (should not be rolled up)
	seedPageViewHours(t, s, "new-doc.md", "ui", hour, 20)

	// Rollup with 7-day retention
	deleted, err := s.RollupHourlyToDaily(ctx, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("RollupHourlyToDaily: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected 3 hourly rows deleted, got %d", deleted)
	}

	// Verify daily aggregates exist
	var dailyCount int
	err = s.readDB.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_view_days WHERE path = 'old-doc.md'`).Scan(&dailyCount)
	if err != nil {
		t.Fatalf("query daily: %v", err)
	}
	if dailyCount != 18 { // 10 + 5 + 3
		t.Fatalf("expected daily aggregate = 18, got %d", dailyCount)
	}

	// Verify recent data is untouched in hourly table
	var recentCount int
	err = s.readDB.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_view_hours WHERE path = 'new-doc.md'`).Scan(&recentCount)
	if err != nil {
		t.Fatalf("query recent: %v", err)
	}
	if recentCount != 20 {
		t.Fatalf("expected recent hourly = 20, got %d", recentCount)
	}

	// Verify old hourly rows are gone
	var oldCount int
	err = s.readDB.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_view_hours WHERE path = 'old-doc.md'`).Scan(&oldCount)
	if err != nil {
		t.Fatalf("query old hourly: %v", err)
	}
	if oldCount != 0 {
		t.Fatalf("expected old hourly data deleted, got %d", oldCount)
	}
}

func TestRollupHourlyToDaily_UnionQuerySeamless(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	// Seed some old data that we'll roll up (8 days ago)
	oldHour := hour - 8*86400
	seedPageViewHours(t, s, "doc.md", "ui", oldHour, 10)

	// Seed recent data
	seedPageViewHours(t, s, "doc.md", "ui", hour, 5)

	// Use a wide enough range to cover the truncated day boundary.
	// After rollup, the daily bucket is at truncateDay(oldHour) which could
	// be up to 23h59m before oldHour. Use 9-day-old start to be safe.
	rangeStart := hour - 9*86400

	// Before rollup: query should return 15 total
	stats, err := s.PageViewsInRange(ctx, "", rangeStart, now)
	if err != nil {
		t.Fatalf("before rollup: %v", err)
	}
	if len(stats) != 1 || stats[0].Count != 15 {
		t.Fatalf("before rollup expected 15, got %+v", stats)
	}

	// Rollup
	_, err = s.RollupHourlyToDaily(ctx, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("rollup: %v", err)
	}

	// After rollup: query should STILL return 15 total (UNION of hourly + daily)
	stats, err = s.PageViewsInRange(ctx, "", rangeStart, now)
	if err != nil {
		t.Fatalf("after rollup: %v", err)
	}
	if len(stats) != 1 || stats[0].Count != 15 {
		t.Fatalf("after rollup expected 15 (union), got %+v", stats)
	}
}

func TestRecordPageView_WritesToBothTables(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	if err := s.RecordPageView(ctx, "test.md", "ui"); err != nil {
		t.Fatalf("RecordPageView: %v", err)
	}

	// Check legacy table
	var legacyCount int
	err := s.readDB.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_views WHERE path = 'test.md'`).Scan(&legacyCount)
	if err != nil {
		t.Fatalf("query legacy: %v", err)
	}
	if legacyCount != 1 {
		t.Fatalf("expected legacy count 1, got %d", legacyCount)
	}

	// Check hourly table
	var hourlyCount int
	err = s.readDB.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_view_hours WHERE path = 'test.md'`).Scan(&hourlyCount)
	if err != nil {
		t.Fatalf("query hourly: %v", err)
	}
	if hourlyCount != 1 {
		t.Fatalf("expected hourly count 1, got %d", hourlyCount)
	}
}

func TestRecordSearch_WritesToBothTables(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	// Failed search
	if err := s.RecordSearch(ctx, "missing", "search", false); err != nil {
		t.Fatalf("RecordSearch failed: %v", err)
	}

	// Successful search
	if err := s.RecordSearch(ctx, "found", "search", true); err != nil {
		t.Fatalf("RecordSearch success: %v", err)
	}

	// Check legacy failed_searches: only "missing" should be there
	var legacyCount int
	err := s.readDB.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM failed_searches`).Scan(&legacyCount)
	if err != nil {
		t.Fatalf("query legacy failed: %v", err)
	}
	if legacyCount != 1 {
		t.Fatalf("expected 1 legacy failed search, got %d", legacyCount)
	}

	// Check search_hours: both should be there
	var hourlyTotal int
	err = s.readDB.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM search_hours`).Scan(&hourlyTotal)
	if err != nil {
		t.Fatalf("query hourly searches: %v", err)
	}
	if hourlyTotal != 2 {
		t.Fatalf("expected 2 hourly search records, got %d", hourlyTotal)
	}
}

func TestRecordSearch_EmptyQuery(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	if err := s.RecordSearch(ctx, "", "search", false); err != nil {
		t.Fatalf("expected no error for empty query, got %v", err)
	}

	var count int
	err := s.readDB.QueryRow(`SELECT COUNT(*) FROM search_hours`).Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no rows for empty query, got %d", count)
	}
}

func TestLeastViewed(t *testing.T) {
	s := newAnalyticsTestSQLite(t)
	ctx := context.Background()

	now := time.Now().Unix()
	hour := now - now%3600

	// Directly seed file_meta (LeastViewed joins against this table)
	nowStr := time.Now().Format(time.RFC3339)
	for _, path := range []string{"viewed.md", "unviewed.md"} {
		_, err := s.writeDB.Exec(`
INSERT OR REPLACE INTO file_meta(path, frontmatter, tasks, updated_at)
VALUES (?, '{}', '[]', ?)`, path, nowStr)
		if err != nil {
			t.Fatalf("seed file_meta %s: %v", path, err)
		}
	}

	// Record views for viewed.md only
	seedPageViewHours(t, s, "viewed.md", "ui", hour, 5)

	least, err := s.LeastViewed(ctx, hour-3600, 10)
	if err != nil {
		t.Fatalf("LeastViewed: %v", err)
	}
	if len(least) < 1 {
		t.Fatal("expected at least 1 unviewed page")
	}
	found := false
	for _, p := range least {
		if p.Path == "unviewed.md" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected unviewed.md in least viewed, got %+v", least)
	}
}

func TestDeltaPercent(t *testing.T) {
	tests := []struct {
		current, prev int
		want          float64
	}{
		{10, 5, 100.0},   // doubled
		{5, 10, -50.0},   // halved
		{0, 0, 0.0},      // zero to zero
		{5, 0, 100.0},    // from zero to something
		{0, 5, -100.0},   // from something to zero
		{100, 100, 0.0},  // stable
	}
	for _, tt := range tests {
		got := deltaPercent(tt.current, tt.prev)
		if got < tt.want-0.1 || got > tt.want+0.1 {
			t.Errorf("deltaPercent(%d, %d) = %f, want %f", tt.current, tt.prev, got, tt.want)
		}
	}
}

func TestTruncateHour(t *testing.T) {
	// 1736950200 % 3600 = 600, so truncated = 1736950200 - 600 = 1736949600
	ts := int64(1736950200)
	truncated := truncateHour(ts)
	expected := int64(1736949600)
	if truncated != expected {
		t.Fatalf("truncateHour(%d) = %d, want %d", ts, truncated, expected)
	}
}

func TestTruncateDay(t *testing.T) {
	// 2025-01-15 14:30:00 UTC = 1736950200
	ts := int64(1736950200)
	truncated := truncateDay(ts)
	// Should be 2025-01-15 00:00:00 UTC = 1736899200
	expected := int64(1736899200)
	if truncated != expected {
		t.Fatalf("truncateDay(%d) = %d, want %d", ts, truncated, expected)
	}
}
