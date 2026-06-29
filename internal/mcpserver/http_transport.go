package mcpserver

import (
	"bytes"
	"io"
	"net/http"

	"github.com/kiwifs/kiwifs/internal/bootstrap"
	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/mark3labs/mcp-go/server"
)

// AuthTokenFromConfig returns the bearer token for MCP HTTP auth when apikey auth is enabled.
func AuthTokenFromConfig(cfg *config.Config) string {
	if cfg == nil || cfg.Auth.Type != "apikey" || cfg.Auth.APIKey == "" {
		return ""
	}
	return cfg.Auth.APIKey
}

// StreamableHTTPHandler serves MCP over Streamable HTTP with MCP 2026-07-28 routing headers,
// server/discover, cache hints, and stateless transport (no Mcp-Session-Id dependency).
func StreamableHTTPHandler(mcpSrv *server.MCPServer, authToken string) http.Handler {
	inner := server.NewStreamableHTTPServer(
		mcpSrv,
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
		httpErrorJSON(w, nil, -32700, "read request body failed")
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	env, err := parseRPCRequest(body)
	if err != nil {
		httpErrorJSON(w, nil, -32700, "invalid JSON-RPC request")
		return
	}
	if env.Method == MethodServerDiscover {
		writeDiscoverResponse(w, env.ID, h.mcpSrv)
		return
	}

	name := rpcResourceName(env.Method, env.Params)
	if err := validateRoutingHeaders(r, env.Method, name); err != nil {
		httpErrorJSON(w, env.ID, -32600, err.Error())
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

// stackBackend reuses a live bootstrap stack without closing it on MCP shutdown.
type stackBackend struct {
	*LocalBackend
}

// NewStackBackend reuses a live bootstrap stack (e.g. from kiwifs serve) without owning its lifetime.
func NewStackBackend(stack *bootstrap.Stack) Backend {
	return &stackBackend{LocalBackend: newLocalBackendFromStack(stack)}
}

func newLocalBackendFromStack(stack *bootstrap.Stack) *LocalBackend {
	b := &LocalBackend{root: stack.Root, stack: stack}
	b.once.Do(func() {}) // skip bootstrap rebuild; stack is injected
	return b
}

func (b *stackBackend) Close() error { return nil }

var _ Backend = (*stackBackend)(nil)
