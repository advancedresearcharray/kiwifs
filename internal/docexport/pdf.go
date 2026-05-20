package docexport

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PDFExporter renders markdown to PDF using Pandoc.
type PDFExporter struct {
	provider FileProvider
	root     string
}

// NewPDFExporter creates a PDF exporter backed by the given file provider.
func NewPDFExporter(provider FileProvider, root string) *PDFExporter {
	return &PDFExporter{provider: provider, root: root}
}

// Formats returns the formats this exporter handles.
func (e *PDFExporter) Formats() []Format { return []Format{FormatPDF} }

// Export renders markdown to PDF via Pandoc.
func (e *PDFExporter) Export(ctx context.Context, opts ExportOpts) (*ExportResult, error) {
	if err := RequireTool("Pandoc", "pandoc"); err != nil {
		return nil, err
	}

	// Determine PDF engine.
	engine := opts.PDFEngine
	if engine == "" {
		// Prefer typst (lighter weight, faster), fall back to xelatex.
		if IsAvailable("typst") {
			engine = "typst"
		} else if IsAvailable("xelatex") {
			engine = "xelatex"
		} else {
			return nil, fmt.Errorf("no PDF engine available; install typst or xelatex")
		}
	}

	// Create temp working directory.
	tmpDir, err := os.MkdirTemp("", "kiwi-pdf-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Prepare input: single file or multi-file compilation.
	inputPath, metadata, err := prepareInputGeneric(ctx, e.provider, e.root, tmpDir, opts)
	if err != nil {
		return nil, fmt.Errorf("prepare input: %w", err)
	}

	outputPath := filepath.Join(tmpDir, "output.pdf")

	// Build Pandoc command.
	args := []string{
		inputPath,
		"-o", outputPath,
		"--pdf-engine=" + engine,
		"--standalone",
	}

	// Apply theme template.
	if tmpl := resolveTemplate(opts.Theme, FormatPDF, engine); tmpl != "" {
		args = append(args, "--template="+tmpl)
	}

	// Inject metadata.
	args = appendMetadataArgs(args, metadata, opts.Metadata)

	// Bibliography support.
	args = appendBibliographyArgs(args, opts, e.root)

	// Cross-reference support.
	if opts.CrossRef && IsAvailable("pandoc-crossref") {
		args = append(args, "--filter", "pandoc-crossref")
	}

	// Resource path for images.
	resourcePath := e.root
	if opts.InputPath != "" {
		resourcePath = filepath.Join(e.root, filepath.Dir(opts.InputPath))
	}
	args = append(args, "--resource-path="+resourcePath+":"+e.root)

	// Table of contents for multi-file/long documents.
	args = append(args, "--toc")

	// Execute Pandoc with timeout.
	cmd := exec.CommandContext(ctx, "pandoc", args...)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "HOME="+os.TempDir())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pandoc failed: %w\noutput: %s", err, string(output))
	}

	// Read the PDF.
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}

	filename := suggestFilename(opts.InputPath, ".pdf")

	return &ExportResult{
		Data:        data,
		ContentType: "application/pdf",
		Filename:    filename,
	}, nil
}


// appendMetadataArgs adds -M key=value flags for Pandoc metadata.
// File-level metadata takes precedence over ExportOpts metadata.
// A date is always injected if neither source provides one.
func appendMetadataArgs(args []string, fromFile map[string]string, fromOpts map[string]string) []string {
	seen := make(map[string]bool)

	for k, v := range fromFile {
		args = append(args, "-M", k+"="+v)
		seen[k] = true
	}
	for k, v := range fromOpts {
		if !seen[k] {
			args = append(args, "-M", k+"="+v)
			seen[k] = true
		}
	}

	if !seen["date"] {
		args = append(args, "-M", "date="+time.Now().Format("2006-01-02"))
	}

	return args
}

// appendBibliographyArgs adds bibliography and CSL arguments if configured.
func appendBibliographyArgs(args []string, opts ExportOpts, root string) []string {
	bib := opts.Bibliography
	if bib == "" {
		// Auto-detect bibliography files.
		for _, candidate := range []string{
			filepath.Join(root, ".kiwi", "references", "refs.bib"),
			filepath.Join(root, ".kiwi", "references", "references.bib"),
			filepath.Join(root, "refs.bib"),
			filepath.Join(root, "references.bib"),
		} {
			if _, err := os.Stat(candidate); err == nil {
				bib = candidate
				break
			}
		}
	} else if !filepath.IsAbs(bib) {
		bib = filepath.Join(root, bib)
	}

	if bib == "" {
		return args
	}

	args = append(args, "--citeproc", "--bibliography="+bib)

	// CSL style.
	csl := resolveCSLStyle(opts.CSLStyle, root)
	if csl != "" {
		args = append(args, "--csl="+csl)
	}

	return args
}

// suggestFilename generates a download filename from the input path.
func suggestFilename(inputPath, ext string) string {
	if inputPath == "" {
		return "export" + ext
	}
	base := filepath.Base(inputPath)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" || base == "." {
		base = filepath.Base(filepath.Dir(inputPath))
	}
	if base == "" || base == "." {
		base = "export"
	}
	return base + ext
}
