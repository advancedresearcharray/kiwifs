package mcpserver_test

import (
	"bytes"
	"encoding/json"
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

func setupServeMCP(t *testing.T) (*bootstrap.Stack, string) {
	t.Helper()
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
`), 0o644); err != nil {
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
	stack.Server.SetMCPHandler(mcpserver.StreamableHTTPHandler(mcpSrv, ""))
	return stack, dir
}

func TestMCPStreamableGETNotServesUI(t *testing.T) {
	stack, _ := setupServeMCP(t)

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Accept", "text/event-stream")
	rec := httptest.NewRecorder()

	go stack.Server.ServeHTTP(rec, req)
	time.Sleep(100 * time.Millisecond)

	body := rec.Body.String()
	if strings.Contains(body, "<!DOCTYPE") || strings.Contains(strings.ToLower(body), "<html") {
		t.Fatalf("GET /mcp served UI HTML: status=%d body=%q", rec.Code, body)
	}
	if ct := rec.Header().Get("Content-Type"); strings.Contains(ct, "text/html") {
		t.Fatalf("GET /mcp Content-Type = %q, want MCP transport not UI", ct)
	}
}

func TestMCPStreamablePOSTNot405(t *testing.T) {
	stack, _ := setupServeMCP(t)

	body := `{"jsonrpc":"2.0","id":"d1","method":"server/discover","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	stack.Server.ServeHTTP(rec, req)

	if rec.Code == http.StatusMethodNotAllowed {
		t.Fatal("POST /mcp returned 405 — MCP route lost to UI catch-all")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /mcp status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func TestMCP2026DiscoverViaServe(t *testing.T) {
	stack, _ := setupServeMCP(t)

	body := `{"jsonrpc":"2.0","id":"d1","method":"server/discover","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	stack.Server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get(mcpserver.HeaderMCPMethod); got != mcpserver.MethodServerDiscover {
		t.Fatalf("Mcp-Method = %q, want %q", got, mcpserver.MethodServerDiscover)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("result = %T", resp["result"])
	}
	if result["cacheScope"] != mcpserver.CacheScopePublic {
		t.Fatalf("cacheScope = %v", result["cacheScope"])
	}
	if _, ok := result["ttlMs"]; !ok {
		t.Fatal("expected ttlMs on discover result")
	}
	versions, ok := result["supportedVersions"].([]any)
	if !ok || len(versions) == 0 {
		t.Fatal("expected supportedVersions")
	}
	if versions[0] != mcpserver.ProtocolVersion20260728 {
		t.Fatalf("first supported version = %v, want %s", versions[0], mcpserver.ProtocolVersion20260728)
	}
}

func TestMCP2026ToolsListRoutingAndCache(t *testing.T) {
	stack, _ := setupServeMCP(t)

	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	initReq := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(initBody)))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("Accept", "application/json, text/event-stream")
	initRec := httptest.NewRecorder()
	stack.Server.ServeHTTP(initRec, initReq)
	if initRec.Code != http.StatusOK {
		t.Fatalf("initialize status = %d, body=%s", initRec.Code, initRec.Body.String())
	}

	listBody := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	listReq := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(listBody)))
	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("Accept", "application/json, text/event-stream")
	listReq.Header.Set(mcpserver.HeaderMCPMethod, "tools/list")
	listRec := httptest.NewRecorder()
	stack.Server.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("tools/list status = %d, body=%s", listRec.Code, listRec.Body.String())
	}
	if got := listRec.Header().Get(mcpserver.HeaderMCPMethod); got != "tools/list" {
		t.Fatalf("response Mcp-Method = %q", got)
	}

	var resp map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	result := resp["result"].(map[string]any)
	if result["ttlMs"] == nil {
		t.Fatalf("missing ttlMs: %v", result)
	}
	if result["cacheScope"] != mcpserver.CacheScopePublic {
		t.Fatalf("cacheScope = %v", result["cacheScope"])
	}
}

func TestSetMCPHandlerIdempotent(t *testing.T) {
	stack, dir := setupServeMCP(t)

	mcpSrv, _, err := mcpserver.New(mcpserver.Options{Root: dir})
	if err != nil {
		t.Fatalf("mcpserver.New: %v", err)
	}
	stack.Server.SetMCPHandler(mcpserver.StreamableHTTPHandler(mcpSrv, "other-token"))

	body := `{"jsonrpc":"2.0","id":"d1","method":"server/discover","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	stack.Server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("second SetMCPHandler broke route: status=%d body=%s", rec.Code, rec.Body.String())
	}
}
