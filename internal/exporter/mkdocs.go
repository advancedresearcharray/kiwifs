package exporter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/storage"
	"gopkg.in/yaml.v3"
)

// MkDocsOptions configures static MkDocs tree export.
type MkDocsOptions struct {
	OutputDir string
	SiteName  string
	SiteURL   string
}

var wikiLinkRe = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)

// ConvertWikiLinkForMkDocs rewrites [[target|label]] to [label](relative-path.md).
func ConvertWikiLinkForMkDocs(content, targetPath string) string {
	return wikiLinkRe.ReplaceAllStringFunc(content, func(match string) string {
		sub := wikiLinkRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		target := strings.TrimSpace(sub[1])
		label := target
		if len(sub) >= 3 && sub[2] != "" {
			label = strings.TrimSpace(sub[2])
		}
		linkPath := targetPathForWikiTarget(target, targetPath)
		if linkPath == "" {
			return match
		}
		return fmt.Sprintf("[%s](%s)", label, linkPath)
	})
}

func targetPathForWikiTarget(target, sourcePath string) string {
	base := strings.TrimSuffix(strings.TrimSpace(target), ".md")
	if base == "" {
		return ""
	}
	targetFile := base + ".md"
	sourceDir := filepath.Dir(sourcePath)
	if sourceDir == "." || sourceDir == "" {
		return targetFile
	}
	targetInDir := filepath.Join(sourceDir, targetFile)
	if sourcePath != targetInDir {
		return targetFile
	}
	return targetFile
}

type mkdocsNavEntry struct {
	title string
	path  string
	order int
}

func mkdocsExtractOrder(fm map[string]any) int {
	for _, key := range []string{"nav_order", "order"} {
		if v, ok := fm[key]; ok {
			switch n := v.(type) {
			case int:
				return n
			case float64:
				return int(n)
			}
		}
	}
	return -1
}

func sortMkdocsNav(entries []mkdocsNavEntry) {
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].order < entries[i].order ||
				(entries[j].order == entries[i].order && entries[j].title < entries[i].title) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

// ExportMkDocs writes docs/ + mkdocs.yml under opts.OutputDir.
func ExportMkDocs(ctx context.Context, store storage.Storage, opts MkDocsOptions) (int, error) {
	if opts.OutputDir == "" {
		return 0, fmt.Errorf("output dir required")
	}
	docsDir := filepath.Join(opts.OutputDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		return 0, err
	}

	var nav []mkdocsNavEntry
	count := 0

	err := storage.Walk(ctx, store, "/", func(entry storage.Entry) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !strings.HasSuffix(strings.ToLower(entry.Path), ".md") {
			return nil
		}
		base := filepath.Base(entry.Path)
		if strings.HasPrefix(base, ".") || strings.Contains(entry.Path, "/.kiwi/") {
			return nil
		}

		content, err := store.Read(ctx, entry.Path)
		if err != nil {
			return nil
		}

		parsed, _ := markdown.Parse(content)
		title := strings.TrimSuffix(base, ".md")
		order := 9999
		if parsed.Frontmatter != nil {
			if t, ok := parsed.Frontmatter["title"].(string); ok && t != "" {
				title = t
			}
			if o := mkdocsExtractOrder(parsed.Frontmatter); o >= 0 {
				order = o
			}
		}

		rel := strings.TrimPrefix(entry.Path, "/")
		fm, bodyBytes, fmErr := markdown.SplitFrontmatter(content)
		body := string(bodyBytes)
		if fmErr != nil {
			body = string(content)
			fm = nil
		}
		converted := ConvertWikiLinkForMkDocs(body, rel)
		var outBytes []byte
		if fm != nil {
			outBytes = append(outBytes, []byte("---\n")...)
			outBytes = append(outBytes, fm...)
			outBytes = append(outBytes, []byte("---\n")...)
		}
		outBytes = append(outBytes, []byte(converted)...)

		destPath := filepath.Join(docsDir, rel)
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(destPath), err)
		}
		if err := os.WriteFile(destPath, outBytes, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", destPath, err)
		}

		nav = append(nav, mkdocsNavEntry{title: title, path: rel, order: order})
		count++
		return nil
	})
	if err != nil {
		return count, err
	}

	sortMkdocsNav(nav)

	siteName := opts.SiteName
	if siteName == "" {
		siteName = "Knowledge Base"
	}
	cfg := map[string]any{
		"site_name": siteName,
		"theme": map[string]any{
			"name": "material",
		},
		"plugins": []string{"search"},
	}
	if opts.SiteURL != "" {
		cfg["site_url"] = opts.SiteURL
	}
	if len(nav) > 0 {
		navYaml := make([]map[string]string, len(nav))
		for i, n := range nav {
			navYaml[i] = map[string]string{n.title: n.path}
		}
		cfg["nav"] = navYaml
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return count, err
	}
	if err := os.WriteFile(filepath.Join(opts.OutputDir, "mkdocs.yml"), out, 0o644); err != nil {
		return count, err
	}
	return count, nil
}
