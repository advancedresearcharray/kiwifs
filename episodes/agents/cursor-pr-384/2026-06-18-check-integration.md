---
memory_kind: episodic
episode_id: cursor-pr-384-check-integration-2026-06-18
title: "PR #384 — integrate sequence checks into existing kiwifs check"
tags: [kiwifs, pr-384, issue-338, pipeline, sequences, check, ci]
date: 2026-06-18
---

## Context

Merge-first work on kiwifs/kiwifs#384. Prior commits added `internal/checkcmd` as a separate `kiwifs check` implementation, but main already ships `cmd/check.go` (#263) for janitor hygiene scans. The feat commit deleted `cmd/check.go`, regressing hygiene CI checks.

## Fix

- Restored `cmd/check.go` with sequence checking layered on the existing janitor scan (`runKnowledgeScan` shared with `cmd/janitor.go`).
- Removed `internal/checkcmd/` — one `kiwifs check` command covers hygiene + sequence integrity.
- Dropped `checkcmd.Command` registration from `cmd/root.go` (check registers via its own `init()`).
- JSON output wraps both scans: `{ "janitor": ..., "sequences": ... }`.

## Tests

```bash
go test ./cmd/... ./internal/pipeline/... ./internal/config/... -race -count=1
go test ./internal/... ./cmd/... ./pkg/... ./tests/... -race -count=1
# all PASS
```

## Fleet publish

Local commit only (no push). Ready for fleet to publish to PR #384.
