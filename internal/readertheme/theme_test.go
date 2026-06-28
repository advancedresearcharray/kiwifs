package readertheme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/config"
)

func TestBuildCSS_LightMode(t *testing.T) {
	theme := map[string]any{
		"mode": "light",
		"light": map[string]any{
			"background": "hsl(0 0% 100%)",
			"primary":    "hsl(65 80% 55%)",
		},
	}
	css := BuildCSS(theme)
	if !strings.Contains(css, "--background: hsl(0 0% 100%)") {
		t.Fatalf("missing background token: %s", css)
	}
	if !strings.Contains(css, "--primary: hsl(65 80% 55%)") {
		t.Fatalf("missing primary token: %s", css)
	}
	if strings.Contains(css, "@media") {
		t.Fatalf("light mode should not include dark media query: %s", css)
	}
}

func TestBuildCSS_DarkMode(t *testing.T) {
	theme := map[string]any{
		"mode": "dark",
		"dark": map[string]any{
			"background": "hsl(0 0% 5%)",
			"foreground": "hsl(0 0% 95%)",
		},
	}
	css := BuildCSS(theme)
	if !strings.Contains(css, "--background: hsl(0 0% 5%)") {
		t.Fatalf("missing dark background: %s", css)
	}
	if strings.Contains(css, "@media") {
		t.Fatalf("forced dark mode should not use media query: %s", css)
	}
}

func TestBuildCSS_SystemMode(t *testing.T) {
	theme := map[string]any{
		"mode": "system",
		"light": map[string]any{
			"background": "#fff",
		},
		"dark": map[string]any{
			"background": "#111",
		},
	}
	css := BuildCSS(theme)
	if !strings.Contains(css, "--background: #fff") {
		t.Fatalf("missing light background: %s", css)
	}
	if !strings.Contains(css, "@media (prefers-color-scheme: dark)") {
		t.Fatalf("system mode should include dark media query: %s", css)
	}
	if !strings.Contains(css, "--background: #111") {
		t.Fatalf("missing dark background: %s", css)
	}
}

func TestBuildCSS_DefaultsToLight(t *testing.T) {
	theme := map[string]any{
		"light": map[string]any{
			"accent": "hsl(120 50% 50%)",
		},
	}
	css := BuildCSS(theme)
	if strings.Contains(css, "@media") {
		t.Fatalf("default mode should be light without media query: %s", css)
	}
}

func TestBuildCSS_EmptyTheme(t *testing.T) {
	if css := BuildCSS(nil); css != "" {
		t.Fatalf("nil theme should produce empty CSS, got %q", css)
	}
	if css := BuildCSS(map[string]any{"mode": "light"}); css != "" {
		t.Fatalf("theme without tokens should produce empty CSS, got %q", css)
	}
}

func TestBuildCSS_RejectsUnsafeValues(t *testing.T) {
	theme := map[string]any{
		"light": map[string]any{
			"background": "#fff; }</style><script>alert(1)</script><style>",
			"primary":    "hsl(65 80% 55%)",
		},
	}
	css := BuildCSS(theme)
	if strings.Contains(css, "script") {
		t.Fatalf("unsafe value should be filtered: %s", css)
	}
	if !strings.Contains(css, "--primary:") {
		t.Fatalf("safe token should remain: %s", css)
	}
}

func TestBuildCSS_RejectsUnsafeKeys(t *testing.T) {
	theme := map[string]any{
		"light": map[string]any{
			"primary": "hsl(65 80% 55%)",
			"foo\n}\n</style><script>alert(1)</script><style>": "red",
			"123bad": "blue",
		},
	}
	css := BuildCSS(theme)
	if strings.Contains(css, "script") {
		t.Fatalf("unsafe key should be filtered: %s", css)
	}
	if strings.Contains(css, "123bad") {
		t.Fatalf("key must start with a letter: %s", css)
	}
	if !strings.Contains(css, "--primary:") {
		t.Fatalf("safe token should remain: %s", css)
	}
}

func TestBrandingFromConfig_TitlePrefix(t *testing.T) {
	ctx := BrandingFromConfig(config.BrandingConfig{
		Name:       "Acme Docs",
		LogoURL:    "brand/logo.png",
		FaviconURL: "brand/favicon.ico",
	}, "Getting Started")
	if ctx.DocumentTitle != "Acme Docs | Getting Started" {
		t.Fatalf("document title = %q", ctx.DocumentTitle)
	}
	if ctx.PageTitle != "Getting Started" {
		t.Fatalf("page title = %q, want page title without brand prefix", ctx.PageTitle)
	}
	if !ctx.HasCustomBranding {
		t.Fatal("expected custom branding")
	}
	if ctx.FaviconURL != "/raw/brand/favicon.ico" {
		t.Fatalf("favicon = %q", ctx.FaviconURL)
	}
	if ctx.FaviconType != "image/x-icon" {
		t.Fatalf("favicon type = %q", ctx.FaviconType)
	}
	if ctx.LogoURL != "/raw/brand/logo.png" {
		t.Fatalf("logo = %q", ctx.LogoURL)
	}
}

func TestBrandingFromConfig_Defaults(t *testing.T) {
	ctx := BrandingFromConfig(config.BrandingConfig{}, "Hello")
	if ctx.DocumentTitle != "Hello" {
		t.Fatalf("document title = %q", ctx.DocumentTitle)
	}
	if ctx.PageTitle != "Hello" {
		t.Fatalf("page title = %q", ctx.PageTitle)
	}
	if ctx.HasCustomBranding {
		t.Fatal("default branding should not be custom")
	}
	if ctx.FaviconURL != config.DefaultBrandingFaviconURL {
		t.Fatalf("favicon = %q", ctx.FaviconURL)
	}
	if ctx.FaviconType != "image/svg+xml" {
		t.Fatalf("favicon type = %q", ctx.FaviconType)
	}
}

func TestCache_Get(t *testing.T) {
	dir := t.TempDir()
	kiwi := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwi, 0o755); err != nil {
		t.Fatal(err)
	}
	themePath := filepath.Join(kiwi, "theme.json")
	body := `{"mode":"light","light":{"primary":"hsl(1 2% 3%)"}}`
	if err := os.WriteFile(themePath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var cache Cache
	first := cache.Get(dir)
	if first == nil {
		t.Fatal("expected theme")
	}
	if first["mode"] != "light" {
		t.Fatalf("mode = %v", first["mode"])
	}

	second := cache.Get(dir)
	if second == nil || second["mode"] != "light" {
		t.Fatal("cache should return same theme")
	}

	if err := os.WriteFile(themePath, []byte(`{"mode":"dark"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	updated := cache.Get(dir)
	if updated == nil || updated["mode"] != "dark" {
		t.Fatalf("cache should reload after mtime change, got %v", updated)
	}

	missing := cache.Get(filepath.Join(dir, "missing"))
	if missing != nil {
		t.Fatalf("missing theme should be nil, got %v", missing)
	}
}

func TestApplyTheme(t *testing.T) {
	ctx := PageContext{PageTitle: "Page", DocumentTitle: "Page"}
	theme := map[string]any{
		"light": map[string]any{"accent": "red"},
	}
	out := ApplyTheme(ctx, theme)
	if out.ThemeCSS == "" {
		t.Fatal("expected theme CSS")
	}
	if !strings.Contains(out.ThemeCSS, "--accent: red") {
		t.Fatalf("css = %q", out.ThemeCSS)
	}
}
