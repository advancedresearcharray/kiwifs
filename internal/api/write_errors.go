package api

import (
	"errors"
	"net/http"

	"github.com/kiwifs/kiwifs/internal/pipeline"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/labstack/echo/v4"
)

// pipelineWriteHTTPError maps pipeline write failures to HTTP status codes.
func pipelineWriteHTTPError(err error) *echo.HTTPError {
	if errors.Is(err, storage.ErrPathDenied) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if errors.Is(err, pipeline.ErrConflict) {
		return echo.NewHTTPError(http.StatusConflict, "file modified since last read — re-fetch and retry")
	}
	if errors.Is(err, pipeline.ErrAppendOnlyDenied) {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}
	if errors.Is(err, pipeline.ErrTransitionDenied) {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}
	if errors.Is(err, pipeline.ErrValidationFailed) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
