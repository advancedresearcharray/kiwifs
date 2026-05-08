package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kiwifs/kiwifs/internal/links"
	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/labstack/echo/v4"
)

// Peek returns a lightweight summary of a file.
func (h *Handlers) Peek(c echo.Context) error {
	path, err := requirePath(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()

	raw, err := readFileOr404(ctx, h.store, path)
	if err != nil {
		return err
	}

	_, body, _ := markdown.SplitFrontmatter(raw)
	if body == nil {
		body = raw
	}

	parsed, _ := markdown.Parse(raw)

	title := ""
	if parsed != nil && len(parsed.Headings) > 0 {
		title = parsed.Headings[0].Text
	}
	if title == "" {
		parts := strings.Split(path, "/")
		title = parts[len(parts)-1]
	}

	snippet := extractFirstParagraphAPI(body, 300)

	linksOut := links.Extract(body)
	linksOut = links.Unique(linksOut)
	if linksOut == nil {
		linksOut = []string{}
	}

	var linksIn []string
	if h.linker != nil {
		entries, _ := h.linker.Backlinks(ctx, path)
		for _, e := range entries {
			linksIn = append(linksIn, e.Path)
		}
	}
	if linksIn == nil {
		linksIn = []string{}
	}

	var headings []string
	if parsed != nil {
		for _, hd := range parsed.Headings {
			headings = append(headings, hd.Text)
		}
	}
	if headings == nil {
		headings = []string{}
	}

	wordCount := len(strings.Fields(string(body)))

	return c.JSON(http.StatusOK, map[string]any{
		"path":        path,
		"title":       title,
		"frontmatter": sanitizeForJSON(parsed.Frontmatter),
		"snippet":     snippet,
		"links_out":   linksOut,
		"links_in":    linksIn,
		"word_count":  wordCount,
		"headings":    headings,
	})
}

// SectionRead returns a single heading section from a file.
func (h *Handlers) SectionRead(c echo.Context) error {
	path, err := requirePath(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()

	raw, err := readFileOr404(ctx, h.store, path)
	if err != nil {
		return err
	}

	_, body, _ := markdown.SplitFrontmatter(raw)
	if body == nil {
		body = raw
	}

	heading := c.QueryParam("heading")
	indexStr := c.QueryParam("index")

	var section *markdown.Section
	if heading != "" {
		section, err = markdown.ExtractSection(body, heading)
	} else if indexStr != "" {
		idx, parseErr := strconv.Atoi(indexStr)
		if parseErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid index")
		}
		section, err = markdown.ExtractSectionByIndex(body, idx)
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "heading or index is required")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":       path,
		"heading":    section.Heading,
		"level":      section.Level,
		"content":    section.Content,
		"line_start": section.LineStart,
		"line_end":   section.LineEnd,
	})
}

// GraphWalk performs a one-hop graph traversal from a page.
func (h *Handlers) GraphWalk(c echo.Context) error {
	path, err := requirePath(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()

	raw, err := readFileOr404(ctx, h.store, path)
	if err != nil {
		return err
	}

	_, body, _ := markdown.SplitFrontmatter(raw)
	if body == nil {
		body = raw
	}

	includeSiblings := c.QueryParam("include_siblings") != "false"

	type neighbor struct {
		Path      string `json:"path"`
		Relation  string `json:"relation"`
		SharedTag string `json:"shared_tag,omitempty"`
	}

	outLinks := links.Extract(body)
	outLinks = links.Unique(outLinks)
	if outLinks == nil {
		outLinks = []string{}
	}

	var inLinks []string
	if h.linker != nil {
		entries, _ := h.linker.Backlinks(ctx, path)
		for _, e := range entries {
			inLinks = append(inLinks, e.Path)
		}
	}
	if inLinks == nil {
		inLinks = []string{}
	}

	var siblings []neighbor
	if includeSiblings {
		dir := dirOf(path)
		var fileTags []string
		fm, _ := markdown.Frontmatter(raw)
		if fm != nil {
			fileTags = extractFrontmatterTags(raw)
		}

		_ = storage.Walk(ctx, h.store, "/", func(e storage.Entry) error {
			if e.Path == path {
				return nil
			}
			if dirOf(e.Path) == dir {
				siblings = append(siblings, neighbor{
					Path:     e.Path,
					Relation: "sibling_dir",
				})
			}
			if len(fileTags) > 0 {
				raw2, err2 := h.store.Read(ctx, e.Path)
				if err2 == nil {
					otherTags := extractFrontmatterTags(raw2)
					for _, ft := range fileTags {
						for _, ot := range otherTags {
							if strings.EqualFold(ft, ot) {
								siblings = append(siblings, neighbor{
									Path:      e.Path,
									Relation:  "sibling_tag",
									SharedTag: ft,
								})
							}
						}
					}
				}
			}
			return nil
		})
	}
	if siblings == nil {
		siblings = []neighbor{}
	}

	var hubScore float64
	if h.linker != nil {
		edges, _ := h.linker.AllEdges(ctx)
		if len(edges) > 0 {
			nodeSet := make(map[string]struct{})
			inCount := 0
			for _, e := range edges {
				nodeSet[e.Source] = struct{}{}
				nodeSet[e.Target] = struct{}{}
				for _, form := range links.TargetForms(path) {
					if strings.EqualFold(e.Target, form) {
						inCount++
						break
					}
				}
			}
			if len(nodeSet) > 0 {
				hubScore = float64(inCount) / float64(len(nodeSet))
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":       path,
		"links_out":  outLinks,
		"links_in":   inLinks,
		"siblings":   siblings,
		"hub_score":  hubScore,
		"in_degree":  len(inLinks),
		"out_degree": len(outLinks),
	})
}

func dirOf(path string) string {
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[:idx]
	}
	return ""
}

// sanitizeForJSON recursively converts map[interface{}]interface{} (YAML native)
// to map[string]interface{} (JSON-compatible). Go's encoding/json rejects the former.
func sanitizeForJSON(v any) any {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]any, len(val))
		for k, v2 := range val {
			m[fmt.Sprint(k)] = sanitizeForJSON(v2)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(val))
		for k, v2 := range val {
			m[k] = sanitizeForJSON(v2)
		}
		return m
	case []interface{}:
		for i, v2 := range val {
			val[i] = sanitizeForJSON(v2)
		}
		return val
	default:
		return v
	}
}

func extractFirstParagraphAPI(body []byte, maxLen int) string {
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
		if idx := strings.LastIndex(result, " "); idx > 0 {
			result = result[:idx]
		}
		result += "…"
	}
	return result
}
