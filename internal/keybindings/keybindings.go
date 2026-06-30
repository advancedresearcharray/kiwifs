package keybindings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Action IDs for app-level shortcuts. Unknown keys in config are ignored.
var knownActions = map[string]struct{}{
	"search":             {},
	"new_page":           {},
	"toggle_editor":      {},
	"save":               {},
	"toggle_sidebar":     {},
	"graph":              {},
	"toggle_bases":       {},
	"toggle_timeline":    {},
	"toggle_kanban":      {},
	"toggle_mode":        {},
	"shortcuts_help":     {},
	"undo":               {},
	"focus_tree_filter":  {},
	"close_overlay":      {},
	"toggle_split":       {},
}

// DefaultBindings are used when no config overrides are present.
var DefaultBindings = map[string]string{
	"search":            "Mod+K",
	"new_page":          "Mod+N",
	"toggle_editor":     "Mod+E",
	"save":              "Mod+S",
	"toggle_sidebar":    "Mod+B",
	"graph":             "Mod+G",
	"toggle_bases":      "Mod+Shift+B",
	"toggle_timeline":   "Mod+Shift+T",
	"toggle_kanban":     "Mod+Shift+W",
	"toggle_mode":       "Mod+Shift+E",
	"shortcuts_help":    "Mod+/",
	"undo":              "Mod+Z",
	"focus_tree_filter": "Mod+Alt+F",
	"close_overlay":     "Escape",
	"toggle_split":      "Mod+\\",
}

// Conflict describes two or more actions bound to the same chord.
type Conflict struct {
	Chord   string   `json:"chord"`
	Actions []string `json:"actions"`
}

// Resolved holds merged bindings plus validation warnings.
type Resolved struct {
	Bindings  map[string]string `json:"bindings"`
	Defaults  map[string]string `json:"defaults"`
	Conflicts []Conflict        `json:"conflicts"`
}

// Options configures how workspace keybindings are loaded.
type Options struct {
	Root              string
	KeybindingsFile   string            // relative path, default .kiwi/keybindings.json
	ConfigKeybindings map[string]string // from [ui.keybindings] in config.toml
}

func keybindingsRelPath(rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return ".kiwi/keybindings.json"
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	if filepath.IsAbs(rel) || strings.Contains(rel, "..") {
		return ".kiwi/keybindings.json"
	}
	return rel
}

// NormalizeChord canonicalizes a chord string for comparison and storage.
func NormalizeChord(chord string) (string, error) {
	chord = strings.TrimSpace(chord)
	if chord == "" {
		return "", fmt.Errorf("empty chord")
	}
	parts := strings.Split(chord, "+")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid chord %q", chord)
	}

	var mods []string
	key := ""
	for _, raw := range parts {
		p := strings.ToLower(strings.TrimSpace(raw))
		switch p {
		case "ctrl", "control":
			mods = appendUnique(mods, "mod")
		case "cmd", "command", "meta", "mod":
			mods = appendUnique(mods, "mod")
		case "shift":
			mods = appendUnique(mods, "shift")
		case "alt", "option":
			mods = appendUnique(mods, "alt")
		case "":
			return "", fmt.Errorf("invalid chord %q", chord)
		default:
			if key != "" {
				return "", fmt.Errorf("multiple keys in chord %q", chord)
			}
			key = normalizeKey(p)
		}
	}
	if key == "" {
		return "", fmt.Errorf("missing key in chord %q", chord)
	}

	sort.Strings(mods)
	out := append(mods, key)
	return strings.Join(out, "+"), nil
}

func normalizeKey(key string) string {
	switch key {
	case "esc", "escape":
		return "escape"
	case "slash", "/":
		return "/"
	case "backslash", `\`:
		return `\`
	case "question", "?":
		return "?"
	default:
		if len(key) == 1 {
			return key
		}
		return key
	}
}

func appendUnique(list []string, item string) []string {
	for _, v := range list {
		if v == item {
			return list
		}
	}
	return append(list, item)
}

func cloneDefaults() map[string]string {
	out := make(map[string]string, len(DefaultBindings))
	for k, v := range DefaultBindings {
		out[k] = v
	}
	return out
}

func filterKnown(src map[string]string) map[string]string {
	out := make(map[string]string)
	for action, chord := range src {
		if _, ok := knownActions[action]; !ok {
			continue
		}
		if normalized, err := NormalizeChord(chord); err == nil {
			out[action] = normalized
		}
	}
	return out
}

func readFileBindings(root, relPath string) (map[string]string, error) {
	p := filepath.Join(root, keybindingsRelPath(relPath))
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid keybindings.json: %w", err)
	}
	return filterKnown(raw), nil
}

// Resolve merges defaults, file overrides, and inline config overrides.
func Resolve(opts Options) (Resolved, error) {
	bindings := cloneDefaults()

	fileBindings, err := readFileBindings(opts.Root, opts.KeybindingsFile)
	if err != nil {
		return Resolved{}, err
	}
	for action, chord := range fileBindings {
		bindings[action] = chord
	}
	for action, chord := range filterKnown(opts.ConfigKeybindings) {
		normalized, err := NormalizeChord(chord)
		if err != nil {
			continue
		}
		bindings[action] = normalized
	}

	normalized := normalizeBindingMap(bindings)
	conflicts := detectConflicts(normalized)
	defaults := normalizeBindingMap(cloneDefaults())
	return Resolved{
		Bindings:  normalized,
		Defaults:  defaults,
		Conflicts: conflicts,
	}, nil
}

func normalizeBindingMap(src map[string]string) map[string]string {
	out := make(map[string]string, len(src))
	for action, chord := range src {
		if normalized, err := NormalizeChord(chord); err == nil {
			out[action] = normalized
		} else {
			out[action] = chord
		}
	}
	return out
}

func detectConflicts(bindings map[string]string) []Conflict {
	byChord := map[string][]string{}
	for action, chord := range bindings {
		byChord[chord] = append(byChord[chord], action)
	}
	var conflicts []Conflict
	for chord, actions := range byChord {
		if len(actions) < 2 {
			continue
		}
		sort.Strings(actions)
		conflicts = append(conflicts, Conflict{Chord: chord, Actions: actions})
	}
	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].Chord < conflicts[j].Chord
	})
	return conflicts
}
