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

func TestGetCustomCSS_EmptyWhenMissing(t *testing.T) {
	s := buildTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/custom.css", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "" {
		t.Errorf("expected empty body, got %q", rec.Body.String())
	}
}

func TestGetCustomCSS_ReturnsContent(t *testing.T) {
	s, dir := buildTestServerWithRoot(t)

	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	css := ".kiwi-admonition-note { border-color: hotpink; }\n"
	if err := os.WriteFile(filepath.Join(kiwiDir, "custom.css"), []byte(css), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/custom.css", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != css {
		t.Errorf("body = %q, want %q", rec.Body.String(), css)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/css; charset=utf-8", ct)
	}
}

func TestGetCustomCSS_StripsScriptTags(t *testing.T) {
	s, dir := buildTestServerWithRoot(t)

	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := ".foo { color: red; }\n<script>alert(1)</script>\n.bar { color: blue; }\n"
	if err := os.WriteFile(filepath.Join(kiwiDir, "custom.css"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/custom.css", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(strings.ToLower(body), "<script") {
		t.Errorf("script tag not stripped: %q", body)
	}
	if !strings.Contains(body, ".foo { color: red; }") || !strings.Contains(body, ".bar { color: blue; }") {
		t.Errorf("expected CSS rules preserved, got %q", body)
	}
}

func TestGetCustomCSS_ConfigurablePath(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	css := "body { outline: 1px dashed green; }\n"
	if err := os.WriteFile(filepath.Join(kiwiDir, "brand.css"), []byte(css), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.CustomCSS = ".kiwi/brand.css"
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/custom.css", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Body.String() != css {
		t.Errorf("body = %q, want %q", rec.Body.String(), css)
	}
}

func TestSanitizeCustomCSS_CaseInsensitive(t *testing.T) {
	in := "a{}<SCRIPT>x</SCRIPT>b{}"
	out := sanitizeCustomCSS(in)
	if strings.Contains(strings.ToLower(out), "script") {
		t.Fatalf("sanitizeCustomCSS(%q) = %q", in, out)
	}
	if out != "a{}b{}" {
		t.Fatalf("sanitizeCustomCSS(%q) = %q, want a{}b{}", in, out)
	}
}
