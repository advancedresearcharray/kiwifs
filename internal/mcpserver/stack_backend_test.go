package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/bootstrap"
	"github.com/kiwifs/kiwifs/internal/config"
)

func TestNewStackBackendDoesNotCloseSharedStack(t *testing.T) {
	dir := t.TempDir()
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kiwiDir, "config.toml"), []byte(`
[search]
engine = "grep"
[versioning]
strategy = "none"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "page.md"), []byte("# Page\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Search:     config.SearchConfig{Engine: "grep"},
		Versioning: config.VersioningConfig{Strategy: "none"},
	}
	stack, err := bootstrap.Build("default", dir, cfg)
	if err != nil {
		t.Fatalf("bootstrap.Build: %v", err)
	}
	defer stack.Close()

	backend := NewStackBackend(stack)
	content, _, err := backend.ReadFile(context.Background(), "page.md")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if content == "" {
		t.Fatal("expected content from shared stack backend")
	}
	if err := backend.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, _, err := backend.ReadFile(context.Background(), "page.md"); err != nil {
		t.Fatalf("stack should remain usable after backend Close: %v", err)
	}
}

func TestAuthTokenFromConfig(t *testing.T) {
	if got := AuthTokenFromConfig(nil); got != "" {
		t.Fatalf("nil cfg = %q, want empty", got)
	}
	got := AuthTokenFromConfig(&config.Config{Auth: config.AuthConfig{Type: "apikey", APIKey: "k"}})
	if got != "k" {
		t.Fatalf("token = %q, want k", got)
	}
	if got := AuthTokenFromConfig(&config.Config{Auth: config.AuthConfig{Type: "none"}}); got != "" {
		t.Fatalf("none auth = %q, want empty", got)
	}
}
