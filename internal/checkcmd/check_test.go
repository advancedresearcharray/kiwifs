package checkcmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckCommandSequenceGap(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".kiwi"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `[sequences]
directories = ["events/"]
`
	if err := os.WriteFile(filepath.Join(dir, ".kiwi", "config.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	eventsDir := filepath.Join(dir, "events")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eventsDir, "log.md"),
		[]byte("<!-- seq:1 -->\na\n<!-- seq:3 -->\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(dir, ".kiwi", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "sequences.json"),
		[]byte("{\n  \"counter\": 3\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := Command
	cmd.SetArgs([]string{"--root", dir})
	err := cmd.Execute()
	if !errors.Is(err, ErrFailed) {
		t.Fatalf("expected ErrFailed, got %v", err)
	}
}

func TestCheckCommandNoSequencesConfigured(t *testing.T) {
	dir := t.TempDir()
	cmd := Command
	cmd.SetArgs([]string{"--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected success when no sequences configured, got %v", err)
	}
}
