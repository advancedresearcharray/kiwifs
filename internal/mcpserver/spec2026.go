package mcpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type rpcEnvelope struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type rpcParams struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

func parseRPCRequest(body []byte) (rpcEnvelope, error) {
	var env rpcEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return rpcEnvelope{}, err
	}
	return env, nil
}

func rpcResourceName(method string, params json.RawMessage) string {
	switch method {
	case "tools/call":
		var p rpcParams
		if json.Unmarshal(params, &p) == nil {
			return p.Name
		}
	case "resources/read":
		var p rpcParams
		if json.Unmarshal(params, &p) == nil {
			return p.URI
		}
	case "prompts/get":
		var p rpcParams
		if json.Unmarshal(params, &p) == nil {
			return p.Name
		}
	}
	return ""
}

func validateRoutingHeaders(r *http.Request, method, name string) error {
	headerMethod := r.Header.Get(HeaderMCPMethod)
	if headerMethod == "" {
		return nil
	}
	if headerMethod != method {
		return fmt.Errorf("header %s=%q does not match body method %q", HeaderMCPMethod, headerMethod, method)
	}
	headerName := r.Header.Get(HeaderMCPName)
	if headerName == "" {
		return nil
	}
	if name == "" {
		return nil
	}
	if headerName != name {
		return fmt.Errorf("header %s=%q does not match body name %q", HeaderMCPName, headerName, name)
	}
	return nil
}

func setRoutingHeaders(w http.ResponseWriter, method, name string) {
	if method != "" {
		w.Header().Set(HeaderMCPMethod, method)
	}
	if name != "" {
		w.Header().Set(HeaderMCPName, name)
	}
}

func augmentCacheableResult(body []byte, method string) []byte {
	switch method {
	case "tools/list", "resources/list", "resources/templates/list", "prompts/list":
		return injectCacheHints(body, int64(defaultListCacheTTL/time.Millisecond), CacheScopePublic)
	case "resources/read":
		return injectCacheHints(body, int64(defaultListCacheTTL/time.Millisecond), CacheScopePublic)
	default:
		return body
	}
}

func injectCacheHints(body []byte, ttlMs int64, scope string) []byte {
	var msg map[string]any
	if err := json.Unmarshal(body, &msg); err != nil {
		return body
	}
	result, ok := msg["result"].(map[string]any)
	if !ok || result == nil {
		return body
	}
	result["ttlMs"] = ttlMs
	result["cacheScope"] = scope
	out, err := json.Marshal(msg)
	if err != nil {
		return body
	}
	return out
}

func augmentToolListSchemas(body []byte) []byte {
	var msg map[string]any
	if err := json.Unmarshal(body, &msg); err != nil {
		return body
	}
	result, ok := msg["result"].(map[string]any)
	if !ok {
		return body
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		return body
	}
	for _, item := range tools {
		tool, ok := item.(map[string]any)
		if !ok {
			continue
		}
		normalizeSchemaField(tool, "inputSchema")
		normalizeSchemaField(tool, "outputSchema")
	}
	out, err := json.Marshal(msg)
	if err != nil {
		return body
	}
	return out
}

func normalizeSchemaField(tool map[string]any, field string) {
	schema, ok := tool[field].(map[string]any)
	if !ok || schema == nil {
		return
	}
	if _, has := schema["$schema"]; !has {
		schema["$schema"] = jsonSchema2020URI
	}
}

func hasExternalSchemaRef(v any) bool {
	switch t := v.(type) {
	case map[string]any:
		if ref, ok := t["$ref"].(string); ok && isExternalRef(ref) {
			return true
		}
		for _, child := range t {
			if hasExternalSchemaRef(child) {
				return true
			}
		}
	case []any:
		for _, child := range t {
			if hasExternalSchemaRef(child) {
				return true
			}
		}
	}
	return false
}

func isExternalRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return false
	}
	return strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://")
}

type spec2026ResponseWriter struct {
	http.ResponseWriter
	method string
	name   string
	buf    bytes.Buffer
	status int
}

func (w *spec2026ResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *spec2026ResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *spec2026ResponseWriter) flush() {
	setRoutingHeaders(w.ResponseWriter, w.method, w.name)
	if w.status > 0 {
		w.ResponseWriter.WriteHeader(w.status)
	} else if w.buf.Len() > 0 {
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	body := w.buf.Bytes()
	if len(body) > 0 && strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		body = augmentCacheableResult(body, w.method)
		if w.method == "tools/list" {
			body = augmentToolListSchemas(body)
		}
	}
	if len(body) > 0 {
		_, _ = w.ResponseWriter.Write(body)
	}
}
