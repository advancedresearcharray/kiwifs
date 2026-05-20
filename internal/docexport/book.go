package docexport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// BookManifest describes a multi-file document compilation order.
// It is read from _book.yaml or _index.md frontmatter at a directory root.
type BookManifest struct {
	Title    string   `yaml:"title"`
	Author   string   `yaml:"author"`
	Date     string   `yaml:"date"`
	Lang     string   `yaml:"lang"`
	Parts    []string `yaml:"parts"`    // ordered list of .md files
	Template string   `yaml:"template"` // optional Pandoc/Typst template
	Theme    string   `yaml:"theme"`
}

// LoadManifest looks for a _book.yaml or _book.yml in the given directory.
// If not found, it returns nil (no error) — callers fall back to auto-ordering.
func LoadManifest(dir string) (*BookManifest, error) {
	for _, name := range []string{"_book.yaml", "_book.yml"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var m BookManifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		// Resolve relative paths in parts list.
		for i, p := range m.Parts {
			if !filepath.IsAbs(p) {
				m.Parts[i] = filepath.Join(dir, p)
			}
		}
		return &m, nil
	}
	return nil, nil
}

// StitchFiles reads multiple markdown files and concatenates them with
// page break markers between chapters. Returns the combined markdown and
// the merged metadata from the first file (or manifest).
func StitchFiles(ctx context.Context, provider FileProvider, paths []string, manifest *BookManifest) ([]byte, map[string]string, error) {
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("no files to stitch")
	}

	metadata := make(map[string]string)
	if manifest != nil {
		if manifest.Title != "" {
			metadata["title"] = manifest.Title
		}
		if manifest.Author != "" {
			metadata["author"] = manifest.Author
		}
		if manifest.Date != "" {
			metadata["date"] = manifest.Date
		}
		if manifest.Lang != "" {
			metadata["lang"] = manifest.Lang
		}
	}

	var buf strings.Builder
	for i, p := range paths {
		content, err := provider.ReadFile(ctx, p)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", p, err)
		}

		body := stripFrontmatter(content)

		if i > 0 {
			// Pandoc page break (works for PDF via LaTeX/Typst and HTML).
			buf.WriteString("\n\n\\newpage\n\n")
		}
		buf.Write(body)
		buf.WriteString("\n")
	}

	return []byte(buf.String()), metadata, nil
}

// stripFrontmatter removes YAML frontmatter from markdown content,
// returning only the body.
func stripFrontmatter(content []byte) []byte {
	s := string(content)
	if !strings.HasPrefix(s, "---") {
		return content
	}
	// Find the closing ---
	rest := s[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return content
	}
	// Skip past the closing --- and any immediate newline.
	body := rest[idx+4:]
	if strings.HasPrefix(body, "\n") {
		body = body[1:]
	}
	return []byte(body)
}
