package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBackupStatusRoutes(t *testing.T) {
	s := buildTestServer(t)

	for _, path := range []string{"/api/kiwi/backup/status", "/api/kiwi/sync/status"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s: status %d body %s", path, rec.Code, rec.Body.String())
		}

		var got struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("GET %s: decode JSON: %v", path, err)
		}
		if got.Enabled {
			t.Fatalf("GET %s: enabled = true, want false before backup sync is configured", path)
		}
	}
}

func TestSetBackupStatusUpdatesHandlersAfterRoutesSetup(t *testing.T) {
	s := buildTestServer(t)
	s.SetBackupStatus(func() any {
		return map[string]any{"success": true}
	})

	for _, path := range []string{"/api/kiwi/backup/status", "/api/kiwi/sync/status"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s: status %d body %s", path, rec.Code, rec.Body.String())
		}

		var got struct {
			Enabled bool `json:"enabled"`
			Status  struct {
				Success bool `json:"success"`
			} `json:"status"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("GET %s: decode JSON: %v", path, err)
		}
		if !got.Enabled || !got.Status.Success {
			t.Fatalf("GET %s: got enabled=%v success=%v, want both true", path, got.Enabled, got.Status.Success)
		}
	}
}
