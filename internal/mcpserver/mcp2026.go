package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ProtocolVersion2026 is the MCP 2026-07-28 specification version.
const ProtocolVersion2026 = "2026-07-28"

// HTTP routing headers required by SEP-2243 for Streamable HTTP transport.
const (
	HeaderMcpMethod = "Mcp-Method"
	HeaderMcpName   = "Mcp-Name"
	HeaderMcpProto  = "Mcp-Protocol-Version"

	metaProtocolVersion = "io.modelcontextprotocol/protocolVersion"
)

// Supported protocol versions advertised via server/discover.
var supportedProtocolVersions = []string{
	ProtocolVersion2026,
	mcp.LATEST_PROTOCOL_VERSION,
	"2025-06-18",
	"2025-03-26",
	"2024-11-05",
}

const (
	defaultListTTLMs    = 300000  // 5 minutes — tools/resources list
	defaultDiscoverTTL  = 3600000 // 1 hour — server/discover
	defaultCacheScope   = "public"
	discoverResultType  = "complete"
)

// DiscoverResult is the server/discover response shape (MCP 2026-07-28).
type DiscoverResult struct {
	mcp.Result
	ResultType        string                 `json:"resultType"`
	SupportedVersions []string               `json:"supportedVersions"`
	Capabilities      mcp.ServerCapabilities `json:"capabilities"`
	ServerInfo        mcp.Implementation     `json:"serverInfo"`
	Instructions      string                 `json:"instructions,omitempty"`
	TTLMs             int64                  `json:"ttlMs"`
	CacheScope        string                 `json:"cacheScope"`
}

type jsonRPCEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Result  json.RawMessage `json:"result"`
	Error   *jsonRPCError   `json:"error"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type discoverParams struct {
	Meta *mcp.Meta `json:"_meta,omitempty"`
}

// mcp2026Transport wraps Streamable HTTP to add 2026-07-28 spec features while
// preserving backward compatibility with legacy initialize-based clients.
type mcp2026Transport struct {
	inner  http.Handler
	mcpSrv *server.MCPServer
}

func wrapMCP2026(inner http.Handler, mcpSrv *server.MCPServer) http.Handler {
	return &mcp2026Transport{inner: inner, mcpSrv: mcpSrv}
}

func (t *mcp2026Transport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		t.inner.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	method, name, parseErr := parseMCPMethodName(body)
	if parseErr != nil {
		t.inner.ServeHTTP(w, r)
		return
	}

	proto := clientProtocolVersion(r, body)
	if proto == ProtocolVersion2026 {
		if err := validateRoutingHeaders(r, method, name); err != nil {
			writeJSONRPCError(w, nil, mcp.INVALID_REQUEST, err.Error())
			return
		}
	}

	if method == "server/discover" {
		t.handleDiscover(w, r, body)
		return
	}

	buf := &bufferingResponseWriter{ResponseWriter: w, header: make(http.Header)}
	r.Body = io.NopCloser(bytes.NewReader(body))
	t.inner.ServeHTTP(buf, r)

	if buf.statusCode == 0 {
		buf.statusCode = http.StatusOK
	}

	respBody := buf.body
	if len(respBody) > 0 && !strings.Contains(buf.header.Get("Content-Type"), "text/event-stream") {
		respBody = enhanceListCaching(respBody, method)
	}

	for k, vals := range buf.header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	emitRoutingHeaders(w, method, name)
	w.WriteHeader(buf.statusCode)
	if len(respBody) > 0 {
		_, _ = w.Write(respBody)
	}
}

func (t *mcp2026Transport) handleDiscover(w http.ResponseWriter, r *http.Request, body []byte) {
	var envelope jsonRPCEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		writeJSONRPCError(w, nil, mcp.PARSE_ERROR, "failed to parse discover request")
		return
	}

	result, err := buildDiscoverResult(t.mcpSrv, r.Context())
	if err != nil {
		writeJSONRPCError(w, envelope.ID, mcp.INTERNAL_ERROR, err.Error())
		return
	}

	resp := map[string]any{
		"jsonrpc": mcp.JSONRPC_VERSION,
		"id":      envelope.ID,
		"result":  result,
	}

	w.Header().Set("Content-Type", "application/json")
	emitRoutingHeaders(w, "server/discover", "")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func buildDiscoverResult(mcpSrv *server.MCPServer, ctx context.Context) (*DiscoverResult, error) {
	initMsg, err := json.Marshal(map[string]any{
		"jsonrpc": mcp.JSONRPC_VERSION,
		"id":      "discover-init",
		"method":  mcp.MethodInitialize,
		"params": map[string]any{
			"protocolVersion": ProtocolVersion2026,
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "kiwifs-discover",
				"version": "1.0.0",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal initialize for discover: %w", err)
	}

	sessionID := mcpSrv.GenerateInProcessSessionID()
	session := server.NewInProcessSession(sessionID, nil)
	ctx = mcpSrv.WithContext(ctx, session)

	raw := mcpSrv.HandleMessage(ctx, initMsg)
	if raw == nil {
		return nil, fmt.Errorf("initialize returned no response for discover")
	}

	respBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal initialize response: %w", err)
	}

	var initResp struct {
		Result mcp.InitializeResult `json:"result"`
		Error  *jsonRPCError        `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &initResp); err != nil {
		return nil, fmt.Errorf("unmarshal initialize response: %w", err)
	}
	if initResp.Error != nil {
		return nil, fmt.Errorf("initialize error: %s", initResp.Error.Message)
	}

	return &DiscoverResult{
		ResultType:        discoverResultType,
		SupportedVersions: supportedProtocolVersions,
		Capabilities:      initResp.Result.Capabilities,
		ServerInfo:        initResp.Result.ServerInfo,
		Instructions:      initResp.Result.Instructions,
		TTLMs:             defaultDiscoverTTL,
		CacheScope:        defaultCacheScope,
	}, nil
}

func parseMCPMethodName(body []byte) (method, name string, err error) {
	var envelope jsonRPCEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", "", err
	}
	method = envelope.Method
	if envelope.Params == nil {
		return method, "", nil
	}

	var params map[string]any
	if err := json.Unmarshal(envelope.Params, &params); err != nil {
		return method, "", nil
	}
	if n, ok := params["name"].(string); ok {
		name = n
	} else if uri, ok := params["uri"].(string); ok {
		name = uri
	}
	return method, name, nil
}

func clientProtocolVersion(r *http.Request, body []byte) string {
	if v := r.Header.Get(HeaderMcpProto); v != "" {
		return v
	}

	var envelope jsonRPCEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.Params == nil {
		return ""
	}

	var params discoverParams
	if err := json.Unmarshal(envelope.Params, &params); err != nil || params.Meta == nil {
		return ""
	}
	if params.Meta.AdditionalFields == nil {
		return ""
	}
	if v, ok := params.Meta.AdditionalFields[metaProtocolVersion]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func validateRoutingHeaders(r *http.Request, method, name string) error {
	headerMethod := r.Header.Get(HeaderMcpMethod)
	if headerMethod == "" {
		return fmt.Errorf("missing %s header", HeaderMcpMethod)
	}
	if headerMethod != method {
		return fmt.Errorf("%s header %q does not match body method %q", HeaderMcpMethod, headerMethod, method)
	}

	if requiresMcpName(method) {
		headerName := r.Header.Get(HeaderMcpName)
		if headerName == "" {
			return fmt.Errorf("missing %s header for %s", HeaderMcpName, method)
		}
		if headerName != name {
			return fmt.Errorf("%s header %q does not match body name %q", HeaderMcpName, headerName, name)
		}
	}
	return nil
}

func requiresMcpName(method string) bool {
	switch method {
	case string(mcp.MethodToolsCall), string(mcp.MethodResourcesRead), string(mcp.MethodPromptsGet):
		return true
	default:
		return false
	}
}

func emitRoutingHeaders(w http.ResponseWriter, method, name string) {
	if method != "" {
		w.Header().Set(HeaderMcpMethod, method)
	}
	if name != "" {
		w.Header().Set(HeaderMcpName, name)
	}
}

func enhanceListCaching(body []byte, method string) []byte {
	switch method {
	case string(mcp.MethodToolsList), string(mcp.MethodResourcesList), string(mcp.MethodResourcesTemplatesList), string(mcp.MethodPromptsList):
		return addCachingFields(body, defaultListTTLMs)
	case string(mcp.MethodResourcesRead):
		return addCachingFields(body, defaultListTTLMs)
	default:
		return body
	}
}

func addCachingFields(body []byte, ttlMs int64) []byte {
	var msg map[string]any
	if err := json.Unmarshal(body, &msg); err != nil {
		return body
	}
	result, ok := msg["result"].(map[string]any)
	if !ok {
		return body
	}
	if _, exists := result["ttlMs"]; !exists {
		result["ttlMs"] = ttlMs
	}
	if _, exists := result["cacheScope"]; !exists {
		result["cacheScope"] = defaultCacheScope
	}
	out, err := json.Marshal(msg)
	if err != nil {
		return body
	}
	// Preserve trailing newline from json.Encoder if present.
	if strings.HasSuffix(string(body), "\n") {
		return append(out, '\n')
	}
	return out
}

func writeJSONRPCError(w http.ResponseWriter, id any, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": mcp.JSONRPC_VERSION,
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

type bufferingResponseWriter struct {
	http.ResponseWriter
	header     http.Header
	statusCode int
	body       []byte
}

func (w *bufferingResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *bufferingResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}
