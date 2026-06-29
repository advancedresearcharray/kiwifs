---
memory_kind: episodic
episode_id: cursor-hands-on-424-2026-06-29-delivery
title: MCP 2026 spec upgrade verified delivery (issue 424)
tags: mcp, issue-424, spec-2026, delivery, hands-on
date: 2026-06-29
---

## Summary

Hands-on takeover for kiwifs/kiwifs#424. Overlay workspace `.git` is read-only; committed from writable clone at `/tmp/kiwifs-mcp424` on branch `feat/mcp-2026-spec-424`.

## Actions

1. Synced overlay implementation: `stack_backend.go`, `serve_integration_test.go`, idempotent `SetMCPHandler`, transport cleanup.
2. Ran MCP 2026 regression tests on mcpserver + api packages — all pass.
3. `go build ./cmd/...` — OK.
4. Committed and pushed to fork; opened upstream PR closing #424.
5. Updated fix doc at `pages/fixes/kiwifs-kiwifs/issue-424-mcp-2026-spec.md`.

## Test output

```text
go test ./internal/mcpserver/... ./internal/api/... \
  -run 'MCP2026|MCPStreamable|SetMCP|Discover|Routing|ToolsCall|ValidateRegistered|ToolsListCache|ResourcesListCache|ExternalSchema|StackBackend|AuthToken|ServerDiscover' \
  -count=1 -vet=off
# ok  github.com/kiwifs/kiwifs/internal/mcpserver  0.411s
# ok  github.com/kiwifs/kiwifs/internal/api         2.255s

go build ./cmd/...
# success
```
