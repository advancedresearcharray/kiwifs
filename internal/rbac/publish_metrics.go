package rbac

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PublishMetrics holds per-page view tracking data for published pages.
type PublishMetrics struct {
	Views       int       `json:"views"`
	FirstViewed time.Time `json:"first_viewed"`
	LastViewed  time.Time `json:"last_viewed"`
}

// PublishMetricsStore manages persistence of published page view counts
// in .kiwi/state/publish-metrics.json, following the same pattern as ShareStore.
type PublishMetricsStore struct {
	path    string
	mu      sync.RWMutex
	metrics map[string]*PublishMetrics // path → metrics
}

// NewPublishMetricsStore loads existing metrics from disk (or starts empty).
func NewPublishMetricsStore(root string) (*PublishMetricsStore, error) {
	dir := filepath.Join(root, ".kiwi", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("publish-metrics: mkdir: %w", err)
	}
	s := &PublishMetricsStore{
		path:    filepath.Join(dir, "publish-metrics.json"),
		metrics: make(map[string]*PublishMetrics),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("publish-metrics: load: %w", err)
	}
	return s, nil
}

// Increment atomically increments the view count for a given path and
// updates first/last viewed timestamps.
func (s *PublishMetricsStore) Increment(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	m, ok := s.metrics[path]
	if !ok {
		m = &PublishMetrics{FirstViewed: now}
		s.metrics[path] = m
	}
	m.Views++
	m.LastViewed = now
	_ = s.save()
}

// Get returns the metrics for a given path. Returns nil if no metrics exist.
func (s *PublishMetricsStore) Get(path string) *PublishMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.metrics[path]
	if !ok {
		return nil
	}
	// Return a copy to avoid data races.
	copy := *m
	return &copy
}

// List returns all metrics indexed by path.
func (s *PublishMetricsStore) List() map[string]*PublishMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]*PublishMetrics, len(s.metrics))
	for k, v := range s.metrics {
		copy := *v
		out[k] = &copy
	}
	return out
}

func (s *PublishMetricsStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.metrics)
}

func (s *PublishMetricsStore) save() error {
	data, err := json.MarshalIndent(s.metrics, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
