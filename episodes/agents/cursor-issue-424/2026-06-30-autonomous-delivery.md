---
memory_kind: episodic
episode_id: cursor-issue-424-autonomous-delivery
title: "MCP 2026-07-28 spec — autonomous delivery (#424)"
tags: [mcp, spec-2026, issue-424, transport, discover, cache, regression-tests]
date: 2026-06-30
---

## Work item

kiwifs/kiwifs#424 — feat(mcp): upgrade to MCP 2026-07-28 spec

## Actions

1. Searched Kiwi depot at `192.168.167.240:3333` — unreachable; read local fix doc `pages/fixes/kiwifs-kiwifs/issue-424-mcp-2026-spec.md`
2. Reproduced gap on upstream main: `StreamableHTTPHandler` was thin mcp-go wrapper with no `server/discover`, routing headers, or cache hints
3. Verified overlay implementation (9 new/modified Go files) and ran full MCP test suite — all green
4. Cloned `kiwifs/kiwifs` to `/tmp/kiwifs-git-work`, applied changes on branch `feat/mcp-2026-07-28-424`
5. Committed locally (fleet publishes PR); did not push per fleet policy

## Root cause

`github.com/mark3labs/mcp-go` v0.49.0 implements JSON-RPC over Streamable HTTP but not the 2026-07-28 transport layer (SEP-2575 routing headers, SEP-2243 discover, SEP-2549 cache hints). KiwiFS delegated entirely to `WithStateLess(true)` without response shaping.

## Test output

```
go test ./internal/mcpserver/ -count=1      → ok  9.425s
go test ./internal/api/ -run MCP -count=1   → ok  2.227s
go test ./tests/ -run MCP -count=1          → ok  0.160s
```

## Deliverables

- Local commit on `/tmp/kiwifs-git-work` branch `feat/mcp-2026-07-28-424` (11 files, +863 lines)
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-424-mcp-2026-spec.md`
- Fleet agent: push + open PR closing #424
