package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	standardwebhooks "github.com/standard-webhooks/standard-webhooks/libraries/go"
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

func TestSignPayloadVerifiesWithStandardWebhooks(t *testing.T) {
	rawKey := []byte("test-secret-key-for-webhooks1234")
	b64Secret := base64.StdEncoding.EncodeToString(rawKey)

	msgID := "msg_abc123"
	now := time.Now().UTC()
	payload := []byte(`{"type":"write","path":"pages/test.md"}`)

	sig, err := signPayload(b64Secret, msgID, now, payload)
	if err != nil {
		t.Fatalf("signPayload: %v", err)
	}

	verifier, err := standardwebhooks.NewWebhookRaw(rawKey)
	if err != nil {
		t.Fatalf("NewWebhookRaw: %v", err)
	}
	headers := http.Header{}
	headers.Set("webhook-id", msgID)
	headers.Set("webhook-timestamp", strconv.FormatInt(now.Unix(), 10))
	headers.Set("webhook-signature", sig)

	if err := verifier.Verify(payload, headers); err != nil {
		t.Fatalf("standard-webhooks verification failed: %v", err)
	}
}

func TestDeliverAddsKiwiSignatureAndRecordsSuccess(t *testing.T) {
	db := testDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	secret := "plain-secret"
	event := Event{Type: "write", Path: "pages/test.md", Actor: "alice", Timestamp: time.Now().UTC().Format(time.RFC3339)}
	var gotSignature string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		gotSignature = r.Header.Get("X-Kiwi-Signature-256")
		if gotSignature == "" {
			t.Errorf("missing X-Kiwi-Signature-256")
		}
		want := expectedBodySignature(secret, body)
		if gotSignature != want {
			t.Errorf("signature = %q, want %q", gotSignature, want)
		}
		if r.Header.Get("webhook-signature") == "" {
			t.Errorf("standard webhook-signature header missing")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	wh, err := store.RegisterWithSecret(context.Background(), ts.URL, "pages/**", secret, "write")
	if err != nil {
		t.Fatal(err)
	}
	d := NewDispatcher(store, Config{MaxWorkers: 1, MaxRetries: 0})
	defer d.Close()
	d.deliver(dispatchJob{webhook: *wh, event: event})

	deliveries, err := store.ListDeliveries(context.Background(), wh.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("deliveries = %d, want 1", len(deliveries))
	}
	if got := deliveries[0]; got.Status != "delivered" || got.StatusCode != http.StatusNoContent || got.Attempt != 1 {
		t.Fatalf("delivery = %+v, want delivered 204 attempt 1", got)
	}
}

func TestDeliverRetriesFailedNon2xxAndRecordsAttempts(t *testing.T) {
	db := testDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	wh, err := store.RegisterWithSecret(context.Background(), ts.URL, "pages/**", "retry-secret", "write")
	if err != nil {
		t.Fatal(err)
	}
	d := NewDispatcher(store, Config{
		MaxWorkers:  1,
		MaxRetries:  1,
		RetryDelays: []time.Duration{time.Millisecond},
	})
	defer d.Close()

	d.deliver(dispatchJob{
		webhook: *wh,
		event:   Event{Type: "write", Path: "pages/retry.md", Actor: "alice", Timestamp: time.Now().UTC().Format(time.RFC3339)},
	})
	deadline := time.Now().Add(time.Second)
	for attempts.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d, want 2", attempts.Load())
	}

	deliveries, err := store.ListDeliveries(context.Background(), wh.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 2 {
		t.Fatalf("deliveries = %d, want 2: %+v", len(deliveries), deliveries)
	}
	var failed, delivered bool
	for _, delivery := range deliveries {
		if delivery.Status == "failed" && delivery.StatusCode == http.StatusInternalServerError && delivery.Attempt == 1 {
			failed = true
		}
		if delivery.Status == "delivered" && delivery.StatusCode == http.StatusNoContent && delivery.Attempt == 2 {
			delivered = true
		}
	}
	if !failed || !delivered {
		t.Fatalf("missing failed/delivered attempts: %+v", deliveries)
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

func expectedBodySignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
