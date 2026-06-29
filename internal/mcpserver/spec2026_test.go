package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestServerDiscover(t *testing.T) {
	s := server.NewMCPServer("kiwifs", "1.0.0")
	s.AddTool(mcp.NewTool("ping"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("pong"), nil
	})

	h := StreamableHTTPHandler(s, "")
	body := `{"jsonrpc":"2.0","id":"d1","method":"server/discover","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get(HeaderMCPMethod); got != MethodServerDiscover {
		t.Fatalf("Mcp-Method = %q, want %q", got, MethodServerDiscover)
	}
	if rec.Header().Get(server.HeaderKeySessionID) != "" {
		t.Fatal("stateless transport must not emit Mcp-Session-Id")
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("result = %T", resp["result"])
	}
	if result["cacheScope"] != CacheScopePublic {
		t.Fatalf("cacheScope = %v", result["cacheScope"])
	}
	if _, ok := result["ttlMs"]; !ok {
		t.Fatal("expected ttlMs on discover result")
	}
	versions, ok := result["supportedVersions"].([]any)
	if !ok || len(versions) == 0 {
		t.Fatal("expected supportedVersions")
	}
	if versions[0] != ProtocolVersion20260728 {
		t.Fatalf("first supported version = %v, want %s", versions[0], ProtocolVersion20260728)
	}
}

func TestRoutingHeadersRequiredWhenProvided(t *testing.T) {
	s := server.NewMCPServer("kiwifs", "1.0.0")
	h := StreamableHTTPHandler(s, "")

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderMCPMethod, "tools/list")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for header/body mismatch", rec.Code)
	}
	if got := rec.Header().Get(HeaderMCPMethod); got != "initialize" {
		t.Fatalf("error response Mcp-Method = %q, want initialize", got)
	}
}

func TestRoutingHeadersEmittedOnResponse(t *testing.T) {
	s := server.NewMCPServer("kiwifs", "1.0.0")
	s.AddTool(mcp.NewTool("demo"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	})
	h := StreamableHTTPHandler(s, "")

	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	initReq := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(initBody)))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("Accept", "application/json, text/event-stream")
	initRec := httptest.NewRecorder()
	h.ServeHTTP(initRec, initReq)
	if initRec.Code != http.StatusOK {
		t.Fatalf("initialize status = %d", initRec.Code)
	}

	listBody := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	listReq := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(listBody)))
	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("Accept", "application/json, text/event-stream")
	listReq.Header.Set(HeaderMCPMethod, "tools/list")
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("tools/list status = %d, body=%s", listRec.Code, listRec.Body.String())
	}
	if got := listRec.Header().Get(HeaderMCPMethod); got != "tools/list" {
		t.Fatalf("response Mcp-Method = %q", got)
	}

	var resp map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	result := resp["result"].(map[string]any)
	if result["ttlMs"] != float64(defaultListCacheTTL/time.Millisecond) {
		t.Fatalf("ttlMs = %v", result["ttlMs"])
	}
	if result["cacheScope"] != CacheScopePublic {
		t.Fatalf("cacheScope = %v", result["cacheScope"])
	}

	tools := result["tools"].([]any)
	tool := tools[0].(map[string]any)
	schema := tool["inputSchema"].(map[string]any)
	if schema["$schema"] != jsonSchema2020URI {
		t.Fatalf("$schema = %v", schema["$schema"])
	}
}

func TestToolsListCacheHints(t *testing.T) {
	body := []byte(`{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"x","inputSchema":{"type":"object","properties":{}}}]}}`)
	out := augmentCacheableResult(body, "tools/list")
	if !strings.Contains(string(out), `"ttlMs":`) {
		t.Fatalf("missing ttlMs: %s", out)
	}
	out = augmentToolListSchemas(out)
	if !strings.Contains(string(out), jsonSchema2020URI) {
		t.Fatalf("missing 2020-12 schema URI: %s", out)
	}
}

func TestResourcesListCacheHints(t *testing.T) {
	body := []byte(`{"jsonrpc":"2.0","id":1,"result":{"resources":[{"uri":"page://x","name":"x"}]}}`)
	out := augmentCacheableResult(body, "resources/list")
	if !strings.Contains(string(out), `"ttlMs":`) {
		t.Fatalf("missing ttlMs: %s", out)
	}
	if !strings.Contains(string(out), `"cacheScope":"public"`) {
		t.Fatalf("missing cacheScope: %s", out)
	}
}

func TestToolsCallEmitsMcpNameHeader(t *testing.T) {
	s := server.NewMCPServer("kiwifs", "1.0.0")
	s.AddTool(mcp.NewTool("ping"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("pong"), nil
	})
	h := StreamableHTTPHandler(s, "")

	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2026-07-28","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	initReq := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(initBody)))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("Accept", "application/json, text/event-stream")
	initRec := httptest.NewRecorder()
	h.ServeHTTP(initRec, initReq)
	if initRec.Code != http.StatusOK {
		t.Fatalf("initialize status = %d", initRec.Code)
	}
	if got := initRec.Header().Get(HeaderMCPMethod); got != "initialize" {
		t.Fatalf("initialize Mcp-Method = %q", got)
	}

	callBody := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"ping","arguments":{}}}`
	callReq := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(callBody)))
	callReq.Header.Set("Content-Type", "application/json")
	callReq.Header.Set("Accept", "application/json, text/event-stream")
	callReq.Header.Set(HeaderMCPMethod, "tools/call")
	callReq.Header.Set(HeaderMCPName, "ping")
	callRec := httptest.NewRecorder()
	h.ServeHTTP(callRec, callReq)
	if callRec.Code != http.StatusOK {
		t.Fatalf("tools/call status = %d, body=%s", callRec.Code, callRec.Body.String())
	}
	if got := callRec.Header().Get(HeaderMCPMethod); got != "tools/call" {
		t.Fatalf("Mcp-Method = %q, want tools/call", got)
	}
	if got := callRec.Header().Get(HeaderMCPName); got != "ping" {
		t.Fatalf("Mcp-Name = %q, want ping", got)
	}
}

func TestValidateRegisteredToolSchemasRejectsExternalRef(t *testing.T) {
	s := server.NewMCPServer("kiwifs", "1.0.0")
	raw := []byte(`{"type":"object","properties":{"x":{"$ref":"https://example.com/bad.json"}}}`)
	s.AddTools(server.ServerTool{
		Tool: mcp.Tool{
			Name:           "bad",
			InputSchema:    mcp.ToolInputSchema{Type: "object"},
			RawInputSchema: raw,
		},
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		},
	})
	if err := validateRegisteredToolSchemas(s); err == nil {
		t.Fatal("expected external $ref rejection at registration time")
	}
}

func TestExternalSchemaRefRejected(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"x": map[string]any{"$ref": "https://example.com/schema.json"},
		},
	}
	if !hasExternalSchemaRef(schema) {
		t.Fatal("expected external ref detection")
	}
	local := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"x": map[string]any{"$ref": "#/$defs/foo"},
		},
	}
	if hasExternalSchemaRef(local) {
		t.Fatal("local $ref should be allowed")
	}
}
