package importer

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// convertConfluenceMacros handles Confluence storage format macros and converts
// them to KiwiFS markdown equivalents.
//
// Mapping:
//   {status}         → frontmatter status field + badge rendering
//   {info}           → :::info
//   {warning}        → :::warning
//   {note}           → :::note
//   {tip}            → :::tip
//   {toc}            → [[toc]]
//   {children}       → kiwi-query block listing child pages
//   {code}           → fenced code block
//   {expand}         → <details> block
//   {panel}          → blockquote with title
func convertConfluenceMacros(storageXML string) string {
	result := storageXML

	// Admonition macros: info, warning, note, tip
	result = convertAdmonitionMacro(result, "info")
	result = convertAdmonitionMacro(result, "warning")
	result = convertAdmonitionMacro(result, "note")
	result = convertAdmonitionMacro(result, "tip")

	// Table of Contents
	result = convertTocMacro(result)

	// Children macro
	result = convertChildrenMacro(result)

	// Status macro
	result = convertStatusMacro(result)

	// Code macro
	result = convertCodeMacro(result)

	// Expand/collapse macro
	result = convertExpandMacro(result)

	// Panel macro
	result = convertPanelMacro(result)

	// Excerpt macro (just extract content)
	result = convertExcerptMacro(result)

	return result
}

var admonitionRegex = map[string]*regexp.Regexp{}

func init() {
	for _, kind := range []string{"info", "warning", "note", "tip"} {
		// Match <ac:structured-macro ac:name="info">...<ac:rich-text-body>CONTENT</ac:rich-text-body>...</ac:structured-macro>
		pattern := fmt.Sprintf(
			`(?s)<ac:structured-macro[^>]*ac:name="%s"[^>]*>.*?<ac:rich-text-body>(.*?)</ac:rich-text-body>.*?</ac:structured-macro>`,
			kind,
		)
		admonitionRegex[kind] = regexp.MustCompile(pattern)
	}
}

func convertAdmonitionMacro(input, kind string) string {
	re := admonitionRegex[kind]
	return re.ReplaceAllStringFunc(input, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		content := strings.TrimSpace(submatch[1])

		// Extract optional title from ac:parameter
		title := extractMacroParam(match, "title")

		var buf strings.Builder
		if title != "" {
			buf.WriteString(fmt.Sprintf("\n\n:::%s %s\n", kind, title))
		} else {
			buf.WriteString(fmt.Sprintf("\n\n:::%s\n", kind))
		}
		buf.WriteString(content)
		buf.WriteString("\n:::\n\n")
		return buf.String()
	})
}

var tocRegex = regexp.MustCompile(`(?s)<ac:structured-macro[^>]*ac:name="toc"[^>]*>.*?</ac:structured-macro>|<ac:structured-macro[^>]*ac:name="toc"[^>]*/>`)

func convertTocMacro(input string) string {
	return tocRegex.ReplaceAllString(input, "\n\n[[toc]]\n\n")
}

var childrenRegex = regexp.MustCompile(`(?s)<ac:structured-macro[^>]*ac:name="children"[^>]*>.*?</ac:structured-macro>|<ac:structured-macro[^>]*ac:name="children"[^>]*/>`)

func convertChildrenMacro(input string) string {
	return childrenRegex.ReplaceAllString(input, "\n\n```kiwi-query\ntype: children\n```\n\n")
}

var statusRegex = regexp.MustCompile(`(?s)<ac:structured-macro[^>]*ac:name="status"[^>]*>(.*?)</ac:structured-macro>`)

func convertStatusMacro(input string) string {
	return statusRegex.ReplaceAllStringFunc(input, func(match string) string {
		title := extractMacroParam(match, "title")
		colour := extractMacroParam(match, "colour")
		if title == "" {
			title = "STATUS"
		}
		if colour == "" {
			colour = "Grey"
		}
		return fmt.Sprintf(`<span class="status status-%s">%s</span>`, strings.ToLower(colour), title)
	})
}

var codeRegex = regexp.MustCompile(`(?s)<ac:structured-macro[^>]*ac:name="code"[^>]*>(.*?)</ac:structured-macro>`)

func convertCodeMacro(input string) string {
	return codeRegex.ReplaceAllStringFunc(input, func(match string) string {
		lang := extractMacroParam(match, "language")
		title := extractMacroParam(match, "title")

		// Extract the code body from <ac:plain-text-body>
		bodyRe := regexp.MustCompile(`(?s)<ac:plain-text-body><!\[CDATA\[(.*?)\]\]></ac:plain-text-body>`)
		bodyMatch := bodyRe.FindStringSubmatch(match)
		code := ""
		if len(bodyMatch) >= 2 {
			code = bodyMatch[1]
		}

		var buf strings.Builder
		if title != "" {
			buf.WriteString(fmt.Sprintf("\n\n**%s:**\n", title))
		}
		buf.WriteString("\n```")
		buf.WriteString(lang)
		buf.WriteByte('\n')
		buf.WriteString(code)
		buf.WriteString("\n```\n\n")
		return buf.String()
	})
}

var expandRegex = regexp.MustCompile(`(?s)<ac:structured-macro[^>]*ac:name="expand"[^>]*>(.*?)</ac:structured-macro>`)

func convertExpandMacro(input string) string {
	return expandRegex.ReplaceAllStringFunc(input, func(match string) string {
		title := extractMacroParam(match, "title")
		if title == "" {
			title = "Click to expand"
		}

		bodyRe := regexp.MustCompile(`(?s)<ac:rich-text-body>(.*?)</ac:rich-text-body>`)
		bodyMatch := bodyRe.FindStringSubmatch(match)
		content := ""
		if len(bodyMatch) >= 2 {
			content = strings.TrimSpace(bodyMatch[1])
		}

		return fmt.Sprintf("\n\n<details>\n<summary>%s</summary>\n\n%s\n\n</details>\n\n", title, content)
	})
}

var panelRegex = regexp.MustCompile(`(?s)<ac:structured-macro[^>]*ac:name="panel"[^>]*>(.*?)</ac:structured-macro>`)

func convertPanelMacro(input string) string {
	return panelRegex.ReplaceAllStringFunc(input, func(match string) string {
		title := extractMacroParam(match, "title")

		bodyRe := regexp.MustCompile(`(?s)<ac:rich-text-body>(.*?)</ac:rich-text-body>`)
		bodyMatch := bodyRe.FindStringSubmatch(match)
		content := ""
		if len(bodyMatch) >= 2 {
			content = strings.TrimSpace(bodyMatch[1])
		}

		var buf strings.Builder
		if title != "" {
			buf.WriteString(fmt.Sprintf("\n\n> **%s**\n>\n", title))
		} else {
			buf.WriteString("\n\n")
		}
		for _, line := range strings.Split(content, "\n") {
			buf.WriteString("> ")
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
		buf.WriteString("\n\n")
		return buf.String()
	})
}

var excerptRegex = regexp.MustCompile(`(?s)<ac:structured-macro[^>]*ac:name="excerpt"[^>]*>(.*?)</ac:structured-macro>`)

func convertExcerptMacro(input string) string {
	return excerptRegex.ReplaceAllStringFunc(input, func(match string) string {
		bodyRe := regexp.MustCompile(`(?s)<ac:rich-text-body>(.*?)</ac:rich-text-body>`)
		bodyMatch := bodyRe.FindStringSubmatch(match)
		if len(bodyMatch) >= 2 {
			return strings.TrimSpace(bodyMatch[1])
		}
		return ""
	})
}

// extractMacroParam extracts a named parameter value from a Confluence macro XML block.
func extractMacroParam(macroXML, paramName string) string {
	pattern := fmt.Sprintf(`<ac:parameter ac:name="%s">(.*?)</ac:parameter>`, regexp.QuoteMeta(paramName))
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(macroXML)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

// parseHTMLString parses an HTML string into an *html.Node for tree traversal.
func parseHTMLString(s string) *html.Node {
	doc, err := html.Parse(bytes.NewReader([]byte(s)))
	if err != nil {
		return nil
	}
	return doc
}
