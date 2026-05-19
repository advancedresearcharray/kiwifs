package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	Path        string          `json:"path"` // Directory path (markdown, obsidian, confluence)
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

	// Airbyte-specific fields
	AirbyteConfig map[string]any `json:"airbyte_config,omitempty"` // Raw Airbyte connector config
	AirbyteImage  string         `json:"airbyte_image,omitempty"`  // Custom Docker image override
	Streams       []string       `json:"streams,omitempty"`        // Specific streams to sync
	Via           string         `json:"via,omitempty"`            // "airbyte" | "airbyte-cloud" | "builtin" (auto if empty)
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
	// If explicit via=airbyte or airbyte_config provided, use Airbyte directly
	if req.Via == "airbyte" || req.AirbyteConfig != nil {
		if importer.DockerAvailable() {
			return buildAirbyteSource(req)
		}
		// No Docker — try Airbyte Cloud if config is present
		return buildAirbyteCloudSource(req)
	}

	// For builtin sources, always use native implementation
	if importer.IsBuiltinSource(req.From) {
		return buildBuiltinSource(req)
	}

	// For network sources: prefer Airbyte Docker, then Cloud, then legacy
	if req.Via != "builtin" {
		if importer.DockerAvailable() {
			cfg := req.AirbyteConfig
			if cfg == nil {
				cfg = legacyToAirbyteConfig(req)
			}
			if cfg != nil {
				return buildAirbyteSource(req)
			}
		}
		// No Docker but have Cloud key?
		src, err := buildAirbyteCloudSource(req)
		if err == nil {
			return src, nil
		}
	}

	// Fallback to legacy built-in connectors
	return buildBuiltinSource(req)
}

func buildAirbyteCloudSource(req importRequest) (importer.Source, error) {
	apiKey := os.Getenv("AIRBYTE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("neither Docker nor Airbyte Cloud API key available")
	}
	config := req.AirbyteConfig
	if config == nil {
		config = legacyToAirbyteConfig(req)
	}
	if config == nil {
		return nil, fmt.Errorf("airbyte_config is required")
	}
	workspaceID := os.Getenv("AIRBYTE_WORKSPACE_ID")
	return importer.NewAirbyteCloudSource(importer.AirbyteCloudSourceOpts{
		APIKey:      apiKey,
		SourceType:  req.From,
		Config:      config,
		Streams:     req.Streams,
		WorkspaceID: workspaceID,
		SourceName:  req.From,
	})
}

func buildAirbyteSource(req importRequest) (importer.Source, error) {
	image := req.AirbyteImage
	if image == "" {
		image = importer.LookupAirbyteImage(req.From)
	}
	if image == "" {
		return nil, fmt.Errorf("no Airbyte connector image found for %q", req.From)
	}

	config := req.AirbyteConfig
	if config == nil {
		// Try to build Airbyte config from legacy request fields
		config = legacyToAirbyteConfig(req)
	}
	if config == nil {
		return nil, fmt.Errorf("airbyte_config is required for Airbyte-based import")
	}

	return importer.NewAirbyteSource(importer.AirbyteSourceOpts{
		Image:      image,
		Config:     config,
		Streams:    req.Streams,
		SourceName: req.From,
	})
}

// legacyToAirbyteConfig translates traditional import request fields into
// Airbyte connector config format for backward compatibility.
func legacyToAirbyteConfig(req importRequest) map[string]any {
	switch req.From {
	case "postgres":
		if req.DSN == "" {
			return nil
		}
		return map[string]any{"host": req.DSN, "port": 5432, "database": req.Database}
	case "notion":
		if req.APIKey == "" {
			return nil
		}
		return map[string]any{"credentials": map[string]any{"auth_type": "token", "token": req.APIKey}}
	case "airtable":
		if req.APIKey == "" {
			return nil
		}
		return map[string]any{"credentials": map[string]any{"auth_method": "api_key", "api_key": req.APIKey}}
	case "firestore":
		if len(req.Credentials) == 0 {
			return nil
		}
		return map[string]any{"project_id": req.Project, "service_account_key_json": string(req.Credentials)}
	default:
		return nil
	}
}

func buildBuiltinSource(req importRequest) (importer.Source, error) {
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
	case "markdown":
		if req.Path == "" {
			return nil, fmt.Errorf("path is required for markdown")
		}
		return importer.NewMarkdown(req.Path, importer.MarkdownOpts{})
	default:
		// Try Airbyte as final fallback for unknown source types
		if importer.DockerAvailable() {
			image := importer.LookupAirbyteImage(req.From)
			if image != "" && req.AirbyteConfig != nil {
				return importer.NewAirbyteSource(importer.AirbyteSourceOpts{
					Image:      image,
					Config:     req.AirbyteConfig,
					Streams:    req.Streams,
					SourceName: req.From,
				})
			}
		}
		supported := strings.Join([]string{"markdown", "postgres", "mysql", "firestore", "sqlite", "mongodb", "csv", "json", "jsonl", "notion", "airtable"}, ", ")
		return nil, fmt.Errorf("unknown source type %q (supported: %s). For Airbyte connectors, provide airbyte_config", req.From, supported)
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

	AirbyteConfig map[string]any `json:"airbyte_config,omitempty"`
	AirbyteImage  string         `json:"airbyte_image,omitempty"`
	Streams       []string       `json:"streams,omitempty"`
	Via           string         `json:"via,omitempty"`
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
		From:          req.From,
		DSN:           req.DSN,
		URI:           req.URI,
		DB:            req.DB,
		Table:         req.Table,
		Collection:    req.Collection,
		Database:      req.Database,
		DatabaseID:    req.DatabaseID,
		BaseID:        req.BaseID,
		TableID:       req.TableID,
		Project:       req.Project,
		Credentials:   req.Credentials,
		APIKey:        req.APIKey,
		Limit:         limit,
		AirbyteConfig: req.AirbyteConfig,
		AirbyteImage:  req.AirbyteImage,
		Streams:       req.Streams,
		Via:           req.Via,
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

// --- Airbyte endpoints (Phase 4) ---

type airbyteRequest struct {
	From         string         `json:"from"`
	AirbyteImage string         `json:"airbyte_image,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

// ImportSources lists all available import sources with their backend type.
func (h *Handlers) ImportSources(c echo.Context) error {
	dockerOK := importer.DockerAvailable()
	cloudKey := h.cfg.Import.AirbyteAPIKey
	if cloudKey == "" {
		cloudKey = os.Getenv("AIRBYTE_API_KEY")
	}
	airbyteAvailable := dockerOK || cloudKey != ""
	sources := importer.ListAvailableSources(airbyteAvailable)
	return c.JSON(http.StatusOK, map[string]any{
		"builtin":           sources["builtin"],
		"airbyte":           sources["airbyte"],
		"docker_available":  dockerOK,
		"cloud_key_present": cloudKey != "",
	})
}

// ImportAirbyteSpec returns the connector specification for an Airbyte source.
func (h *Handlers) ImportAirbyteSpec(c echo.Context) error {
	var req airbyteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	image := req.AirbyteImage
	if image == "" {
		image = importer.LookupAirbyteImage(req.From)
	}
	if image == "" {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no Airbyte connector for %q", req.From))
	}

	// Prefer Docker if available
	if importer.DockerAvailable() {
		src, err := importer.NewAirbyteSource(importer.AirbyteSourceOpts{
			Image:  image,
			Config: map[string]any{},
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		ctx := c.Request().Context()
		spec, err := src.Spec(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, spec)
	}

	// No Docker — check for Airbyte Cloud API key
	apiKey := h.cfg.Import.AirbyteAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("AIRBYTE_API_KEY")
	}
	if apiKey != "" {
		return c.JSON(http.StatusOK, map[string]any{
			"mode": "cloud",
			"message": "Connector spec not available without Docker. " +
				"Airbyte Cloud key is configured — use the connection wizard to configure and sync.",
		})
	}

	return echo.NewHTTPError(http.StatusServiceUnavailable,
		"Docker is not available and no Airbyte Cloud API key is configured. "+
			"Either start Docker or set AIRBYTE_API_KEY in your config/environment.")
}

// ImportAirbyteCheck validates connection config against an Airbyte connector.
func (h *Handlers) ImportAirbyteCheck(c echo.Context) error {
	var req airbyteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Config == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "config is required")
	}

	image := req.AirbyteImage
	if image == "" {
		image = importer.LookupAirbyteImage(req.From)
	}
	if image == "" {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no Airbyte connector for %q", req.From))
	}

	if !importer.DockerAvailable() {
		return h.airbyteUnavailableError(c)
	}

	src, err := importer.NewAirbyteSource(importer.AirbyteSourceOpts{
		Image:  image,
		Config: req.Config,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	ctx := c.Request().Context()
	status, err := src.Check(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, status)
}

// ImportAirbyteDiscover returns available streams from an Airbyte connector.
func (h *Handlers) ImportAirbyteDiscover(c echo.Context) error {
	var req airbyteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Config == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "config is required")
	}

	image := req.AirbyteImage
	if image == "" {
		image = importer.LookupAirbyteImage(req.From)
	}
	if image == "" {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no Airbyte connector for %q", req.From))
	}

	if !importer.DockerAvailable() {
		return h.airbyteUnavailableError(c)
	}

	src, err := importer.NewAirbyteSource(importer.AirbyteSourceOpts{
		Image:  image,
		Config: req.Config,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	ctx := c.Request().Context()
	catalog, err := src.Discover(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, catalog)
}

// airbyteUnavailableError returns a structured error when Docker is missing.
// If an Airbyte Cloud key is configured, hints at that path instead.
func (h *Handlers) airbyteUnavailableError(c echo.Context) error {
	apiKey := h.cfg.Import.AirbyteAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("AIRBYTE_API_KEY")
	}
	if apiKey != "" {
		return echo.NewHTTPError(http.StatusServiceUnavailable,
			"Docker is not available. Airbyte Cloud API key is configured — "+
				"use via=airbyte-cloud or configure your connection through the Cloud dashboard.")
	}
	return echo.NewHTTPError(http.StatusServiceUnavailable,
		"Docker is not available and no Airbyte Cloud API key is configured. "+
			"Either start Docker or set AIRBYTE_API_KEY in your config/environment.")
}

// --- Airbyte Cloud handlers ---

// getAirbyteCloudClient returns a configured Airbyte Cloud client or an error.
func (h *Handlers) getAirbyteCloudClient() (*importer.AirbyteCloudClient, error) {
	apiKey := h.cfg.Import.AirbyteAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("AIRBYTE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no Airbyte Cloud API key configured (set airbyte_api_key in config or AIRBYTE_API_KEY env)")
	}
	return importer.NewAirbyteCloudClient(apiKey), nil
}

func (h *Handlers) getAirbyteWorkspaceID() string {
	ws := h.cfg.Import.AirbyteWorkspaceID
	if ws == "" {
		ws = os.Getenv("AIRBYTE_WORKSPACE_ID")
	}
	return ws
}

// ImportAirbyteCloudCheck creates a temporary source in Airbyte Cloud and validates the connection.
func (h *Handlers) ImportAirbyteCloudCheck(c echo.Context) error {
	client, err := h.getAirbyteCloudClient()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, err.Error())
	}

	var req struct {
		From   string         `json:"from"`
		Config map[string]any `json:"config"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Config == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "config is required")
	}

	wsID := h.getAirbyteWorkspaceID()
	if wsID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "airbyte_workspace_id not configured")
	}

	defID := importer.LookupAirbyteDefinitionID(req.From)
	if defID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no Airbyte Cloud definition for %q", req.From))
	}

	ctx := c.Request().Context()

	sourceID, err := client.CreateSource(ctx, wsID, "kiwifs-check-"+req.From, defID, req.Config)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("Airbyte Cloud create source: %v", err))
	}

	result, err := client.CheckSourceConnection(ctx, sourceID)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"status":    "failed",
			"message":   err.Error(),
			"source_id": sourceID,
		})
	}
	result["source_id"] = sourceID
	return c.JSON(http.StatusOK, result)
}

// ImportAirbyteCloudDiscover uses Airbyte Cloud API to discover streams from a source.
func (h *Handlers) ImportAirbyteCloudDiscover(c echo.Context) error {
	client, err := h.getAirbyteCloudClient()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, err.Error())
	}

	var req struct {
		SourceID string `json:"source_id"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.SourceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "source_id is required")
	}

	ctx := c.Request().Context()
	streams, err := client.DiscoverSourceSchema(ctx, req.SourceID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("Airbyte Cloud discover: %v", err))
	}
	return c.JSON(http.StatusOK, map[string]any{"streams": streams})
}

// ImportAirbyteCloudConnections lists existing Airbyte Cloud connections.
func (h *Handlers) ImportAirbyteCloudConnections(c echo.Context) error {
	client, err := h.getAirbyteCloudClient()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, err.Error())
	}

	wsID := h.getAirbyteWorkspaceID()
	var wsIDs []string
	if wsID != "" {
		wsIDs = []string{wsID}
	}

	ctx := c.Request().Context()
	conns, err := client.ListConnections(ctx, wsIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("Airbyte Cloud: %v", err))
	}
	return c.JSON(http.StatusOK, map[string]any{"connections": conns})
}

// ImportAirbyteCloudSync triggers a sync for an existing Airbyte Cloud connection.
func (h *Handlers) ImportAirbyteCloudSync(c echo.Context) error {
	client, err := h.getAirbyteCloudClient()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, err.Error())
	}

	var req struct {
		ConnectionID string `json:"connection_id"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.ConnectionID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "connection_id is required")
	}

	ctx := c.Request().Context()
	job, err := client.TriggerSync(ctx, req.ConnectionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("Airbyte Cloud sync: %v", err))
	}
	return c.JSON(http.StatusOK, job)
}

// --- File upload import endpoint ---

const maxImportUploadSize = 256 << 20 // 256 MB

// ImportUpload accepts a multipart file upload, writes it to a temp file, then
// runs the import pipeline against it. This lets browser users drag-and-drop a
// file (CSV, JSON, JSONL, YAML, Excel, SQLite) instead of typing server paths.
func (h *Handlers) ImportUpload(c echo.Context) error {
	from := c.FormValue("from")
	if from == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "from is required")
	}
	supported := map[string]bool{"csv": true, "json": true, "jsonl": true, "yaml": true, "excel": true, "sqlite": true}
	if !supported[from] {
		return echo.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("file upload not supported for %q — use the path-based import", from))
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file field is required")
	}
	if file.Size > maxImportUploadSize {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge,
			fmt.Sprintf("file exceeds %d MB limit", maxImportUploadSize>>20))
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open upload")
	}
	defer src.Close()

	// Write to temp file so the importer can read it
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		switch from {
		case "csv":
			ext = ".csv"
		case "json":
			ext = ".json"
		case "jsonl":
			ext = ".jsonl"
		case "yaml":
			ext = ".yaml"
		case "excel":
			ext = ".xlsx"
		case "sqlite":
			ext = ".db"
		}
	}
	tmp, err := os.CreateTemp("", "kiwifs-import-*"+ext)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create temp file")
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to write temp file")
	}
	tmp.Close()

	// Parse optional form fields
	prefix := c.FormValue("prefix")
	idColumn := c.FormValue("id_column")
	table := c.FormValue("table") // for sqlite
	query := c.FormValue("query") // for sqlite

	// Determine what to call the data if no prefix given
	if prefix == "" {
		prefix = strings.TrimSuffix(filepath.Base(file.Filename), ext)
	}

	// Run the preview (dry_run) or actual import
	mode := c.FormValue("mode") // "preview" or "import" (default: import)

	var ir importRequest
	ir.From = from
	ir.Prefix = prefix
	ir.IDColumn = idColumn
	ir.Table = table
	ir.Query = query

	switch from {
	case "csv", "json", "jsonl", "yaml", "excel":
		ir.File = tmpPath
	case "sqlite":
		ir.DB = tmpPath
		if ir.Table == "" && ir.Query == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "table or query is required for sqlite")
		}
	}

	if mode == "preview" {
		ir.Limit = 5
		apiSrc, err := buildAPISource(ir)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		defer apiSrc.Close()

		ctx := c.Request().Context()
		records, errs := apiSrc.Stream(ctx)

		var previews []previewRecord
		count := 0
		for rec := range records {
			if count >= 5 {
				break
			}
			fm := make(map[string]any, len(rec.Fields)+2)
			for k, v := range rec.Fields {
				fm[k] = v
			}
			fm["_source"] = apiSrc.Name()
			fm["_source_id"] = rec.SourceID

			title := rec.PrimaryKey
			if t, ok := rec.Fields["title"].(string); ok && t != "" {
				title = t
			} else if t, ok := rec.Fields["name"].(string); ok && t != "" {
				title = t
			}

			path := fmt.Sprintf("%s/%s.md", prefix, importer.SanitizePath(rec.PrimaryKey))
			body := fmt.Sprintf("# %s\n\n> Auto-imported from %s (row %s)", title, rec.Table, rec.SourceID)

			previews = append(previews, previewRecord{
				Path:        path,
				Frontmatter: fm,
				BodyPreview: body,
			})
			count++
		}
		for err := range errs {
			if err != nil && len(previews) == 0 {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
		}
		return c.JSON(http.StatusOK, previewResponse{Records: previews})
	}

	// Run actual import
	apiSrc, err := buildAPISource(ir)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer apiSrc.Close()

	actor := sanitizeActor(c.Request().Header.Get("X-Actor"))
	if actor == "anonymous" {
		actor = "api-import"
	}

	opts := importer.Options{
		Prefix:   ir.Prefix,
		IDColumn: ir.IDColumn,
		Actor:    actor,
		Limit:    ir.Limit,
	}

	ctx := c.Request().Context()
	stats, err := importer.Run(ctx, apiSrc, h.pipe, opts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("import failed: %v", err))
	}

	return c.JSON(http.StatusOK, importResponse{
		Imported: stats.Imported,
		Skipped:  stats.Skipped,
		Errors:   stats.Errors,
	})
}
