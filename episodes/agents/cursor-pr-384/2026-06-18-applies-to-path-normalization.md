---
memory_kind: episodic
episode_id: cursor-pr-384-applies-to-path-normalization-2026-06-18
title: "PR #384 — peer review fix: AppliesTo path normalization"
tags: [kiwifs, pr-384, issue-338, pipeline, sequences, peer-review, delivery]
date: 2026-06-18
---

## Context

Hands-on takeover delivery for kiwifs/kiwifs#384. Prior agent committed only episodic markdown (`c091861`); delivery check failed (`no_committed_diff`, `peer_review_not_passed`).

## Peer review finding

`SequenceStore.AppliesTo` compared raw paths against configured directory prefixes. API append uses paths like `/events/log.md` (leading slash per OpenAPI), while storage normalizes via `path.Clean` before read/write. Result: appends succeeded but **no** `<!-- seq:N -->` marker was injected.

## Fix

- Added `normalizeSequencePath` in `internal/pipeline/sequences.go` (mirrors storage path cleaning).
- Extended `TestSequenceAppliesTo` with leading-slash, `./` prefix, and prefix-sibling cases.
- Added `TestSequenceAppendWithLeadingSlashPath` integration test.

## Tests

```bash
go test ./cmd/... ./internal/pipeline/... ./internal/config/... ./internal/bootstrap/... -count=1 -race
```

All packages **PASS**.

## Commit

Source commit on `feat/pipeline-sequence-numbering-338` with Go changes only (this run).
