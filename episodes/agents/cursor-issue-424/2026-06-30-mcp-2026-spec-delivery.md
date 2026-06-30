---
memory_kind: episodic
episode_id: cursor-issue-424-mcp-2026-spec-delivery
title: "MCP 2026-07-28 spec transport upgrade (#424)"
tags: [mcp, spec-2026, issue-424, transport, discover, cache]
date: 2026-06-30
---

## Work item

kiwifs/kiwifs#424 — feat(mcp): upgrade to MCP 2026-07-28 spec

## Actions

1. Searched Kiwi depot at `192.168.167.240:3333` — unreachable (timeout); read local fix doc `pages/fixes/kiwifs-kiwifs/issue-424-mcp-2026-spec.md`
2. Found prior implementation commit `9920eaab` on GitHub (closed PR #429); workspace lacked spec2026 files
3. Restored implementation from commit: `discover.go`, `http_transport.go`, `protocol.go`, `spec2026.go`, `schema2020.go`, tests
4. Wired `validateRegisteredToolSchemas` in `mcpserver.New()`; moved `StreamableHTTPHandler` to `http_transport.go`
5. Upgraded `github.com/mark3labs/mcp-go` v0.49.0 → v0.55.1 (tests pass)
6. Wrote updated fix doc locally — Kiwi MCP gateway unreachable

## Test output

```
go test ./internal/mcpserver/ -count=1   → ok (9.7s)
go test ./internal/api/ -run MCP -count=1 → ok (2.2s)
go test ./tests/ -run MCP -count=1        → ok (0.2s)
```

## Deliverables

- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-424-mcp-2026-spec.md`
- Fleet agent to publish branch + PR (local `.git` empty/unusable)
