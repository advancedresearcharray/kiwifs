package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestContext_WithPlaybook(t *testing.T) {
	s, dir := buildSQLiteTestServer(t)

	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(kiwiDir, "playbook.md"), []byte("# Playbook\nTest playbook"), 0644)
	os.WriteFile(filepath.Join(dir, "SCHEMA.md"), []byte("# Schema\nTest schema"), 0644)
	os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Index\nTest index"), 0644)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/context", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result["playbook"] != "# Playbook\nTest playbook" {
		t.Errorf("playbook = %q", result["playbook"])
	}
	if result["schema"] != "# Schema\nTest schema" {
		t.Errorf("schema = %q", result["schema"])
	}
	if result["index"] != "# Index\nTest index" {
		t.Errorf("index = %q", result["index"])
	}
}

func TestContext_NoPlaybook(t *testing.T) {
	s := buildTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/context", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result["playbook"] != "" {
		t.Errorf("expected empty playbook, got %q", result["playbook"])
	}
	if result["schema"] != "" {
		t.Errorf("expected empty schema, got %q", result["schema"])
	}
	if result["index"] != "" {
		t.Errorf("expected empty index, got %q", result["index"])
	}
}

func TestContext_WikiTemplate(t *testing.T) {
	s, dir := buildSQLiteTestServer(t)

	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(kiwiDir, "playbook.md"), []byte("# Agent Playbook — Team Wiki"), 0644)
	os.WriteFile(filepath.Join(dir, "SCHEMA.md"), []byte("# Schema — Team Wiki"), 0644)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/context", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result["playbook"] != "# Agent Playbook — Team Wiki" {
		t.Errorf("playbook = %q", result["playbook"])
	}
	if result["schema"] != "# Schema — Team Wiki" {
		t.Errorf("schema = %q", result["schema"])
	}
}
