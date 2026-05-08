package mcpserver

import (
	"strings"
	"unicode"
)

// extractFirstParagraph returns the first non-heading, non-empty paragraph
// from a markdown body, truncated to maxLen characters.
func extractFirstParagraph(body []byte, maxLen int) string {
	lines := strings.Split(string(body), "\n")
	var para strings.Builder
	inParagraph := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inParagraph {
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			inParagraph = true
		}

		if inParagraph && trimmed == "" {
			break
		}

		if para.Len() > 0 {
			para.WriteByte(' ')
		}
		para.WriteString(trimmed)

		if para.Len() >= maxLen {
			break
		}
	}

	result := para.String()
	if len(result) > maxLen {
		result = result[:maxLen]
		if idx := strings.LastIndexFunc(result, unicode.IsSpace); idx > 0 {
			result = result[:idx]
		}
		result += "…"
	}
	return result
}
