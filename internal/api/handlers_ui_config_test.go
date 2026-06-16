package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/config"
)

func TestUIConfig_Defaults(t *testing.T) {
	s := buildTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/ui-config", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res uiConfigResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.ThemeLocked {
		t.Fatal("expected themeLocked false by default")
	}
	if res.Branding.Name != "" || res.Branding.LogoURL != "" {
		t.Fatalf("expected empty branding fields, got %+v", res.Branding)
	}
}

func TestUIConfig_Branding(t *testing.T) {
	s, dir := buildTestServerWithRoot(t)

	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `
[ui.branding]
name = "Acme KB"
logo_url = ".kiwi/assets/logo.png"
favicon_url = ".kiwi/assets/favicon.svg"
welcome_title = "Welcome to Acme"
welcome_message = "Get started."
`
	if err := os.WriteFile(filepath.Join(kiwiDir, "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	s.handlers.ui = cfg.UI

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/ui-config", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res uiConfigResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	b := res.Branding
	if b.Name != "Acme KB" || b.LogoURL != ".kiwi/assets/logo.png" {
		t.Fatalf("unexpected branding: %+v", b)
	}
	if b.WelcomeTitle != "Welcome to Acme" || b.WelcomeMessage != "Get started." {
		t.Fatalf("unexpected welcome copy: %+v", b)
	}
}
