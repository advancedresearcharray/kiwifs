// Package analytics provides a buffered, deduplicating writer for page view
// and search events. Events are collected in memory and flushed to SQLite in
// batches every few seconds, reducing write contention on the single-writer
// connection pool.
package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"log"
	"sync"
	"time"
)

// ViewEvent represents a single page-view or search recording request.
type ViewEvent struct {
	Kind      EventKind
	Path      string // page path (for views)
	Source    string // "ui", "api", etc.
	Query     string // search query (for searches)
	SearchType string
	HadResults bool
	ActorHash  string // opaque hash of the requesting actor (IP, API key, etc.)
}

// EventKind discriminates between view and search events.
type EventKind int

const (
	EventPageView EventKind = iota
	EventSearch
)

const (
	defaultFlushInterval = 5 * time.Second
	defaultChanCapacity  = 1024
	defaultDedupWindow   = 15 * time.Minute
)

// Writer buffers analytics events and flushes them to SQLite in batches.
type Writer struct {
	writeDB *sql.DB
	ch      chan ViewEvent
	done    chan struct{}
	wg      sync.WaitGroup

	// In-memory dedup: fnv(actor+path) -> last-seen unix timestamp.
	// Protected by mu. Entries older than dedupWindow are lazily evicted.
	mu          sync.Mutex
	dedup       map[uint64]int64
	dedupWindow time.Duration
}

// NewWriter creates a buffered analytics writer. The writeDB must be the
// single-writer SQLite connection pool (MaxOpenConns=1).
func NewWriter(writeDB *sql.DB) *Writer {
	return &Writer{
		writeDB:     writeDB,
		ch:          make(chan ViewEvent, defaultChanCapacity),
		done:        make(chan struct{}),
		dedup:       make(map[uint64]int64),
		dedupWindow: defaultDedupWindow,
	}
}

// Start launches the background consumer goroutine. Call Stop() to drain.
func (w *Writer) Start() {
	w.wg.Add(1)
	go w.loop()
}

// Stop signals the writer to flush remaining events and exit. Blocks until
// the flush completes or the context expires.
func (w *Writer) Stop() {
	close(w.done)
	w.wg.Wait()
}

// Record enqueues an event for batched writing. Non-blocking: if the channel
// is full the event is silently dropped (backpressure shedding).
func (w *Writer) Record(ctx context.Context, ev ViewEvent) {
	// In-memory dedup for page views: skip if same actor+path within window.
	if ev.Kind == EventPageView && ev.ActorHash != "" {
		if w.isDuplicate(ev.ActorHash, ev.Path) {
			return
		}
	}
	select {
	case w.ch <- ev:
	default:
		// Channel full — drop event. This is intentional backpressure.
	}
}

func (w *Writer) isDuplicate(actor, path string) bool {
	h := fnv.New64a()
	h.Write([]byte(actor))
	h.Write([]byte{0})
	h.Write([]byte(path))
	key := h.Sum64()

	now := time.Now().Unix()
	w.mu.Lock()
	defer w.mu.Unlock()

	if last, ok := w.dedup[key]; ok {
		if now-last < int64(w.dedupWindow.Seconds()) {
			return true
		}
	}
	w.dedup[key] = now
	return false
}

func (w *Writer) loop() {
	defer w.wg.Done()

	ticker := time.NewTicker(defaultFlushInterval)
	defer ticker.Stop()

	var buf []ViewEvent

	for {
		select {
		case ev := <-w.ch:
			buf = append(buf, ev)
			// Batch drain: read as many as available without blocking.
			for {
				select {
				case ev2 := <-w.ch:
					buf = append(buf, ev2)
				default:
					goto drained
				}
			}
		drained:
			if len(buf) >= 100 {
				w.flush(buf)
				buf = buf[:0]
			}
		case <-ticker.C:
			if len(buf) > 0 {
				w.flush(buf)
				buf = buf[:0]
			}
			w.evictStaleDedup()
		case <-w.done:
			// Drain remaining events from channel.
			for {
				select {
				case ev := <-w.ch:
					buf = append(buf, ev)
				default:
					goto shutdown
				}
			}
		shutdown:
			if len(buf) > 0 {
				w.flush(buf)
			}
			return
		}
	}
}

func (w *Writer) flush(events []ViewEvent) {
	if len(events) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := w.writeDB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("analytics: begin flush tx: %v", err)
		return
	}
	defer tx.Rollback()

	pvStmt, err := tx.PrepareContext(ctx, `
INSERT INTO page_view_hours(path, source, hour, count, unique_actors)
VALUES (?, ?, ?, 1, 0)
ON CONFLICT(path, source, hour) DO UPDATE SET count = count + 1`)
	if err != nil {
		log.Printf("analytics: prepare page_view_hours: %v", err)
		return
	}
	defer pvStmt.Close()

	pvLegacy, err := tx.PrepareContext(ctx, `
INSERT INTO page_views(path, source, count, first_seen, last_seen)
VALUES (?, ?, 1, ?, ?)
ON CONFLICT(path, source) DO UPDATE SET
	count = count + 1,
	last_seen = excluded.last_seen`)
	if err != nil {
		log.Printf("analytics: prepare page_views legacy: %v", err)
		return
	}
	defer pvLegacy.Close()

	shStmt, err := tx.PrepareContext(ctx, `
INSERT INTO search_hours(query, search_type, hour, count, had_results)
VALUES (?, ?, ?, 1, ?)
ON CONFLICT(query, search_type, hour) DO UPDATE SET count = count + 1`)
	if err != nil {
		log.Printf("analytics: prepare search_hours: %v", err)
		return
	}
	defer shStmt.Close()

	fsLegacy, err := tx.PrepareContext(ctx, `
INSERT INTO failed_searches(query, search_type, count, first_seen, last_seen)
VALUES (?, ?, 1, ?, ?)
ON CONFLICT(query, search_type) DO UPDATE SET
	count = count + 1,
	last_seen = excluded.last_seen`)
	if err != nil {
		log.Printf("analytics: prepare failed_searches legacy: %v", err)
		return
	}
	defer fsLegacy.Close()

	for _, ev := range events {
		now := time.Now().Unix()
		hour := now - now%3600

		switch ev.Kind {
		case EventPageView:
			if _, err := pvStmt.ExecContext(ctx, ev.Path, ev.Source, hour); err != nil {
				log.Printf("analytics: write page_view_hours: %v", err)
			}
			if _, err := pvLegacy.ExecContext(ctx, ev.Path, ev.Source, now, now); err != nil {
				log.Printf("analytics: write page_views legacy: %v", err)
			}
		case EventSearch:
			hadResults := 0
			if ev.HadResults {
				hadResults = 1
			}
			if _, err := shStmt.ExecContext(ctx, ev.Query, ev.SearchType, hour, hadResults); err != nil {
				log.Printf("analytics: write search_hours: %v", err)
			}
			if !ev.HadResults {
				if _, err := fsLegacy.ExecContext(ctx, ev.Query, ev.SearchType, now, now); err != nil {
					log.Printf("analytics: write failed_searches legacy: %v", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("analytics: commit flush: %v", err)
	}
}

func (w *Writer) evictStaleDedup() {
	cutoff := time.Now().Unix() - int64(w.dedupWindow.Seconds())
	w.mu.Lock()
	defer w.mu.Unlock()
	for k, ts := range w.dedup {
		if ts < cutoff {
			delete(w.dedup, k)
		}
	}
}

// ActorHash computes an opaque FNV hash suitable for dedup from an actor
// identifier (e.g., IP address, API key prefix).
func ActorHash(parts ...string) string {
	h := fnv.New64a()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return fmt.Sprintf("%016x", h.Sum64())
}
