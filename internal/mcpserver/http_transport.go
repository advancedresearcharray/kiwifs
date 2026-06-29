package mcpserver

import (
	"bytes"
	"io"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
)

// StreamableHTTPHandler serves MCP over Streamable HTTP with MCP 2026-07-28 routing headers,
// server/discover, cache hints, and stateless transport (no Mcp-Session-Id dependency).
func StreamableHTTPHandler(mcpSrv *server.MCPServer, authToken string) http.Handler {
	inner := server.NewStreamableHTTPServer(
		mcpSrv,
		server.WithEndpointPath("/mcp"),
		server.WithStateLess(true),
	)
	return bearerAuth(authToken, spec2026Handler{mcpSrv: mcpSrv, inner: inner})
}

type spec2026Handler struct {
	mcpSrv *server.MCPServer
	inner  http.Handler
}

func (h spec2026Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.inner.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpErrorJSON(w, nil, -32700, "read request body failed", "", "")
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	env, err := parseRPCRequest(body)
	if err != nil {
		httpErrorJSON(w, nil, -32700, "invalid JSON-RPC request", "", "")
		return
	}
	if env.Method == MethodServerDiscover {
		writeDiscoverResponse(w, env.ID, h.mcpSrv)
		return
	}

	name := rpcResourceName(env.Method, env.Params)
	if err := validateRoutingHeaders(r, env.Method, name); err != nil {
		httpErrorJSON(w, env.ID, -32600, err.Error(), env.Method, name)
		return
	}

	capture := &spec2026ResponseWriter{
		ResponseWriter: w,
		method:         env.Method,
		name:           name,
	}
	h.inner.ServeHTTP(capture, r)
	capture.flush()

	// Stateless: strip session id if upstream emitted one (2026 spec).
	w.Header().Del(server.HeaderKeySessionID)
}
