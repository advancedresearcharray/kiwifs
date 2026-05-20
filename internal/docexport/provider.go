package docexport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/storage"
)

// StorageProvider implements FileProvider using a KiwiFS Storage backend.
type StorageProvider struct {
	store storage.Storage
	root  string // filesystem root for resolving asset paths
}

// NewStorageProvider creates a FileProvider backed by KiwiFS storage.
func NewStorageProvider(store storage.Storage, root string) *StorageProvider {
	return &StorageProvider{store: store, root: root}
}

// ReadFile reads a markdown file from the storage layer.
func (p *StorageProvider) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return p.store.Read(ctx, path)
}

// orderedFile is a markdown file with its sort key for multi-file compilation.
type orderedFile struct {
	Path  string
	Order int
	Name  string
}

// ListFiles returns all .md files under a directory, sorted by:
// 1. Frontmatter "order" field (if present)
// 2. Numeric filename prefix (e.g. 01-intro.md)
// 3. Alphabetical filename
func (p *StorageProvider) ListFiles(ctx context.Context, dir string) ([]string, error) {
	var files []orderedFile

	err := storage.Walk(ctx, p.store, dir, func(e storage.Entry) error {
		if !strings.HasSuffix(strings.ToLower(e.Path), ".md") {
			return nil
		}
		// Skip hidden files and .kiwi directory.
		base := filepath.Base(e.Path)
		if strings.HasPrefix(base, ".") || strings.Contains(e.Path, "/.kiwi/") {
			return nil
		}

		of := orderedFile{
			Path:  e.Path,
			Order: 9999, // default: sort last
			Name:  base,
		}

		// Try to extract order from frontmatter.
		content, err := p.store.Read(ctx, e.Path)
		if err == nil {
			parsed, _ := markdown.Parse(content)
			if parsed.Frontmatter != nil {
				if o, ok := parsed.Frontmatter["order"]; ok {
					switch v := o.(type) {
					case int:
						of.Order = v
					case float64:
						of.Order = int(v)
					case string:
						if n, err := strconv.Atoi(v); err == nil {
							of.Order = n
						}
					}
				}
				if o, ok := parsed.Frontmatter["nav_order"]; ok {
					switch v := o.(type) {
					case int:
						of.Order = v
					case float64:
						of.Order = int(v)
					case string:
						if n, err := strconv.Atoi(v); err == nil {
							of.Order = n
						}
					}
				}
			}
		}

		// Try to extract numeric prefix from filename (e.g. "01-intro.md").
		if of.Order == 9999 {
			if idx := strings.IndexByte(base, '-'); idx > 0 {
				if n, err := strconv.Atoi(base[:idx]); err == nil {
					of.Order = n
				}
			}
		}

		files = append(files, of)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, err)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Order != files[j].Order {
			return files[i].Order < files[j].Order
		}
		return files[i].Name < files[j].Name
	})

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	return paths, nil
}

// ResolveAsset resolves a relative asset path to an absolute filesystem path.
func (p *StorageProvider) ResolveAsset(_ context.Context, relativeTo, asset string) (string, error) {
	// If asset is already absolute, use it directly.
	if filepath.IsAbs(asset) {
		if _, err := os.Stat(asset); err != nil {
			return "", fmt.Errorf("asset not found: %s", asset)
		}
		return asset, nil
	}

	// Resolve relative to the directory of the source file.
	dir := filepath.Dir(relativeTo)
	candidate := filepath.Join(p.root, dir, asset)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	// Try from root.
	candidate = filepath.Join(p.root, asset)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	// Try assets directory.
	candidate = filepath.Join(p.root, ".kiwi", "assets", asset)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	return "", fmt.Errorf("asset %q not found relative to %q", asset, relativeTo)
}
