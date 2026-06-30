package themepresets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiwifs/kiwifs/internal/config"
)

// Preset is a workspace-defined color scheme loaded from JSON.
type Preset struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Light       map[string]string `json:"light"`
	Dark        map[string]string `json:"dark"`
	Source      string            `json:"source,omitempty"`
	File        string            `json:"file,omitempty"`
}

// LoadError records a validation or read failure for one preset file.
type LoadError struct {
	File  string `json:"file"`
	Error string `json:"error"`
}

// Result holds loaded workspace presets and any per-file errors.
type Result struct {
	Presets []Preset    `json:"presets"`
	Errors  []LoadError `json:"errors,omitempty"`
}

// ValidatePreset checks that raw JSON decodes to a usable theme preset.
func ValidatePreset(raw map[string]any) (Preset, error) {
	name, _ := raw["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" {
		return Preset{}, fmt.Errorf("missing or empty name")
	}
	desc, _ := raw["description"].(string)
	light, err := tokenMap(raw["light"])
	if err != nil {
		return Preset{}, fmt.Errorf("light: %w", err)
	}
	dark, err := tokenMap(raw["dark"])
	if err != nil {
		return Preset{}, fmt.Errorf("dark: %w", err)
	}
	if len(light) == 0 && len(dark) == 0 {
		return Preset{}, fmt.Errorf("light and dark must include at least one token")
	}
	return Preset{
		Name:        name,
		Description: desc,
		Light:       light,
		Dark:        dark,
		Source:      "workspace",
	}, nil
}

func tokenMap(v any) (map[string]string, error) {
	raw, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("must be an object")
	}
	out := make(map[string]string, len(raw))
	for k, val := range raw {
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("token %q must be a string", k)
		}
		out[k] = s
	}
	return out, nil
}

// LoadFromDir reads *.json presets from the configured workspace directory.
func LoadFromDir(root string, themeCfg config.UIThemeConfig) Result {
	rel := themeCfg.ResolvedPresetsDir()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	entries, err := os.ReadDir(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}
		}
		return Result{Errors: []LoadError{{File: rel, Error: err.Error()}}}
	}

	var result Result
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(strings.ToLower(ent.Name()), ".json") {
			continue
		}
		fileRel := filepath.ToSlash(filepath.Join(rel, ent.Name()))
		data, err := os.ReadFile(filepath.Join(abs, ent.Name()))
		if err != nil {
			result.Errors = append(result.Errors, LoadError{File: fileRel, Error: err.Error()})
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			result.Errors = append(result.Errors, LoadError{File: fileRel, Error: "invalid JSON: " + err.Error()})
			continue
		}
		preset, err := ValidatePreset(raw)
		if err != nil {
			result.Errors = append(result.Errors, LoadError{File: fileRel, Error: err.Error()})
			continue
		}
		preset.File = fileRel
		result.Presets = append(result.Presets, preset)
	}
	return result
}

// FilterByAllowed returns presets whose names match allowed (case-insensitive).
// When allowed is empty, all presets are returned.
func FilterByAllowed(presets []Preset, allowed []string) []Preset {
	if len(allowed) == 0 {
		return presets
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		allowedSet[strings.ToLower(a)] = struct{}{}
	}
	if len(allowedSet) == 0 {
		return presets
	}
	out := make([]Preset, 0, len(presets))
	for _, p := range presets {
		if _, ok := allowedSet[strings.ToLower(p.Name)]; ok {
			out = append(out, p)
		}
	}
	return out
}
