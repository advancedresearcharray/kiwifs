package exporter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/storage"
)

func TestConvertWikiLinkForMkDocs(t *testing.T) {
	in := "See [[getting-started|Start here]] for details."
	got := ConvertWikiLinkForMkDocs(in, "guides/index.md")
	if !strings.Contains(got, "[Start here](getting-started.md)") {
		t.Fatalf("got %q", got)
	}
}

func TestExportMkDocsSampleWorkspace(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store, err := storage.NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Write(ctx, "pages/hello.md", []byte("---\ntitle: Hello\n---\n# Hello\n\nSee [[world]]\n")); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(ctx, "pages/world.md", []byte("---\ntitle: World\n---\n# World\n")); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(t.TempDir(), "site")
	count, err := ExportMkDocs(ctx, store, MkDocsOptions{
		OutputDir: outDir,
		SiteName:  "Test KB",
		SiteURL:   "https://example.com",
	})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if count < 2 {
		t.Fatalf("count=%d want >=2", count)
	}
	if _, err := os.Stat(filepath.Join(outDir, "mkdocs.yml")); err != nil {
		t.Fatalf("mkdocs.yml: %v", err)
	}
	helloPath := filepath.Join(outDir, "docs", "pages", "hello.md")
	if _, err := os.Stat(helloPath); err != nil {
		t.Fatalf("hello.md: %v", err)
	}
	body, err := os.ReadFile(helloPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "[world](world.md)") {
		t.Fatalf("wiki link not converted in export: %q", string(body))
	}
}
