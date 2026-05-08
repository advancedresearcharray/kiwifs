package api

import (
	"net/http"
	"path/filepath"

	"github.com/kiwifs/kiwifs/internal/importer"
	"github.com/labstack/echo/v4"
)

type ingestRequest struct {
	File             string `json:"file"`
	SplitMode        string `json:"split_mode"`
	Prefix           string `json:"prefix"`
	ExtractKeywords  bool   `json:"extract_keywords"`
	MaxKeywords      int    `json:"max_keywords"`
	ConvertCrossRefs bool   `json:"convert_crossrefs"`
	Actor            string `json:"actor"`
}

func (h *Handlers) Ingest(c echo.Context) error {
	var req ingestRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.File == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	ext := filepath.Ext(req.File)
	if !importer.IsMarkItDownFormat(ext) {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported file format: "+ext)
	}

	if req.SplitMode == "" {
		req.SplitMode = "single"
	}

	opts := importer.IngestOptions{
		SplitMode:        req.SplitMode,
		Prefix:           req.Prefix,
		ExtractKeywords:  req.ExtractKeywords,
		MaxKeywords:      req.MaxKeywords,
		ConvertCrossRefs: req.ConvertCrossRefs,
		Actor:            req.Actor,
	}

	result, err := importer.Ingest(c.Request().Context(), req.File, h.pipe, opts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
