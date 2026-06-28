// Package docexport renders markdown documents into publication-ready
// formats: PDF, standalone HTML, slide decks, and static documentation
// sites. It shells out to industry-standard tools (Pandoc, Marp, MkDocs)
// rather than reimplementing typesetting in Go.
package docexport

import (
	"context"
	"fmt"
	"strings"
)

// Format enumerates the supported output formats.
type Format string

const (
	FormatPDF    Format = "pdf"
	FormatHTML   Format = "html"
	FormatSlides Format = "slides"
	FormatSite   Format = "site"
)

// ParseFormat normalises a user-supplied format string.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "pdf":
		return FormatPDF, nil
	case "html":
		return FormatHTML, nil
	case "slides", "slide", "presentation":
		return FormatSlides, nil
	case "site", "docs", "mkdocs":
		return FormatSite, nil
	default:
		return "", fmt.Errorf("unsupported document export format %q (supported: pdf, html, slides, site)", s)
	}
}

// ExportOpts controls document export behaviour.
type ExportOpts struct {
	// Format selects the output format (pdf, html, slides, site).
	Format Format

	// InputPath is the file or directory to export.
	// A single .md file exports that file; a directory exports all
	// markdown files found within it according to ordering rules.
	InputPath string

	// Theme selects a named theme (paper, modern, minimal, dark, presentation).
	// Empty string uses the default theme for the format.
	Theme string

	// SelfContained embeds all resources (images, CSS, fonts) into a
	// single output file. Only applies to HTML and PDF formats.
	SelfContained bool

	// Bibliography is the path to a .bib / .json / .ris file.
	// Empty means no bibliography processing.
	Bibliography string

	// CSLStyle selects a citation style (apa, ieee, chicago, vancouver, harvard).
	// Empty defaults to "apa".
	CSLStyle string

	// CrossRef enables pandoc-crossref for numbered figures/tables/equations.
	CrossRef bool

	// PDFEngine selects the PDF engine for Pandoc (typst, xelatex, weasyprint).
	// Empty defaults to "typst" if available, then "xelatex".
	PDFEngine string

	// SlideFormat selects slide output: "html" (default), "pdf", or "pptx".
	SlideFormat string

	// Metadata is extra key-value pairs injected into Pandoc metadata
	// (title, author, date, lang, etc.). Frontmatter values take precedence.
	Metadata map[string]string

	// SiteConfig overrides for MkDocs site generation.
	SiteName string
	SiteURL  string
	RepoURL  string
}

// ExportResult holds the output from a successful export.
type ExportResult struct {
	// Data contains the rendered output bytes (PDF, HTML, PPTX, or zip for site).
	Data []byte

	// ContentType is the MIME type of the output.
	ContentType string

	// Filename is a suggested download filename.
	Filename string
}

// Exporter renders markdown into a publication-ready format.
type Exporter interface {
	// Export renders the input document(s) and returns the result.
	Export(ctx context.Context, opts ExportOpts) (*ExportResult, error)

	// Formats returns the formats this exporter supports.
	Formats() []Format
}

// FileProvider abstracts filesystem access so exporters can read
// files from the KiwiFS storage layer.
type FileProvider interface {
	// ReadFile returns the content of a file at the given path.
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// ListFiles returns all markdown file paths under a directory,
	// sorted by the canonical ordering (frontmatter order > filename).
	ListFiles(ctx context.Context, dir string) ([]string, error)

	// ResolveAsset resolves a relative asset path (image, bib file)
	// to its absolute filesystem path.
	ResolveAsset(ctx context.Context, relativeTo, asset string) (string, error)
}

// Registry holds format-to-exporter mappings and dispatches export requests.
type Registry struct {
	exporters map[Format]Exporter
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{exporters: make(map[Format]Exporter)}
}

// Register adds an exporter for one or more formats.
func (r *Registry) Register(e Exporter) {
	for _, f := range e.Formats() {
		r.exporters[f] = e
	}
}

// Export dispatches to the appropriate exporter.
func (r *Registry) Export(ctx context.Context, opts ExportOpts) (*ExportResult, error) {
	e, ok := r.exporters[opts.Format]
	if !ok {
		return nil, fmt.Errorf("no exporter registered for format %q", opts.Format)
	}
	return e.Export(ctx, opts)
}

// SupportedFormats returns all registered formats.
func (r *Registry) SupportedFormats() []Format {
	formats := make([]Format, 0, len(r.exporters))
	for f := range r.exporters {
		formats = append(formats, f)
	}
	return formats
}
