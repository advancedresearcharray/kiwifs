package importer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Airbyte protocol message types.
const (
	AirbyteMessageRecord           = "RECORD"
	AirbyteMessageState            = "STATE"
	AirbyteMessageLog              = "LOG"
	AirbyteMessageSpec             = "SPEC"
	AirbyteMessageConnectionStatus = "CONNECTION_STATUS"
	AirbyteMessageCatalog          = "CATALOG"
	AirbyteMessageTrace            = "TRACE"
)

// AirbyteMessage is the top-level envelope for all Airbyte protocol messages.
type AirbyteMessage struct {
	Type             string                    `json:"type"`
	Record           *AirbyteRecordMessage     `json:"record,omitempty"`
	State            *AirbyteStateMessage      `json:"state,omitempty"`
	Log              *AirbyteLogMessage        `json:"log,omitempty"`
	Spec             *AirbyteSpecMessage       `json:"spec,omitempty"`
	ConnectionStatus *AirbyteConnectionStatus  `json:"connectionStatus,omitempty"`
	Catalog          *AirbyteCatalog           `json:"catalog,omitempty"`
	Trace            *AirbyteTraceMessage      `json:"trace,omitempty"`
}

type AirbyteRecordMessage struct {
	Stream    string         `json:"stream"`
	Data      map[string]any `json:"data"`
	EmittedAt int64          `json:"emitted_at"`
	Namespace string         `json:"namespace,omitempty"`
}

type AirbyteStateMessage struct {
	Data json.RawMessage `json:"data,omitempty"`
	Type string          `json:"type,omitempty"`
}

type AirbyteLogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type AirbyteSpecMessage struct {
	ConnectionSpecification json.RawMessage `json:"connectionSpecification"`
	DocumentationURL        string          `json:"documentationUrl,omitempty"`
}

type AirbyteConnectionStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type AirbyteCatalog struct {
	Streams []AirbyteStream `json:"streams"`
}

type AirbyteStream struct {
	Name                    string          `json:"name"`
	Namespace               string          `json:"namespace,omitempty"`
	JSONSchema              json.RawMessage `json:"json_schema,omitempty"`
	SupportedSyncModes      []string        `json:"supported_sync_modes,omitempty"`
	DefaultCursorField      []string        `json:"default_cursor_field,omitempty"`
	SourceDefinedPrimaryKey [][]string      `json:"source_defined_primary_key,omitempty"`
}

// ConfiguredCatalog is what you send to the read command.
type ConfiguredCatalog struct {
	Streams []ConfiguredStream `json:"streams"`
}

type ConfiguredStream struct {
	Stream              AirbyteStream `json:"stream"`
	SyncMode            string        `json:"sync_mode"`
	DestinationSyncMode string        `json:"destination_sync_mode"`
}

// AirbyteTraceMessage handles trace/error messages from connectors.
type AirbyteTraceMessage struct {
	Type      string                `json:"type"`
	EmittedAt float64               `json:"emitted_at"`
	Error     *AirbyteTraceError    `json:"error,omitempty"`
	Estimate  *AirbyteTraceEstimate `json:"estimate,omitempty"`
}

type AirbyteTraceError struct {
	Message       string `json:"message"`
	InternalMsg   string `json:"internal_message,omitempty"`
	StackTrace    string `json:"stack_trace,omitempty"`
	FailureType   string `json:"failure_type,omitempty"`
}

type AirbyteTraceEstimate struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	RowCount  int64  `json:"row_estimate,omitempty"`
	ByteCount int64  `json:"byte_estimate,omitempty"`
}

// AirbyteSource implements Source by running an Airbyte connector Docker image.
type AirbyteSource struct {
	image       string
	config      map[string]any
	streams     []string // if empty, sync all discovered streams
	sourceName  string
	dockerHost  string // optional DOCKER_HOST override
}

// AirbyteSourceOpts configures how to run an Airbyte source connector.
type AirbyteSourceOpts struct {
	Image      string         // Docker image, e.g. "airbyte/source-postgres:latest"
	Config     map[string]any // Connector config (matches spec's connectionSpecification)
	Streams    []string       // Specific streams to sync; empty = all
	SourceName string         // Display name for the source
	DockerHost string         // Optional DOCKER_HOST override
}

// NewAirbyteSource creates a source that runs an Airbyte connector via Docker.
func NewAirbyteSource(opts AirbyteSourceOpts) (*AirbyteSource, error) {
	if opts.Image == "" {
		return nil, fmt.Errorf("airbyte: image is required")
	}
	if opts.Config == nil {
		return nil, fmt.Errorf("airbyte: config is required")
	}
	name := opts.SourceName
	if name == "" {
		parts := strings.Split(opts.Image, "/")
		last := parts[len(parts)-1]
		name = strings.TrimPrefix(strings.Split(last, ":")[0], "source-")
	}
	return &AirbyteSource{
		image:      opts.Image,
		config:     opts.Config,
		streams:    opts.Streams,
		sourceName: name,
		dockerHost: opts.DockerHost,
	}, nil
}

func (s *AirbyteSource) Name() string { return s.sourceName }

func (s *AirbyteSource) Close() error { return nil }

// Stream runs the Airbyte connector and emits Records.
func (s *AirbyteSource) Stream(ctx context.Context) (<-chan Record, <-chan error) {
	records := make(chan Record, 128)
	errs := make(chan error, 1)

	go func() {
		defer close(records)
		defer close(errs)

		configFile, err := s.writeTempJSON("airbyte-config-*.json", s.config)
		if err != nil {
			errs <- fmt.Errorf("airbyte: write config: %w", err)
			return
		}
		defer os.Remove(configFile)

		// Discover catalog if no specific streams provided, or build configured catalog
		catalog, err := s.discoverOrBuild(ctx, configFile)
		if err != nil {
			errs <- fmt.Errorf("airbyte: catalog: %w", err)
			return
		}

		catalogFile, err := s.writeTempJSON("airbyte-catalog-*.json", catalog)
		if err != nil {
			errs <- fmt.Errorf("airbyte: write catalog: %w", err)
			return
		}
		defer os.Remove(catalogFile)

		// Run the read command
		args := s.dockerArgs("read",
			"--config", "/tmp/config.json",
			"--catalog", "/tmp/catalog.json",
		)
		args = append(s.dockerRunPrefix(configFile, catalogFile), args...)

		cmd := exec.CommandContext(ctx, "docker", args...)
		if s.dockerHost != "" {
			cmd.Env = append(os.Environ(), "DOCKER_HOST="+s.dockerHost)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errs <- fmt.Errorf("airbyte: stdout pipe: %w", err)
			return
		}
		cmd.Stderr = io.Discard

		if err := cmd.Start(); err != nil {
			errs <- fmt.Errorf("airbyte: start docker: %w", err)
			return
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 4*1024*1024), 16*1024*1024) // 16MB line buffer for large records

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var msg AirbyteMessage
			if err := json.Unmarshal(line, &msg); err != nil {
				continue // skip non-JSON lines (Docker logs, connector debug output)
			}

			switch msg.Type {
			case AirbyteMessageRecord:
				if msg.Record == nil {
					continue
				}
				recs := s.recordsFromAirbyte(msg.Record)
				for _, rec := range recs {
					select {
					case records <- rec:
					case <-ctx.Done():
						_ = cmd.Process.Kill()
						return
					}
				}
			case AirbyteMessageTrace:
				if msg.Trace != nil && msg.Trace.Error != nil {
					errs <- fmt.Errorf("airbyte connector error: %s", msg.Trace.Error.Message)
					_ = cmd.Process.Kill()
					return
				}
			case AirbyteMessageLog:
				// could log these if needed
			}
		}

		if err := cmd.Wait(); err != nil {
			if ctx.Err() == nil {
				errs <- fmt.Errorf("airbyte: docker exited: %w", err)
			}
		}
	}()

	return records, errs
}

func (s *AirbyteSource) recordsFromAirbyte(rec *AirbyteRecordMessage) []Record {
	// Detect Firebase RTDB key/value pattern: {"key": "...", "value": "{...json...}"}
	if keyVal, hasKey := rec.Data["key"]; hasKey {
		if valStr, hasVal := rec.Data["value"]; hasVal {
			if keyStr, ok := keyVal.(string); ok {
				return s.explodeRTDBRecord(rec, keyStr, valStr)
			}
		}
	}
	return []Record{s.recordToImporterRecord(rec)}
}

// explodeRTDBRecord takes a Firebase RTDB key/value record and explodes nested
// objects into individual records. For example, key="users" with value containing
// {user1: {...}, user2: {...}} produces separate records for each user.
func (s *AirbyteSource) explodeRTDBRecord(rec *AirbyteRecordMessage, key string, rawValue any) []Record {
	var valueMap map[string]any

	switch v := rawValue.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &valueMap); err != nil {
			// Not a JSON object string — treat as a simple leaf value
			return []Record{s.makeLeafRecord(rec, key, rawValue)}
		}
	case map[string]any:
		valueMap = v
	default:
		return []Record{s.makeLeafRecord(rec, key, rawValue)}
	}

	// Check if the value map contains nested objects (collection pattern)
	// e.g., users: {user1: {name: "Alice", ...}, user2: {name: "Bob", ...}}
	hasNestedObjects := false
	for _, v := range valueMap {
		if _, isMap := v.(map[string]any); isMap {
			hasNestedObjects = true
			break
		}
	}

	if !hasNestedObjects {
		// Flat object — render as a single record with fields as frontmatter
		return []Record{{
			SourceID:   fmt.Sprintf("airbyte:%s:%s:%s", s.sourceName, rec.Stream, key),
			SourceDSN:  s.image,
			Table:      rec.Stream,
			Fields:     valueMap,
			PrimaryKey: key,
		}}
	}

	// Collection of nested objects — explode into one record per child
	var records []Record
	for childKey, childVal := range valueMap {
		childMap, isMap := childVal.(map[string]any)
		if !isMap {
			// Scalar child — create a simple record
			childMap = map[string]any{"value": childVal}
		}
		childMap["_parent"] = key
		pk := key + "/" + childKey
		records = append(records, Record{
			SourceID:   fmt.Sprintf("airbyte:%s:%s:%s", s.sourceName, rec.Stream, pk),
			SourceDSN:  s.image,
			Table:      rec.Stream,
			Fields:     childMap,
			PrimaryKey: pk,
		})
	}
	return records
}

func (s *AirbyteSource) makeLeafRecord(rec *AirbyteRecordMessage, key string, value any) Record {
	fields := map[string]any{"value": value}
	return Record{
		SourceID:   fmt.Sprintf("airbyte:%s:%s:%s", s.sourceName, rec.Stream, key),
		SourceDSN:  s.image,
		Table:      rec.Stream,
		Fields:     fields,
		PrimaryKey: key,
	}
}

func (s *AirbyteSource) recordToImporterRecord(rec *AirbyteRecordMessage) Record {
	pk := ""
	if id, ok := rec.Data["id"]; ok {
		pk = fmt.Sprintf("%v", id)
	} else if id, ok := rec.Data["_id"]; ok {
		pk = fmt.Sprintf("%v", id)
	} else if id, ok := rec.Data["ID"]; ok {
		pk = fmt.Sprintf("%v", id)
	}
	if pk == "" {
		pk = fmt.Sprintf("%d", rec.EmittedAt)
	}

	stream := rec.Stream
	if rec.Namespace != "" {
		stream = rec.Namespace + "." + rec.Stream
	}

	return Record{
		SourceID:   fmt.Sprintf("airbyte:%s:%s:%s", s.sourceName, rec.Stream, pk),
		SourceDSN:  s.image,
		Table:      stream,
		Fields:     rec.Data,
		PrimaryKey: pk,
	}
}

// Spec returns the connector's specification (what config it needs).
func (s *AirbyteSource) Spec(ctx context.Context) (*AirbyteSpecMessage, error) {
	args := s.dockerRunPrefix("", "")
	args = append(args, s.dockerArgs("spec")...)

	out, err := s.runDocker(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("airbyte spec: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		var msg AirbyteMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Type == AirbyteMessageSpec && msg.Spec != nil {
			return msg.Spec, nil
		}
	}
	return nil, fmt.Errorf("airbyte spec: no SPEC message in output")
}

// Check validates the connection config.
func (s *AirbyteSource) Check(ctx context.Context) (*AirbyteConnectionStatus, error) {
	configFile, err := s.writeTempJSON("airbyte-config-*.json", s.config)
	if err != nil {
		return nil, err
	}
	defer os.Remove(configFile)

	args := s.dockerRunPrefix(configFile, "")
	args = append(args, s.dockerArgs("check", "--config", "/tmp/config.json")...)

	out, err := s.runDocker(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("airbyte check: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		var msg AirbyteMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Type == AirbyteMessageConnectionStatus && msg.ConnectionStatus != nil {
			return msg.ConnectionStatus, nil
		}
	}
	return nil, fmt.Errorf("airbyte check: no CONNECTION_STATUS in output")
}

// Discover returns available streams from the source.
func (s *AirbyteSource) Discover(ctx context.Context) (*AirbyteCatalog, error) {
	configFile, err := s.writeTempJSON("airbyte-config-*.json", s.config)
	if err != nil {
		return nil, err
	}
	defer os.Remove(configFile)

	args := s.dockerRunPrefix(configFile, "")
	args = append(args, s.dockerArgs("discover", "--config", "/tmp/config.json")...)

	out, err := s.runDocker(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("airbyte discover: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		var msg AirbyteMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Type == AirbyteMessageCatalog && msg.Catalog != nil {
			return msg.Catalog, nil
		}
	}
	return nil, fmt.Errorf("airbyte discover: no CATALOG in output")
}

func (s *AirbyteSource) discoverOrBuild(ctx context.Context, configFile string) (*ConfiguredCatalog, error) {
	args := s.dockerRunPrefix(configFile, "")
	args = append(args, s.dockerArgs("discover", "--config", "/tmp/config.json")...)

	out, err := s.runDocker(ctx, args)
	if err != nil {
		return nil, err
	}

	var catalog *AirbyteCatalog
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		var msg AirbyteMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Type == AirbyteMessageCatalog && msg.Catalog != nil {
			catalog = msg.Catalog
			break
		}
	}
	if catalog == nil {
		return nil, fmt.Errorf("no catalog discovered")
	}

	configured := &ConfiguredCatalog{}
	for _, stream := range catalog.Streams {
		if len(s.streams) > 0 && !contains(s.streams, stream.Name) {
			continue
		}
		syncMode := "full_refresh"
		if len(stream.SupportedSyncModes) > 0 {
			syncMode = stream.SupportedSyncModes[0]
		}
		configured.Streams = append(configured.Streams, ConfiguredStream{
			Stream:              stream,
			SyncMode:            syncMode,
			DestinationSyncMode: "append",
		})
	}

	if len(configured.Streams) == 0 {
		return nil, fmt.Errorf("no streams matched filter (available: %d)", len(catalog.Streams))
	}

	return configured, nil
}

func (s *AirbyteSource) dockerRunPrefix(configFile, catalogFile string) []string {
	args := []string{"run", "--rm", "-i"}

	if configFile != "" {
		args = append(args, "-v", configFile+":/tmp/config.json:ro")
	}
	if catalogFile != "" {
		args = append(args, "-v", catalogFile+":/tmp/catalog.json:ro")
	}

	return args
}

func (s *AirbyteSource) dockerArgs(command string, extra ...string) []string {
	args := []string{s.image, command}
	args = append(args, extra...)
	return args
}

func (s *AirbyteSource) runDocker(ctx context.Context, args []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", args...)
	if s.dockerHost != "" {
		cmd.Env = append(os.Environ(), "DOCKER_HOST="+s.dockerHost)
	}

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			detail := string(exitErr.Stderr)
			if detail == "" && len(out) > 0 {
				detail = string(out)
			}
			if len(detail) > 500 {
				detail = detail[:500]
			}
			return nil, fmt.Errorf("docker exit %d: %s", exitErr.ExitCode(), detail)
		}
		return nil, err
	}
	return out, nil
}

func (s *AirbyteSource) writeTempJSON(pattern string, data any) (string, error) {
	tmp, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	enc := json.NewEncoder(tmp)
	if err := enc.Encode(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	tmp.Close()
	os.Chmod(tmp.Name(), 0644)
	return tmp.Name(), nil
}

// DockerAvailable checks if Docker is installed and responsive.
func DockerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "info")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
