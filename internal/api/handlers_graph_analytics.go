package api

import (
	"net/http"
	"strconv"

	"github.com/kiwifs/kiwifs/internal/graphutil"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) GraphAnalytics(c echo.Context) error {
	if h.linker == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "link indexing is not enabled")
	}

	ctx := c.Request().Context()
	edges, err := h.linker.AllEdges(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	result := graphutil.Analyze(edges, limit)
	return c.JSON(http.StatusOK, result)
}

// GraphCentrality returns PageRank and betweenness centrality for all pages.
//
//	GET /api/kiwi/graph/centrality?limit=50
func (h *Handlers) GraphCentrality(c echo.Context) error {
	if h.linker == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "link indexing is not enabled")
	}

	ctx := c.Request().Context()
	edges, err := h.linker.AllEdges(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	entries := graphutil.Centrality(edges)

	limit := len(entries)
	if l := c.QueryParam("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v < limit {
			limit = v
		}
	}
	if limit < len(entries) {
		entries = entries[:limit]
	}

	return c.JSON(http.StatusOK, map[string]any{
		"pages": entries,
	})
}

// GraphCommunities returns community clusters detected via the Louvain algorithm.
//
//	GET /api/kiwi/graph/communities
func (h *Handlers) GraphCommunities(c echo.Context) error {
	if h.linker == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "link indexing is not enabled")
	}

	ctx := c.Request().Context()
	edges, err := h.linker.AllEdges(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	communities := graphutil.CommunitiesFromEdges(edges)

	return c.JSON(http.StatusOK, map[string]any{
		"communities": communities,
	})
}

// GraphPath returns the shortest path between two pages in the link graph.
//
//	GET /api/kiwi/graph/path?from=X&to=Y
func (h *Handlers) GraphPath(c echo.Context) error {
	if h.linker == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "link indexing is not enabled")
	}

	from := c.QueryParam("from")
	to := c.QueryParam("to")
	if from == "" || to == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "from and to query parameters are required")
	}

	ctx := c.Request().Context()
	edges, err := h.linker.AllEdges(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	path, err := graphutil.ShortestPathFromEdges(edges, from, to)
	if err != nil {
		if err == graphutil.ErrNodeNotFound {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if err == graphutil.ErrNoPath {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path": path,
	})
}
