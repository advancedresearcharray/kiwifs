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
