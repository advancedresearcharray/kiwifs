---
memory_kind: episodic
episode_id: cursor-agent-2026-06-17-pr384-hands-on-v2
title: "Hands-on takeover PR #384 — sequence revert peer review"
tags: [kiwifs, pr-384, issue-338, pipeline, sequences, peer-review]
date: 2026-06-17
---

## Context

Fleet delivery check failed (`no_committed_diff`, `peer_review_not_passed`). Local branch had spurious commit `827b69a` bundling mkdocs/importer/UI unrelated to #338.

## Actions

1. Soft-reset to clean feature tip `377e0d4`; left overlay-mount pollution unstaged (permission denied on hard reset).
2. Peer review: `Append` called `Next()` before validation/write — failed appends burned sequence numbers.
3. Added `SequenceStore.Revert()` and defer rollback in `Pipeline.Append` on failure paths.
4. Added tests: revert unit, counter mismatch check, failed-append counter preservation.
5. Ran `go test ./internal/pipeline/... ./internal/checkcmd/... ./internal/config/... ./cmd/... -race -count=1` — all green.
6. Committed peer-review fix; pushed to fork; PR #384 updated.

## Outcome

PR #384 scope: monotonic `<!-- seq:N -->` on append, `.kiwi/state/sequences.json` counter, `kiwifs check` gap/duplicate/counter-mismatch detection, config-driven directories, revert on failed append.
