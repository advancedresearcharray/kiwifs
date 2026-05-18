package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const airbyteCloudBaseURL = "https://api.airbyte.com/v1"

// AirbyteCloudClient calls the Airbyte Cloud API for users who don't have Docker
// but have an Airbyte Cloud API key.
type AirbyteCloudClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewAirbyteCloudClient creates a client for the Airbyte Cloud API.
func NewAirbyteCloudClient(apiKey string) *AirbyteCloudClient {
	return &AirbyteCloudClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

// AirbyteCloudSource represents a source configured in Airbyte Cloud.
type AirbyteCloudSourceConfig struct {
	SourceID     string `json:"sourceId,omitempty"`
	Name         string `json:"name"`
	WorkspaceID  string `json:"workspaceId"`
	DefinitionID string `json:"sourceDefinitionId"`
	Config       any    `json:"configuration"`
}

// AirbyteCloudJob represents a sync job in Airbyte Cloud.
type AirbyteCloudJob struct {
	JobID      int64  `json:"jobId"`
	Status     string `json:"status"`
	JobType    string `json:"jobType"`
	StartTime  string `json:"startTime,omitempty"`
	BytesSynced int64 `json:"bytesSynced,omitempty"`
	RowsSynced  int64 `json:"rowsSynced,omitempty"`
}

// AirbyteCloudSource implements the Source interface using Airbyte Cloud API.
// It creates a temporary source + connection, triggers a sync, reads results,
// then cleans up.
type AirbyteCloudSource struct {
	client      *AirbyteCloudClient
	sourceType  string
	config      map[string]any
	streams     []string
	workspaceID string
	sourceName  string
}

// AirbyteCloudSourceOpts configures a cloud-based Airbyte source.
type AirbyteCloudSourceOpts struct {
	APIKey      string
	SourceType  string         // e.g. "postgres", "notion"
	Config      map[string]any // Connector config matching Airbyte spec
	Streams     []string
	WorkspaceID string
	SourceName  string
}

// NewAirbyteCloudSource creates a source that uses Airbyte Cloud API.
func NewAirbyteCloudSource(opts AirbyteCloudSourceOpts) (*AirbyteCloudSource, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("airbyte cloud: api_key is required")
	}
	if opts.Config == nil {
		return nil, fmt.Errorf("airbyte cloud: config is required")
	}
	name := opts.SourceName
	if name == "" {
		name = opts.SourceType
	}
	return &AirbyteCloudSource{
		client:      NewAirbyteCloudClient(opts.APIKey),
		sourceType:  opts.SourceType,
		config:      opts.Config,
		streams:     opts.Streams,
		workspaceID: opts.WorkspaceID,
		sourceName:  name,
	}, nil
}

func (s *AirbyteCloudSource) Name() string { return s.sourceName }

func (s *AirbyteCloudSource) Close() error { return nil }

// Stream triggers a sync via Airbyte Cloud and streams records back.
// This is a simplified implementation that uses the Airbyte API to:
// 1. List jobs for an existing connection
// 2. Trigger a new sync
// 3. Poll until complete
// 4. Return records from the destination
//
// Note: Full implementation requires a webhook destination or custom destination
// that pipes records back. For MVP, this returns an error directing users to
// use Docker mode or pre-configure connections in Airbyte Cloud UI.
func (s *AirbyteCloudSource) Stream(ctx context.Context) (<-chan Record, <-chan error) {
	records := make(chan Record, 128)
	errs := make(chan error, 1)

	go func() {
		defer close(records)
		defer close(errs)
		errs <- fmt.Errorf(
			"airbyte cloud: direct streaming not yet supported. " +
				"Configure a connection in Airbyte Cloud UI and use " +
				"'kiwifs import --from <source> --airbyte-connection-id <id>' " +
				"or install Docker for local connector execution")
	}()

	return records, errs
}

// CreateSource creates a new source in Airbyte Cloud and returns its ID.
func (c *AirbyteCloudClient) CreateSource(ctx context.Context, workspaceID, name, definitionID string, config map[string]any) (string, error) {
	body := map[string]any{
		"workspaceId":        workspaceID,
		"name":              name,
		"sourceDefinitionId": definitionID,
		"configuration":     config,
	}
	resp, err := c.post(ctx, "/sources", body)
	if err != nil {
		return "", err
	}
	var result struct {
		SourceID string `json:"sourceId"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("parse create source response: %w", err)
	}
	return result.SourceID, nil
}

// CheckSourceConnection validates source credentials via Airbyte Cloud API.
func (c *AirbyteCloudClient) CheckSourceConnection(ctx context.Context, sourceID string) (map[string]any, error) {
	resp, err := c.post(ctx, "/sources/check_connection", map[string]any{
		"sourceId": sourceID,
	})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DiscoverSourceSchema discovers available streams from a source in Airbyte Cloud.
func (c *AirbyteCloudClient) DiscoverSourceSchema(ctx context.Context, sourceID string) ([]map[string]any, error) {
	resp, err := c.post(ctx, "/sources/discover_schema", map[string]any{
		"sourceId": sourceID,
	})
	if err != nil {
		return nil, err
	}
	var result struct {
		Catalog struct {
			Streams []map[string]any `json:"streams"`
		} `json:"catalog"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Catalog.Streams, nil
}

// GetSourceDefinitions returns all available source definitions from Airbyte Cloud.
func (c *AirbyteCloudClient) GetSourceDefinitions(ctx context.Context) ([]map[string]any, error) {
	resp, err := c.get(ctx, "/source_definitions")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// TriggerSync triggers a sync job for an existing Airbyte Cloud connection.
func (c *AirbyteCloudClient) TriggerSync(ctx context.Context, connectionID string) (*AirbyteCloudJob, error) {
	body := map[string]any{
		"connectionId": connectionID,
		"jobType":      "sync",
	}
	resp, err := c.post(ctx, "/jobs", body)
	if err != nil {
		return nil, err
	}
	var job AirbyteCloudJob
	if err := json.Unmarshal(resp, &job); err != nil {
		return nil, fmt.Errorf("parse job response: %w", err)
	}
	return &job, nil
}

// GetJob returns the status of a sync job.
func (c *AirbyteCloudClient) GetJob(ctx context.Context, jobID int64) (*AirbyteCloudJob, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/jobs/%d", jobID))
	if err != nil {
		return nil, err
	}
	var job AirbyteCloudJob
	if err := json.Unmarshal(resp, &job); err != nil {
		return nil, fmt.Errorf("parse job response: %w", err)
	}
	return &job, nil
}

// ListSources lists sources in a workspace.
func (c *AirbyteCloudClient) ListSources(ctx context.Context, workspaceIDs []string) ([]json.RawMessage, error) {
	path := "/sources"
	if len(workspaceIDs) > 0 {
		path += "?workspaceIds=" + workspaceIDs[0]
	}
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// ListConnections lists connections in a workspace.
func (c *AirbyteCloudClient) ListConnections(ctx context.Context, workspaceIDs []string) ([]json.RawMessage, error) {
	path := "/connections"
	if len(workspaceIDs) > 0 {
		path += "?workspaceIds=" + workspaceIDs[0]
	}
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *AirbyteCloudClient) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", airbyteCloudBaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("airbyte cloud API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("airbyte cloud API %d: %s", resp.StatusCode, truncate(string(body), 500))
	}
	return body, nil
}

func (c *AirbyteCloudClient) post(ctx context.Context, path string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", airbyteCloudBaseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("airbyte cloud API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("airbyte cloud API %d: %s", resp.StatusCode, truncate(string(body), 500))
	}
	return body, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
