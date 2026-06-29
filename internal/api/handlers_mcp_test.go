package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kiwifs/kiwifs/internal/bootstrap"
	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/kiwifs/kiwifs/internal/mcpserver"
)

func setupMCPAPIServer(t *testing.T) http.Handler {
	t.Helper()
	dir := t.TempDir()
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgToml := `[search]
engine = "grep"
[versioning]
strategy = "none"
`
	if err := os.WriteFile(filepath.Join(kiwiDir, "config.toml"), []byte(cfgToml), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Search:     config.SearchConfig{Engine: "grep"},
		Versioning: config.VersioningConfig{Strategy: "none"},
	}
	stack, err := bootstrap.Build("default", dir, cfg)
	if err != nil {
		t.Fatalf("bootstrap.Build: %v", err)
	}
	t.Cleanup(func() { _ = stack.Close() })

	mcpSrv, _, err := mcpserver.New(mcpserver.Options{
		Backend: mcpserver.NewStackBackend(stack),
		Emitter: stack.Emitter,
	})
	if err != nil {
		t.Fatalf("mcpserver.New: %v", err)
	}
	stack.Server.SetMCPHandler(mcpserver.StreamableHTTPHandler(mcpSrv, mcpserver.AuthTokenFromConfig(stack.Config)))
	return stack.Server
}

func TestMCPStreamableHTTPInitialize(t *testing.T) {
	srv := setupMCPAPIServer(t)

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code == http.StatusMethodNotAllowed {
		t.Fatalf("POST /mcp returned 405 — MCP route not mounted before UI catch-all")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /mcp status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	resp := rec.Body.String()
	if !strings.Contains(resp, `"result"`) {
		t.Fatalf("expected JSON-RPC result, got: %s", resp)
	}
	if strings.Contains(resp, "<html") || strings.Contains(resp, "<!DOCTYPE") {
		t.Fatalf("expected JSON, got HTML: %s", resp[:min(200, len(resp))])
	}
	if got := rec.Header().Get("Mcp-Method"); got != "initialize" {
		t.Fatalf("Mcp-Method = %q, want initialize", got)
	}
	if rec.Header().Get("Mcp-Session-Id") != "" {
		t.Fatal("stateless response must not include Mcp-Session-Id")
	}
}

func TestMCPStreamableHTTPGetIsNotHTML(t *testing.T) {
	srv := setupMCPAPIServer(t)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/mcp", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if resp == nil {
		t.Fatalf("GET /mcp: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusMethodNotAllowed {
		t.Fatalf("GET /mcp returned 405")
	}
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want SSE not HTML", ct)
	}
	if !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	if len(body) > 0 && (strings.Contains(string(body), "<html") || strings.Contains(string(body), "<!DOCTYPE")) {
		t.Fatalf("GET /mcp returned HTML body")
	}
}

func TestMCPStreamableHTTPRequiresAuthWhenConfigured(t *testing.T) {
	dir := t.TempDir()
	kiwiDir := filepath.Join(dir, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kiwiDir, "config.toml"), []byte(`
[search]
engine = "grep"
[versioning]
strategy = "none"
[auth]
type = "apikey"
api_key = "secret-key"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	stack, err := bootstrap.Build("default", dir, cfg)
	if err != nil {
		t.Fatalf("bootstrap.Build: %v", err)
	}
	t.Cleanup(func() { _ = stack.Close() })

	mcpSrv, _, err := mcpserver.New(mcpserver.Options{
		Backend: mcpserver.NewStackBackend(stack),
		Emitter: stack.Emitter,
	})
	if err != nil {
		t.Fatalf("mcpserver.New: %v", err)
	}
	stack.Server.SetMCPHandler(mcpserver.StreamableHTTPHandler(mcpSrv, mcpserver.AuthTokenFromConfig(stack.Config)))

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	stack.Server.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing auth status = %d, want 401", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer secret-key")
	rec2 := httptest.NewRecorder()
	stack.Server.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("authed status = %d, want 200; body: %s", rec2.Code, rec2.Body.String())
	}
}

func TestMCPStreamableHTTPServerDiscover(t *testing.T) {
	srv := setupMCPAPIServer(t)

	body := `{"jsonrpc":"2.0","id":"api-disc-1","method":"server/discover","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /mcp server/discover status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Mcp-Method"); got != "server/discover" {
		t.Fatalf("Mcp-Method = %q, want server/discover", got)
	}
	if got := rec.Header().Get("MCP-Protocol-Version"); got != "2026-07-28" {
		t.Fatalf("MCP-Protocol-Version = %q, want 2026-07-28", got)
	}
	if rec.Header().Get("Mcp-Session-Id") != "" {
		t.Fatal("stateless response must not include Mcp-Session-Id")
	}

	var resp struct {
		Result struct {
			SupportedVersions []string       `json:"supportedVersions"`
			Capabilities      map[string]any `json:"capabilities"`
			TTLMs             int            `json:"ttlMs"`
			CacheScope        string         `json:"cacheScope"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Result.SupportedVersions) == 0 || resp.Result.SupportedVersions[0] != "2026-07-28" {
		t.Fatalf("supportedVersions = %v", resp.Result.SupportedVersions)
	}
	if resp.Result.TTLMs != 3_600_000 {
		t.Fatalf("ttlMs = %d, want 3600000", resp.Result.TTLMs)
	}
	if resp.Result.CacheScope != "public" {
		t.Fatalf("cacheScope = %q, want public", resp.Result.CacheScope)
	}
	if _, ok := resp.Result.Capabilities["tools"]; !ok {
		t.Fatalf("expected tools capability, got %v", resp.Result.Capabilities)
	}
}

func TestMCPStreamableHTTPToolsListCachingHeaders(t *testing.T) {
	srv := setupMCPAPIServer(t)

	body := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Mcp-Method", "tools/list")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /mcp tools/list status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Mcp-Method"); got != "tools/list" {
		t.Fatalf("Mcp-Method = %q, want tools/list", got)
	}
	if rec.Header().Get("Mcp-Session-Id") != "" {
		t.Fatal("stateless response must not include Mcp-Session-Id")
	}

	var resp struct {
		Result struct {
			TTLMs      int `json:"ttlMs"`
			CacheScope string `json:"cacheScope"`
			Tools      []struct {
				InputSchema map[string]any `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Result.TTLMs != 300_000 {
		t.Fatalf("ttlMs = %d, want 300000", resp.Result.TTLMs)
	}
	if resp.Result.CacheScope != "public" {
		t.Fatalf("cacheScope = %q, want public", resp.Result.CacheScope)
	}
	if len(resp.Result.Tools) == 0 {
		t.Fatal("expected tools")
	}
	if got := resp.Result.Tools[0].InputSchema["$schema"]; got != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("inputSchema.$schema = %v", got)
	}
}

func TestMCPRouteDoesNotShadowUI(t *testing.T) {
	srv := setupMCPAPIServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	// UI may be not-built in test env; just ensure /mcp is not served as /.
	reqMCP := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`))
	reqMCP.Header.Set("Content-Type", "application/json")
	recMCP := httptest.NewRecorder()
	srv.ServeHTTP(recMCP, reqMCP)
	if recMCP.Code == http.StatusMethodNotAllowed {
		t.Fatal("POST /mcp should not fall through to UI catch-all")
	}
	_, _ = io.ReadAll(rec.Body)
}
