package markdown

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	mdLinkURLRe    = regexp.MustCompile(`\[[^\]]*\]\((https?://[^)\s]+)\)`)
	angleLinkURLRe = regexp.MustCompile(`<(https?://[^>\s]+)>`)
	bareURLRe      = regexp.MustCompile(`https?://[^\s<>)\]"']+`)
	inlineCodeRe   = regexp.MustCompile("`[^`]*`")
)

// ExtractExternalURLs returns unique http(s) URLs from markdown body text.
// Fenced code blocks, inline code spans, and frontmatter are excluded by the
// caller passing BodyAfterFrontmatter output only.
func ExtractExternalURLs(body string) []string {
	body = stripFencedCodeBlocks(body)
	body = inlineCodeRe.ReplaceAllString(body, " ")

	seen := make(map[string]struct{})
	add := func(raw string) {
		raw = strings.TrimRight(raw, ".,;:!?)")
		if raw == "" {
			return
		}
		if _, err := url.Parse(raw); err != nil {
			return
		}
		if _, ok := seen[raw]; ok {
			return
		}
		seen[raw] = struct{}{}
	}

	for _, m := range mdLinkURLRe.FindAllStringSubmatch(body, -1) {
		if len(m) > 1 {
			add(m[1])
		}
	}
	for _, m := range angleLinkURLRe.FindAllStringSubmatch(body, -1) {
		if len(m) > 1 {
			add(m[1])
		}
	}
	for _, m := range bareURLRe.FindAllString(body, -1) {
		add(m)
	}

	out := make([]string, 0, len(seen))
	for u := range seen {
		out = append(out, u)
	}
	return out
}
