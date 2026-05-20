# Analytics v2 — Implementation Checklist

> Zero new Go dependencies. Optional: one small chart lib for frontend sparklines (or raw SVG).

---

## Phase 1: Backend — Time-bucketed storage

- [ ] **1.1** Add `page_view_hours` table: `(path TEXT, source TEXT, hour INTEGER, count INTEGER, unique_actors INTEGER, PRIMARY KEY(path, source, hour))`
- [ ] **1.2** Add `search_hours` table: `(query TEXT, search_type TEXT, hour INTEGER, count INTEGER, had_results INTEGER, PRIMARY KEY(query, search_type, hour))`
- [ ] **1.3** Add indexes: `idx_pvh_hour`, `idx_pvh_path`, `idx_sh_hour`
- [ ] **1.4** Auto-migration in `createSchema`: detect old `page_views`/`failed_searches` tables, migrate existing counts into the hour of `last_seen`, then drop old tables
- [ ] **1.5** Update `RecordPageView` to write to `page_view_hours` (truncate `time.Now().Unix()` to hour: `now - now%3600`)
- [ ] **1.6** Update `RecordFailedSearch` to write to `search_hours` with `had_results=0`
- [ ] **1.7** Record successful searches too (`had_results=1`) for search success rate

## Phase 2: Backend — Write coalescing & dedup

- [ ] **2.1** Create `internal/analytics/writer.go` with a buffered channel (`chan ViewEvent`, capacity 1024)
- [ ] **2.2** Consumer goroutine: flush buffer to SQLite every 5 seconds via batch `INSERT ... ON CONFLICT DO UPDATE SET count = count + ?`
- [ ] **2.3** In-memory dedup map: `map[uint64]int64` keyed by `fnv(actor+path)` → last-seen unix. Skip if within 15 minutes.
- [ ] **2.4** Expose `Writer.Record(ctx, path, source, actorHash)` — non-blocking (drop if channel full)
- [ ] **2.5** Wire into `handlers_file.go` replacing direct `RecordPageView` call
- [ ] **2.6** Wire into `handlers_search.go` for both success and failure recording
- [ ] **2.7** Graceful shutdown: flush remaining buffer on `SIGTERM`

## Phase 3: Backend — Query layer & trends

- [ ] **3.1** `PageViewsInRange(ctx, pathPrefix, since, until int64) ([]PageViewStat, error)` — sum counts from `page_view_hours` within range
- [ ] **3.2** `PageViewTimeSeries(ctx, path, since, until, bucketSize int64) ([]TimePoint, error)` — returns `{timestamp, count}` per bucket
- [ ] **3.3** `SearchSuccessRate(ctx, since, until int64) (float64, error)` — `SUM(CASE had_results WHEN 1 THEN count END) / SUM(count)`
- [ ] **3.4** `TrendingPages(ctx, periodDays int) ([]TrendStat, error)` — compare current period vs previous period, compute `(current - previous) / max(previous, 1)`
- [ ] **3.5** `DecliningPages(ctx, periodDays int) ([]TrendStat, error)` — same but negative trend + zero views in current period
- [ ] **3.6** `ContentGaps(ctx, limit int) ([]FailedSearchStat, error)` — failed searches not yet dismissed

## Phase 4: Backend — Data retention rollup

- [ ] **4.1** Add `page_view_days` table: `(path TEXT, source TEXT, day INTEGER, count INTEGER, unique_actors INTEGER, PRIMARY KEY(path, source, day))`
- [ ] **4.2** Add rollup function: collapse hourly rows older than 90 days into daily, delete hourly originals
- [ ] **4.3** Schedule rollup via existing `janitor.Scheduler` — run once per day at low-traffic hour
- [ ] **4.4** Analytics queries: UNION `page_view_hours` (recent) with `page_view_days` (old) transparently

## Phase 5: Backend — New API endpoints

- [ ] **5.1** `GET /api/kiwi/analytics/overview?period=7d` — returns: total_views, total_searches, search_success_rate, unique_pages_viewed, each with delta_percent vs previous period
- [ ] **5.2** `GET /api/kiwi/analytics/views?period=30d&path=&source=` — time-series array + top pages
- [ ] **5.3** `GET /api/kiwi/analytics/searches?period=30d` — time-series + success rate + top failed
- [ ] **5.4** `GET /api/kiwi/analytics/trends?period=7d` — trending up + declining pages
- [ ] **5.5** `GET /api/kiwi/analytics/content-gaps` — actionable failed searches (with dismiss support)
- [ ] **5.6** `POST /api/kiwi/analytics/content-gaps/:id/dismiss` — mark a failed search as noise
- [ ] **5.7** Keep old `/api/kiwi/analytics` working (backward compat) — populate `engagement` from new tables

## Phase 6: Frontend — Dashboard component

- [ ] **6.1** Replace `KiwiEngagement.tsx` with `KiwiAnalytics.tsx` (new component)
- [ ] **6.2** Period selector dropdown: 7d / 30d / 90d (stored in local state, passed to all API calls)
- [ ] **6.3** Summary cards row: Views (with delta %), Searches (with delta %), Search success rate (with delta pp)
- [ ] **6.4** Sparkline component: tiny inline SVG `<polyline>` from time-series data (no dependency needed — ~30 lines)
- [ ] **6.5** Trending pages section: list with ↑/↓/→ indicators and percentage change
- [ ] **6.6** Content gaps section: failed searches with [Create page] and [Dismiss] action buttons
- [ ] **6.7** Source breakdown: horizontal stacked bar (ui / api / mcp / s3 / webdav)
- [ ] **6.8** "Least viewed" section: pages with 0 views in selected period + last edit date

## Phase 7: Frontend — Per-page inline analytics

- [ ] **7.1** In `KiwiPage.tsx`: replace raw view count with "X views this week (↑Y%)" using `/analytics/views?path=&period=7d`
- [ ] **7.2** Expandable detail card: 30-day sparkline, source breakdown, first/last viewed
- [ ] **7.3** Graceful fallback: if endpoint returns 404/501 (old server), hide analytics section silently

## Phase 8: Robustness & edge cases

- [ ] **8.1** All frontend analytics access uses `?.` / `??` with fallback defaults (never crash on missing fields)
- [ ] **8.2** Backend: empty-state responses always return `[]` not `null` for arrays
- [ ] **8.3** API version detection: if response missing expected fields, show "upgrade KiwiFS" hint instead of crashing
- [ ] **8.4** Rate-limit view recording: max 1 count per (actor, path) per 15min window
- [ ] **8.5** Bot filtering: skip recording for requests with known bot User-Agents
- [ ] **8.6** Tests: `TestAnalyticsTimeSeries`, `TestTrendComputation`, `TestRollup`, `TestWriteCoalescing`
- [ ] **8.7** Benchmark: ensure analytics queries complete in <50ms on 100k-row tables

---

## Execution order

Start with **Phase 1 + 2** (storage + writes) → then **Phase 3 + 5** (queries + endpoints) → then **Phase 6 + 7** (UI) → then **Phase 4 + 8** (retention + hardening).

Phase 4 and 8 can be deferred — they're important but the feature is usable without them.
