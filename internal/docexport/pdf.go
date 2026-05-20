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
	inputPath, metadata, err := e.prepareInput(ctx, tmpDir, opts)
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

// prepareInput handles both single-file and multi-file (directory/book) inputs.
// It writes the processed markdown to tmpDir and returns the path to the input
// file, along with any metadata extracted from frontmatter or manifests.
func (e *PDFExporter) prepareInput(ctx context.Context, tmpDir string, opts ExportOpts) (string, map[string]string, error) {
	inputPath := opts.InputPath

	// Check if input is a directory (multi-file compilation).
	absInput := filepath.Join(e.root, inputPath)
	info, err := os.Stat(absInput)

	if err == nil && info.IsDir() {
		return e.prepareMultiFile(ctx, tmpDir, inputPath, opts)
	}

	// Single file.
	content, err := e.provider.ReadFile(ctx, inputPath)
	if err != nil {
		return "", nil, fmt.Errorf("read %s: %w", inputPath, err)
	}

	// Copy assets referenced in the markdown to tmpDir.
	content = e.resolveLocalAssets(ctx, inputPath, content, tmpDir)

	mdPath := filepath.Join(tmpDir, "input.md")
	if err := os.WriteFile(mdPath, content, 0644); err != nil {
		return "", nil, fmt.Errorf("write input: %w", err)
	}

	return mdPath, nil, nil
}

// prepareMultiFile handles directory-based multi-file compilation.
func (e *PDFExporter) prepareMultiFile(ctx context.Context, tmpDir, dirPath string, opts ExportOpts) (string, map[string]string, error) {
	// Check for manifest.
	absDir := filepath.Join(e.root, dirPath)
	manifest, err := LoadManifest(absDir)
	if err != nil {
		return "", nil, err
	}

	var paths []string
	if manifest != nil && len(manifest.Parts) > 0 {
		paths = manifest.Parts
	} else {
		// Auto-discover and order files.
		paths, err = e.provider.ListFiles(ctx, dirPath)
		if err != nil {
			return "", nil, err
		}
	}

	if len(paths) == 0 {
		return "", nil, fmt.Errorf("no markdown files found in %s", dirPath)
	}

	combined, metadata, err := StitchFiles(ctx, e.provider, paths, manifest)
	if err != nil {
		return "", nil, err
	}

	mdPath := filepath.Join(tmpDir, "input.md")
	if err := os.WriteFile(mdPath, combined, 0644); err != nil {
		return "", nil, fmt.Errorf("write combined: %w", err)
	}

	return mdPath, metadata, nil
}

// resolveLocalAssets copies referenced images to the temp directory and
// rewrites paths in the markdown. This ensures Pandoc can find them.
func (e *PDFExporter) resolveLocalAssets(ctx context.Context, sourcePath string, content []byte, tmpDir string) []byte {
	// For now, we rely on Pandoc's --resource-path to find assets.
	// This method is a hook for future enhancement where we could
	// copy assets into tmpDir for truly isolated builds.
	_ = ctx
	_ = sourcePath
	_ = tmpDir
	return content
}

// appendMetadataArgs adds -M key=value flags for Pandoc metadata.
func appendMetadataArgs(args []string, fromFile map[string]string, fromOpts map[string]string) []string {
	seen := make(map[string]bool)

	// File-level metadata takes precedence.
	for k, v := range fromFile {
		args = append(args, "-M", k+"="+v)
		seen[k] = true
	}
	for k, v := range fromOpts {
		if !seen[k] {
			args = append(args, "-M", k+"="+v)
		}
	}

	// Ensure date is always set.
	if !seen["date"] {
		hasDate := false
		for k := range fromOpts {
			if k == "date" {
				hasDate = true
				break
			}
		}
		if !hasDate {
			args = append(args, "-M", "date="+time.Now().Format("2006-01-02"))
		}
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
