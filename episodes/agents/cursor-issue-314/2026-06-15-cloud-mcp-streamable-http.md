---
memory_kind: episodic
episode_id: cursor-issue-314-2026-06-15
title: "Issue #314 — mount MCP Streamable HTTP on serve"
tags: [kiwifs, issue-314, mcp, streamable-http, bug-fix]
date: 2026-06-15
---

Fixed kiwifs/kiwifs#314: Cloud MCP endpoint returned 405 on POST and HTML on GET because `kiwifs serve` never mounted Streamable HTTP MCP on the main Echo server — `/mcp` fell through to the UI `GET /*` catch-all.

Changes: `wireMCPHTTP` in `cmd/serve.go`, `SetMCPHandler` route ordering in `internal/api/server.go`, `NewStackBackend` + `StreamableHTTPHandler` in mcpserver.

Tests passed (hands-on takeover 2026-06-15):
- `go test ./internal/api/ -run TestMCP -count=1` — 4/4 PASS
- `go test ./internal/mcpserver/ -run 'TestNewStack|TestAuthToken' -count=1` — 2/2 PASS
- `go test ./tests/ -run MCP -count=1` — 2/2 PASS
- `go test ./internal/api/ ./internal/mcpserver/ ./cmd/ -count=1` — all PASS

Peer review (hands-on):
- `wireMCPHTTP` reuses live stack via `NewStackBackend`; `Close` no-op when `ownStack=false` — correct lifetime
- `SetMCPHandler` registers `echo.Any("/mcp")` idempotently; route wins over UI `GET /*` (verified by tests)
- `StreamableHTTPHandler` + `AuthTokenFromConfig` shared between `mcp --http` and colocated serve MCP
- No issues found; peer review passed

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-314-cloud-mcp-streamable-http.md`
Kiwi cluster: fix doc indexed at CT934 (`pages/fixes/kiwifs-kiwifs/issue-314-cloud-mcp-endpoint-rejects-streamable-ht.md`)
