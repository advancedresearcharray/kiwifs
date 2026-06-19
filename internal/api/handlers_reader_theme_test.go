package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/config"
)

func TestPublishedPage_WorkspaceTheme(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	theme := `{"mode":"light","light":{"primary":"hsl(200 80% 45%)","background":"hsl(210 20% 98%)"}}`
	if err := os.WriteFile(filepath.Join(kiwiDir, "theme.json"), []byte(theme), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	pageContent := "---\npublished: true\ntitle: Themed Page\n---\n# Themed Page\n"
	mustPutFile(t, s, "docs/themed.md", pageContent)

	req := httptest.NewRequest(http.MethodGet, "/p/docs/themed.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "--primary: hsl(200 80% 45%)") {
		t.Fatalf("expected workspace theme primary token in HTML, got snippet: %s", body[:min(500, len(body))])
	}
	if !strings.Contains(body, "--background: hsl(210 20% 98%)") {
		t.Fatalf("expected workspace theme background token in HTML")
	}
	if !strings.Contains(body, "var(--background)") {
		t.Fatalf("expected template to use themed CSS variables")
	}
}

func TestPublishedPage_Branding(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.Branding = config.BrandingConfig{
		Name:       "Acme Docs",
		LogoURL:    "brand/logo.png",
		FaviconURL: "brand/favicon.ico",
	}
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	pageContent := "---\npublished: true\ntitle: Getting Started\n---\n# Getting Started\n"
	mustPutFile(t, s, "docs/start.md", pageContent)

	req := httptest.NewRequest(http.MethodGet, "/p/docs/start.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "<title>Acme Docs | Getting Started</title>") {
		t.Fatalf("expected branded document title in HTML")
	}
	if !strings.Contains(body, "<h1>Getting Started</h1>") {
		t.Fatalf("expected page heading without brand prefix in h1")
	}
	if !strings.Contains(body, `href="/raw/brand/favicon.ico"`) {
		t.Fatalf("expected custom favicon link")
	}
	if !strings.Contains(body, `type="image/x-icon"`) {
		t.Fatalf("expected correct favicon MIME type")
	}
	if !strings.Contains(body, `class="brand-mark"`) {
		t.Fatalf("expected custom branding footer")
	}
	if !strings.Contains(body, `src="/raw/brand/logo.png"`) {
		t.Fatalf("expected custom logo in footer")
	}
	if strings.Contains(body, "Published with") {
		t.Fatalf("default KiwiFS footer should be replaced by custom branding")
	}
}

func TestPublishedPage_DefaultThemeFallback(t *testing.T) {
	s := buildTestServer(t)
	pageContent := "---\npublished: true\ntitle: Plain Page\n---\n# Plain Page\n"
	mustPutFile(t, s, "docs/plain.md", pageContent)

	req := httptest.NewRequest(http.MethodGet, "/p/docs/plain.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "--background: #ffffff") {
		t.Fatalf("expected default fallback CSS variables")
	}
	if strings.Contains(body, "--primary: hsl(") {
		t.Fatalf("unexpected workspace theme tokens without theme.json")
	}
	if !strings.Contains(body, "Published with") {
		t.Fatalf("expected default KiwiFS footer without custom branding")
	}
}
