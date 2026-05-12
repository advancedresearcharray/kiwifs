package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/kiwifs/kiwifs/internal/importer"
	"github.com/labstack/echo/v4"
)

type importRequest struct {
	From        string          `json:"from"`
	DSN         string          `json:"dsn"`
	URI         string          `json:"uri"`
	DB          string          `json:"db"`
	File        string          `json:"file"`
	Table       string          `json:"table"`
	Collection  string          `json:"collection"`
	Database    string          `json:"database"`
	DatabaseID  string          `json:"database_id"`
	BaseID      string          `json:"base_id"`
	TableID     string          `json:"table_id"`
	Project     string          `json:"project"`
	Query       string          `json:"query"`
	Columns     []string        `json:"columns"`
	IDColumn    string          `json:"id_column"`
	Prefix      string          `json:"prefix"`
	DryRun      bool            `json:"dry_run"`
	Limit       int             `json:"limit"`
	Credentials json.RawMessage `json:"credentials,omitempty"` // Service account JSON (Firestore)
	APIKey      string          `json:"api_key,omitempty"`     // API key (Notion, Airtable)
}

type importResponse struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}

func (h *Handlers) Import(c echo.Context) error {
	var req importRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.From == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "from is required")
	}

	src, err := buildAPISource(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer src.Close()

	var columns []string
	if len(req.Columns) > 0 {
		columns = req.Columns
	}

	actor := sanitizeActor(c.Request().Header.Get("X-Actor"))
	if actor == "anonymous" {
		actor = "api-import"
	}

	opts := importer.Options{
		Prefix:   req.Prefix,
		IDColumn: req.IDColumn,
		Columns:  columns,
		DryRun:   req.DryRun,
		Limit:    req.Limit,
		Actor:    actor,
	}

	ctx := c.Request().Context()
	stats, err := importer.Run(ctx, src, h.pipe, opts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("import failed: %v", err))
	}

	// Auto-save connection metadata on successful import (no credentials stored)
	if h.connStore != nil && stats.Imported > 0 {
		connMeta := &importer.ConnectionMeta{
			From:       req.From,
			Name:       req.From + ":" + coalesce(req.Collection, req.Table, req.DatabaseID, req.TableID),
			Project:    req.Project,
			Table:      req.Table,
			Collection: req.Collection,
			Database:   req.Database,
			DatabaseID: req.DatabaseID,
			BaseID:     req.BaseID,
			TableID:    req.TableID,
			DSN:        req.DSN,
			URI:        req.URI,
			Prefix:     opts.Prefix,
			IDColumn:   req.IDColumn,
			Columns:    req.Columns,
			LastStats:  &importer.ConnectionStats{Imported: stats.Imported, Skipped: stats.Skipped, Errors: stats.Errors},
		}
		_ = h.connStore.Save(connMeta)
	}

	return c.JSON(http.StatusOK, importResponse{
		Imported: stats.Imported,
		Skipped:  stats.Skipped,
		Errors:   stats.Errors,
	})
}

func coalesce(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func buildAPISource(req importRequest) (importer.Source, error) {
	switch req.From {
	case "postgres":
		if req.DSN == "" {
			return nil, fmt.Errorf("dsn is required for postgres")
		}
		if req.Table == "" && req.Query == "" {
			return nil, fmt.Errorf("table or query is required for postgres")
		}
		return importer.NewPostgres(req.DSN, req.Table, req.Query, req.Columns)
	case "mysql":
		if req.DSN == "" {
			return nil, fmt.Errorf("dsn is required for mysql")
		}
		if req.Table == "" && req.Query == "" {
			return nil, fmt.Errorf("table or query is required for mysql")
		}
		return importer.NewMySQL(req.DSN, req.Table, req.Query, req.Columns)
	case "firestore":
		if req.Project == "" {
			return nil, fmt.Errorf("project is required for firestore")
		}
		if req.Collection == "" {
			return nil, fmt.Errorf("collection is required for firestore")
		}
		if len(req.Credentials) > 0 {
			return importer.NewFirestoreWithCredentials(req.Project, req.Collection, []byte(req.Credentials))
		}
		return importer.NewFirestore(req.Project, req.Collection)
	case "sqlite":
		if req.DB == "" {
			return nil, fmt.Errorf("db is required for sqlite")
		}
		if req.Table == "" && req.Query == "" {
			return nil, fmt.Errorf("table or query is required for sqlite")
		}
		return importer.NewSQLiteSource(req.DB, req.Table, req.Query)
	case "mongodb":
		uri := req.URI
		if uri == "" {
			uri = req.DSN
		}
		if uri == "" {
			return nil, fmt.Errorf("uri is required for mongodb")
		}
		if req.Collection == "" {
			return nil, fmt.Errorf("collection is required for mongodb")
		}
		db := req.Database
		if db == "" {
			return nil, fmt.Errorf("database is required for mongodb")
		}
		return importer.NewMongoDB(uri, db, req.Collection)
	case "csv":
		if req.File == "" {
			return nil, fmt.Errorf("file is required for csv")
		}
		return importer.NewCSV(req.File, true)
	case "json", "jsonl":
		if req.File == "" {
			return nil, fmt.Errorf("file is required for json/jsonl")
		}
		return importer.NewJSON(req.File)
	case "notion":
		apiKey := req.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("NOTION_API_KEY")
		}
		if req.DatabaseID == "" {
			return nil, fmt.Errorf("database_id is required for notion")
		}
		return importer.NewNotion(apiKey, req.DatabaseID)
	case "airtable":
		apiKey := req.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("AIRTABLE_API_KEY")
		}
		if req.BaseID == "" {
			return nil, fmt.Errorf("base_id is required for airtable")
		}
		if req.TableID == "" {
			return nil, fmt.Errorf("table_id is required for airtable")
		}
		return importer.NewAirtable(apiKey, req.BaseID, req.TableID)
	default:
		supported := strings.Join([]string{"postgres", "mysql", "firestore", "sqlite", "mongodb", "csv", "json", "jsonl", "notion", "airtable"}, ", ")
		return nil, fmt.Errorf("unknown source type %q (supported: %s)", req.From, supported)
	}
}

// --- Browse & Preview endpoints (Phase 2) ---

type browseRequest struct {
	From        string          `json:"from"`
	DSN         string          `json:"dsn"`
	URI         string          `json:"uri"`
	DB          string          `json:"db"`
	Project     string          `json:"project"`
	Database    string          `json:"database"`
	Credentials json.RawMessage `json:"credentials,omitempty"`
	APIKey      string          `json:"api_key,omitempty"`
}

type browseResponse struct {
	Tables []importer.BrowseTable `json:"tables"`
}

// ImportBrowse lists available tables/collections for a given data source.
func (h *Handlers) ImportBrowse(c echo.Context) error {
	var req browseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.From == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "from is required")
	}

	ctx := c.Request().Context()
	tables, err := browseSource(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, browseResponse{Tables: tables})
}

func browseSource(ctx context.Context, req browseRequest) ([]importer.BrowseTable, error) {
	switch req.From {
	case "firestore":
		return browseFirestore(ctx, req)
	case "postgres":
		return browsePostgres(ctx, req)
	case "mysql":
		return browseMySQL(ctx, req)
	case "mongodb":
		return browseMongoDB(ctx, req)
	default:
		return nil, fmt.Errorf("browse not supported for source type %q", req.From)
	}
}

func browseFirestore(ctx context.Context, req browseRequest) ([]importer.BrowseTable, error) {
	if req.Project == "" {
		return nil, fmt.Errorf("project is required for firestore")
	}

	var src *importer.FirestoreSource
	var err error
	if len(req.Credentials) > 0 {
		src, err = importer.NewFirestoreWithCredentials(req.Project, "__browse__", []byte(req.Credentials))
	} else {
		src, err = importer.NewFirestore(req.Project, "__browse__")
	}
	if err != nil {
		return nil, err
	}
	defer src.Close()

	names, err := src.BrowseCollections(ctx)
	if err != nil {
		return nil, err
	}
	tables := make([]importer.BrowseTable, len(names))
	for i, name := range names {
		tables[i] = importer.BrowseTable{Name: name}
	}
	return tables, nil
}

func browsePostgres(ctx context.Context, req browseRequest) ([]importer.BrowseTable, error) {
	if req.DSN == "" {
		return nil, fmt.Errorf("dsn is required for postgres")
	}
	src, err := importer.NewPostgres(req.DSN, "", "SELECT 1", nil)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	return importer.BrowsePostgresTables(ctx, src)
}

func browseMySQL(ctx context.Context, req browseRequest) ([]importer.BrowseTable, error) {
	if req.DSN == "" {
		return nil, fmt.Errorf("dsn is required for mysql")
	}
	src, err := importer.NewMySQL(req.DSN, "", "SELECT 1", nil)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	return importer.BrowseMySQLTables(ctx, src)
}

func browseMongoDB(ctx context.Context, req browseRequest) ([]importer.BrowseTable, error) {
	uri := req.URI
	if uri == "" {
		uri = req.DSN
	}
	if uri == "" {
		return nil, fmt.Errorf("uri is required for mongodb")
	}
	db := req.Database
	if db == "" {
		return nil, fmt.Errorf("database is required for mongodb")
	}
	src, err := importer.NewMongoDB(uri, db, "__browse__")
	if err != nil {
		return nil, err
	}
	defer src.Close()

	return importer.BrowseMongoCollections(ctx, src)
}

// --- Preview ---

type previewRequest struct {
	From        string          `json:"from"`
	DSN         string          `json:"dsn"`
	URI         string          `json:"uri"`
	DB          string          `json:"db"`
	Table       string          `json:"table"`
	Collection  string          `json:"collection"`
	Database    string          `json:"database"`
	DatabaseID  string          `json:"database_id"`
	BaseID      string          `json:"base_id"`
	TableID     string          `json:"table_id"`
	Project     string          `json:"project"`
	Credentials json.RawMessage `json:"credentials,omitempty"`
	APIKey      string          `json:"api_key,omitempty"`
	Limit       int             `json:"limit"`
}

type previewRecord struct {
	Path        string         `json:"path"`
	Frontmatter map[string]any `json:"frontmatter"`
	BodyPreview string         `json:"body_preview"`
}

type previewResponse struct {
	Records []previewRecord `json:"records"`
}

// ImportPreview fetches sample records from a source and returns rendered markdown previews.
func (h *Handlers) ImportPreview(c echo.Context) error {
	var req previewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.From == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "from is required")
	}

	limit := req.Limit
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	// Build an importRequest to reuse buildAPISource
	ir := importRequest{
		From:        req.From,
		DSN:         req.DSN,
		URI:         req.URI,
		DB:          req.DB,
		Table:       req.Table,
		Collection:  req.Collection,
		Database:    req.Database,
		DatabaseID:  req.DatabaseID,
		BaseID:      req.BaseID,
		TableID:     req.TableID,
		Project:     req.Project,
		Credentials: req.Credentials,
		APIKey:      req.APIKey,
		Limit:       limit,
	}

	src, err := buildAPISource(ir)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer src.Close()

	ctx := c.Request().Context()
	records, errs := src.Stream(ctx)

	var previews []previewRecord
	count := 0
	for rec := range records {
		if count >= limit {
			break
		}
		fm := make(map[string]any, len(rec.Fields)+2)
		for k, v := range rec.Fields {
			fm[k] = v
		}
		fm["_source"] = src.Name()
		fm["_source_id"] = rec.SourceID

		title := rec.PrimaryKey
		if t, ok := rec.Fields["title"].(string); ok && t != "" {
			title = t
		} else if t, ok := rec.Fields["name"].(string); ok && t != "" {
			title = t
		}

		path := fmt.Sprintf("%s/%s.md", src.Name(), importer.SanitizePath(rec.PrimaryKey))
		body := fmt.Sprintf("# %s\n\n> Auto-imported from %s (row %s)", title, rec.Table, rec.SourceID)

		previews = append(previews, previewRecord{
			Path:        path,
			Frontmatter: fm,
			BodyPreview: body,
		})
		count++
	}

	// Drain any errors
	for err := range errs {
		if err != nil && len(previews) == 0 {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return c.JSON(http.StatusOK, previewResponse{Records: previews})
}

// --- Connection CRUD endpoints (Phase 3) ---

// ListConnections returns all saved import connections.
func (h *Handlers) ListConnections(c echo.Context) error {
	if h.connStore == nil {
		return c.JSON(http.StatusOK, []any{})
	}
	return c.JSON(http.StatusOK, h.connStore.List())
}

// SaveConnection saves connection metadata (no credentials stored).
func (h *Handlers) SaveConnection(c echo.Context) error {
	if h.connStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "connection store not available")
	}

	var conn importer.ConnectionMeta
	if err := c.Bind(&conn); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if conn.From == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "from is required")
	}
	if conn.Name == "" {
		conn.Name = conn.From
	}

	if err := h.connStore.Save(&conn); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, conn)
}

// DeleteConnection removes a saved connection.
func (h *Handlers) DeleteConnection(c echo.Context) error {
	if h.connStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "connection store not available")
	}

	id := c.Param("id")
	if err := h.connStore.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// RunConnection re-runs an import for a saved connection.
// Credentials must be provided in the request body (they are not stored).
func (h *Handlers) RunConnection(c echo.Context) error {
	if h.connStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "connection store not available")
	}

	id := c.Param("id")
	conn, ok := h.connStore.Get(id)
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "connection not found")
	}

	// Parse runtime credentials from body
	var creds struct {
		Credentials json.RawMessage `json:"credentials,omitempty"`
		APIKey      string          `json:"api_key,omitempty"`
	}
	if err := c.Bind(&creds); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Build importRequest from saved connection + runtime creds
	ir := importRequest{
		From:        conn.From,
		DSN:         conn.DSN,
		URI:         conn.URI,
		Table:       conn.Table,
		Collection:  conn.Collection,
		Database:    conn.Database,
		DatabaseID:  conn.DatabaseID,
		BaseID:      conn.BaseID,
		TableID:     conn.TableID,
		Project:     conn.Project,
		Prefix:      conn.Prefix,
		IDColumn:    conn.IDColumn,
		Columns:     conn.Columns,
		Credentials: creds.Credentials,
		APIKey:      creds.APIKey,
	}

	src, err := buildAPISource(ir)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer src.Close()

	actor := sanitizeActor(c.Request().Header.Get("X-Actor"))
	if actor == "anonymous" {
		actor = "api-import"
	}

	opts := importer.Options{
		Prefix:   conn.Prefix,
		IDColumn: conn.IDColumn,
		Columns:  conn.Columns,
		Actor:    actor,
	}

	ctx := c.Request().Context()
	stats, err := importer.Run(ctx, src, h.pipe, opts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("import failed: %v", err))
	}

	// Update last_run on the connection
	connStats := &importer.ConnectionStats{
		Imported: stats.Imported,
		Skipped:  stats.Skipped,
		Errors:   stats.Errors,
	}
	_ = h.connStore.UpdateLastRun(id, connStats)

	return c.JSON(http.StatusOK, importResponse{
		Imported: stats.Imported,
		Skipped:  stats.Skipped,
		Errors:   stats.Errors,
	})
}

// RunImport is used by the MCP tool to trigger an import programmatically.
func RunImport(ctx context.Context, req importRequest, pipe interface {
	Write(ctx context.Context, path string, content []byte, actor string) (interface{ ETag() string }, error)
}) (*importResponse, error) {
	return nil, fmt.Errorf("use the REST API for import")
}
