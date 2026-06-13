package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintFlagsOrphanAndBrokenLinks(t *testing.T) {
	root := t.TempDir()
	// SCHEMA.md references a missing page (orphan).
	if err := os.WriteFile(filepath.Join(root, "SCHEMA.md"),
		[]byte("# Schema\n\nExpected: [[index]] and [[missing-page]]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// index.md exists; concepts/a.md links to a broken target.
	if err := os.WriteFile(filepath.Join(root, "index.md"),
		[]byte("# Index\n\nsee [[concepts/a]]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "concepts"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "concepts/a.md"),
		[]byte("# A\n\nsee [[nowhere]]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Lint(root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	kinds := map[string]int{}
	for _, is := range res.Issues {
		kinds[is.Kind]++
	}
	if kinds["orphan"] < 1 {
		t.Fatalf("expected orphan issue, got %v", kinds)
	}
	if kinds["broken-link"] < 1 {
		t.Fatalf("expected broken-link issue, got %v", kinds)
	}
}

func TestLintIgnoresWikiLinksInFencedCode(t *testing.T) {
	root := t.TempDir()
	content := `# Example TOML Configuration

` + "```toml\n[server]\nhost = \"localhost\"\n\n[[routes]]\npath = \"/api\"\n```\n"
	if err := os.WriteFile(filepath.Join(root, "SCHEMA.md"), []byte("# Schema\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "clips"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "clips/test.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Lint(root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	for _, is := range res.Issues {
		if is.Kind == "broken-link" {
			t.Fatalf("unexpected broken-link for code-block syntax: %+v", is)
		}
	}
}

func TestLintIgnoresWikiLinksInIndentedAndInlineCode(t *testing.T) {
	root := t.TempDir()
	content := "# Indented example\n\n    [[indented-routes]]\n\nUse " + "`[[inline-routes]]`" + " in prose.\n"
	if err := os.WriteFile(filepath.Join(root, "SCHEMA.md"), []byte("# Schema\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "indented.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Lint(root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	for _, is := range res.Issues {
		if is.Kind == "broken-link" {
			t.Fatalf("unexpected broken-link for indented/inline code syntax: %+v", is)
		}
	}
}

func TestLintMissingSchema(t *testing.T) {
	root := t.TempDir()
	res, err := Lint(root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	found := false
	for _, is := range res.Issues {
		if is.Kind == "missing-schema" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected missing-schema issue, got %v", res.Issues)
	}
}
