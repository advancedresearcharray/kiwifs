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

func TestGetThemePresets_LoadsWorkspacePresets(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	themesDir := filepath.Join(dir, ".kiwi", "themes")
	if err := os.MkdirAll(themesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	valid := `{"name":"corporate-light","description":"Brand light","light":{"background":"#fff"},"dark":{"background":"#111"}}`
	if err := os.WriteFile(filepath.Join(themesDir, "corporate-light.json"), []byte(valid), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themesDir, "bad.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/theme/presets", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res struct {
		Presets []struct {
			Name  string `json:"name"`
			Light map[string]string `json:"light"`
		} `json:"presets"`
		Errors []struct {
			File  string `json:"file"`
			Error string `json:"error"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Presets) != 1 || res.Presets[0].Name != "corporate-light" {
		t.Fatalf("presets = %+v", res.Presets)
	}
	if len(res.Errors) != 1 {
		t.Fatalf("errors = %+v, want 1", res.Errors)
	}
}

func TestGetThemePresets_RejectsPathTraversalPresetsDir(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	outside := filepath.Join(dir, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	valid := `{"name":"outside-preset","light":{"background":"#fff"},"dark":{"background":"#111"}}`
	if err := os.WriteFile(filepath.Join(outside, "outside-preset.json"), []byte(valid), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.Theme.PresetsDir = "../outside"
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/theme/presets", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res struct {
		Presets []struct {
			Name string `json:"name"`
		} `json:"presets"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Presets) != 0 {
		t.Fatalf("path traversal presets_dir should fall back to default dir, got presets %+v", res.Presets)
	}
}

func TestUIConfig_ThemeAllowedPresets(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.Theme.AllowedPresets = []string{"kiwi", "corporate-dark"}
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/ui-config", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res struct {
		Theme struct {
			AllowedPresets []string `json:"allowedPresets"`
		} `json:"theme"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	want := []string{"kiwi", "corporate-dark"}
	if len(res.Theme.AllowedPresets) != len(want) {
		t.Fatalf("allowedPresets = %+v, want %+v", res.Theme.AllowedPresets, want)
	}
	for i, v := range want {
		if res.Theme.AllowedPresets[i] != v {
			t.Fatalf("allowedPresets[%d] = %q, want %q", i, res.Theme.AllowedPresets[i], v)
		}
	}
}
