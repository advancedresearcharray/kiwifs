---
memory_kind: episodic
episode_id: cursor-issue-424-2026-06-29
title: Issue #424 MCP 2026-07-28 spec upgrade delivery
tags: [kiwifs, issue-424, mcp, spec-2026, clawwork]
date: 2026-06-29
---

## Task

Implement kiwifs/kiwifs#424 — upgrade MCP server to 2026-07-28 spec (stateless transport, routing headers, discover, cache hints, JSON Schema 2020-12).

## Investigation

1. Searched Kiwi depot at `http://192.168.167.240:3333` — unreachable (connection refused).
2. Read local fix doc draft at `pages/fixes/kiwifs-kiwifs/issue-424-mcp-2026-spec.md` — no prior committed fix.
3. Found in-progress implementation in `internal/mcpserver/{protocol,spec2026,http_transport,discover,schema2020}.go` plus tracked diffs in `mcpserver.go`, `serve.go`, `server.go`.

## Root cause

Prior MCP stack used bare `mcp-go` `StreamableHTTPServer` without spec-2026 response shaping: no `server/discover`, no routing headers, no list cache hints, no external `$ref` guard.

## Implementation

- `StreamableHTTPHandler` wraps stateless transport with `spec2026Handler` + `spec2026ResponseWriter`
- `server/discover` returns capabilities, supported versions, cache hints
- Response headers: `Mcp-Method`, `Mcp-Name` (tools/call); strips `Mcp-Session-Id`
- `tools/list` responses augmented with `ttlMs`, `cacheScope`, JSON Schema 2020-12 `$schema`
- `validateRegisteredToolSchemas` rejects external HTTP(S) `$ref` at startup
- `wireMCPHTTP` + `SetMCPHandler` mount `/mcp` on colocated `kiwifs serve`

## Test results

```
GOCACHE=/tmp/gocache go test ./internal/mcpserver/... -run 'Discover|Routing|ToolsCall|ValidateRegistered|ToolsListCache|ExternalSchema' -count=1 -vet=off  → PASS
GOCACHE=/tmp/gocache go test ./internal/api/... -run 'MCP2026|MCPStreamable' -count=1 -vet=off  → PASS
GOCACHE=/tmp/gocache go build ./cmd/... ./internal/mcpserver/... ./internal/api/...  → OK
```

## Outcome

Local commit ready for fleet publish. Closes must-have acceptance criteria for #424. OAuth PKCE and MCP Tasks deferred per issue priority.
