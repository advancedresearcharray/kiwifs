package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kiwifs/kiwifs/internal/memory"
	"github.com/labstack/echo/v4"
)

// MemoryReport returns episodic vs merged-from coverage for consolidation pipelines.
// Query param episodes_prefix overrides [memory] episodes_path_prefix from config.
func (h *Handlers) MemoryReport(c echo.Context) error {
	ctx := c.Request().Context()
	prefix := c.QueryParam("episodes_prefix")
	if prefix == "" {
		prefix = h.memoryEpisodesPrefix
	}
	limit, err := nonNegativeIntQuery(c, "limit")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	offset, err := nonNegativeIntQuery(c, "offset")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	opt := memory.Options{EpisodesPathPrefix: prefix, Limit: limit, Offset: offset}
	rep, err := memory.Scan(ctx, h.store, opt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, rep)
}

func nonNegativeIntQuery(c echo.Context, name string) (int, error) {
	raw := c.QueryParam(name)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", name)
	}
	return n, nil
}
