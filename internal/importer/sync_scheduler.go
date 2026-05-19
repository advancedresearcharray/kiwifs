package importer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kiwifs/kiwifs/internal/pipeline"
)

// SyncScheduler runs periodic import syncs for connections with sync_enabled=true.
type SyncScheduler struct {
	store    *ConnectionStore
	pipe     *pipeline.Pipeline
	buildSrc func(conn *ConnectionMeta) (Source, error)

	mu      sync.Mutex
	started bool
	stopCh  chan struct{}
}

// NewSyncScheduler creates a scheduler that checks for due syncs every minute.
func NewSyncScheduler(store *ConnectionStore, pipe *pipeline.Pipeline, buildSrc func(*ConnectionMeta) (Source, error)) *SyncScheduler {
	return &SyncScheduler{
		store:    store,
		pipe:     pipe,
		buildSrc: buildSrc,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the sync loop in a goroutine.
func (s *SyncScheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	s.mu.Unlock()

	go s.loop(ctx)
}

// Stop signals the scheduler to exit.
func (s *SyncScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return
	}
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
}

func (s *SyncScheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkAndRunDueSyncs(ctx)
		}
	}
}

func (s *SyncScheduler) checkAndRunDueSyncs(ctx context.Context) {
	conns := s.store.List()
	now := time.Now().UTC()

	for _, conn := range conns {
		if !conn.SyncEnabled || conn.SyncStatus == "running" {
			continue
		}

		if !s.isDue(conn, now) {
			continue
		}

		go s.runSync(ctx, conn)
	}
}

func (s *SyncScheduler) isDue(conn *ConnectionMeta, now time.Time) bool {
	if conn.NextSync == "" {
		return true
	}
	next, err := time.Parse(time.RFC3339, conn.NextSync)
	if err != nil {
		return true
	}
	return now.After(next)
}

func (s *SyncScheduler) runSync(ctx context.Context, conn *ConnectionMeta) {
	_ = s.store.UpdateSyncStatus(conn.ID, "running", "")

	src, err := s.buildSrc(conn)
	if err != nil {
		errMsg := fmt.Sprintf("build source: %v", err)
		log.Printf("sync[%s]: %s", conn.ID, errMsg)
		_ = s.store.UpdateSyncStatus(conn.ID, "error", errMsg)
		s.scheduleNext(conn)
		return
	}
	defer src.Close()

	opts := Options{
		Prefix:   conn.Prefix,
		IDColumn: conn.IDColumn,
		Columns:  conn.Columns,
		Actor:    "sync-scheduler",
	}

	stats, err := Run(ctx, src, s.pipe, opts)
	if err != nil {
		errMsg := fmt.Sprintf("run: %v", err)
		log.Printf("sync[%s]: %s", conn.ID, errMsg)
		_ = s.store.UpdateSyncStatus(conn.ID, "error", errMsg)
		s.scheduleNext(conn)
		return
	}

	connStats := &ConnectionStats{
		Imported: stats.Imported,
		Skipped:  stats.Skipped,
		Errors:   stats.Errors,
	}
	_ = s.store.UpdateLastRun(conn.ID, connStats)
	_ = s.store.UpdateSyncStatus(conn.ID, "idle", "")
	s.scheduleNext(conn)
	log.Printf("sync[%s/%s]: imported=%d skipped=%d", conn.ID, conn.Name, stats.Imported, stats.Skipped)
}

func (s *SyncScheduler) scheduleNext(conn *ConnectionMeta) {
	interval := parseSyncInterval(conn.SyncInterval)
	next := time.Now().UTC().Add(interval).Format(time.RFC3339)
	_ = s.store.UpdateNextSync(conn.ID, next)
}

func parseSyncInterval(interval string) time.Duration {
	if d, err := time.ParseDuration(interval); err == nil && d > 0 {
		return d
	}
	return 1 * time.Hour // default
}
