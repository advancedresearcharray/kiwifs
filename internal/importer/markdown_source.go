package importer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiwifs/kiwifs/internal/markdown"
)

// MarkdownOpts configures the MarkdownSource behavior.
type MarkdownOpts struct {
	// NonRecursive, when true, skips subdirectories. Default is recursive.
	NonRecursive bool
	// SkipAssets, when true, ignores non-markdown files. Default copies them.
	SkipAssets bool
}

// MarkdownSource imports markdown files from a file or directory.
// Unlike ObsidianSource, it does not rewrite any syntax - files are
// imported as-is with their existing frontmatter and body.
type MarkdownSource struct {
	path   string // file or directory path
	isFile bool   // true if path is a single file
	opts   MarkdownOpts
	files  []mdFile
	assets []mdAsset
}

type mdFile struct {
	relPath string
	content []byte
	fm      map[string]any
}

type mdAsset struct {
	srcPath string
	relDest string
}

// NewMarkdown creates a source for importing markdown files.
// The path can be a single .md file or a directory of markdown files.
func NewMarkdown(path string, opts MarkdownOpts) (*MarkdownSource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("markdown path: %w", err)
	}

	s := &MarkdownSource{path: path, opts: opts}

	if info.IsDir() {
		if err := s.walkDir(); err != nil {
			return nil, err
		}
	} else {
		s.isFile = true
		if err := s.readSingleFile(path); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *MarkdownSource) readSingleFile(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".md" && ext != ".mdx" {
		return fmt.Errorf("not a markdown file: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	fm, _ := markdown.Frontmatter(data)
	if fm == nil {
		fm = map[string]any{}
	}

	// Use filename without extension as the relative path
	baseName := filepath.Base(path)
	relPath := strings.TrimSuffix(baseName, ext)

	s.files = append(s.files, mdFile{
		relPath: relPath,
		content: data,
		fm:      fm,
	})
	return nil
}

func (s *MarkdownSource) Name() string {
	base := filepath.Base(s.path)
	if s.isFile {
		// Strip extension for single file
		ext := filepath.Ext(base)
		return strings.TrimSuffix(base, ext)
	}
	return base
}

func (s *MarkdownSource) walkDir() error {
	return filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(s.path, path)

		if info.IsDir() {
			// Skip the root directory itself
			if path == s.path {
				return nil
			}
			base := filepath.Base(path)
			// Skip hidden directories and common internal folders
			if strings.HasPrefix(base, ".") || base == "node_modules" {
				return filepath.SkipDir
			}
			// Skip subdirectories if NonRecursive is set
			if s.opts.NonRecursive {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".md" || ext == ".mdx" {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", rel, err)
			}

			fm, _ := markdown.Frontmatter(data)
			if fm == nil {
				fm = map[string]any{}
			}

			// Strip the .md extension from the relative path
			relPath := strings.TrimSuffix(rel, ext)

			s.files = append(s.files, mdFile{
				relPath: relPath,
				content: data,
				fm:      fm,
			})
		} else if !s.opts.SkipAssets && isAssetFile(path) {
			s.assets = append(s.assets, mdAsset{
				srcPath: path,
				relDest: filepath.Join("assets", filepath.Base(path)),
			})
		}

		return nil
	})
}

func (s *MarkdownSource) Stream(ctx context.Context) (<-chan Record, <-chan error) {
	records := make(chan Record, 64)
	errs := make(chan error, 1)

	go func() {
		defer close(records)
		defer close(errs)

		name := s.Name()
		for i, f := range s.files {
			if ctx.Err() != nil {
				return
			}

			fields := make(map[string]any, len(f.fm)+1)
			for k, v := range f.fm {
				fields[k] = v
			}
			// Store the full original content - Run() will use renderRawContent
			fields["_raw_content"] = string(f.content)

			pk := sanitizePath(f.relPath)

			rec := Record{
				SourceID:   fmt.Sprintf("markdown:%s:%d", name, i),
				SourceDSN:  s.path,
				Table:      name,
				Fields:     fields,
				PrimaryKey: pk,
			}
			select {
			case records <- rec:
			case <-ctx.Done():
				return
			}
		}
	}()
	return records, errs
}

func (s *MarkdownSource) Close() error { return nil }

// Assets returns the list of non-markdown files found that should be copied
// to the assets/ folder. The caller (importer.Run or a wrapper) can use this
// to copy assets after the import completes.
func (s *MarkdownSource) Assets() []mdAsset {
	return s.assets
}

func isAssetFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".bmp", ".ico",
		".pdf", ".mp3", ".mp4", ".wav", ".webm", ".ogg", ".mov",
		".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return true
	}
	return false
}
