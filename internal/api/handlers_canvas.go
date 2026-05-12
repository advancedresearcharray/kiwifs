package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/labstack/echo/v4"
)

// CanvasData represents a canvas file structure
type CanvasData struct {
	Nodes []CanvasNode `json:"nodes"`
	Edges []CanvasEdge `json:"edges"`
}

type CanvasNode struct {
	ID   string                 `json:"id"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	X    float64                `json:"x,omitempty"`
	Y    float64                `json:"y,omitempty"`
}

type CanvasEdge struct {
	ID     string                 `json:"id"`
	Source string                 `json:"source"`
	Target string                 `json:"target"`
	Type   string                 `json:"type,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// ListCanvas returns all canvas files in the knowledge base
func (h *Handlers) ListCanvas(c echo.Context) error {
	var canvasFiles []string

	err := storage.WalkAll(c.Request().Context(), h.store, "/", func(e storage.Entry) error {
		if storage.IsCanvasFile(e.Path) {
			canvasFiles = append(canvasFiles, e.Path)
		}
		return nil
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if canvasFiles == nil {
		canvasFiles = []string{}
	}

	return c.JSON(http.StatusOK, map[string]any{"canvases": canvasFiles})
}

// ReadCanvas reads a canvas file and returns its JSON content
func (h *Handlers) ReadCanvas(c echo.Context) error {
	path := c.QueryParam("path")
	if path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path required")
	}

	if !strings.HasSuffix(path, ".canvas.json") {
		return echo.NewHTTPError(http.StatusBadRequest, "path must end with .canvas.json")
	}

	content, err := h.store.Read(c.Request().Context(), path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, "canvas not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Validate JSON structure
	var canvas CanvasData
	if err := json.Unmarshal(content, &canvas); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid canvas JSON")
	}

	c.Response().Header().Set("Content-Type", "application/json")
	return c.JSONBlob(http.StatusOK, content)
}

// WriteCanvas writes a canvas file
func (h *Handlers) WriteCanvas(c echo.Context) error {
	path := c.QueryParam("path")
	if path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path required")
	}

	if !strings.HasSuffix(path, ".canvas.json") {
		return echo.NewHTTPError(http.StatusBadRequest, "path must end with .canvas.json")
	}

	content, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate JSON structure
	var canvas CanvasData
	if err := json.Unmarshal(content, &canvas); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid canvas JSON: must have nodes and edges arrays")
	}

	// Ensure required fields exist
	if canvas.Nodes == nil {
		canvas.Nodes = []CanvasNode{}
	}
	if canvas.Edges == nil {
		canvas.Edges = []CanvasEdge{}
	}

	// Re-marshal to ensure clean JSON
	cleanContent, err := json.MarshalIndent(canvas, "", "  ")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "system"
	}

	result, err := h.pipe.Write(c.Request().Context(), path, cleanContent, actor)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"etag": result.ETag,
		"path": path,
	})
}
