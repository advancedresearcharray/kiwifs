package search

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// truncateHour truncates a unix timestamp to the start of its hour.
func truncateHour(ts int64) int64 { return ts - ts%3600 }

// truncateDay truncates a unix timestamp to the start of its day (UTC).
func truncateDay(ts int64) int64 { return ts - ts%86400 }

// periodBounds returns (since, until, prevSince) for a period in seconds,
// ending at now.
func periodBounds(periodSeconds int64) (since, until, prevSince int64) {
	now := time.Now().Unix()
	until = now
	since = now - periodSeconds
	prevSince = since - periodSeconds
	return
}

// PageViewsInRange returns page view stats summed within [since, until],
// querying both page_view_hours (recent) and page_view_days (rolled up).
func (s *SQLite) PageViewsInRange(ctx context.Context, pathPrefix string, since, until int64) ([]PageViewStat, error) {
	sqlQ := `
SELECT path, SUM(count) AS views, MIN(hour) AS first, MAX(hour) AS last
FROM (
	SELECT path, count, hour FROM page_view_hours WHERE hour >= ? AND hour <= ?
	UNION ALL
	SELECT path, count, day AS hour FROM page_view_days WHERE day >= ? AND day <= ?
) sub`
	args := []any{since, until, since, until}
	if pathPrefix != "" {
		sqlQ += ` WHERE path LIKE ?`
		args = append(args, pathPrefix+"%")
	}
	sqlQ += ` GROUP BY path ORDER BY views DESC, last DESC, path ASC`

	rows, err := s.readDB.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, fmt.Errorf("page views in range: %w", err)
	}
	defer rows.Close()

	var stats []PageViewStat
	for rows.Next() {
		var st PageViewStat
		if err := rows.Scan(&st.Path, &st.Count, &st.FirstSeen, &st.LastSeen); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// PageViewTimeSeries returns time-bucketed view counts for a specific path
// (or all paths if empty) within [since, until].
func (s *SQLite) PageViewTimeSeries(ctx context.Context, path string, since, until, bucketSize int64) ([]TimePoint, error) {
	if bucketSize <= 0 {
		bucketSize = 3600
	}

	sqlQ := `
SELECT (hour / ? * ?) AS bucket, SUM(count) AS views
FROM (
	SELECT hour, count, path FROM page_view_hours WHERE hour >= ? AND hour <= ?
	UNION ALL
	SELECT day AS hour, count, path FROM page_view_days WHERE day >= ? AND day <= ?
) sub`
	args := []any{bucketSize, bucketSize, since, until, since, until}
	if path != "" {
		sqlQ += ` WHERE path = ?`
		args = append(args, path)
	}
	sqlQ += ` GROUP BY bucket ORDER BY bucket ASC`

	rows, err := s.readDB.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, fmt.Errorf("page view time series: %w", err)
	}
	defer rows.Close()

	var points []TimePoint
	for rows.Next() {
		var pt TimePoint
		if err := rows.Scan(&pt.Timestamp, &pt.Count); err != nil {
			return nil, err
		}
		points = append(points, pt)
	}
	return points, rows.Err()
}

// SearchSuccessRate returns the ratio of successful searches in [since, until].
// Returns 0 if no searches recorded.
func (s *SQLite) SearchSuccessRate(ctx context.Context, since, until int64) (float64, error) {
	var total, successful int
	err := s.readDB.QueryRowContext(ctx, `
SELECT
	COALESCE(SUM(count), 0),
	COALESCE(SUM(CASE WHEN had_results = 1 THEN count ELSE 0 END), 0)
FROM search_hours WHERE hour >= ? AND hour <= ?
`, since, until).Scan(&total, &successful)
	if err != nil {
		return 0, fmt.Errorf("search success rate: %w", err)
	}
	if total == 0 {
		return 0, nil
	}
	return float64(successful) / float64(total), nil
}

// TrendingPages compares the current period with the previous period and
// returns pages sorted by positive delta percentage.
func (s *SQLite) TrendingPages(ctx context.Context, periodDays int) ([]TrendStat, error) {
	periodSec := int64(periodDays) * 86400
	since, until, prevSince := periodBounds(periodSec)

	rows, err := s.readDB.QueryContext(ctx, `
SELECT
	COALESCE(cur.path, prev.path) AS page,
	COALESCE(cur.views, 0) AS current_views,
	COALESCE(prev.views, 0) AS previous_views
FROM (
	SELECT path, SUM(count) AS views
	FROM page_view_hours WHERE hour >= ? AND hour <= ?
	GROUP BY path
) cur
LEFT JOIN (
	SELECT path, SUM(count) AS views
	FROM page_view_hours WHERE hour >= ? AND hour < ?
	GROUP BY path
) prev ON cur.path = prev.path
WHERE COALESCE(cur.views, 0) > COALESCE(prev.views, 0)
ORDER BY (CAST(COALESCE(cur.views, 0) - COALESCE(prev.views, 0) AS REAL) / MAX(COALESCE(prev.views, 0), 1)) DESC
LIMIT 20
`, since, until, prevSince, since)
	if err != nil {
		return nil, fmt.Errorf("trending pages: %w", err)
	}
	defer rows.Close()

	var stats []TrendStat
	for rows.Next() {
		var st TrendStat
		if err := rows.Scan(&st.Path, &st.CurrentViews, &st.PreviousViews); err != nil {
			return nil, err
		}
		prev := st.PreviousViews
		if prev == 0 {
			prev = 1
		}
		st.DeltaPercent = float64(st.CurrentViews-st.PreviousViews) / float64(prev) * 100
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// DecliningPages returns pages with negative trends or zero views in the
// current period that had views in the previous period.
func (s *SQLite) DecliningPages(ctx context.Context, periodDays int) ([]TrendStat, error) {
	periodSec := int64(periodDays) * 86400
	since, until, prevSince := periodBounds(periodSec)

	rows, err := s.readDB.QueryContext(ctx, `
SELECT
	prev.path AS page,
	COALESCE(cur.views, 0) AS current_views,
	prev.views AS previous_views
FROM (
	SELECT path, SUM(count) AS views
	FROM page_view_hours WHERE hour >= ? AND hour < ?
	GROUP BY path
) prev
LEFT JOIN (
	SELECT path, SUM(count) AS views
	FROM page_view_hours WHERE hour >= ? AND hour <= ?
	GROUP BY path
) cur ON prev.path = cur.path
WHERE COALESCE(cur.views, 0) < prev.views
ORDER BY (CAST(COALESCE(cur.views, 0) - prev.views AS REAL) / MAX(prev.views, 1)) ASC
LIMIT 20
`, prevSince, since, since, until)
	if err != nil {
		return nil, fmt.Errorf("declining pages: %w", err)
	}
	defer rows.Close()

	var stats []TrendStat
	for rows.Next() {
		var st TrendStat
		if err := rows.Scan(&st.Path, &st.CurrentViews, &st.PreviousViews); err != nil {
			return nil, err
		}
		prev := st.PreviousViews
		if prev == 0 {
			prev = 1
		}
		st.DeltaPercent = float64(st.CurrentViews-st.PreviousViews) / float64(prev) * 100
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// ContentGaps returns failed searches that haven't been dismissed, sorted by
// count descending.
func (s *SQLite) ContentGaps(ctx context.Context, limit int) ([]FailedSearchStat, error) {
	limit = NormalizeLimit(limit)

	rows, err := s.readDB.QueryContext(ctx, `
SELECT sh.query, sh.search_type, SUM(sh.count) AS total, MIN(sh.hour) AS first, MAX(sh.hour) AS last
FROM search_hours sh
LEFT JOIN content_gap_dismissals cgd ON sh.query = cgd.query AND sh.search_type = cgd.search_type
WHERE sh.had_results = 0 AND cgd.query IS NULL
GROUP BY sh.query, sh.search_type
ORDER BY total DESC, last DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("content gaps: %w", err)
	}
	defer rows.Close()

	var stats []FailedSearchStat
	for rows.Next() {
		var st FailedSearchStat
		if err := rows.Scan(&st.Query, &st.SearchType, &st.Count, &st.FirstSeen, &st.LastSeen); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// DismissContentGap marks a content gap as noise so it no longer appears.
func (s *SQLite) DismissContentGap(ctx context.Context, query, searchType string) error {
	query = strings.TrimSpace(query)
	searchType = strings.TrimSpace(searchType)
	if query == "" {
		return fmt.Errorf("query is required")
	}
	if searchType == "" {
		searchType = "search"
	}
	now := time.Now().Unix()
	_, err := s.writeDB.ExecContext(ctx, `
INSERT OR REPLACE INTO content_gap_dismissals(query, search_type, dismissed_at)
VALUES (?, ?, ?)
`, query, searchType, now)
	return err
}

// AnalyticsOverview returns summary stats with deltas compared to the previous period.
func (s *SQLite) AnalyticsOverview(ctx context.Context, periodSeconds int64) (*OverviewStats, error) {
	since, until, prevSince := periodBounds(periodSeconds)

	var stats OverviewStats

	// Current period views
	err := s.readDB.QueryRowContext(ctx, `
SELECT COALESCE(SUM(count), 0), COUNT(DISTINCT path)
FROM page_view_hours WHERE hour >= ? AND hour <= ?
`, since, until).Scan(&stats.TotalViews, &stats.UniquePages)
	if err != nil {
		return nil, fmt.Errorf("overview current views: %w", err)
	}

	// Previous period views
	var prevViews, prevUnique int
	err = s.readDB.QueryRowContext(ctx, `
SELECT COALESCE(SUM(count), 0), COUNT(DISTINCT path)
FROM page_view_hours WHERE hour >= ? AND hour < ?
`, prevSince, since).Scan(&prevViews, &prevUnique)
	if err != nil {
		return nil, fmt.Errorf("overview prev views: %w", err)
	}

	stats.ViewsDelta = deltaPercent(stats.TotalViews, prevViews)
	stats.UniquePagesDelta = deltaPercent(stats.UniquePages, prevUnique)

	// Current period searches
	var totalSearches, successfulSearches int
	err = s.readDB.QueryRowContext(ctx, `
SELECT COALESCE(SUM(count), 0),
	COALESCE(SUM(CASE WHEN had_results = 1 THEN count ELSE 0 END), 0)
FROM search_hours WHERE hour >= ? AND hour <= ?
`, since, until).Scan(&totalSearches, &successfulSearches)
	if err != nil {
		return nil, fmt.Errorf("overview current searches: %w", err)
	}
	stats.TotalSearches = totalSearches
	if totalSearches > 0 {
		stats.SearchSuccessRate = float64(successfulSearches) / float64(totalSearches)
	}

	// Previous period searches
	var prevSearches, prevSuccess int
	err = s.readDB.QueryRowContext(ctx, `
SELECT COALESCE(SUM(count), 0),
	COALESCE(SUM(CASE WHEN had_results = 1 THEN count ELSE 0 END), 0)
FROM search_hours WHERE hour >= ? AND hour < ?
`, prevSince, since).Scan(&prevSearches, &prevSuccess)
	if err != nil {
		return nil, fmt.Errorf("overview prev searches: %w", err)
	}
	stats.SearchesDelta = deltaPercent(totalSearches, prevSearches)

	var prevRate float64
	if prevSearches > 0 {
		prevRate = float64(prevSuccess) / float64(prevSearches)
	}
	stats.SuccessRateDelta = (stats.SearchSuccessRate - prevRate) * 100 // percentage points

	return &stats, nil
}

// SearchTimeSeries returns time-bucketed search counts.
func (s *SQLite) SearchTimeSeries(ctx context.Context, since, until, bucketSize int64) ([]TimePoint, error) {
	if bucketSize <= 0 {
		bucketSize = 3600
	}

	rows, err := s.readDB.QueryContext(ctx, `
SELECT (hour / ? * ?) AS bucket, SUM(count) AS total
FROM search_hours WHERE hour >= ? AND hour <= ?
GROUP BY bucket ORDER BY bucket ASC
`, bucketSize, bucketSize, since, until)
	if err != nil {
		return nil, fmt.Errorf("search time series: %w", err)
	}
	defer rows.Close()

	var points []TimePoint
	for rows.Next() {
		var pt TimePoint
		if err := rows.Scan(&pt.Timestamp, &pt.Count); err != nil {
			return nil, err
		}
		points = append(points, pt)
	}
	return points, rows.Err()
}

// TopSearches returns the most popular search queries in a period.
func (s *SQLite) TopSearches(ctx context.Context, limit int, since int64, failedOnly bool) ([]SearchStat, error) {
	limit = NormalizeLimit(limit)

	sqlQ := `
SELECT query, search_type, SUM(count) AS total, had_results
FROM search_hours WHERE hour >= ?`
	args := []any{since}
	if failedOnly {
		sqlQ += ` AND had_results = 0`
	}
	sqlQ += ` GROUP BY query, search_type, had_results ORDER BY total DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.readDB.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, fmt.Errorf("top searches: %w", err)
	}
	defer rows.Close()

	var stats []SearchStat
	for rows.Next() {
		var st SearchStat
		if err := rows.Scan(&st.Query, &st.SearchType, &st.Count, &st.HadResults); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// SourceBreakdown returns view counts grouped by source within [since, until].
func (s *SQLite) SourceBreakdown(ctx context.Context, since, until int64) (map[string]int, error) {
	rows, err := s.readDB.QueryContext(ctx, `
SELECT source, SUM(count) AS views
FROM page_view_hours WHERE hour >= ? AND hour <= ?
GROUP BY source ORDER BY views DESC
`, since, until)
	if err != nil {
		return nil, fmt.Errorf("source breakdown: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var src string
		var count int
		if err := rows.Scan(&src, &count); err != nil {
			return nil, err
		}
		result[src] = count
	}
	return result, rows.Err()
}

// LeastViewed returns pages from file_meta that have zero views in the
// given period, ordered by last update date.
func (s *SQLite) LeastViewed(ctx context.Context, since int64, limit int) ([]PageViewStat, error) {
	limit = NormalizeLimit(limit)

	rows, err := s.readDB.QueryContext(ctx, `
SELECT fm.path, 0, 0, 0
FROM file_meta fm
LEFT JOIN (
	SELECT path, SUM(count) AS views
	FROM page_view_hours WHERE hour >= ?
	GROUP BY path
) pv ON fm.path = pv.path
WHERE pv.views IS NULL OR pv.views = 0
ORDER BY fm.updated_at DESC
LIMIT ?
`, since, limit)
	if err != nil {
		return nil, fmt.Errorf("least viewed: %w", err)
	}
	defer rows.Close()

	var stats []PageViewStat
	for rows.Next() {
		var st PageViewStat
		if err := rows.Scan(&st.Path, &st.Count, &st.FirstSeen, &st.LastSeen); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// RollupHourlyToDaily collapses hourly rows older than maxAge into daily
// aggregates and deletes the hourly originals.
func (s *SQLite) RollupHourlyToDaily(ctx context.Context, maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Unix() - int64(maxAge.Seconds())
	cutoffDay := truncateDay(cutoff)

	// Insert daily aggregates from hourly data older than cutoff.
	res, err := s.writeDB.ExecContext(ctx, `
INSERT INTO page_view_days(path, source, day, count, unique_actors)
SELECT path, source, (hour / 86400 * 86400) AS day, SUM(count), MAX(unique_actors)
FROM page_view_hours
WHERE hour < ?
GROUP BY path, source, day
ON CONFLICT(path, source, day) DO UPDATE SET
	count = count + excluded.count,
	unique_actors = MAX(unique_actors, excluded.unique_actors)
`, cutoffDay)
	if err != nil {
		return 0, fmt.Errorf("rollup insert: %w", err)
	}

	// Delete rolled-up hourly rows.
	del, err := s.writeDB.ExecContext(ctx, `
DELETE FROM page_view_hours WHERE hour < ?
`, cutoffDay)
	if err != nil {
		return 0, fmt.Errorf("rollup delete: %w", err)
	}
	n, _ := del.RowsAffected()
	_ = res
	return n, nil
}

func deltaPercent(current, previous int) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100 // new activity from zero baseline
	}
	return float64(current-previous) / float64(previous) * 100
}
