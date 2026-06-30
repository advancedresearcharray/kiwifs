package keybindings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeChord(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Ctrl+N", "mod+n"},
		{"Mod+Shift+B", "mod+shift+b"},
		{"Ctrl+Alt+F", "alt+mod+f"},
		{"Escape", "escape"},
		{"Ctrl+/", "mod+/"},
		{"Mod+\\", "mod+\\"},
	}
	for _, tc := range tests {
		got, err := NormalizeChord(tc.in)
		if err != nil {
			t.Fatalf("NormalizeChord(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeChord(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestResolveDefaultsWhenMissing(t *testing.T) {
	root := t.TempDir()
	res, err := Resolve(Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if res.Bindings["search"] != "mod+k" {
		t.Fatalf("search = %q, want mod+k", res.Bindings["search"])
	}
	if res.Bindings["toggle_split"] != `mod+\` {
		t.Fatalf("toggle_split = %q, want mod+\\", res.Bindings["toggle_split"])
	}
	if len(res.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %+v", res.Conflicts)
	}
}

func TestResolveFileOverrides(t *testing.T) {
	root := t.TempDir()
	kiwiDir := filepath.Join(root, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"search":"Ctrl+J","new_page":"Ctrl+Shift+N"}`
	if err := os.WriteFile(filepath.Join(kiwiDir, "keybindings.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Resolve(Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if res.Bindings["search"] != "mod+j" {
		t.Fatalf("search = %q, want mod+j", res.Bindings["search"])
	}
	if res.Bindings["new_page"] != "mod+shift+n" {
		t.Fatalf("new_page = %q, want mod+shift+n", res.Bindings["new_page"])
	}
	if res.Bindings["save"] != "mod+s" {
		t.Fatalf("save default missing: %q", res.Bindings["save"])
	}
}

func TestResolveConfigOverridesFile(t *testing.T) {
	root := t.TempDir()
	kiwiDir := filepath.Join(root, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kiwiDir, "keybindings.json"), []byte(`{"search":"Ctrl+J"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Resolve(Options{
		Root:              root,
		ConfigKeybindings: map[string]string{"search": "Ctrl+K"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Bindings["search"] != "mod+k" {
		t.Fatalf("config override lost: %q", res.Bindings["search"])
	}
}

func TestResolveDetectsConflicts(t *testing.T) {
	root := t.TempDir()
	res, err := Resolve(Options{
		Root: root,
		ConfigKeybindings: map[string]string{
			"search":   "Ctrl+K",
			"new_page": "Ctrl+K",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %+v", res.Conflicts)
	}
	if res.Conflicts[0].Chord != "mod+k" {
		t.Fatalf("conflict chord = %q", res.Conflicts[0].Chord)
	}
}

func TestResolveIgnoresUnknownActions(t *testing.T) {
	root := t.TempDir()
	res, err := Resolve(Options{
		Root: root,
		ConfigKeybindings: map[string]string{
			"not_real": "Ctrl+Q",
			"search":   "Ctrl+J",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := res.Bindings["not_real"]; ok {
		t.Fatalf("unknown action should be ignored")
	}
}
