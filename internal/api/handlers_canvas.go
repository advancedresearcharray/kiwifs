package api

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiwifs/kiwifs/internal/jsoncanvas"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/labstack/echo/v4"
)

type canvasListEntry struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// ListCanvas returns all canvas files in the knowledge base.
func (h *Handlers) ListCanvas(c echo.Context) error {
	var entries []canvasListEntry

	err := storage.WalkAll(c.Request().Context(), h.store, "/", func(e storage.Entry) error {
		if storage.IsCanvasFile(e.Path) {
			entries = append(entries, canvasListEntry{
				Path: e.Path,
				Name: canvasDisplayName(e.Path),
			})
		}
		return nil
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if entries == nil {
		entries = []canvasListEntry{}
	}

	return c.JSON(http.StatusOK, map[string]any{"canvases": entries})
}

// ReadCanvas reads a canvas file and returns its JSON content unchanged.
func (h *Handlers) ReadCanvas(c echo.Context) error {
	path, err := requireCanvasPath(c)
	if err != nil {
		return err
	}

	content, err := h.store.Read(c.Request().Context(), path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, "canvas not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := jsoncanvas.Validate(content); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid canvas JSON")
	}

	c.Response().Header().Set("Content-Type", "application/json")
	return c.JSONBlob(http.StatusOK, content)
}

// WriteCanvas writes a canvas file, preserving the request body after validation.
func (h *Handlers) WriteCanvas(c echo.Context) error {
	path, err := requireCanvasPath(c)
	if err != nil {
		return err
	}

	content, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := jsoncanvas.Validate(content); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid canvas JSON: must include nodes and edges arrays")
	}

	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "system"
	}

	result, err := h.pipe.Write(c.Request().Context(), path, content, actor)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"etag": result.ETag,
		"path": path,
	})
}

func requireCanvasPath(c echo.Context) (string, error) {
	path := c.QueryParam("path")
	if path == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, "path required")
	}
	if !strings.HasSuffix(strings.ToLower(path), ".canvas.json") {
		return "", echo.NewHTTPError(http.StatusBadRequest, "path must end with .canvas.json")
	}
	return path, nil
}

func canvasDisplayName(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".canvas.json")
	if name == "" {
		return base
	}
	return name
}
