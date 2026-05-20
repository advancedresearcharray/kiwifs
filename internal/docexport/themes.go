package docexport

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
)

//go:embed themes/*
var embeddedThemes embed.FS

// Theme names available out of the box.
const (
	ThemePaper        = "paper"
	ThemeModern       = "modern"
	ThemeMinimal      = "minimal"
	ThemeDark         = "dark"
	ThemePresentation = "presentation"
)

// AvailableThemes returns the list of built-in theme names.
func AvailableThemes() []string {
	return []string{
		ThemePaper,
		ThemeModern,
		ThemeMinimal,
		ThemeDark,
		ThemePresentation,
	}
}

// resolveTemplate finds the Pandoc template file for a given theme and format.
// Returns empty string if no template is available (Pandoc will use defaults).
func resolveTemplate(theme string, format Format, engine string) string {
	if theme == "" {
		return ""
	}

	// Check for user-defined templates in .kiwi/themes/ first.
	// (the root path is resolved at the call site, not here)

	// For typst engine, look for .typ template; otherwise .latex or .html.
	var ext string
	switch format {
	case FormatPDF:
		if engine == "typst" {
			ext = ".typ"
		} else {
			ext = ".latex"
		}
	case FormatHTML:
		ext = ".html"
	default:
		return ""
	}

	// Try to extract embedded template.
	name := "themes/" + theme + ext
	if _, err := embeddedThemes.ReadFile(name); err == nil {
		// Write to temp file (Pandoc needs a file path).
		return writeEmbeddedTemp(name)
	}

	return ""
}

// resolveThemeCSS finds the CSS file for a given theme (used by HTML export).
func resolveThemeCSS(theme string) string {
	if theme == "" {
		return ""
	}

	name := "themes/" + theme + ".css"
	if _, err := embeddedThemes.ReadFile(name); err == nil {
		return writeEmbeddedTemp(name)
	}

	return ""
}

// resolveCSLStyle finds a CSL style file for citations.
// Built-in styles: apa, ieee, chicago, vancouver, harvard.
func resolveCSLStyle(style string, root string) string {
	if style == "" {
		return ""
	}

	// Check user-provided CSL files.
	for _, candidate := range []string{
		filepath.Join(root, ".kiwi", "references", style+".csl"),
		filepath.Join(root, style+".csl"),
		filepath.Join(root, ".kiwi", "references", style),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Check embedded CSL styles.
	name := "themes/" + style + ".csl"
	if _, err := embeddedThemes.ReadFile(name); err == nil {
		return writeEmbeddedTemp(name)
	}

	return ""
}

// writeEmbeddedTemp extracts an embedded file to a temporary location and
// returns the path. The file lives for the duration of the process. For a
// long-running server this is fine — the set of themes is small and fixed.
func writeEmbeddedTemp(name string) string {
	data, err := embeddedThemes.ReadFile(name)
	if err != nil {
		return ""
	}

	// Use a deterministic path based on the name to avoid re-extraction.
	tmpDir := filepath.Join(os.TempDir(), "kiwi-themes")
	os.MkdirAll(tmpDir, 0755)

	safeName := strings.ReplaceAll(name, "/", "_")
	path := filepath.Join(tmpDir, safeName)

	// Write only if not already present.
	if _, err := os.Stat(path); err != nil {
		os.WriteFile(path, data, 0644)
	}

	return path
}
