package api

import (
	"log"
	"net/http"

	"github.com/kiwifs/kiwifs/internal/vectorstore"
	"github.com/labstack/echo/v4"
)

type suggestionItem struct {
	Target     string  `json:"target"`
	Similarity float64 `json:"similarity"`
	Snippet    string  `json:"snippet"`
}

func (h *Handlers) Suggestions(c echo.Context) error {
	path := c.QueryParam("path")
	if path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}
	if h.vectors == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, vectorstore.ErrDisabled.Error())
	}

	ctx := c.Request().Context()

	content, err := h.store.Read(ctx, path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "page not found")
	}

	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if n, err := parseInt(l); err == nil && n > 0 {
			limit = n
		}
	}

	topK := limit * 3
	if topK < 20 {
		topK = 20
	}
	results, err := h.vectors.Search(ctx, string(content), topK)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Get existing outbound + inbound links to filter them out
	linked := make(map[string]bool)
	linked[path] = true
	if h.linker != nil {
		edges, err := h.linker.AllEdges(ctx)
		if err != nil {
			log.Printf("suggestions: AllEdges: %v", err)
		}
		for _, e := range edges {
			if e.Source == path {
				linked[e.Target] = true
			}
			if e.Target == path {
				linked[e.Source] = true
			}
		}
		backlinks, _ := h.linker.Backlinks(ctx, path)
		for _, bl := range backlinks {
			linked[bl.Path] = true
		}
	}

	var suggestions []suggestionItem
	for _, r := range results {
		if linked[r.Path] {
			continue
		}
		suggestions = append(suggestions, suggestionItem{
			Target:     r.Path,
			Similarity: r.Score,
			Snippet:    r.Snippet,
		})
		if len(suggestions) >= limit {
			break
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"suggestions": suggestions,
	})
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid number")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
