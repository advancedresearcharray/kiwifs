package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kiwifs/kiwifs/internal/memory"
	"github.com/labstack/echo/v4"
)

const memoryCacheTTL = 30 * time.Second

type memoryCacheEntry struct {
	report    *memory.Report
	prefix    string
	createdAt time.Time
}

// MemoryReport godoc
//
//	@Summary		Get memory consolidation report
//	@Description	Returns a consolidation report summarizing episodic memory coverage across the knowledge base.
//	@Tags			memory
//	@Security		BearerAuth
//	@Param			episodes_prefix	query		string	false	"Override the default episodes path prefix"
//	@Param			limit			query		int		false	"Limit the number of files returned"
//	@Param			offset			query		int		false	"Skip the first N files"
//	@Success		200				{object}	memory.Report
//	@Failure		400				{object}	map[string]string
//	@Failure		500				{object}	map[string]string
//	@Router			/api/kiwi/memory/report [get]
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

	if entry := h.memoryCache.Load(); entry != nil {
		if e, ok := entry.(*memoryCacheEntry); ok && e != nil && e.prefix == prefix && time.Since(e.createdAt) < memoryCacheTTL {
			rep := paginateReport(e.report, limit, offset)
			return c.JSON(http.StatusOK, rep)
		}
	}

	val, err, _ := h.memoryGroup.Do("scan:"+prefix, func() (any, error) {
		opt := memory.Options{EpisodesPathPrefix: prefix}
		return memory.Scan(ctx, h.store, opt)
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	fullReport := val.(*memory.Report)
	h.memoryCache.Store(&memoryCacheEntry{report: fullReport, prefix: prefix, createdAt: time.Now()})

	rep := paginateReport(fullReport, limit, offset)
	return c.JSON(http.StatusOK, rep)
}

func (h *Handlers) invalidateMemoryCache() {
	h.memoryCache.Store((*memoryCacheEntry)(nil))
}

func paginateReport(full *memory.Report, limit, offset int) *memory.Report {
	if limit == 0 && offset == 0 {
		return full
	}
	clone := *full
	clone.Episodes = paginateSlice(full.Episodes, limit, offset)
	clone.Unmerged = paginateSlice(full.Unmerged, limit, offset)
	return &clone
}

func paginateSlice(s []memory.EpisodicFile, limit, offset int) []memory.EpisodicFile {
	if offset >= len(s) {
		return nil
	}
	s = s[offset:]
	if limit > 0 && limit < len(s) {
		s = s[:limit]
	}
	return s
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
