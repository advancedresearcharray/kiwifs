package claims

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestClaimAndRelease(t *testing.T) {
	db := testDB(t)
	s, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	c, err := s.Claim(ctx, "tasks/fix-bug.md", "agent-A", 30*time.Minute)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if c.ClaimedBy != "agent-A" {
		t.Fatalf("expected agent-A, got %s", c.ClaimedBy)
	}

	_, err = s.Claim(ctx, "tasks/fix-bug.md", "agent-B", 30*time.Minute)
	if err != ErrAlreadyClaimed {
		t.Fatalf("expected ErrAlreadyClaimed, got %v", err)
	}

	// Same agent can renew
	c2, err := s.Claim(ctx, "tasks/fix-bug.md", "agent-A", 1*time.Hour)
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	if c2.ClaimedBy != "agent-A" {
		t.Fatalf("expected agent-A on renew, got %s", c2.ClaimedBy)
	}

	if err := s.Release(ctx, "tasks/fix-bug.md", "agent-A"); err != nil {
		t.Fatalf("release: %v", err)
	}

	// Now agent-B can claim
	c3, err := s.Claim(ctx, "tasks/fix-bug.md", "agent-B", 30*time.Minute)
	if err != nil {
		t.Fatalf("claim after release: %v", err)
	}
	if c3.ClaimedBy != "agent-B" {
		t.Fatalf("expected agent-B, got %s", c3.ClaimedBy)
	}
}

func TestReleaseWrongHolder(t *testing.T) {
	db := testDB(t)
	s, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_, _ = s.Claim(ctx, "tasks/t.md", "agent-A", 30*time.Minute)
	if err := s.Release(ctx, "tasks/t.md", "agent-B"); err != ErrNotHolder {
		t.Fatalf("expected ErrNotHolder, got %v", err)
	}
}

func TestListActive(t *testing.T) {
	db := testDB(t)
	s, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_, _ = s.Claim(ctx, "tasks/a.md", "agent-A", 30*time.Minute)
	_, _ = s.Claim(ctx, "tasks/b.md", "agent-B", 30*time.Minute)

	active, err := s.ListActive(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("expected 2 active claims, got %d", len(active))
	}
}

func TestExpireStale(t *testing.T) {
	db := testDB(t)
	s, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// Claim with very short lease (must cross a second boundary for RFC3339)
	_, _ = s.Claim(ctx, "tasks/expired.md", "agent-A", 1*time.Second)
	time.Sleep(2 * time.Second)

	n, err := s.ExpireStale(ctx)
	if err != nil {
		t.Fatalf("expire: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 expired, got %d", n)
	}

	// agent-B can now claim
	c, err := s.Claim(ctx, "tasks/expired.md", "agent-B", 30*time.Minute)
	if err != nil {
		t.Fatalf("claim after expiry: %v", err)
	}
	if c.ClaimedBy != "agent-B" {
		t.Fatalf("expected agent-B, got %s", c.ClaimedBy)
	}
}
