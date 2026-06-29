---
memory_kind: episodic
episode_id: cursor-hands-on-424-2026-06-29
title: Issue #424 MCP 2026 spec — hands-on delivery
tags: [kiwifs, issue-424, mcp, spec-2026, clawwork]
date: 2026-06-29
---

## Summary

Delivered kiwifs/kiwifs#424 on clean branch `feat/mcp-2026-spec-424-clean` rebased from upstream main. Prior fork branch was massively diverged (515 files); extracted only the 2026 transport layer on top of existing #314 colocated `/mcp` wiring.

## Root cause of delivery failure

Overlay workspace `.git` is read-only (permission denied on index.lock). Fleet agent had uncommitted overlay changes but no commit. Old fork branch included unrelated deletions from stale main.

## Changes

- Added `protocol.go`, `spec2026.go`, `http_transport.go`, `discover.go`, `schema2020.go`
- Added `spec2026_test.go`, `serve_integration_test.go`
- Updated `mcpserver.go`: `validateRegisteredToolSchemas` at startup; delegate to 2026 `StreamableHTTPHandler`
- Updated `handlers_mcp_test.go`: `TestMCP2026ServerDiscover`, `TestMCP2026ToolsListCacheHints`

## Tests

```
GOCACHE=/tmp/gocache424 go test ./internal/mcpserver/... ./internal/api/... \
  -run 'MCP2026|MCPStreamable|SetMCP|Discover|Routing|ToolsCall|ValidateRegistered|ToolsListCache|ResourcesListCache|ExternalSchema|StackBackend|AuthToken|ServerDiscover' \
  -count=1 -vet=off → PASS (mcpserver 0.45s, api 2.3s)
go build ./cmd/... → PASS
```

## Acceptance criteria met

- [x] `Mcp-Method` / `Mcp-Name` on responses
- [x] `server/discover` implemented
- [x] `tools/list` + `resources/list` include `ttlMs` and `cacheScope`
- [x] JSON Schema 2020-12 with external `$ref` rejection
- [x] Stateless — no `Mcp-Session-Id` dependency
- [x] Legacy `2024-11-05` clients still work
- [x] Integration tests
