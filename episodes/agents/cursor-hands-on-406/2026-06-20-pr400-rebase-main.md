---
memory_kind: episodic
episode_id: cursor-hands-on-406-2026-06-20-pr400-rebase-main
title: PR #400 rebase onto main — append_only without ValidateWrite regression
tags: [kiwifs, pr-400, append_only, rebase, merge-nurture]
date: 2026-06-20
---

# PR #400 rebase — append_only enforcement

**Target:** [kiwifs/kiwifs#400](https://github.com/kiwifs/kiwifs/pull/400) (closes #337)

## Problem found

PR branch had diverged from `main` and regressed:

- Removed `WriteKind`, `ErrWriteRejected`, `validate.go`, and `[[validate_write]]` config wiring
- GitHub reported `mergeable: CONFLICTING`
- Overlay git index write failures blocked normal checkout

## Work performed

1. Checked out pr-400 tree via alternate index (`GIT_INDEX_FILE=.git/index.pr400`).
2. Restored main versions of `pipeline.go`, `validate.go`, `config.go`, `bootstrap.go`, `handlers_file.go`.
3. Re-applied append_only hooks on top of main's ValidateWrite API:
   - `rejectAppendOnlyOverwrite` in WriteWithOpts and WriteStream (under writeMu)
   - `rejectAppendOnlyBulkOverwrite` in BulkWrite (under writeMu)
4. Added `ErrAppendOnly` → 409 in all four API write error paths.
5. Updated `TestPipelineValidateWriteRulesIntegration` to expect `ErrAppendOnly` (hardcoded guard runs before config rules).
6. Kept workspace-compatible `search.NewSQLite` in bootstrap (typed-link search not in overlay base).

## Test results

```
go test ./internal/pipeline/... -run AppendOnly     — PASS (7)
go test ./internal/api/... -run AppendOnly           — PASS (7)
go test ./internal/mcpserver/... -run AppendOnly     — PASS (1)
go test ./internal/pipeline/... -run ValidateWrite   — PASS
go test ./internal/...                               — PASS
```

## Deliverables

- Durable fix doc: `pages/fixes/kiwifs-kiwifs/issue-337-append-only-frontmatter.md`
- Local commit only (fleet publishes)
