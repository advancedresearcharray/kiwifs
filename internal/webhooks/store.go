package webhooks

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Webhook struct {
	ID         string   `json:"id"`
	URL        string   `json:"url"`
	PathGlob   string   `json:"path_glob"`
	Secret     string   `json:"secret,omitempty"`
	CreatedAt  string   `json:"created_at"`
	Enabled    bool     `json:"enabled"`
	EventTypes []string `json:"event_types,omitempty"`
}

type Delivery struct {
	ID         string `json:"id"`
	WebhookID  string `json:"webhook_id"`
	EventType  string `json:"event_type"`
	Path       string `json:"path"`
	Attempt    int    `json:"attempt"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) (*Store, error) {
	s := &Store{db: db}
	if err := s.createSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) createSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS webhooks (
			id           TEXT PRIMARY KEY,
			url          TEXT NOT NULL,
			path_glob    TEXT NOT NULL,
			secret       TEXT NOT NULL,
			created_at   TEXT NOT NULL,
			enabled      INTEGER NOT NULL DEFAULT 1,
			event_types  TEXT
		)
	`)
	if err != nil {
		return err
	}
	s.db.Exec(`ALTER TABLE webhooks ADD COLUMN event_types TEXT`)
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS webhook_deliveries (
			id          TEXT PRIMARY KEY,
			webhook_id  TEXT NOT NULL,
			event_type  TEXT NOT NULL,
			path        TEXT NOT NULL,
			attempt     INTEGER NOT NULL,
			status      TEXT NOT NULL,
			status_code INTEGER NOT NULL DEFAULT 0,
			error       TEXT NOT NULL DEFAULT '',
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_created ON webhook_deliveries(webhook_id, created_at DESC)`)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) Register(ctx context.Context, url, pathGlob string, eventTypes ...string) (*Webhook, error) {
	id := generateID()
	secret := generateSecret()

	wh := &Webhook{
		ID:         id,
		URL:        url,
		PathGlob:   pathGlob,
		Secret:     secret,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Enabled:    true,
		EventTypes: eventTypes,
	}

	var etJSON *string
	if len(eventTypes) > 0 {
		data, _ := json.Marshal(eventTypes)
		s := string(data)
		etJSON = &s
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO webhooks(id, url, path_glob, secret, created_at, enabled, event_types) VALUES (?, ?, ?, ?, ?, 1, ?)`,
		wh.ID, wh.URL, wh.PathGlob, wh.Secret, wh.CreatedAt, etJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("register webhook: %w", err)
	}
	return wh, nil
}

func (s *Store) List(ctx context.Context) ([]Webhook, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, url, path_glob, created_at, enabled, event_types FROM webhooks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Webhook
	for rows.Next() {
		var wh Webhook
		var enabled int
		var etJSON sql.NullString
		if err := rows.Scan(&wh.ID, &wh.URL, &wh.PathGlob, &wh.CreatedAt, &enabled, &etJSON); err != nil {
			return nil, err
		}
		wh.Enabled = enabled == 1
		if etJSON.Valid && etJSON.String != "" {
			_ = json.Unmarshal([]byte(etJSON.String), &wh.EventTypes)
		}
		out = append(out, wh)
	}
	if out == nil {
		out = []Webhook{}
	}
	return out, rows.Err()
}

func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM webhooks WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("webhook not found: %s", id)
	}
	return nil
}

func (s *Store) FindMatching(ctx context.Context, path string, eventType ...string) ([]Webhook, error) {
	all, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	et := ""
	if len(eventType) > 0 {
		et = eventType[0]
	}
	var matched []Webhook
	for _, wh := range all {
		if !wh.Enabled {
			continue
		}
		if !matchGlob(wh.PathGlob, path) {
			continue
		}
		if et != "" && len(wh.EventTypes) > 0 {
			found := false
			for _, t := range wh.EventTypes {
				if t == et {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		matched = append(matched, wh)
	}
	return matched, nil
}

func (s *Store) GetSecret(ctx context.Context, id string) (string, error) {
	var secret string
	err := s.db.QueryRowContext(ctx, `SELECT secret FROM webhooks WHERE id = ?`, id).Scan(&secret)
	return secret, err
}

func (s *Store) RecordDelivery(ctx context.Context, d Delivery) (*Delivery, error) {
	if d.ID == "" {
		d.ID = generateDeliveryID()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if d.CreatedAt == "" {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO webhook_deliveries(id, webhook_id, event_type, path, attempt, status, status_code, error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.WebhookID, d.EventType, d.Path, d.Attempt, d.Status, d.StatusCode, d.Error, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("record webhook delivery: %w", err)
	}
	return &d, nil
}

func (s *Store) UpdateDelivery(ctx context.Context, id, status string, statusCode int, errMsg string) error {
	updatedAt := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		UPDATE webhook_deliveries
		SET status = ?, status_code = ?, error = ?, updated_at = ?
		WHERE id = ?
	`, status, statusCode, errMsg, updatedAt, id)
	if err != nil {
		return fmt.Errorf("update webhook delivery: %w", err)
	}
	return nil
}

func (s *Store) ListDeliveries(ctx context.Context, webhookID string, limit int) ([]Delivery, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, webhook_id, event_type, path, attempt, status, status_code, error, created_at, updated_at
		FROM webhook_deliveries
		WHERE webhook_id = ?
		ORDER BY created_at DESC, attempt DESC
		LIMIT ?
	`, webhookID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Delivery
	for rows.Next() {
		var d Delivery
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.EventType, &d.Path, &d.Attempt, &d.Status, &d.StatusCode, &d.Error, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if out == nil {
		out = []Delivery{}
	}
	return out, rows.Err()
}

// RegisterWithSecret registers a webhook with a caller-provided secret.
// Used by bootstrap to auto-register config-driven [[webhook_entries]].
func (s *Store) RegisterWithSecret(ctx context.Context, url, pathGlob, secret string, eventTypes ...string) (*Webhook, error) {
	id := generateID()

	wh := &Webhook{
		ID:         id,
		URL:        url,
		PathGlob:   pathGlob,
		Secret:     secret,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Enabled:    true,
		EventTypes: eventTypes,
	}

	var etJSON *string
	if len(eventTypes) > 0 {
		data, _ := json.Marshal(eventTypes)
		s := string(data)
		etJSON = &s
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO webhooks(id, url, path_glob, secret, created_at, enabled, event_types) VALUES (?, ?, ?, ?, ?, 1, ?)`,
		wh.ID, wh.URL, wh.PathGlob, wh.Secret, wh.CreatedAt, etJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("register webhook: %w", err)
	}
	return wh, nil
}

// FindByURL checks if a webhook with the given URL already exists.
func (s *Store) FindByURL(ctx context.Context, url string) (*Webhook, error) {
	var wh Webhook
	var enabled int
	var etJSON sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, url, path_glob, created_at, enabled, event_types FROM webhooks WHERE url = ?`, url,
	).Scan(&wh.ID, &wh.URL, &wh.PathGlob, &wh.CreatedAt, &enabled, &etJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	wh.Enabled = enabled == 1
	if etJSON.Valid && etJSON.String != "" {
		_ = json.Unmarshal([]byte(etJSON.String), &wh.EventTypes)
	}
	return &wh, nil
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "wh_" + hex.EncodeToString(b)
}

func generateDeliveryID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "wd_" + hex.EncodeToString(b)
}

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "whsec_" + hex.EncodeToString(b)
}
