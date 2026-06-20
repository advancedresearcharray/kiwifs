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

func TestPublishedPage_DarkModeTheme(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	theme := `{"mode":"dark","dark":{"background":"hsl(0 0% 5%)","foreground":"hsl(0 0% 95%)"}}`
	if err := os.WriteFile(filepath.Join(kiwiDir, "theme.json"), []byte(theme), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	pageContent := "---\npublished: true\ntitle: Dark Page\n---\n# Dark Page\n"
	mustPutFile(t, s, "docs/dark.md", pageContent)

	req := httptest.NewRequest(http.MethodGet, "/p/docs/dark.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "--background: hsl(0 0% 5%)") {
		t.Fatalf("expected dark mode background token in HTML")
	}
	if !strings.Contains(body, "--foreground: hsl(0 0% 95%)") {
		t.Fatalf("expected dark mode foreground token in HTML")
	}
}

func TestPublishedPage_SystemModeTheme(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	theme := `{"mode":"system","light":{"background":"#fff"},"dark":{"background":"#111"}}`
	if err := os.WriteFile(filepath.Join(kiwiDir, "theme.json"), []byte(theme), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	pageContent := "---\npublished: true\ntitle: System Page\n---\n# System Page\n"
	mustPutFile(t, s, "docs/system.md", pageContent)

	req := httptest.NewRequest(http.MethodGet, "/p/docs/system.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "--background: #fff") {
		t.Fatalf("expected light background token in HTML")
	}
	if !strings.Contains(body, "@media (prefers-color-scheme: dark)") {
		t.Fatalf("system mode should include dark media query")
	}
	if !strings.Contains(body, "--background: #111") {
		t.Fatalf("expected dark background token in media query block")
	}
}

func TestPublishedPage_InvalidThemeJSONFallback(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kiwiDir, "theme.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	pageContent := "---\npublished: true\ntitle: Broken Theme\n---\n# Broken Theme\n"
	mustPutFile(t, s, "docs/broken-theme.md", pageContent)

	req := httptest.NewRequest(http.MethodGet, "/p/docs/broken-theme.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "--background: #ffffff") {
		t.Fatalf("invalid theme.json should fall back to default CSS variables")
	}
}

func TestPublishedPage_ThemeOnlyInHTMLResponse(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	theme := `{"mode":"light","light":{"primary":"hsl(200 80% 45%)"}}`
	if err := os.WriteFile(filepath.Join(kiwiDir, "theme.json"), []byte(theme), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	pageContent := "---\npublished: true\ntitle: Negotiated\n---\n# Negotiated\n"
	mustPutFile(t, s, "docs/negotiated.md", pageContent)

	mdReq := httptest.NewRequest(http.MethodGet, "/p/docs/negotiated.md", nil)
	mdReq.Header.Set("Accept", "text/markdown")
	mdRec := httptest.NewRecorder()
	s.echo.ServeHTTP(mdRec, mdReq)
	if mdRec.Code != http.StatusOK {
		t.Fatalf("markdown GET: %d", mdRec.Code)
	}
	if strings.Contains(mdRec.Body.String(), "--primary:") {
		t.Fatalf("markdown response must not include injected theme CSS")
	}

	jsonReq := httptest.NewRequest(http.MethodGet, "/p/docs/negotiated.md", nil)
	jsonReq.Header.Set("Accept", "application/json")
	jsonRec := httptest.NewRecorder()
	s.echo.ServeHTTP(jsonRec, jsonReq)
	if jsonRec.Code != http.StatusOK {
		t.Fatalf("json GET: %d", jsonRec.Code)
	}
	if strings.Contains(jsonRec.Body.String(), "--primary:") {
		t.Fatalf("json response must not include injected theme CSS")
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
