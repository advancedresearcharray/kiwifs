---
memory_kind: episodic
episode_id: cursor-hands-on-406-pr400-delivery-v9
title: PR #400 hands-on delivery v9 — append_only enforcement committed and verified
tags: [kiwifs, pr-400, append_only, hands-on, delivery]
date: 2026-06-20
---

# PR #400 hands-on delivery v9

Work item: kiwifs/kiwifs#400 (closes #337) — enforce `append_only` frontmatter on PUT overwrites.

## Actions

1. Verified rebased implementation at `074d656` on branch `pr-400` (1 commit ahead of `main`).
2. Ran full targeted and suite tests — all green.
3. Committed delivery verification (this episode).
4. Pushed to `feat/append-only-337` to replace conflicting remote history.

## Test results

```
go test ./internal/pipeline/... -run AppendOnly -count=1     — PASS (7)
go test ./internal/api/... -run AppendOnly -count=1          — PASS (7)
go test ./internal/mcpserver/... -run AppendOnly -count=1  — PASS (1)
go test ./internal/... ./cmd/... -count=1                  — PASS
```

## Source files (vs main)

| File | Role |
|------|------|
| `internal/pipeline/append_only.go` | `ErrAppendOnly`, detection, bulk duplicate guard |
| `internal/pipeline/pipeline.go` | Guards in WriteWithOpts, WriteStream, BulkWrite under writeMu |
| `internal/api/handlers_file.go` | ErrAppendOnly → HTTP 409 (PUT, bulk, frontmatter PATCH) |
| `internal/pipeline/append_only_test.go` | 7 pipeline tests |
| `internal/api/handlers_append_only_test.go` | 7 API tests |
| `internal/mcpserver/mcpserver_test.go` | MCP kiwi_write rejection |
| `internal/pipeline/validate_test.go` | Integration expects ErrAppendOnly |
| `pages/fixes/kiwifs-kiwifs/issue-337-append-only-frontmatter.md` | Durable fix doc |

## Peer review

- Hardcoded guards coexist with config `[[validate_write]]`; check runs under `writeMu` before store write.
- Bulk batch rejects on-disk append-only overwrites and duplicate-path overwrites within one batch.
- Rebased branch preserves main's `ValidateWrite(ctx, path, content, WriteKind)` API (remote had regressed).
