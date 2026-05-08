package api

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/kiwifs/kiwifs/internal/draft"
	"github.com/kiwifs/kiwifs/internal/pipeline"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/labstack/echo/v4"
)

type draftResponse struct {
	ID        string    `json:"id"`
	Branch    string    `json:"branch"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
}

func toDraftResponse(d *draft.Draft) draftResponse {
	return draftResponse{
		ID:        d.ID,
		Branch:    d.Branch,
		Actor:     d.Actor,
		CreatedAt: d.CreatedAt,
	}
}

func (h *Handlers) CreateDraft(c echo.Context) error {
	if h.draftMgr == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "drafts not enabled")
	}
	var body struct {
		Actor string `json:"actor"`
	}
	_ = c.Bind(&body)
	if body.Actor == "" {
		body.Actor = c.Request().Header.Get("X-Actor")
	}
	if body.Actor == "" {
		body.Actor = "api"
	}
	d, err := h.draftMgr.Create(c.Request().Context(), body.Actor)
	if err != nil {
		if errors.Is(err, draft.ErrMaxActive) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		if errors.Is(err, draft.ErrEmptyRepo) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, toDraftResponse(d))
}

func (h *Handlers) ListDrafts(c echo.Context) error {
	if h.draftMgr == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "drafts not enabled")
	}
	drafts := h.draftMgr.List()
	out := make([]draftResponse, len(drafts))
	for i, d := range drafts {
		out[i] = toDraftResponse(d)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handlers) GetDraft(c echo.Context) error {
	if h.draftMgr == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "drafts not enabled")
	}
	d, err := h.draftMgr.Get(c.Param("id"))
	if err != nil {
		if errors.Is(err, draft.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "draft not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, toDraftResponse(d))
}

func (h *Handlers) DraftDiff(c echo.Context) error {
	if h.draftMgr == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "drafts not enabled")
	}
	diff, err := h.draftMgr.Diff(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, draft.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "draft not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"diff": diff})
}

func (h *Handlers) MergeDraft(c echo.Context) error {
	if h.draftMgr == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "drafts not enabled")
	}
	err := h.draftMgr.Merge(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, draft.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "draft not found")
		}
		if errors.Is(err, draft.ErrConflict) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "merged"})
}

func (h *Handlers) DiscardDraft(c echo.Context) error {
	if h.draftMgr == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "drafts not enabled")
	}
	err := h.draftMgr.Discard(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, draft.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "draft not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) draftPipeline(id string) (*pipeline.Pipeline, error) {
	if h.draftMgr == nil {
		return nil, errors.New("drafts not enabled")
	}
	return h.draftMgr.Pipeline(id)
}

func (h *Handlers) DraftReadFile(c echo.Context) error {
	pipe, err := h.draftPipeline(c.Param("id"))
	if err != nil {
		if errors.Is(err, draft.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "draft not found")
		}
		return echo.NewHTTPError(http.StatusNotImplemented, err.Error())
	}
	path := c.QueryParam("path")
	if path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}
	content, rerr := pipe.Store.Read(c.Request().Context(), path)
	if rerr != nil {
		if errors.Is(rerr, storage.ErrPathDenied) {
			return echo.NewHTTPError(http.StatusBadRequest, rerr.Error())
		}
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}
	etag := pipeline.ETag(content)
	c.Response().Header().Set("ETag", `"`+etag+`"`)
	return c.Blob(http.StatusOK, "text/markdown; charset=utf-8", content)
}

func (h *Handlers) DraftWriteFile(c echo.Context) error {
	pipe, err := h.draftPipeline(c.Param("id"))
	if err != nil {
		if errors.Is(err, draft.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "draft not found")
		}
		return echo.NewHTTPError(http.StatusNotImplemented, err.Error())
	}
	path := c.QueryParam("path")
	if path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}
	body, rerr := io.ReadAll(c.Request().Body)
	if rerr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to read body")
	}
	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "draft-api"
	}
	res, werr := pipe.Write(c.Request().Context(), path, body, actor)
	if werr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, werr.Error())
	}
	c.Response().Header().Set("ETag", `"`+res.ETag+`"`)
	return c.JSON(http.StatusOK, map[string]string{"path": res.Path, "etag": res.ETag})
}

func (h *Handlers) DraftDeleteFile(c echo.Context) error {
	pipe, err := h.draftPipeline(c.Param("id"))
	if err != nil {
		if errors.Is(err, draft.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "draft not found")
		}
		return echo.NewHTTPError(http.StatusNotImplemented, err.Error())
	}
	path := c.QueryParam("path")
	if path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}
	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "draft-api"
	}
	if derr := pipe.Delete(c.Request().Context(), path, actor); derr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, derr.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) DraftTree(c echo.Context) error {
	pipe, err := h.draftPipeline(c.Param("id"))
	if err != nil {
		if errors.Is(err, draft.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "draft not found")
		}
		return echo.NewHTTPError(http.StatusNotImplemented, err.Error())
	}
	path := c.QueryParam("path")
	if path == "" {
		path = "/"
	}
	st, serr := storage.BuildTree(c.Request().Context(), pipe.Store, path, maxTreeDepth)
	if serr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, serr.Error())
	}
	return c.JSON(http.StatusOK, toTreeEntry(st))
}
