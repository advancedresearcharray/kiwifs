---
memory_kind: episodic
episode_id: cursor-pr-384-hands-on-delivery-2026-06-18
title: "PR #384 — hands-on delivery: check integration + push"
tags: [kiwifs, pr-384, issue-338, pipeline, sequences, check, delivery]
date: 2026-06-18
---

## Context

Hands-on takeover for kiwifs/kiwifs#384. Prior agent integrated sequence checks into `cmd/check.go` (commit `38b7566`) but fleet publish failed: branch diverged from `fork/feat/pipeline-sequence-numbering-338`, no push.

## Work done

1. Verified `cmd/check.go` runs janitor hygiene + `pipeline.CheckSequences` when `[sequences]` is configured; removed duplicate `internal/checkcmd`.
2. Refactored `runCheck` → `runCheckWithCode` for testable exit codes (peer review).
3. Added `TestRunCheckWithCode_SequenceGapFails` and `TestRunCheckWithCode_CleanWhenSequencesDisabled`.
4. Full test suite green with `-race`.
5. Force-pushed rebased branch (5 feature commits on `origin/main`) to fork for PR #384.

## Tests

```bash
go test ./cmd/... ./internal/pipeline/... ./internal/config/... -race -count=1
go test ./internal/... ./cmd/... ./pkg/... ./tests/... -race -count=1
```

## Kiwi docs

- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-338-pipeline-sequence-numbering.md` (local, gitignored `kiwifs-*` pattern)
