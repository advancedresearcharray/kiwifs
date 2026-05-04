package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/kiwifs/kiwifs/internal/exporter"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) Export(c echo.Context) error {
	format := c.QueryParam("format")
	if format == "" {
		format = "jsonl"
	}
	if format != "jsonl" && format != "csv" && format != "parquet" {
		return echo.NewHTTPError(http.StatusBadRequest, "format must be jsonl, csv, or parquet")
	}

	pathPrefix := c.QueryParam("path")
	includeContent := c.QueryParam("include_content") == "true"
	includeLinks := c.QueryParam("include_links") == "true"
	includeEmb := c.QueryParam("include_embeddings") == "true"
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	var columns []string
	if cols := c.QueryParam("columns"); cols != "" {
		columns = strings.Split(cols, ",")
		for i := range columns {
			columns[i] = strings.TrimSpace(columns[i])
		}
	}

	if limit < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "limit must be non-negative")
	}
	if format == "parquet" && includeEmb {
		return echo.NewHTTPError(http.StatusBadRequest, "embeddings are not supported in Parquet format; use format=jsonl instead")
	}

	contentType := "application/x-ndjson"
	switch format {
	case "csv":
		contentType = "text/csv"
	case "parquet":
		contentType = "application/vnd.apache.parquet"
	}

	opts := exporter.Options{
		Format:            format,
		PathPrefix:        pathPrefix,
		Columns:           columns,
		IncludeContent:    includeContent,
		IncludeLinks:      includeLinks,
		IncludeEmbeddings: includeEmb,
		Output:            c.Response().Writer,
		Limit:             limit,
	}

	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Transfer-Encoding", "chunked")
	c.Response().WriteHeader(http.StatusOK)

	ctx := c.Request().Context()
	_, err := exporter.Export(ctx, h.store, h.searcher, h.vectors, opts)
	if err != nil {
		return err
	}
	c.Response().Flush()
	return nil
}
