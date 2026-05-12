package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConnectionMeta stores non-secret metadata about a saved import connection.
// Credentials are NEVER stored — only connection parameters and last-run info.
type ConnectionMeta struct {
	ID         string          `json:"id"`
	From       string          `json:"from"`
	Name       string          `json:"name"`
	Project    string          `json:"project,omitempty"`
	Table      string          `json:"table,omitempty"`
	Collection string          `json:"collection,omitempty"`
	Database   string          `json:"database,omitempty"`
	DatabaseID string          `json:"database_id,omitempty"`
	BaseID     string          `json:"base_id,omitempty"`
	TableID    string          `json:"table_id,omitempty"`
	DSN        string          `json:"dsn,omitempty"`
	URI        string          `json:"uri,omitempty"`
	Prefix     string          `json:"prefix"`
	IDColumn   string          `json:"id_column,omitempty"`
	Columns    []string        `json:"columns,omitempty"`
	LastRun    string          `json:"last_run,omitempty"`
	LastStats  *ConnectionStats `json:"last_stats,omitempty"`
	CreatedAt  string          `json:"created_at"`
}

// ConnectionStats records the result of the most recent import run.
type ConnectionStats struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

// ConnectionStore manages persistence of import connection metadata
// in .kiwi/state/imports.json. Thread-safe via RWMutex.
type ConnectionStore struct {
	path  string
	mu    sync.RWMutex
	conns map[string]*ConnectionMeta
}

// NewConnectionStore creates a store rooted at the given kiwifs root directory.
// Creates .kiwi/state/ if it does not exist.
func NewConnectionStore(root string) (*ConnectionStore, error) {
	dir := filepath.Join(root, ".kiwi", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("connections: mkdir: %w", err)
	}
	s := &ConnectionStore{
		path:  filepath.Join(dir, "imports.json"),
		conns: make(map[string]*ConnectionMeta),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("connections: load: %w", err)
	}
	return s, nil
}

func (s *ConnectionStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var entries []*ConnectionMeta
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	for _, c := range entries {
		s.conns[c.ID] = c
	}
	return nil
}

func (s *ConnectionStore) save() error {
	all := make([]*ConnectionMeta, 0, len(s.conns))
	for _, c := range s.conns {
		all = append(all, c)
	}
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// List returns all saved connections.
func (s *ConnectionStore) List() []*ConnectionMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*ConnectionMeta, 0, len(s.conns))
	for _, c := range s.conns {
		out = append(out, c)
	}
	return out
}

// Get returns a connection by ID.
func (s *ConnectionStore) Get(id string) (*ConnectionMeta, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.conns[id]
	return c, ok
}

// Save persists a new or updated connection. Generates ID if empty.
func (s *ConnectionStore) Save(c *ConnectionMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.CreatedAt == "" {
		c.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	s.conns[c.ID] = c
	return s.save()
}

// Delete removes a connection by ID.
func (s *ConnectionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.conns[id]; !ok {
		return fmt.Errorf("connection %q not found", id)
	}
	delete(s.conns, id)
	return s.save()
}

// UpdateLastRun updates the last_run timestamp and stats for a connection.
func (s *ConnectionStore) UpdateLastRun(id string, stats *ConnectionStats) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.conns[id]
	if !ok {
		return fmt.Errorf("connection %q not found", id)
	}
	c.LastRun = time.Now().UTC().Format(time.RFC3339)
	c.LastStats = stats
	return s.save()
}
