package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// newTestDB creates an in-memory SQLite DB with the required tables.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Create the tables the writer expects
	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS page_view_hours (
			path TEXT NOT NULL,
			source TEXT NOT NULL DEFAULT 'api',
			hour INTEGER NOT NULL,
			count INTEGER NOT NULL DEFAULT 0,
			unique_actors INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (path, source, hour)
		)`,
		`CREATE TABLE IF NOT EXISTS page_views (
			path TEXT NOT NULL,
			source TEXT NOT NULL DEFAULT 'api',
			count INTEGER NOT NULL DEFAULT 0,
			first_seen INTEGER NOT NULL,
			last_seen INTEGER NOT NULL,
			PRIMARY KEY (path, source)
		)`,
		`CREATE TABLE IF NOT EXISTS search_hours (
			query TEXT NOT NULL,
			search_type TEXT NOT NULL DEFAULT 'search',
			hour INTEGER NOT NULL,
			count INTEGER NOT NULL DEFAULT 0,
			had_results INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (query, search_type, hour)
		)`,
		`CREATE TABLE IF NOT EXISTS failed_searches (
			query TEXT NOT NULL,
			search_type TEXT NOT NULL DEFAULT 'search',
			count INTEGER NOT NULL DEFAULT 0,
			first_seen INTEGER NOT NULL,
			last_seen INTEGER NOT NULL,
			PRIMARY KEY (query, search_type)
		)`,
	} {
		if _, err := db.Exec(ddl); err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	return db
}

func TestWriterFlushPageViews(t *testing.T) {
	db := newTestDB(t)
	w := NewWriter(db)
	w.Start()

	// Record some page views
	for i := 0; i < 5; i++ {
		w.Record(context.Background(), ViewEvent{
			Kind:   EventPageView,
			Path:   "doc.md",
			Source: "ui",
		})
	}

	// Stop triggers flush
	w.Stop()

	// Verify data was written
	var count int
	err := db.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_view_hours WHERE path = 'doc.md'`).Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected 5 page views in page_view_hours, got %d", count)
	}

	// Verify legacy table
	var legacyCount int
	err = db.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_views WHERE path = 'doc.md'`).Scan(&legacyCount)
	if err != nil {
		t.Fatalf("query legacy: %v", err)
	}
	if legacyCount != 5 {
		t.Fatalf("expected 5 page views in page_views, got %d", legacyCount)
	}
}

func TestWriterFlushSearches(t *testing.T) {
	db := newTestDB(t)
	w := NewWriter(db)
	w.Start()

	// Record successful and failed searches
	w.Record(context.Background(), ViewEvent{
		Kind:       EventSearch,
		Query:      "found",
		SearchType: "search",
		HadResults: true,
	})
	w.Record(context.Background(), ViewEvent{
		Kind:       EventSearch,
		Query:      "notfound",
		SearchType: "search",
		HadResults: false,
	})

	w.Stop()

	// Verify search_hours has both
	var total int
	err := db.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM search_hours`).Scan(&total)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 search events, got %d", total)
	}

	// Verify failed_searches only has the failed one
	var failedCount int
	err = db.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM failed_searches`).Scan(&failedCount)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if failedCount != 1 {
		t.Fatalf("expected 1 failed search in legacy table, got %d", failedCount)
	}
}

func TestWriterDedup(t *testing.T) {
	db := newTestDB(t)
	w := NewWriter(db)
	w.Start()

	// Record same actor+path multiple times — should be deduped
	for i := 0; i < 10; i++ {
		w.Record(context.Background(), ViewEvent{
			Kind:      EventPageView,
			Path:      "dedup.md",
			Source:    "ui",
			ActorHash: "same-actor",
		})
	}

	w.Stop()

	// Only 1 should have been recorded (rest deduped within 15min window)
	var count int
	err := db.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_view_hours WHERE path = 'dedup.md'`).Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 page view after dedup, got %d", count)
	}
}

func TestWriterDedup_DifferentActors(t *testing.T) {
	db := newTestDB(t)
	w := NewWriter(db)
	w.Start()

	// Different actors should not be deduped
	for i := 0; i < 5; i++ {
		w.Record(context.Background(), ViewEvent{
			Kind:      EventPageView,
			Path:      "multi.md",
			Source:    "ui",
			ActorHash: fmt.Sprintf("actor-%d", i),
		})
	}

	w.Stop()

	var count int
	err := db.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_view_hours WHERE path = 'multi.md'`).Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected 5 page views from different actors, got %d", count)
	}
}

func TestWriterDedup_NoActorHash(t *testing.T) {
	db := newTestDB(t)
	w := NewWriter(db)
	w.Start()

	// Events without ActorHash should not be deduped
	for i := 0; i < 3; i++ {
		w.Record(context.Background(), ViewEvent{
			Kind:   EventPageView,
			Path:   "nohash.md",
			Source: "api",
		})
	}

	w.Stop()

	var count int
	err := db.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM page_view_hours WHERE path = 'nohash.md'`).Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 page views without dedup, got %d", count)
	}
}

func TestWriterBackpressure(t *testing.T) {
	db := newTestDB(t)
	w := NewWriter(db)
	// Don't start the consumer — channel will fill up

	// Fill the channel
	for i := 0; i < defaultChanCapacity+100; i++ {
		w.Record(context.Background(), ViewEvent{
			Kind:   EventPageView,
			Path:   fmt.Sprintf("page-%d.md", i),
			Source: "ui",
		})
	}

	// Channel should be at capacity; the extra 100 should have been dropped
	if len(w.ch) != defaultChanCapacity {
		t.Fatalf("expected channel at capacity %d, got %d", defaultChanCapacity, len(w.ch))
	}

	// Now start and stop to drain
	w.Start()
	w.Stop()
}

func TestActorHash(t *testing.T) {
	h1 := ActorHash("192.168.1.1")
	h2 := ActorHash("192.168.1.2")
	h3 := ActorHash("192.168.1.1")

	if h1 == h2 {
		t.Fatal("different IPs should produce different hashes")
	}
	if h1 != h3 {
		t.Fatal("same IP should produce same hash")
	}
	if len(h1) != 16 {
		t.Fatalf("expected 16-char hex hash, got %d chars: %s", len(h1), h1)
	}
}

func TestIsDuplicate(t *testing.T) {
	db := newTestDB(t)
	w := NewWriter(db)

	// First call should not be duplicate
	if w.isDuplicate("actor1", "page.md") {
		t.Fatal("first call should not be duplicate")
	}

	// Same actor+path within window should be duplicate
	if !w.isDuplicate("actor1", "page.md") {
		t.Fatal("second call should be duplicate")
	}

	// Different actor should not be duplicate
	if w.isDuplicate("actor2", "page.md") {
		t.Fatal("different actor should not be duplicate")
	}

	// Different path should not be duplicate
	if w.isDuplicate("actor1", "other.md") {
		t.Fatal("different path should not be duplicate")
	}
}

func TestEvictStaleDedup(t *testing.T) {
	db := newTestDB(t)
	w := NewWriter(db)
	w.dedupWindow = 1 * time.Second // Short window for testing

	// Add entries
	w.isDuplicate("actor1", "page.md")
	w.isDuplicate("actor2", "page.md")

	if len(w.dedup) != 2 {
		t.Fatalf("expected 2 dedup entries, got %d", len(w.dedup))
	}

	// Wait for entries to expire. We need 2+ seconds because Unix timestamps
	// are integer seconds and eviction uses strict less-than comparison.
	// Entry at time T, eviction cutoff = now - 1. If now = T+2, cutoff = T+1,
	// and T < T+1 is true → evicted.
	time.Sleep(2100 * time.Millisecond)

	w.evictStaleDedup()

	if len(w.dedup) != 0 {
		t.Fatalf("expected 0 dedup entries after eviction, got %d", len(w.dedup))
	}
}
