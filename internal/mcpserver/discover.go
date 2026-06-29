package mcpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

type discoverResult struct {
	ResultType        string            `json:"resultType"`
	SupportedVersions []string          `json:"supportedVersions"`
	Capabilities      map[string]any    `json:"capabilities"`
	ServerInfo        map[string]string `json:"serverInfo"`
	Instructions      string            `json:"instructions,omitempty"`
	TTLMs             int64             `json:"ttlMs"`
	CacheScope        string            `json:"cacheScope"`
}

func buildDiscoverResult(mcpSrv *server.MCPServer) discoverResult {
	caps := map[string]any{}
	if len(mcpSrv.ListTools()) > 0 {
		caps["tools"] = map[string]any{}
	}
	// KiwiFS always registers page resources; mcp-go does not expose ListResources/ListPrompts.
	caps["resources"] = map[string]any{}

	return discoverResult{
		ResultType:        "complete",
		SupportedVersions: append([]string(nil), supportedProtocolVersions...),
		Capabilities:      caps,
		ServerInfo: map[string]string{
			"name":    "kiwifs",
			"version": "1.0.0",
		},
		Instructions: "KiwiFS knowledge base MCP server. Use kiwi_search before writing; kiwi_read to inspect existing pages.",
		TTLMs:        int64(defaultDiscoverCacheTTL / time.Millisecond),
		CacheScope:   CacheScopePublic,
	}
}

func writeDiscoverResponse(w http.ResponseWriter, id json.RawMessage, mcpSrv *server.MCPServer) {
	result := buildDiscoverResult(mcpSrv)
	payload, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	})
	if err != nil {
		httpErrorJSON(w, id, -32603, "failed to encode discover response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(HeaderMCPMethod, MethodServerDiscover)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func httpErrorJSON(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	payload, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write(payload)
}
