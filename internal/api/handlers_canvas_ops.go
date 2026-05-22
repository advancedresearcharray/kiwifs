package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/kiwifs/kiwifs/internal/jsoncanvas"
	"github.com/labstack/echo/v4"
)

// DeleteCanvas removes a canvas file.
//
// DELETE /api/kiwi/canvas?path=<path>
func (h *Handlers) DeleteCanvas(c echo.Context) error {
	path, err := requireCanvasPath(c)
	if err != nil {
		return err
	}

	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "system"
	}

	if err := h.pipe.Delete(c.Request().Context(), path, actor); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, "canvas not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"deleted": path})
}

// PatchCanvas applies a batch of atomic node/edge operations.
//
// PATCH /api/kiwi/canvas?path=<path>
//
// Agents use this to add, update, or remove individual nodes and edges
// without replacing the entire document.
//
// Body: { "operations": [ { "op": "add_node", "node": {...} }, ... ] }
func (h *Handlers) PatchCanvas(c echo.Context) error {
	path, err := requireCanvasPath(c)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var req jsoncanvas.PatchRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid patch request")
	}
	if len(req.Operations) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "operations array is empty")
	}

	ctx := c.Request().Context()

	doc, err := h.store.Read(ctx, path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			doc = jsoncanvas.EmptyDocument()
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	patched, result, err := jsoncanvas.ApplyPatch(doc, req.Operations)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "system"
	}

	writeResult, err := h.pipe.Write(ctx, path, patched, actor)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":    path,
		"etag":    writeResult.ETag,
		"applied": result,
	})
}

// QueryCanvas searches nodes and edges within a canvas.
//
// GET /api/kiwi/canvas/query?path=<path>&type=<node_type>&q=<search>&connected=<node_id>&nodes_only=true&edges_only=true
func (h *Handlers) QueryCanvas(c echo.Context) error {
	path, err := requireCanvasPath(c)
	if err != nil {
		return err
	}

	doc, err := h.store.Read(c.Request().Context(), path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, "canvas not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	q := jsoncanvas.QueryParams{
		NodeType:  c.QueryParam("type"),
		Search:    c.QueryParam("q"),
		Connected: c.QueryParam("connected"),
		NodesOnly: c.QueryParam("nodes_only") == "true",
		EdgesOnly: c.QueryParam("edges_only") == "true",
	}

	result, err := jsoncanvas.Query(doc, q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// AutoLayoutCanvas repositions existing canvas nodes using Graphviz.
// Unlike generate (which builds from wiki-links), this takes the current
// nodes/edges and computes optimal spatial positions.
//
// POST /api/kiwi/canvas/auto-layout?path=<path>
//
// Optional body: { "layout": "dot"|"neato"|"fdp"|"circo" }
func (h *Handlers) AutoLayoutCanvas(c echo.Context) error {
	path, err := requireCanvasPath(c)
	if err != nil {
		return err
	}

	var req struct {
		Layout string `json:"layout"`
	}
	_ = c.Bind(&req)

	ctx := c.Request().Context()

	doc, err := h.store.Read(ctx, path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, "canvas not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	laid, nodeCount, edgeCount, err := jsoncanvas.AutoLayout(doc, req.Layout)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "system"
	}

	writeResult, err := h.pipe.Write(ctx, path, laid, actor)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":       path,
		"etag":       writeResult.ETag,
		"node_count": nodeCount,
		"edge_count": edgeCount,
	})
}
