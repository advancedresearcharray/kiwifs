package api

import (
	"net/http"
	"time"

	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/rbac"
	"github.com/labstack/echo/v4"
)

type publishRequest struct {
	Path string `json:"path"`
}

type publishResponse struct {
	Path        string  `json:"path"`
	Published   bool    `json:"published"`
	PublishedAt string  `json:"published_at,omitempty"`
	PublicURL   string  `json:"public_url,omitempty"`
}

type publishStatusResponse struct {
	Path        string `json:"path"`
	Published   bool   `json:"published"`
	PublishedAt string `json:"published_at,omitempty"`
	PublicURL   string `json:"public_url,omitempty"`
	ViewCount   int    `json:"view_count"`
}

// Publish sets `published: true` and `published_at` in the page frontmatter.
// POST /api/kiwi/publish
func (h *Handlers) Publish(c echo.Context) error {
	var req publishRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if req.Path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}

	ctx := c.Request().Context()
	content, err := readFileOr404(ctx, h.store, req.Path)
	if err != nil {
		return err
	}

	// Set published: true
	content, err = markdown.SetFrontmatterField(content, "published", true)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update frontmatter: "+err.Error())
	}

	// Set published_at only if not already present (preserve original publish date).
	existingAt := rbac.PagePublishedAt(content)
	now := time.Now().UTC()
	if existingAt == nil {
		content, err = markdown.SetFrontmatterField(content, "published_at", now)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set published_at: "+err.Error())
		}
	} else {
		now = *existingAt
	}

	// Write back through the pipeline (triggers search index, webhooks, events).
	actor := sanitizeActor(c.Request().Header.Get("X-Actor"))
	if _, err := h.pipe.Write(ctx, req.Path, content, actor); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "write failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, publishResponse{
		Path:        req.Path,
		Published:   true,
		PublishedAt: now.Format(time.RFC3339),
		PublicURL:   "/p/" + req.Path,
	})
}

// Unpublish sets `published: false` in the page frontmatter, preserving published_at.
// POST /api/kiwi/unpublish
func (h *Handlers) Unpublish(c echo.Context) error {
	var req publishRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if req.Path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}

	ctx := c.Request().Context()
	content, err := readFileOr404(ctx, h.store, req.Path)
	if err != nil {
		return err
	}

	// Set published: false
	content, err = markdown.SetFrontmatterField(content, "published", false)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update frontmatter: "+err.Error())
	}

	// Write back through the pipeline.
	actor := sanitizeActor(c.Request().Header.Get("X-Actor"))
	if _, err := h.pipe.Write(ctx, req.Path, content, actor); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "write failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, publishResponse{
		Path:      req.Path,
		Published: false,
	})
}

// PublishStatus returns the publish metadata for a page.
// GET /api/kiwi/publish/status?path=...
func (h *Handlers) PublishStatus(c echo.Context) error {
	path, err := requirePath(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	content, err := readFileOr404(ctx, h.store, path)
	if err != nil {
		return err
	}

	published := rbac.PagePublished(content)
	publishedAt := rbac.PagePublishedAt(content)

	resp := publishStatusResponse{
		Path:      path,
		Published: published,
	}

	if publishedAt != nil {
		resp.PublishedAt = publishedAt.Format(time.RFC3339)
	}

	if published {
		resp.PublicURL = "/p/" + path
	}

	// Include view count from metrics store if available.
	if h.publishMetrics != nil {
		if m := h.publishMetrics.Get(path); m != nil {
			resp.ViewCount = m.Views
		}
	}

	return c.JSON(http.StatusOK, resp)
}
