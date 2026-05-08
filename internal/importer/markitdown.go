package importer

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// MarkItDownResult holds the raw markdown output from MarkItDown.
type MarkItDownResult struct {
	Markdown   string
	SourceFile string
}

// ConvertWithMarkItDown shells out to the `markitdown` CLI to convert
// a file to markdown. Requires `pip install markitdown[all]`.
func ConvertWithMarkItDown(ctx context.Context, filePath string) (*MarkItDownResult, error) {
	if _, err := exec.LookPath("markitdown"); err != nil {
		return nil, fmt.Errorf("markitdown not found in PATH — install with: pip install 'markitdown[all]'")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "markitdown", filePath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("markitdown failed: %v — stderr: %s", err, stderr.String())
	}

	return &MarkItDownResult{
		Markdown:   stdout.String(),
		SourceFile: filePath,
	}, nil
}

var markItDownFormats = []string{
	"pdf", "docx", "pptx", "xlsx",
	"html", "htm",
	"csv", "json", "xml",
	"epub",
	"jpg", "jpeg", "png", "gif", "bmp", "tiff",
	"mp3", "wav", "m4a",
	"zip",
}

// IsMarkItDownFormat returns true if the extension is supported by MarkItDown.
func IsMarkItDownFormat(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	for _, f := range markItDownFormats {
		if ext == f {
			return true
		}
	}
	return false
}
