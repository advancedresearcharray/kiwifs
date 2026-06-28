package importer

import (
	"fmt"
	"strings"
)

// GenerateMarkdownFromSection generates a single section file with frontmatter.
func GenerateMarkdownFromSection(section IngestSection, docTitle string) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("source: %q\n", docTitle))
	if len(section.Keywords) > 0 {
		sb.WriteString(fmt.Sprintf("keywords: [%s]\n", strings.Join(section.Keywords, ", ")))
	}
	wordCount := len(strings.Fields(section.Content))
	sb.WriteString(fmt.Sprintf("word_count: %d\n", wordCount))
	sb.WriteString("---\n\n")

	level := strings.Repeat("#", section.Level)
	sb.WriteString(fmt.Sprintf("%s %s\n\n", level, section.Heading))
	sb.WriteString(section.Content)
	sb.WriteString("\n")

	return sb.String()
}

// GenerateMarkdownSingleFile generates one big markdown file with all sections.
func GenerateMarkdownSingleFile(sections []IngestSection, docTitle string, topKeywords []string) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("source: %q\n", docTitle))
	sb.WriteString(fmt.Sprintf("sections: %d\n", len(sections)))
	if len(topKeywords) > 0 {
		max := 15
		if len(topKeywords) < max {
			max = len(topKeywords)
		}
		sb.WriteString(fmt.Sprintf("keywords: [%s]\n", strings.Join(topKeywords[:max], ", ")))
	}
	sb.WriteString("---\n\n")

	sb.WriteString(fmt.Sprintf("# %s\n\n", docTitle))

	for _, section := range sections {
		level := strings.Repeat("#", section.Level)
		sb.WriteString(fmt.Sprintf("%s %s\n\n", level, section.Heading))
		sb.WriteString(section.Content)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func dedup(ss []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
