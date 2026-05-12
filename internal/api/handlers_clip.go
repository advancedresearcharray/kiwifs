package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/kiwifs/kiwifs/internal/clipper"
	"github.com/labstack/echo/v4"
)

// Clip handles web page clipping requests.
//
//	POST /api/kiwi/clip
//	{
//	  "url": "https://example.com/article",
//	  "title": "Optional Title Override",
//	  "tags": ["web", "research"],
//	  "folder": "clips/"
//	}
//	→ { "path": "clips/article-title.md", "title": "Article Title", "excerpt": "..." }
func (h *Handlers) Clip(c echo.Context) error {
	var req struct {
		URL    string   `json:"url"`
		Title  string   `json:"title"`
		Tags   []string `json:"tags"`
		Folder string   `json:"folder"`
	}

	body, err := io.ReadAll(io.LimitReader(c.Request().Body, 10<<20)) // 10MB limit
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to read body")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}

	if req.URL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "url is required")
	}

	// Clip the web page
	clipReq := clipper.ClipRequest{
		URL:    req.URL,
		Title:  req.Title,
		Tags:   req.Tags,
		Folder: req.Folder,
	}

	result, content, err := clipper.Clip(c.Request().Context(), clipReq, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "clip failed: "+err.Error())
	}

	// Write through pipeline
	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "clipper"
	}

	_, writeErr := h.pipe.Write(c.Request().Context(), result.Path, []byte(content), actor)
	if writeErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "write failed: "+writeErr.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":    result.Path,
		"title":   result.Title,
		"excerpt": result.Excerpt,
	})
}
