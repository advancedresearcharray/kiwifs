package importer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kiwifs/kiwifs/internal/pipeline"
)

// IngestOptions controls the document ingestion pipeline.
type IngestOptions struct {
	SplitMode        string // "single" (one big file) or "sections" (one file per top-level heading)
	Prefix           string // output path prefix in kiwifs (e.g. "imports/financial-report/")
	ExtractKeywords  bool
	MaxKeywords      int
	ConvertCrossRefs bool
	Actor            string
}

// IngestResult summarizes what was produced.
type IngestResult struct {
	SourceFile  string   `json:"source_file"`
	Format      string   `json:"format"`
	TotalPages  int      `json:"total_sections"`
	OutputFiles []string `json:"output_files"`
	Keywords    []string `json:"top_keywords"`
}

// IngestSection is a parsed section from the converted markdown.
type IngestSection struct {
	Heading  string
	Level    int
	Content  string
	Keywords []string
}

// Ingest converts a document to markdown via MarkItDown, then runs the
// post-processing pipeline: section splitting, TF-IDF keywords,
// cross-reference detection, and frontmatter generation.
func Ingest(ctx context.Context, filePath string, pipe *pipeline.Pipeline, opts IngestOptions) (*IngestResult, error) {
	raw, err := ConvertWithMarkItDown(ctx, filePath)
	if err != nil {
		return nil, err
	}

	sections := SplitByHeadings([]byte(raw.Markdown))
	if len(sections) == 0 {
		sections = []IngestSection{{
			Heading: strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
			Level:   1,
			Content: raw.Markdown,
		}}
	}

	var allKeywords []string
	if opts.ExtractKeywords {
		sectionTexts := make([]string, len(sections))
		for i, s := range sections {
			sectionTexts[i] = s.Content
		}
		corpusDF := BuildCorpusDF(sectionTexts)
		maxKW := opts.MaxKeywords
		if maxKW == 0 {
			maxKW = 10
		}
		for i := range sections {
			kw := ExtractKeywords(sections[i].Content, corpusDF, len(sections), maxKW)
			sections[i].Keywords = kw
			allKeywords = append(allKeywords, kw...)
		}
	}

	if opts.ConvertCrossRefs {
		docTitle := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		prefix := opts.Prefix
		if prefix == "" {
			prefix = "imports/" + slugify(docTitle) + "/"
		}
		sectionMap := buildSectionMap(sections, prefix)
		for i := range sections {
			sections[i].Content = ConvertCrossRefs(sections[i].Content, sectionMap)
		}
	}

	docTitle := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	prefix := opts.Prefix
	if prefix == "" {
		prefix = "imports/" + slugify(docTitle) + "/"
	}

	actor := opts.Actor
	if actor == "" {
		actor = "kiwi-ingest"
	}

	var outputFiles []string

	if opts.SplitMode == "sections" && len(sections) > 1 {
		for _, section := range sections {
			path := prefix + slugify(section.Heading) + ".md"
			content := GenerateMarkdownFromSection(section, docTitle)
			if pipe != nil {
				if _, werr := pipe.Write(ctx, path, []byte(content), actor); werr != nil {
					return nil, fmt.Errorf("write %s: %w", path, werr)
				}
			}
			outputFiles = append(outputFiles, path)
		}
	} else {
		path := prefix + slugify(docTitle) + ".md"
		content := GenerateMarkdownSingleFile(sections, docTitle, allKeywords)
		if pipe != nil {
			if _, werr := pipe.Write(ctx, path, []byte(content), actor); werr != nil {
				return nil, fmt.Errorf("write %s: %w", path, werr)
			}
		}
		outputFiles = append(outputFiles, path)
	}

	return &IngestResult{
		SourceFile:  filePath,
		Format:      filepath.Ext(filePath),
		TotalPages:  len(sections),
		OutputFiles: outputFiles,
		Keywords:    dedup(allKeywords),
	}, nil
}

// SplitByHeadings splits markdown into sections at heading boundaries.
func SplitByHeadings(body []byte) []IngestSection {
	lines := strings.Split(string(body), "\n")
	var sections []IngestSection
	var current *IngestSection

	for _, line := range lines {
		level := ingestHeadingLevel(line)
		if level > 0 {
			if current != nil {
				current.Content = strings.TrimSpace(current.Content)
				sections = append(sections, *current)
			}
			title := strings.TrimSpace(strings.TrimLeft(line, "#"))
			current = &IngestSection{
				Heading: title,
				Level:   level,
			}
		} else if current != nil {
			current.Content += line + "\n"
		}
	}
	if current != nil {
		current.Content = strings.TrimSpace(current.Content)
		sections = append(sections, *current)
	}

	return sections
}

func ingestHeadingLevel(line string) int {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return 0
	}
	level := 0
	for _, ch := range trimmed {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level > 0 && level < len(trimmed) && trimmed[level] == ' ' {
		return level
	}
	return 0
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}
