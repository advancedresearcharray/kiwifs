package webhooks

import (
	"context"
	"database/sql"
	"testing"

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

func TestStoreRegisterAndList(t *testing.T) {
	db := testDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	wh, err := store.Register(ctx, "https://example.com/hook", "pages/**")
	if err != nil {
		t.Fatal(err)
	}
	if wh.ID == "" || wh.Secret == "" {
		t.Fatal("expected non-empty ID and Secret")
	}

	hooks, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(hooks))
	}
	if hooks[0].URL != "https://example.com/hook" {
		t.Fatalf("unexpected URL: %s", hooks[0].URL)
	}
}

func TestStoreDelete(t *testing.T) {
	db := testDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	wh, _ := store.Register(ctx, "https://example.com/hook", "pages/**")
	if err := store.Delete(ctx, wh.ID); err != nil {
		t.Fatal(err)
	}

	hooks, _ := store.List(ctx)
	if len(hooks) != 0 {
		t.Fatalf("expected 0 webhooks after delete, got %d", len(hooks))
	}
}

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern, path string
		want          bool
	}{
		{"pages/**", "pages/auth.md", true},
		{"pages/**", "episodes/run-1.md", false},
		{"*.md", "test.md", true},
		{"*.md", "test.txt", false},
	}
	for _, tt := range tests {
		got := matchGlob(tt.pattern, tt.path)
		if got != tt.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}
