package webhooks

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func NewStoreFromRoot(root string) (*Store, error) {
	dir := filepath.Join(root, ".kiwi", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("webhooks: mkdir: %w", err)
	}
	dbPath := filepath.Join(dir, "webhooks.db")
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("webhooks: open db: %w", err)
	}
	return NewStore(db)
}
