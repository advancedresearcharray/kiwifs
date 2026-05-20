package api

import (
	"context"
	"net/http"
	"time"

	"github.com/kiwifs/kiwifs/internal/docexport"
	"github.com/labstack/echo/v4"
)

// ExportDocument handles POST /api/kiwi/export/document
// It renders markdown to PDF, HTML, slides, or a static site.
func (h *Handlers) ExportDocument(c echo.Context) error {
	var req struct {
		Format        string            `json:"format"`
		Path          string            `json:"path"`
		Theme         string            `json:"theme"`
		SelfContained bool              `json:"self_contained"`
		Bibliography  string            `json:"bibliography"`
		CSLStyle      string            `json:"csl_style"`
		CrossRef      bool              `json:"crossref"`
		PDFEngine     string            `json:"pdf_engine"`
		SlideFormat   string            `json:"slide_format"`
		Metadata      map[string]string `json:"metadata"`
		SiteName      string            `json:"site_name"`
		SiteURL       string            `json:"site_url"`
		RepoURL       string            `json:"repo_url"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body: "+err.Error())
	}

	if req.Format == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "format is required (pdf, html, slides, site)")
	}
	if req.Path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required (file or directory to export)")
	}

	format, err := docexport.ParseFormat(req.Format)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Create exporter registry.
	provider := docexport.NewStorageProvider(h.store, h.root)
	registry := docexport.NewRegistry()
	registry.Register(docexport.NewPDFExporter(provider, h.root))
	registry.Register(docexport.NewHTMLExporter(provider, h.root))
	registry.Register(docexport.NewSlidesExporter(provider, h.root))
	registry.Register(docexport.NewSiteExporter(provider, h.store, h.root))

	opts := docexport.ExportOpts{
		Format:        format,
		InputPath:     req.Path,
		Theme:         req.Theme,
		SelfContained: req.SelfContained,
		Bibliography:  req.Bibliography,
		CSLStyle:      req.CSLStyle,
		CrossRef:      req.CrossRef,
		PDFEngine:     req.PDFEngine,
		SlideFormat:   req.SlideFormat,
		Metadata:      req.Metadata,
		SiteName:      req.SiteName,
		SiteURL:       req.SiteURL,
		RepoURL:       req.RepoURL,
	}

	// 5-minute timeout for document rendering (can be slow for large docs).
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Minute)
	defer cancel()

	result, err := registry.Export(ctx, opts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "export failed: "+err.Error())
	}

	c.Response().Header().Set("Content-Type", result.ContentType)
	c.Response().Header().Set("Content-Disposition", "attachment; filename=\""+result.Filename+"\"")
	return c.Blob(http.StatusOK, result.ContentType, result.Data)
}
