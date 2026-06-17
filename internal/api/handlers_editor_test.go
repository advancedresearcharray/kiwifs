package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kiwifs/kiwifs/internal/config"
)

func TestGetEditorSlashCommands_EmptyWhenUnset(t *testing.T) {
	s := buildTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/editor/slash-commands", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res editorSlashCommandsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Commands) != 0 {
		t.Fatalf("expected no commands, got %+v", res.Commands)
	}
}

func TestGetEditorSlashCommands_FromConfig(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.Editor.SlashCommands = []config.SlashCommandConfig{
		{
			ID:          "adr",
			Label:       "ADR",
			Icon:        "FileCheck",
			Description: "Insert ADR template",
			Template:    "templates/adr.md",
		},
		{
			ID:       "",
			Template: "templates/skip.md",
		},
		{
			ID:       "no-template",
			Label:    "Skip me",
			Template: "",
		},
	}
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/editor/slash-commands", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res editorSlashCommandsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Commands) != 1 {
		t.Fatalf("expected 1 command, got %+v", res.Commands)
	}
	cmd := res.Commands[0]
	if cmd.ID != "adr" || cmd.Label != "ADR" || cmd.Icon != "FileCheck" {
		t.Fatalf("unexpected command: %+v", cmd)
	}
	if cmd.Description != "Insert ADR template" || cmd.Template != "templates/adr.md" {
		t.Fatalf("unexpected metadata: %+v", cmd)
	}
}

func TestGetEditorSlashCommands_DefaultLabelFromID(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.Editor.SlashCommands = []config.SlashCommandConfig{{
		ID:       "runbook",
		Template: "templates/runbook-step.md",
	}}
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/editor/slash-commands", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	var res editorSlashCommandsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Commands) != 1 || res.Commands[0].Label != "runbook" {
		t.Fatalf("label fallback failed: %+v", res.Commands)
	}
}

func TestGetEditorSlashCommands_SkipsInvalidID(t *testing.T) {
	dir, pipe, cstore := buildTestPipeline(t)
	cfg := &config.Config{}
	cfg.Storage.Root = dir
	cfg.UI.Editor.SlashCommands = []config.SlashCommandConfig{
		{ID: "bad id", Label: "Spaces", Template: "templates/bad.md"},
		{ID: "adr", Label: "ADR", Template: "templates/adr.md"},
	}
	s := NewServer(cfg, pipe, nil, cstore, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/editor/slash-commands", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	var res editorSlashCommandsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Commands) != 1 || res.Commands[0].ID != "adr" {
		t.Fatalf("expected only valid id, got %+v", res.Commands)
	}
}
