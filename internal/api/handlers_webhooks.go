package api

import (
	"net/http"

	"github.com/kiwifs/kiwifs/internal/webhooks"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) CreateWebhook(c echo.Context) error {
	if h.webhookStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "webhooks not enabled")
	}
	var req struct {
		URL      string `json:"url"`
		PathGlob string `json:"path_glob"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if req.URL == "" || req.PathGlob == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "url and path_glob are required")
	}

	wh, err := h.webhookStore.Register(c.Request().Context(), req.URL, req.PathGlob)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, wh)
}

func (h *Handlers) ListWebhooks(c echo.Context) error {
	if h.webhookStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "webhooks not enabled")
	}
	hooks, err := h.webhookStore.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, hooks)
}

func (h *Handlers) DeleteWebhook(c echo.Context) error {
	if h.webhookStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "webhooks not enabled")
	}
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "id is required")
	}
	if err := h.webhookStore.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// Ensure Handlers has a webhookStore field — added in the struct definition.
var _ = (*webhooks.Store)(nil) // type-check import
