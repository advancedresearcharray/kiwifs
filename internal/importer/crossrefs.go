package importer

import (
	"fmt"
	"regexp"
	"strings"
)

var crossRefPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)see\s+section\s+(\d+(?:\.\d+)*)`),
	regexp.MustCompile(`(?i)(?:described|discussed|detailed|explained)\s+in\s+section\s+(\d+(?:\.\d+)*)`),
	regexp.MustCompile(`(?i)chapter\s+(\d+)`),
	regexp.MustCompile(`(?i)appendix\s+([A-Z])`),
	regexp.MustCompile(`(?i)table\s+(\d+(?:\.\d+)*)`),
	regexp.MustCompile(`(?i)figure\s+(\d+(?:\.\d+)*)`),
}

// ConvertCrossRefs finds "See Section X.Y" patterns and converts to [[wiki-links]].
func ConvertCrossRefs(content string, sectionMap map[string]string) string {
	for _, pattern := range crossRefPatterns {
		content = pattern.ReplaceAllStringFunc(content, func(match string) string {
			sub := pattern.FindStringSubmatch(match)
			if len(sub) < 2 {
				return match
			}
			ref := sub[1]
			if target, ok := sectionMap[ref]; ok {
				return fmt.Sprintf("[[%s|%s]]", target, match)
			}
			return match
		})
	}
	return content
}

func buildSectionMap(sections []IngestSection, prefix string) map[string]string {
	m := make(map[string]string)
	for i, s := range sections {
		slug := slugify(s.Heading)
		m[fmt.Sprintf("%d", i+1)] = prefix + slug
		m[strings.ToLower(s.Heading)] = prefix + slug
	}
	return m
}
