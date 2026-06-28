package mcpserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

)

func TestMCP_KiwiRemember(t *testing.T) {
	b, tmp := setupTestBackend(t)
	defer b.Close()

	out := mustCallTool(t, handleRemember(b), "kiwi_remember", map[string]any{
		"content": "User prefers dark mode",
		"scope":   "user:alice",
		"tags":    []any{"preference", "ui"},
	})
	date := time.Now().UTC().Format("2006-01-02")
	if !strings.Contains(out, "episodes/"+date+"/") {
		t.Fatalf("want episodes/%s/ in:\n%s", date, out)
	}
	if !strings.Contains(out, "episode_id:") {
		t.Fatalf("want episode_id in:\n%s", out)
	}

	entries, err := os.ReadDir(filepath.Join(tmp, "episodes", date))
	if err != nil || len(entries) != 1 {
		t.Fatalf("episodes dir: entries=%d err=%v", len(entries), err)
	}
	data, err := os.ReadFile(filepath.Join(tmp, "episodes", date, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"memory_kind: episodic",
		"scope: user:alice",
		"User prefers dark mode",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("want %q in:\n%s", want, text)
		}
	}
}

func TestMCP_KiwiRememberWithEpisodeID(t *testing.T) {
	b, tmp := setupTestBackend(t)
	defer b.Close()

	id := "custom-ep-42"
	out := mustCallTool(t, handleRemember(b), "kiwi_remember", map[string]any{
		"content":    "Note without scope",
		"episode_id": id,
	})
	if !strings.Contains(out, id) {
		t.Fatalf("want episode id in:\n%s", out)
	}
	date := time.Now().UTC().Format("2006-01-02")
	path := filepath.Join(tmp, "episodes", date, id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "episode_id: "+id) {
		t.Fatalf("missing episode_id in:\n%s", data)
	}
}

func TestMCP_KiwiForget(t *testing.T) {
	b, tmp := setupTestBackend(t)
	defer b.Close()

	path := "pages/pref.md"
	body := "---\nmemory_status: active\ntitle: Dark mode\n---\n\nUser prefers dark mode.\n"
	if err := os.MkdirAll(filepath.Join(tmp, "pages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, path), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	out := mustCallTool(t, handleForget(b), "kiwi_forget", map[string]any{
		"path":   path,
		"reason": "outdated preference",
	})
	if !strings.Contains(out, "superseded") {
		t.Fatalf("want superseded in:\n%s", out)
	}

	data, err := os.ReadFile(filepath.Join(tmp, path))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"memory_status: superseded",
		"valid_until:",
		"superseded_reason: outdated preference",
		"User prefers dark mode.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("want %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "memory_status: active") {
		t.Fatalf("active status should be replaced:\n%s", text)
	}
}
