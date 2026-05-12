package api

import (
	"net/http"

	"github.com/kiwifs/kiwifs/internal/views"
	"github.com/labstack/echo/v4"
)

// ListViews returns all view definitions
func (h *Handlers) ListViews(c echo.Context) error {
	viewsList, err := views.List(h.root)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"views": viewsList})
}

// GetView returns a single view definition
func (h *Handlers) GetView(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "view name required")
	}

	view, err := views.Get(h.root, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, view)
}

// SaveView creates or updates a view definition
func (h *Handlers) SaveView(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "view name required")
	}

	var view views.View
	if err := c.Bind(&view); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	view.Name = name

	if err := views.Save(h.root, view); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "saved", "view": view})
}

// DeleteView removes a view definition
func (h *Handlers) DeleteView(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "view name required")
	}

	if err := views.Delete(h.root, name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "deleted"})
}

// ExecuteView runs a view's query using the dataview executor
func (h *Handlers) ExecuteView(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "view name required")
	}

	view, err := views.Get(h.root, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if h.dv == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "dataview executor not available")
	}

	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	result, err := h.dv.Query(c.Request().Context(), view.Query, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Apply formulas to computed columns
	if len(view.Columns) > 0 {
		for i := range result.Rows {
			for _, col := range view.Columns {
				if col.Formula != "" {
					val, err := views.EvalFormula(col.Formula, result.Rows[i])
					if err == nil {
						result.Rows[i][col.Property] = val
					}
				}
			}
		}
	}

	return c.JSON(http.StatusOK, result)
}
