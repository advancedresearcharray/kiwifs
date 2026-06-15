---
memory_kind: episodic
episode_id: cursor-issue-314-2026-06-15
title: "Issue #314 — mount MCP Streamable HTTP on serve"
tags: [kiwifs, issue-314, mcp, streamable-http, bug-fix]
date: 2026-06-15
---

Fixed kiwifs/kiwifs#314: Cloud MCP endpoint returned 405 on POST and HTML on GET because `kiwifs serve` never mounted Streamable HTTP MCP on the main Echo server — `/mcp` fell through to the UI `GET /*` catch-all.

Changes: `wireMCPHTTP` in `cmd/serve.go`, `SetMCPHandler` route ordering in `internal/api/server.go`, `NewStackBackend` + `StreamableHTTPHandler` in mcpserver.

Tests passed:
- `go test ./internal/api/ -run TestMCP -count=1` (4 tests)
- `go test ./internal/mcpserver/ -run 'TestNewStack|TestAuthToken' -count=1`

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-314-cloud-mcp-streamable-http.md`

Note: Kiwi MCP gateway unavailable in agent env; remote Kiwi write at CT934 returned `invalid API key`. Docs written to workspace `pages/` and `episodes/` trees directly.
