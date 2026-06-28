package docexport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- ParseFormat ---

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"pdf", FormatPDF, false},
		{"PDF", FormatPDF, false},
		{"  pdf  ", FormatPDF, false},
		{"html", FormatHTML, false},
		{"slides", FormatSlides, false},
		{"slide", FormatSlides, false},
		{"presentation", FormatSlides, false},
		{"site", FormatSite, false},
		{"docs", FormatSite, false},
		{"mkdocs", FormatSite, false},
		{"", "", true},
		{"docx", "", true},
		{"latex", "", true},
		{"SLIDES\n", FormatSlides, false}, // trailing newline
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- suggestFilename ---

func TestSuggestFilename(t *testing.T) {
	tests := []struct {
		inputPath string
		ext       string
		want      string
	}{
		{"docs/report.md", ".pdf", "report.pdf"},
		{"report.md", ".html", "report.html"},
		{"", ".pdf", "export.pdf"},
		{"docs/", ".pdf", "docs.pdf"},
		{"./", ".pdf", "export.pdf"},
		{"deeply/nested/path/chapter.md", ".pdf", "chapter.pdf"},
		{"file-with-dashes.md", ".pdf", "file-with-dashes.pdf"},
		{"UPPERCASE.MD", ".html", "UPPERCASE.html"},
		{"no-extension", ".pdf", "no-extension.pdf"},
		{".", ".pdf", "export.pdf"},
	}
	for _, tt := range tests {
		t.Run(tt.inputPath+"→"+tt.ext, func(t *testing.T) {
			got := suggestFilename(tt.inputPath, tt.ext)
			if got != tt.want {
				t.Errorf("suggestFilename(%q, %q) = %q, want %q", tt.inputPath, tt.ext, got, tt.want)
			}
		})
	}
}

// --- stripFrontmatter ---

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no frontmatter",
			input: "# Hello\nWorld",
			want:  "# Hello\nWorld",
		},
		{
			name:  "basic frontmatter",
			input: "---\ntitle: Test\n---\n# Hello",
			want:  "# Hello",
		},
		{
			name:  "frontmatter with CRLF",
			input: "---\r\ntitle: Test\r\n---\r\n# Hello",
			want:  "# Hello",
		},
		{
			name:  "frontmatter with mixed CRLF and LF",
			input: "---\r\ntitle: Mixed\n---\n# Hello",
			want:  "# Hello",
		},
		{
			name:  "empty frontmatter",
			input: "---\n---\n# Hello",
			want:  "# Hello",
		},
		{
			name:  "frontmatter with extra newlines",
			input: "---\ntitle: Test\ntags: [a, b]\n---\n\n# Hello\n\nWorld",
			want:  "\n# Hello\n\nWorld",
		},
		{
			name:  "content looks like frontmatter but isn't",
			input: "Some text\n---\ntitle: not frontmatter\n---\n",
			want:  "Some text\n---\ntitle: not frontmatter\n---\n",
		},
		{
			name:  "triple dashes in content after real frontmatter",
			input: "---\ntitle: Real\n---\n# Title\n---\nmore content",
			want:  "# Title\n---\nmore content",
		},
		{
			name:  "only frontmatter no body",
			input: "---\ntitle: Lonely\n---\n",
			want:  "",
		},
		{
			name:  "frontmatter with unicode",
			input: "---\ntitle: 日本語テスト\nauthor: Ñoño\n---\n# 你好",
			want:  "# 你好",
		},
		{
			name:  "unclosed frontmatter",
			input: "---\ntitle: Never Closed\n# Content",
			want:  "---\ntitle: Never Closed\n# Content",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripFrontmatter([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("stripFrontmatter():\n  got:  %q\n  want: %q", got, tt.want)
			}
		})
	}
}

// --- ensureMarpFrontmatter ---

func TestEnsureMarpFrontmatter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "no frontmatter",
			input: "# My Slides\n\nContent here",
			check: func(t *testing.T, output string) {
				if !strings.HasPrefix(output, "---\nmarp: true\n---") {
					t.Errorf("should prepend marp frontmatter, got: %q", output[:50])
				}
				if !strings.Contains(output, "# My Slides") {
					t.Error("original content should be preserved")
				}
			},
		},
		{
			name:  "existing frontmatter without marp",
			input: "---\ntitle: Talk\ntheme: gaia\n---\n# Slide 1",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "marp: true") {
					t.Error("should inject marp: true")
				}
				if !strings.Contains(output, "title: Talk") {
					t.Error("existing frontmatter should be preserved")
				}
				if !strings.Contains(output, "theme: gaia") {
					t.Error("theme should be preserved")
				}
				// Must only have one frontmatter block.
				count := strings.Count(output, "---")
				if count != 2 {
					t.Errorf("should have exactly 2 --- delimiters, got %d in: %q", count, output)
				}
			},
		},
		{
			name:  "already has marp true",
			input: "---\nmarp: true\ntitle: Talk\n---\n# Slide 1",
			check: func(t *testing.T, output string) {
				if strings.Count(output, "marp:") != 1 {
					t.Error("should not duplicate marp directive")
				}
			},
		},
		{
			name:  "marp false should not be duplicated",
			input: "---\nmarp: false\ntitle: Talk\n---\n# Slide 1",
			check: func(t *testing.T, output string) {
				if strings.Count(output, "marp:") != 1 {
					t.Error("should not add duplicate marp line when marp: false exists")
				}
			},
		},
		{
			name:  "empty document",
			input: "",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "marp: true") {
					t.Error("empty doc should get marp frontmatter")
				}
			},
		},
		{
			name:  "CRLF line endings",
			input: "---\r\ntitle: Talk\r\n---\r\n# Slide 1",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "marp: true") {
					t.Error("should inject marp: true with CRLF input")
				}
				if !strings.Contains(output, "# Slide 1") {
					t.Error("content after CRLF frontmatter should survive")
				}
			},
		},
		{
			name:  "frontmatter with --- inside code block after it",
			input: "---\ntitle: Code\n---\n```\n---\n```",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "marp: true") {
					t.Error("should inject marp: true")
				}
				if !strings.Contains(output, "```\n---\n```") {
					t.Error("code block with --- should be preserved")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := string(ensureMarpFrontmatter([]byte(tt.input)))
			tt.check(t, output)
		})
	}
}

// --- BookManifest ---

func TestLoadManifest(t *testing.T) {
	t.Run("no manifest returns nil", func(t *testing.T) {
		dir := t.TempDir()
		m, err := LoadManifest(dir, "docs")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m != nil {
			t.Error("expected nil manifest when no _book.yaml exists")
		}
	})

	t.Run("valid manifest", func(t *testing.T) {
		dir := t.TempDir()
		yaml := `title: "My Book"
author: "Author"
date: "2026-01-01"
lang: en
parts:
  - intro.md
  - chapters/01-setup.md
  - chapters/02-usage.md
  - appendix.md
`
		os.WriteFile(filepath.Join(dir, "_book.yaml"), []byte(yaml), 0644)

		m, err := LoadManifest(dir, "docs")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m == nil {
			t.Fatal("expected non-nil manifest")
		}
		if m.Title != "My Book" {
			t.Errorf("title = %q, want %q", m.Title, "My Book")
		}
		if len(m.Parts) != 4 {
			t.Fatalf("parts count = %d, want 4", len(m.Parts))
		}
		if m.Parts[0] != "docs/intro.md" {
			t.Errorf("parts[0] = %q, want %q", m.Parts[0], "docs/intro.md")
		}
	})

	t.Run("yml extension", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "_book.yml"), []byte("title: YML Test\nparts:\n  - a.md\n"), 0644)

		m, err := LoadManifest(dir, "src")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m == nil || m.Title != "YML Test" {
			t.Error("should parse _book.yml")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "_book.yaml"), []byte("{{invalid yaml"), 0644)

		_, err := LoadManifest(dir, "docs")
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})

	t.Run("absolute paths in parts not re-joined", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "_book.yaml"), []byte("parts:\n  - /abs/path.md\n"), 0644)

		m, err := LoadManifest(dir, "docs")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.Parts[0] != "/abs/path.md" {
			t.Errorf("absolute path should not be re-joined, got %q", m.Parts[0])
		}
	})
}

// --- StitchFiles ---

func TestStitchFiles(t *testing.T) {
	provider := &memProvider{
		files: map[string][]byte{
			"ch1.md": []byte("---\ntitle: Chapter 1\n---\n# Chapter 1\nContent one."),
			"ch2.md": []byte("# Chapter 2\nContent two."),
			"ch3.md": []byte("---\norder: 3\n---\n# Chapter 3\n\nContent three with math: $E=mc^2$"),
		},
	}

	t.Run("basic stitch", func(t *testing.T) {
		combined, _, err := StitchFiles(context.Background(), provider, []string{"ch1.md", "ch2.md", "ch3.md"}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(combined)

		if strings.Contains(s, "title: Chapter 1") {
			t.Error("frontmatter should be stripped from chapters")
		}
		if !strings.Contains(s, "# Chapter 1") {
			t.Error("body should be preserved")
		}
		if !strings.Contains(s, "# Chapter 2") {
			t.Error("second chapter body should be present")
		}
		if !strings.Contains(s, "\\newpage") {
			t.Error("page breaks should be inserted between chapters")
		}
		if strings.HasPrefix(s, "\\newpage") {
			t.Error("page break should not be at the beginning")
		}
		if !strings.Contains(s, "$E=mc^2$") {
			t.Error("math should survive stitching")
		}
	})

	t.Run("empty paths", func(t *testing.T) {
		_, _, err := StitchFiles(context.Background(), provider, nil, nil)
		if err == nil {
			t.Error("expected error for empty paths")
		}
	})

	t.Run("manifest metadata", func(t *testing.T) {
		manifest := &BookManifest{
			Title:  "My Book",
			Author: "Author Name",
			Lang:   "en",
		}
		_, meta, err := StitchFiles(context.Background(), provider, []string{"ch1.md"}, manifest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if meta["title"] != "My Book" {
			t.Errorf("title = %q, want %q", meta["title"], "My Book")
		}
		if meta["author"] != "Author Name" {
			t.Errorf("author = %q, want %q", meta["author"], "Author Name")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, _, err := StitchFiles(context.Background(), provider, []string{"ch1.md", "nonexistent.md"}, nil)
		if err == nil {
			t.Error("expected error for missing file")
		}
	})
}

// --- Registry ---

func TestRegistry(t *testing.T) {
	t.Run("no exporter registered", func(t *testing.T) {
		r := NewRegistry()
		_, err := r.Export(context.Background(), ExportOpts{Format: FormatPDF})
		if err == nil {
			t.Error("expected error for unregistered format")
		}
	})

	t.Run("supported formats", func(t *testing.T) {
		r := NewRegistry()
		r.Register(&mockExporter{formats: []Format{FormatPDF, FormatHTML}})
		formats := r.SupportedFormats()
		if len(formats) != 2 {
			t.Errorf("expected 2 formats, got %d", len(formats))
		}
	})
}

// --- appendMetadataArgs ---

func TestAppendMetadataArgs(t *testing.T) {
	t.Run("file metadata takes precedence", func(t *testing.T) {
		args := appendMetadataArgs(nil,
			map[string]string{"title": "From File"},
			map[string]string{"title": "From Opts", "author": "Author"},
		)
		found := map[string]string{}
		for i := 0; i < len(args)-1; i += 2 {
			if args[i] == "-M" {
				parts := strings.SplitN(args[i+1], "=", 2)
				found[parts[0]] = parts[1]
			}
		}
		if found["title"] != "From File" {
			t.Errorf("title should be from file, got %q", found["title"])
		}
		if found["author"] != "Author" {
			t.Error("author from opts should be included")
		}
	})

	t.Run("auto-injects date", func(t *testing.T) {
		args := appendMetadataArgs(nil, nil, nil)
		hasDate := false
		for _, a := range args {
			if strings.HasPrefix(a, "date=") {
				hasDate = true
			}
		}
		if !hasDate {
			t.Error("should auto-inject date when not provided")
		}
	})

	t.Run("does not override explicit date", func(t *testing.T) {
		args := appendMetadataArgs(nil,
			map[string]string{"date": "2025-01-01"},
			nil,
		)
		dateCount := 0
		for _, a := range args {
			if strings.HasPrefix(a, "date=") {
				dateCount++
			}
		}
		if dateCount != 1 {
			t.Errorf("date should appear exactly once, got %d", dateCount)
		}
	})
}

// --- appendBibliographyArgs ---

func TestAppendBibliographyArgs(t *testing.T) {
	t.Run("no bibliography", func(t *testing.T) {
		args := appendBibliographyArgs(nil, ExportOpts{}, "/nonexistent")
		if len(args) != 0 {
			t.Error("no bib should produce no args")
		}
	})

	t.Run("explicit bibliography path", func(t *testing.T) {
		dir := t.TempDir()
		bibPath := filepath.Join(dir, "refs.bib")
		os.WriteFile(bibPath, []byte("@article{test, title={Test}}"), 0644)

		args := appendBibliographyArgs(nil, ExportOpts{Bibliography: bibPath}, dir)

		hasCiteproc := false
		hasBib := false
		for _, a := range args {
			if a == "--citeproc" {
				hasCiteproc = true
			}
			if strings.HasPrefix(a, "--bibliography=") {
				hasBib = true
			}
		}
		if !hasCiteproc {
			t.Error("should include --citeproc")
		}
		if !hasBib {
			t.Error("should include --bibliography")
		}
	})

	t.Run("auto-detect refs.bib", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "refs.bib"), []byte("@article{a, title={A}}"), 0644)

		args := appendBibliographyArgs(nil, ExportOpts{}, dir)
		hasBib := false
		for _, a := range args {
			if strings.Contains(a, "refs.bib") {
				hasBib = true
			}
		}
		if !hasBib {
			t.Error("should auto-detect refs.bib in root")
		}
	})

	t.Run("auto-detect in .kiwi/references/", func(t *testing.T) {
		dir := t.TempDir()
		refsDir := filepath.Join(dir, ".kiwi", "references")
		os.MkdirAll(refsDir, 0755)
		os.WriteFile(filepath.Join(refsDir, "refs.bib"), []byte("@article{a, title={A}}"), 0644)

		args := appendBibliographyArgs(nil, ExportOpts{}, dir)
		hasBib := false
		for _, a := range args {
			if strings.Contains(a, "refs.bib") {
				hasBib = true
			}
		}
		if !hasBib {
			t.Error("should auto-detect refs.bib in .kiwi/references/")
		}
	})

	t.Run("relative bibliography path resolved", func(t *testing.T) {
		dir := t.TempDir()
		bibPath := filepath.Join(dir, "my.bib")
		os.WriteFile(bibPath, []byte("@article{a, title={A}}"), 0644)

		args := appendBibliographyArgs(nil, ExportOpts{Bibliography: "my.bib"}, dir)
		for _, a := range args {
			if strings.HasPrefix(a, "--bibliography=") {
				path := strings.TrimPrefix(a, "--bibliography=")
				if !filepath.IsAbs(path) {
					t.Error("relative bibliography should be resolved to absolute")
				}
			}
		}
	})
}

// --- Themes ---

func TestAvailableThemes(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) != 5 {
		t.Errorf("expected 5 themes, got %d", len(themes))
	}
	expected := map[string]bool{"paper": true, "modern": true, "minimal": true, "dark": true, "presentation": true}
	for _, th := range themes {
		if !expected[th] {
			t.Errorf("unexpected theme: %s", th)
		}
	}
}

func TestResolveThemeCSS(t *testing.T) {
	t.Run("known theme returns path", func(t *testing.T) {
		path := resolveThemeCSS("paper")
		if path == "" {
			t.Error("paper theme should resolve to a CSS path")
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("resolved CSS should exist on disk: %v", err)
		}
	})

	t.Run("unknown theme returns empty", func(t *testing.T) {
		path := resolveThemeCSS("nonexistent-theme-xyz")
		if path != "" {
			t.Errorf("unknown theme should return empty, got %q", path)
		}
	})

	t.Run("empty theme returns empty", func(t *testing.T) {
		path := resolveThemeCSS("")
		if path != "" {
			t.Errorf("empty theme should return empty, got %q", path)
		}
	})
}

func TestResolveCSLStyle(t *testing.T) {
	t.Run("user CSL file", func(t *testing.T) {
		dir := t.TempDir()
		cslPath := filepath.Join(dir, "custom.csl")
		os.WriteFile(cslPath, []byte("<style/>"), 0644)

		got := resolveCSLStyle("custom", dir)
		if got != cslPath {
			t.Errorf("should find user CSL, got %q", got)
		}
	})

	t.Run("user CSL in .kiwi/references", func(t *testing.T) {
		dir := t.TempDir()
		refsDir := filepath.Join(dir, ".kiwi", "references")
		os.MkdirAll(refsDir, 0755)
		cslPath := filepath.Join(refsDir, "ieee.csl")
		os.WriteFile(cslPath, []byte("<style/>"), 0644)

		got := resolveCSLStyle("ieee", dir)
		if got != cslPath {
			t.Errorf("should find CSL in .kiwi/references, got %q", got)
		}
	})

	t.Run("nonexistent style returns empty", func(t *testing.T) {
		got := resolveCSLStyle("nonexistent", t.TempDir())
		if got != "" {
			t.Errorf("should return empty for nonexistent CSL, got %q", got)
		}
	})
}

// --- DepStatus ---

func TestCheckDeps(t *testing.T) {
	deps := CheckDeps()
	if len(deps) == 0 {
		t.Fatal("CheckDeps should return at least one entry")
	}
	foundPandoc := false
	for _, d := range deps {
		if d.Binary == "pandoc" {
			foundPandoc = true
			// Not asserting availability since test env may not have it.
			break
		}
	}
	if !foundPandoc {
		t.Error("should check for pandoc")
	}
}

func TestIsAvailableCache(t *testing.T) {
	// Call twice to exercise cache path.
	_ = IsAvailable("this-tool-definitely-does-not-exist-xyz-123")
	result := IsAvailable("this-tool-definitely-does-not-exist-xyz-123")
	if result {
		t.Error("nonexistent tool should not be available")
	}
}

func TestRequireTool(t *testing.T) {
	err := RequireTool("NonexistentTool", "this-tool-definitely-does-not-exist-xyz-123")
	if err == nil {
		t.Error("should return error for missing tool")
	}
	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("error message should mention installation, got: %s", err)
	}
}

// --- Edge cases with weird markdown ---

func TestStripFrontmatterEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "frontmatter with colons in values",
			input: "---\ntitle: \"Key: Value: Pair\"\nurl: https://example.com\n---\nBody",
			want:  "Body",
		},
		{
			name:  "frontmatter with multiline string",
			input: "---\ntitle: |\n  Multi\n  Line\n---\nBody",
			want:  "Body",
		},
		{
			name:  "four dashes not treated as frontmatter delimiter",
			input: "----\ntitle: not frontmatter\n---\nBody",
			want:  "----\ntitle: not frontmatter\n---\nBody",
		},
		{
			name:  "content is only dashes",
			input: "---",
			want:  "---",
		},
		{
			name:  "binary-like content after frontmatter",
			input: "---\ntitle: bin\n---\n\x00\x01\x02\x03",
			want:  "\x00\x01\x02\x03",
		},
		{
			name:  "very large frontmatter key count",
			input: generateLargeFrontmatter(100) + "Body content",
			want:  "Body content",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripFrontmatter([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("stripFrontmatter():\n  got:  %q\n  want: %q", truncate(got, 200), truncate(tt.want, 200))
			}
		})
	}
}

func TestEnsureMarpFrontmatterEdgeCases(t *testing.T) {
	t.Run("frontmatter with marp inside a value", func(t *testing.T) {
		input := "---\nnotes: marp is cool\n---\n# Slide"
		output := string(ensureMarpFrontmatter([]byte(input)))
		// "marp:" is not present as a key, only in a value — should still inject.
		// Current impl checks strings.Contains(fm, "marp:") which would match "marp is cool"
		// only if that substring contained "marp:" — let's verify.
		if !strings.Contains(output, "marp:") {
			t.Error("should contain marp: directive")
		}
	})

	t.Run("slide separators in content", func(t *testing.T) {
		input := "# Slide 1\n\n---\n\n# Slide 2\n\n---\n\n# Slide 3"
		output := string(ensureMarpFrontmatter([]byte(input)))
		if !strings.Contains(output, "marp: true") {
			t.Error("should add marp frontmatter")
		}
		if strings.Count(output, "# Slide") != 3 {
			t.Error("all slide content should be preserved")
		}
	})
}

// --- sortNavEntries ---

func TestSortNavEntries(t *testing.T) {
	entries := []navEntry{
		{Title: "Zebra", Order: 3},
		{Title: "Alpha", Order: 1},
		{Title: "Bravo", Order: 1},
		{Title: "Charlie", Order: 2},
		{Title: "Delta", Order: 9999},
	}
	sortNavEntries(entries)

	expected := []string{"Alpha", "Bravo", "Charlie", "Zebra", "Delta"}
	for i, e := range entries {
		if e.Title != expected[i] {
			t.Errorf("position %d: got %q, want %q", i, e.Title, expected[i])
		}
	}
}

func TestSortNavEntriesEmpty(t *testing.T) {
	sortNavEntries(nil)          // should not panic
	sortNavEntries([]navEntry{}) // should not panic
}

// --- ExportOpts validation ---

func TestExportOptsDefaults(t *testing.T) {
	opts := ExportOpts{}
	if opts.Format != "" {
		t.Error("default format should be empty")
	}
	if opts.SelfContained {
		t.Error("default SelfContained should be false")
	}
	if opts.CrossRef {
		t.Error("default CrossRef should be false")
	}
}

// --- zipDirectory ---

func TestZipDirectory(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure.
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>hello</html>"), 0644)
	os.WriteFile(filepath.Join(sub, "page.html"), []byte("<html>page</html>"), 0644)

	data, err := zipDirectory(dir)
	if err != nil {
		t.Fatalf("zipDirectory failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("zip should not be empty")
	}
	// Verify it's a valid zip (starts with PK signature).
	if data[0] != 'P' || data[1] != 'K' {
		t.Error("output should be a valid zip archive")
	}
}

func TestZipDirectoryEmpty(t *testing.T) {
	dir := t.TempDir()
	data, err := zipDirectory(dir)
	if err != nil {
		t.Fatalf("zipDirectory on empty dir failed: %v", err)
	}
	// Empty zip is still a valid zip (just the end-of-central-directory record).
	if len(data) == 0 {
		t.Error("even empty dir should produce valid zip bytes")
	}
}

// --- writeEmbeddedTemp ---

func TestWriteEmbeddedTemp(t *testing.T) {
	path := writeEmbeddedTemp("themes/paper.css")
	if path == "" {
		t.Fatal("should return a path for embedded paper.css")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("extracted file should exist: %v", err)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "Paper theme") {
		t.Error("extracted content should contain paper theme CSS")
	}

	// Call again — should return same path (idempotent).
	path2 := writeEmbeddedTemp("themes/paper.css")
	if path2 != path {
		t.Error("repeated calls should return the same path")
	}
}

func TestWriteEmbeddedTempNonexistent(t *testing.T) {
	path := writeEmbeddedTemp("themes/does-not-exist.css")
	if path != "" {
		t.Error("nonexistent embedded file should return empty")
	}
}

// --- Helpers ---

type memProvider struct {
	files map[string][]byte
}

func (p *memProvider) ReadFile(_ context.Context, path string) ([]byte, error) {
	data, ok := p.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

func (p *memProvider) ListFiles(_ context.Context, _ string) ([]string, error) {
	var paths []string
	for k := range p.files {
		paths = append(paths, k)
	}
	return paths, nil
}

func (p *memProvider) ResolveAsset(_ context.Context, _, asset string) (string, error) {
	return asset, nil
}

type mockExporter struct {
	formats []Format
}

func (e *mockExporter) Export(_ context.Context, _ ExportOpts) (*ExportResult, error) {
	return &ExportResult{Data: []byte("mock"), ContentType: "text/plain", Filename: "mock.txt"}, nil
}

func (e *mockExporter) Formats() []Format { return e.formats }

func generateLargeFrontmatter(n int) string {
	var b strings.Builder
	b.WriteString("---\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "key_%d: value_%d\n", i, i)
	}
	b.WriteString("---\n")
	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
