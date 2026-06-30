package themepresets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/config"
)

func TestValidatePreset(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p, err := ValidatePreset(map[string]any{
			"name":        "Corporate Light",
			"description": "Brand colors",
			"light":       map[string]any{"background": "hsl(0 0% 100%)"},
			"dark":        map[string]any{"background": "hsl(0 0% 5%)"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if p.Name != "Corporate Light" || p.Source != "workspace" {
			t.Fatalf("preset = %+v", p)
		}
	})

	tests := []struct {
		name string
		raw  map[string]any
	}{
		{name: "missing name", raw: map[string]any{"light": map[string]any{"background": "#fff"}}},
		{name: "invalid light", raw: map[string]any{"name": "x", "light": "bad"}},
		{name: "non-string token", raw: map[string]any{
			"name":  "x",
			"light": map[string]any{"background": 123},
		}},
		{name: "empty tokens", raw: map[string]any{"name": "x", "light": map[string]any{}, "dark": map[string]any{}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ValidatePreset(tc.raw); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestLoadFromDir(t *testing.T) {
	root := t.TempDir()
	themesDir := filepath.Join(root, ".kiwi", "themes")
	if err := os.MkdirAll(themesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	valid := `{"name":"corporate-light","description":"Light brand","light":{"background":"#fff"},"dark":{"background":"#111"}}`
	if err := os.WriteFile(filepath.Join(themesDir, "corporate-light.json"), []byte(valid), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themesDir, "broken.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themesDir, "ignored.txt"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := LoadFromDir(root, config.UIThemeConfig{})
	if len(res.Presets) != 1 {
		t.Fatalf("presets = %+v, want 1", res.Presets)
	}
	if res.Presets[0].Name != "corporate-light" {
		t.Fatalf("name = %q", res.Presets[0].Name)
	}
	if len(res.Errors) != 1 {
		t.Fatalf("errors = %+v, want 1", res.Errors)
	}
}

func TestLoadFromDir_CustomPath(t *testing.T) {
	root := t.TempDir()
	customDir := filepath.Join(root, "brand", "themes")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"name":"brand","light":{"primary":"#000"},"dark":{"primary":"#fff"}}`
	if err := os.WriteFile(filepath.Join(customDir, "brand.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	res := LoadFromDir(root, config.UIThemeConfig{PresetsDir: "brand/themes"})
	if len(res.Presets) != 1 || res.Presets[0].Name != "brand" {
		t.Fatalf("presets = %+v", res.Presets)
	}
}

func TestLoadFromDir_RejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside-themes")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(outside) })
	body := `{"name":"leaked","light":{"background":"#000"},"dark":{"background":"#111"}}`
	if err := os.WriteFile(filepath.Join(outside, "leaked.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	res := LoadFromDir(root, config.UIThemeConfig{PresetsDir: "../" + filepath.Base(outside)})
	if len(res.Presets) != 0 {
		t.Fatalf("expected no presets from traversal, got %+v", res.Presets)
	}
}

func TestFilterByAllowed(t *testing.T) {
	presets := []Preset{
		{Name: "Kiwi"},
		{Name: "corporate-light"},
		{Name: "corporate-dark"},
	}
	filtered := FilterByAllowed(presets, []string{"kiwi", "Corporate-Light"})
	if len(filtered) != 2 {
		t.Fatalf("filtered = %+v", filtered)
	}
	if len(FilterByAllowed(presets, nil)) != 3 {
		t.Fatal("nil allowed should return all")
	}
}
