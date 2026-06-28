package markdown

import (
	"regexp"
	"sort"
	"strings"
)

var templateParamRe = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)

// ExtractTemplateParameters returns unique `{{name}}` placeholders from markdown
// body text, ignoring variables inside fenced code blocks.
func ExtractTemplateParameters(body string) []string {
	body = stripFencedCodeBlocks(body)
	seen := map[string]struct{}{}
	var out []string
	for _, m := range templateParamRe.FindAllStringSubmatch(body, -1) {
		if len(m) < 2 {
			continue
		}
		name := m[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func stripFencedCodeBlocks(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if !inFence {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}
