---
memory_kind: episodic
episode_id: cursor-issue-424-2026-06-29-hands-on
title: MCP 2026-07-28 spec upgrade — hands-on delivery for kiwifs#424
tags: [mcp, spec-2026, issue-424, commit, pr]
date: 2026-06-29
---

## Goal

Deliver verified MCP 2026-07-28 upgrade for kiwifs/kiwifs#424 with commit, green tests, and PR.

## Actions

1. Verified uncommitted MCP transport layer in overlay workspace; tests green.
2. Cherry-picked onto `upstream/main` in `/tmp/kiwifs-commit` (overlay `.git` read-only).
3. Resolved merge conflicts: kept `NewStackBackend`, scoped search, branding; upgraded mcp-go v0.55.1.
4. Fixed duplicate `SetMCPHandler` in `internal/api/server.go`.
5. Added fix doc at `pages/fixes/kiwifs-kiwifs/issue-424-mcp-2026-spec.md`.

## Test results

```
go test ./internal/mcpserver/ -count=1  → PASS (9.2s)
go test ./tests/ -run MCP -count=1      → PASS (0.15s)
```

## Branch

`feat/mcp-2026-spec-424` on fork `advancedresearcharray/kiwifs`
