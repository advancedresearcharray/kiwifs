package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestInferSchema_SaveSchema_WritesFile(t *testing.T) {
	dir := t.TempDir()
	prev, _ := os.Getwd()
	_ = os.Chdir(dir)
	t.Cleanup(func() { _ = os.Chdir(prev) })

	if err := os.WriteFile("data.csv", []byte("id,name\n1,Alice\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	c := &cobra.Command{}
	c.Flags().String("file", "", "")
	c.Flags().Bool("save-schema", false, "")
	_ = c.Flags().Set("file", "data.csv")
	_ = c.Flags().Set("save-schema", "true")

	if err := runInferSchema(c, "csv"); err != nil {
		t.Fatalf("runInferSchema: %v", err)
	}

	outPath := filepath.Join(".kiwi", "schemas", "data.json")
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected schema written at %s: %v", outPath, err)
	}
}

func TestInferSchema_SaveSchema_AbsolutePath(t *testing.T) {
	srcDir := t.TempDir()
	csvPath := filepath.Join(srcDir, "sales.csv")
	if err := os.WriteFile(csvPath, []byte("product,price,qty\nWidget,9.99,100\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	workDir := t.TempDir()
	prev, _ := os.Getwd()
	_ = os.Chdir(workDir)
	t.Cleanup(func() { _ = os.Chdir(prev) })

	c := &cobra.Command{}
	c.Flags().String("file", "", "")
	c.Flags().Bool("save-schema", false, "")
	_ = c.Flags().Set("file", csvPath)
	_ = c.Flags().Set("save-schema", "true")

	if err := runInferSchema(c, "csv"); err != nil {
		t.Fatalf("runInferSchema with absolute path: %v", err)
	}

	outPath := filepath.Join(".kiwi", "schemas", "sales.json")
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected schema at %s (basename only, not full path): %v", outPath, err)
	}
}

func TestBuildSource_BibTeXRequiresFile(t *testing.T) {
	c := &cobra.Command{}
	c.Flags().String("file", "", "")
	_, err := buildSource(c, "bibtex")
	if err == nil || err.Error() != "--file is required for bibtex" {
		t.Fatalf("buildSource(bibtex) err = %v, want --file is required for bibtex", err)
	}
}

func TestBuildSource_BibTeXMissingFile(t *testing.T) {
	c := &cobra.Command{}
	c.Flags().String("file", "", "")
	_ = c.Flags().Set("file", filepath.Join(t.TempDir(), "missing.bib"))
	_, err := buildSource(c, "bibtex")
	if err == nil {
		t.Fatal("expected error for missing bibtex file")
	}
}

func TestBuildSource_BibTeX(t *testing.T) {
	bibPath := filepath.Join(t.TempDir(), "refs.bib")
	if err := os.WriteFile(bibPath, []byte(`@article{a, title={T}, author={A}, year={2024}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	c := &cobra.Command{}
	c.Flags().String("file", "", "")
	_ = c.Flags().Set("file", bibPath)

	src, err := buildSource(c, "bibtex")
	if err != nil {
		t.Fatalf("buildSource(bibtex): %v", err)
	}
	if src.Name() != "refs" {
		t.Fatalf("Name() = %q, want refs", src.Name())
	}
	_ = src.Close()
}
