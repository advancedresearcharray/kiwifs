---
memory_kind: episodic
episode_id: cursor-fleet-2026-06-30-issue-422
title: "Deliver kiwi_recall fusion retrieval (#422)"
tags: [kiwifs, recall, rrf, issue-422, mcp, search]
date: 2026-06-30
---

## Task

Implement `kiwi_recall` fusion retrieval for kiwifs/kiwifs#422 — RRF merge of FTS, vector, and graph (backlink) signals with REST and MCP exposure.

## Before

- `kiwi_search` via depot API: cluster `http://192.168.167.240:3333` unreachable (connection refused).
- Local fix doc existed only for go-vet self-assignment sub-fix (`pages/fixes/kiwifs-kiwifs/issue-422-recall-go-vet-self-assignment.md`).
- Feature branch `fork/feat/kiwi-recall-fusion-422` had core implementation; `fork/feat/kiwi-recall-422` had MCP tests and empty-query test.

## Work done

1. Checked out `feat/kiwi-recall-422` from `fork/feat/kiwi-recall-fusion-422`.
2. Cherry-picked MCP integration tests (`eb64a82`) and scope/empty-query hardening (`5cefaaf`) from `fork/feat/kiwi-recall-422`.
3. Verified all recall tests pass across `internal/search`, `internal/api`, `internal/mcpserver`.
4. Wrote durable fix doc at `pages/fixes/kiwifs-kiwifs/issue-422-kiwi-recall-fusion.md`.

## Test results

```
go test ./internal/search/... -run 'Recall|FuseRRF|Recaller' -count=1  → ok
go test ./internal/api/... -run Recall -count=1                        → ok
go test ./internal/mcpserver/... -run Recall -count=1                  → ok
```

## Outcome

Branch `feat/kiwi-recall-422` ready for fleet publish (push + PR closing #422). No push performed per fleet policy.
