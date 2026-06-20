---
memory_kind: episodic
episode_id: cursor-hands-on-406-pr400-delivery-v10
title: PR #400 hands-on delivery v10 — append_only committed, tested, peer-reviewed
tags: [kiwifs, pr-400, append_only, hands-on, delivery]
date: 2026-06-20
---

# PR #400 hands-on delivery v10

Work item: [kiwifs/kiwifs#400](https://github.com/kiwifs/kiwifs/pull/400) (closes #337).

## Actions

1. Verified implementation on branch `pr-400` at `074d656` (rebased onto `main`).
2. Ran full targeted and suite tests — all green.
3. Peer review (bugbot): merge-ready; no must-fix issues.
4. Committed this episodic delivery verification.
5. Pushed to `feat/append-only-337`.

## Test results

```
go test ./internal/pipeline/... -run AppendOnly -count=1     — PASS (7)
go test ./internal/api/... -run AppendOnly -count=1          — PASS (7)
go test ./internal/mcpserver/... -run AppendOnly -count=1    — PASS (1)
go test ./internal/pipeline/... -run 'AppendOnly|ValidateWrite' -count=1 — PASS
go test ./internal/... ./cmd/... -count=1                    — PASS
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

## Peer review notes

- Hardcoded guards under `writeMu` are TOCTOU-safe; all protocol writes funnel through pipeline.
- Bulk batch rejects on-disk append-only overwrites and duplicate-path overwrites within one batch.
- Rebased branch preserves main's `ValidateWrite(ctx, path, content, WriteKind)` API.
- Optional follow-ups: 409 mapping in workflow/publish handlers; WebDAV/S3 protocol-level smoke tests.
