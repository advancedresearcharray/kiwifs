package importer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleBibTeX = `@article{smith2024attention,
  title = {Attention Mechanisms in Neural Networks},
  author = {Smith, John and Jones, Alice},
  year = {2024},
  journal = {NeurIPS},
  doi = {10.1234/example},
  abstract = {We present a survey of attention mechanisms.},
  keywords = {attention, neural-networks}
}

@inproceedings{doe2023ml,
  title = "Deep Learning {\'E}tudes",
  author = "Doe, Jane",
  booktitle = {ICML},
  year = 2023,
  pages = {1--10}
}

@book{knuth1984tex,
  author = {Knuth, Donald E.},
  title = {The {\TeX}book},
  publisher = {Addison-Wesley},
  year = {1984},
  isbn = {0-201-13448-9}
}
`

func TestBibTeXStream(t *testing.T) {
	bibPath := filepath.Join(t.TempDir(), "references.bib")
	if err := os.WriteFile(bibPath, []byte(sampleBibTeX), 0o644); err != nil {
		t.Fatal(err)
	}

	src, err := NewBibTeX(bibPath)
	if err != nil {
		t.Fatalf("new bibtex: %v", err)
	}
	defer src.Close()

	ch, errs := src.Stream(context.Background())
	var recs []Record
	for r := range ch {
		recs = append(recs, r)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream error: %v", err)
		}
	}

	if len(recs) != 3 {
		t.Fatalf("got %d records, want 3", len(recs))
	}

	article := recs[0]
	if article.PrimaryKey != "smith2024attention" {
		t.Fatalf("primary key=%q, want smith2024attention", article.PrimaryKey)
	}
	if article.Fields["bibtex_type"] != "article" {
		t.Fatalf("bibtex_type=%v, want article", article.Fields["bibtex_type"])
	}
	if article.Fields["title"] != "Attention Mechanisms in Neural Networks" {
		t.Fatalf("title=%v", article.Fields["title"])
	}
	authors, ok := article.Fields["authors"].([]string)
	if !ok || len(authors) != 2 || authors[0] != "Smith, John" {
		t.Fatalf("authors=%v", article.Fields["authors"])
	}
	if article.Fields["year"] != 2024 {
		t.Fatalf("year=%v, want 2024", article.Fields["year"])
	}
	if article.Fields["venue"] != "NeurIPS" {
		t.Fatalf("venue=%v, want NeurIPS", article.Fields["venue"])
	}
	if article.Fields["doi"] != "10.1234/example" {
		t.Fatalf("doi=%v", article.Fields["doi"])
	}
	tags, ok := article.Fields["tags"].([]string)
	if !ok || len(tags) != 2 {
		t.Fatalf("tags=%v", article.Fields["tags"])
	}

	raw, ok := article.Fields["_raw_content"].(string)
	if !ok {
		t.Fatal("missing _raw_content")
	}
	if !strings.Contains(raw, "bibtex_key: smith2024attention") {
		t.Fatalf("missing bibtex_key in raw content: %s", raw)
	}
	if !strings.Contains(raw, "# Attention Mechanisms in Neural Networks") {
		t.Fatalf("missing heading: %s", raw)
	}
	if !strings.Contains(raw, "Smith, John and Jones, Alice (2024). *NeurIPS*.") {
		t.Fatalf("missing citation line: %s", raw)
	}

	inproc := recs[1]
	if inproc.Fields["bibtex_type"] != "inproceedings" {
		t.Fatalf("bibtex_type=%v", inproc.Fields["bibtex_type"])
	}
	if inproc.Fields["venue"] != "ICML" {
		t.Fatalf("venue=%v, want ICML from booktitle", inproc.Fields["venue"])
	}
	if inproc.Fields["pages"] != "1--10" {
		t.Fatalf("pages=%v", inproc.Fields["pages"])
	}
	title, _ := inproc.Fields["title"].(string)
	if title != "Deep Learning Études" {
		t.Fatalf("title=%q, want LaTeX unescaped title", title)
	}

	book := recs[2]
	if book.Fields["bibtex_type"] != "book" {
		t.Fatalf("bibtex_type=%v", book.Fields["bibtex_type"])
	}
	if book.Fields["venue"] != "Addison-Wesley" {
		t.Fatalf("venue=%v, want publisher as venue", book.Fields["venue"])
	}
	if book.Fields["isbn"] != "0-201-13448-9" {
		t.Fatalf("isbn=%v", book.Fields["isbn"])
	}
}

func TestBibTeXImportPipeline(t *testing.T) {
	bibPath := filepath.Join(t.TempDir(), "refs.bib")
	if err := os.WriteFile(bibPath, []byte(sampleBibTeX), 0o644); err != nil {
		t.Fatal(err)
	}

	src, err := NewBibTeX(bibPath)
	if err != nil {
		t.Fatalf("new bibtex: %v", err)
	}
	defer src.Close()

	pipe, store := testPipeline(t)
	ctx := context.Background()
	stats, err := Run(ctx, src, pipe, Options{Actor: "test"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if stats.Imported != 3 {
		t.Fatalf("imported=%d, want 3", stats.Imported)
	}

	content, err := store.Read(ctx, "refs/smith2024attention.md")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	s := string(content)
	if strings.Contains(s, "_raw_content") {
		t.Fatalf("_raw_content should not appear in output: %s", s)
	}
	if !strings.Contains(s, "bibtex_key: smith2024attention") {
		t.Fatalf("missing bibtex_key: %s", s)
	}
	if !strings.Contains(s, "_source: refs") {
		t.Fatalf("missing _source tracking: %s", s)
	}
	if !strings.Contains(s, "authors:") {
		t.Fatalf("missing authors array: %s", s)
	}
}

func TestUnescapeBibTeX(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`Deep Learning \'Etudes`, "Deep Learning Études"},
		{`caf\'e`, "café"},
		{`100\%`, "100%"},
		{`a\_b`, "a_b"},
		{`line1\\line2`, `line1\line2`},
	}
	for _, tt := range tests {
		if got := unescapeBibTeX(tt.in); got != tt.want {
			t.Errorf("unescapeBibTeX(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseBibAuthors(t *testing.T) {
	got := parseBibAuthors("Smith, John and Jones, Alice")
	if len(got) != 2 || got[0] != "Smith, John" || got[1] != "Jones, Alice" {
		t.Fatalf("parseBibAuthors: %v", got)
	}
}

func TestBibTeXMissingFile(t *testing.T) {
	_, err := NewBibTeX(filepath.Join(t.TempDir(), "missing.bib"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
