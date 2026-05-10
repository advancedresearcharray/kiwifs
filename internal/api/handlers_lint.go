package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/labstack/echo/v4"
)

// Lint validates markdown content for structural issues and returns a list
// of findings. Accepts JSON body with either "path" (reads from store) or
// "content" (inline markdown).
//
//	POST /api/kiwi/lint
//	{ "path": "pages/foo.md" }         → lint an existing file
//	{ "content": "# Hello\n..." }      → lint inline content
//	→ { "issues": [...] }
func (h *Handlers) Lint(c echo.Context) error {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	body, err := io.ReadAll(io.LimitReader(c.Request().Body, 32<<20+1))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to read body")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}

	var data []byte
	if req.Content != "" {
		data = []byte(req.Content)
	} else if req.Path != "" {
		content, readErr := h.store.Read(c.Request().Context(), req.Path)
		if readErr != nil {
			return echo.NewHTTPError(http.StatusNotFound, "file not found: "+req.Path)
		}
		data = content
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "provide either path or content")
	}

	issues := markdown.LintMarkdown(data)
	if issues == nil {
		issues = []markdown.LintIssue{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"issues": issues,
	})
}
