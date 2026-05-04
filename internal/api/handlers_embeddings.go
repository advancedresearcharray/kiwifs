package api

import (
	"net/http"

	"github.com/kiwifs/kiwifs/internal/vectorstore"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) Embeddings(c echo.Context) error {
	path := c.QueryParam("path")
	if path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}
	if h.vectors == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, vectorstore.ErrDisabled.Error())
	}

	ctx := c.Request().Context()
	chunks, err := h.vectors.GetVectors(ctx, path)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(chunks) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "no embeddings found for path")
	}

	type chunkResp struct {
		ChunkIdx int       `json:"chunk_idx"`
		Text     string    `json:"text"`
		Vector   []float32 `json:"vector"`
	}

	out := make([]chunkResp, len(chunks))
	dims := 0
	for i, ch := range chunks {
		out[i] = chunkResp{
			ChunkIdx: ch.ChunkIdx,
			Text:     ch.Text,
			Vector:   ch.Vector,
		}
		if i == 0 && len(ch.Vector) > 0 {
			dims = len(ch.Vector)
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":       path,
		"dimensions": dims,
		"chunks":     out,
	})
}
