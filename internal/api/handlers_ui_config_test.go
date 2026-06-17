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
	if len(res.Features) != 7 {
		t.Fatalf("expected 7 default features, got %+v", res.Features)
	}
	for _, key := range []string{"graph", "kanban", "canvas", "whiteboard", "timeline", "bases", "data_sources"} {
		if !res.Features[key] {
			t.Fatalf("expected feature %s enabled by default", key)
		}
	}
}

func TestUIConfig_Features(t *testing.T) {
	s, dir := buildTestServerWithRoot(t)

	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `
[ui.features]
kanban = false
graph = true
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
	if res.Features["kanban"] {
		t.Fatal("expected kanban disabled")
	}
	if !res.Features["graph"] || !res.Features["bases"] {
		t.Fatalf("unexpected features: %+v", res.Features)
	}
}
