package mcpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func setupMCPHTTPTest(t *testing.T) (*server.MCPServer, http.Handler) {
	t.Helper()
	tmp := t.TempDir()
	kiwiDir := filepath.Join(tmp, ".kiwi")
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
	if err := os.WriteFile(filepath.Join(tmp, "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mcpSrv, _, err := New(Options{Root: tmp})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return mcpSrv, StreamableHTTPHandler(mcpSrv, "")
}

func mcpPOST(t *testing.T, handler http.Handler, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestMCP2026_ServerDiscover(t *testing.T) {
	_, handler := setupMCPHTTPTest(t)

	body := `{"jsonrpc":"2.0","id":"disc-1","method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`
	rec := mcpPOST(t, handler, body, map[string]string{
		HeaderMcpMethod: "server/discover",
		HeaderMcpProto:  ProtocolVersion2026,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Result DiscoverResult `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.Result.ResultType != discoverResultType {
		t.Errorf("resultType = %q, want %q", resp.Result.ResultType, discoverResultType)
	}
	if len(resp.Result.SupportedVersions) == 0 {
		t.Error("expected supportedVersions")
	}
	if !strings.Contains(strings.Join(resp.Result.SupportedVersions, ","), ProtocolVersion2026) {
		t.Errorf("supportedVersions missing %s: %v", ProtocolVersion2026, resp.Result.SupportedVersions)
	}
	if resp.Result.ServerInfo.Name != "kiwifs" {
		t.Errorf("serverInfo.name = %q, want kiwifs", resp.Result.ServerInfo.Name)
	}
	if resp.Result.TTLMs != defaultDiscoverTTL {
		t.Errorf("ttlMs = %d, want %d", resp.Result.TTLMs, defaultDiscoverTTL)
	}
	if resp.Result.CacheScope != defaultCacheScope {
		t.Errorf("cacheScope = %q, want %q", resp.Result.CacheScope, defaultCacheScope)
	}
	if resp.Result.Capabilities.Tools == nil {
		t.Error("expected tools capability in discover result")
	}

	if got := rec.Header().Get(HeaderMcpMethod); got != "server/discover" {
		t.Errorf("response %s = %q, want server/discover", HeaderMcpMethod, got)
	}
}

func TestMCP2026_ToolsListJSONSchema202012(t *testing.T) {
	_, handler := setupMCPHTTPTest(t)

	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	if rec := mcpPOST(t, handler, initBody, nil); rec.Code != http.StatusOK {
		t.Fatalf("initialize status = %d", rec.Code)
	}

	listBody := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	listRec := mcpPOST(t, handler, listBody, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("tools/list status = %d", listRec.Code)
	}

	var listResp struct {
		Result struct {
			Tools []struct {
				InputSchema map[string]any `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(listResp.Result.Tools) == 0 {
		t.Fatal("expected at least one tool")
	}
	for i, tool := range listResp.Result.Tools {
		if tool.InputSchema == nil {
			continue
		}
		if got := tool.InputSchema["$schema"]; got != jsonSchema202012 {
			t.Errorf("tool[%d].inputSchema.$schema = %v, want %s", i, got, jsonSchema202012)
		}
	}
}

func TestMCP2026_ToolsListCachingMetadata(t *testing.T) {
	_, handler := setupMCPHTTPTest(t)

	// Legacy initialize handshake still works (backward compatibility).
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	initRec := mcpPOST(t, handler, initBody, nil)
	if initRec.Code != http.StatusOK {
		t.Fatalf("initialize status = %d", initRec.Code)
	}

	listBody := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	listRec := mcpPOST(t, handler, listBody, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("tools/list status = %d, body: %s", listRec.Code, listRec.Body.String())
	}

	var listResp struct {
		Result map[string]any `json:"result"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if listResp.Result["ttlMs"] != float64(defaultListTTLMs) {
		t.Errorf("ttlMs = %v, want %d", listResp.Result["ttlMs"], defaultListTTLMs)
	}
	if listResp.Result["cacheScope"] != defaultCacheScope {
		t.Errorf("cacheScope = %v, want %q", listResp.Result["cacheScope"], defaultCacheScope)
	}
	if got := listRec.Header().Get(HeaderMcpMethod); got != string(mcp.MethodToolsList) {
		t.Errorf("response %s = %q, want %q", HeaderMcpMethod, got, mcp.MethodToolsList)
	}
}

func TestMCP2026_RoutingHeaderMismatchRejected(t *testing.T) {
	_, handler := setupMCPHTTPTest(t)

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	rec := mcpPOST(t, handler, body, map[string]string{
		HeaderMcpMethod: "tools/call",
		HeaderMcpProto:  ProtocolVersion2026,
	})

	var errResp struct {
		Error jsonRPCError `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal: %v; body: %s", err, rec.Body.String())
	}
	if errResp.Error.Code != mcp.INVALID_REQUEST {
		t.Errorf("error code = %d, want %d", errResp.Error.Code, mcp.INVALID_REQUEST)
	}
	if !strings.Contains(errResp.Error.Message, HeaderMcpMethod) {
		t.Errorf("error message = %q, want header mismatch detail", errResp.Error.Message)
	}
}

func TestMCP2026_MissingRoutingHeaderRejected(t *testing.T) {
	_, handler := setupMCPHTTPTest(t)

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`
	rec := mcpPOST(t, handler, body, map[string]string{
		HeaderMcpProto: ProtocolVersion2026,
	})

	var errResp struct {
		Error jsonRPCError `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if errResp.Error.Code != mcp.INVALID_REQUEST {
		t.Errorf("error code = %d, want %d", errResp.Error.Code, mcp.INVALID_REQUEST)
	}
}

func TestMCP2026_LegacyInitializeWithoutHeaders(t *testing.T) {
	_, handler := setupMCPHTTPTest(t)

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"legacy","version":"1.0"}}}`
	rec := mcpPOST(t, handler, body, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", rec.Code, rec.Body.String())
	}
	resp := rec.Body.String()
	if !strings.Contains(resp, `"result"`) {
		t.Fatalf("expected result, got: %s", resp)
	}
	// Stateless mode — no session ID required.
	if rec.Header().Get("Mcp-Session-Id") != "" {
		t.Error("stateless transport should not return Mcp-Session-Id")
	}
}

func TestMCP2026_ToolsCallRequiresMcpName(t *testing.T) {
	_, handler := setupMCPHTTPTest(t)

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"kiwi_tree","arguments":{}}}`
	rec := mcpPOST(t, handler, body, map[string]string{
		HeaderMcpMethod: string(mcp.MethodToolsCall),
		HeaderMcpProto:  ProtocolVersion2026,
	})

	var errResp struct {
		Error jsonRPCError `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if errResp.Error.Code != mcp.INVALID_REQUEST {
		t.Errorf("error code = %d, want %d", errResp.Error.Code, mcp.INVALID_REQUEST)
	}
	if !strings.Contains(errResp.Error.Message, HeaderMcpName) {
		t.Errorf("error message = %q, want %s requirement", errResp.Error.Message, HeaderMcpName)
	}
}

func TestEnhanceListCachingPreservesExisting(t *testing.T) {
	in := []byte(`{"jsonrpc":"2.0","id":1,"result":{"tools":[],"ttlMs":999,"cacheScope":"private"}}`)
	out := enhanceListCaching(in, string(mcp.MethodToolsList))
	var msg map[string]any
	if err := json.Unmarshal(out, &msg); err != nil {
		t.Fatal(err)
	}
	result := msg["result"].(map[string]any)
	if result["ttlMs"] != float64(999) {
		t.Errorf("ttlMs overwritten: %v", result["ttlMs"])
	}
	if result["cacheScope"] != "private" {
		t.Errorf("cacheScope overwritten: %v", result["cacheScope"])
	}
}

func TestParseMCPMethodName(t *testing.T) {
	method, name, err := parseMCPMethodName([]byte(`{"method":"tools/call","params":{"name":"kiwi_read"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if method != "tools/call" || name != "kiwi_read" {
		t.Fatalf("got %q/%q", method, name)
	}

	method, name, err = parseMCPMethodName([]byte(`{"method":"resources/read","params":{"uri":"kiwi://pages/test"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if method != "resources/read" || name != "kiwi://pages/test" {
		t.Fatalf("got %q/%q", method, name)
	}
}

func TestBuildDiscoverResult(t *testing.T) {
	mcpSrv, _ := setupMCPHTTPTest(t)
	result, err := buildDiscoverResult(mcpSrv, t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if result.ServerInfo.Name != "kiwifs" {
		t.Errorf("name = %q", result.ServerInfo.Name)
	}
	if len(result.SupportedVersions) < 2 {
		t.Errorf("expected multiple supported versions, got %v", result.SupportedVersions)
	}
}

func TestMCP2026_StatelessNoSessionID(t *testing.T) {
	_, handler := setupMCPHTTPTest(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	rec := mcpPOST(t, handler, body, nil)
	_, _ = io.ReadAll(rec.Body)
	if rec.Header().Get("Mcp-Session-Id") != "" {
		t.Error("expected no Mcp-Session-Id in stateless mode")
	}
}
