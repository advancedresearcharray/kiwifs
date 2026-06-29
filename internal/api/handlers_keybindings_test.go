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

func TestGetKeybindings_DefaultsWhenMissing(t *testing.T) {
	s := buildTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/keybindings", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res struct {
		Bindings map[string]string `json:"bindings"`
		Conflicts []struct {
			Chord string `json:"chord"`
		} `json:"conflicts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Bindings["search"] != "mod+k" {
		t.Fatalf("search = %q, want mod+k", res.Bindings["search"])
	}
	if len(res.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %+v", res.Conflicts)
	}
}

func TestGetKeybindings_FileOverrides(t *testing.T) {
	s, dir := buildTestServerWithRoot(t)

	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"graph":"Ctrl+Shift+G","save":"Ctrl+S"}`
	if err := os.WriteFile(filepath.Join(kiwiDir, "keybindings.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/keybindings", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	var res struct {
		Bindings map[string]string `json:"bindings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Bindings["graph"] != "mod+shift+g" {
		t.Fatalf("graph = %q, want mod+shift+g", res.Bindings["graph"])
	}
}

func TestGetKeybindings_ConfigurablePath(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"toggle_sidebar":"Ctrl+Shift+B"}`
	if err := os.WriteFile(filepath.Join(kiwiDir, "keys.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.KeybindingsFile = ".kiwi/keys.json"
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/keybindings", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	var res struct {
		Bindings map[string]string `json:"bindings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Bindings["toggle_sidebar"] != "mod+shift+b" {
		t.Fatalf("toggle_sidebar = %q", res.Bindings["toggle_sidebar"])
	}
}

func TestGetKeybindings_Conflicts(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.Keybindings = map[string]string{
		"search":   "Ctrl+K",
		"new_page": "Ctrl+K",
	}
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/keybindings", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	var res struct {
		Conflicts []struct {
			Chord   string   `json:"chord"`
			Actions []string `json:"actions"`
		} `json:"conflicts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %+v", res.Conflicts)
	}
}
