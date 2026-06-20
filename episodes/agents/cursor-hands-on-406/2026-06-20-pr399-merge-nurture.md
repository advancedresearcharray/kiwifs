---
memory_kind: episodic
episode_id: cursor-hands-on-406-2026-06-20-pr399
title: "PR #399 merge-nurture — rebase onto main, PathPrefix fix verified"
tags: [kiwifs, pr-399, issue-103, mkdocs, exporter, merge-nurture, hands-on]
date: 2026-06-20
---

## Context

Idle queue merge-first work on kiwifs/kiwifs PR #399 (`feat/mkdocs-export-103`, closes #103). GitHub reported `mergeable: CONFLICTING`. Feature already on `origin/main` via PR #275; only PathPrefix boundary fix needed.

## Actions

1. Reset `pr-399` to `origin/main` and applied PathPrefix fix.
2. Removed Cursor attribution from commits per fleet policy.
3. Ran tests — all green.

## Outcome

Branch `pr-399` is 1 commit ahead of `origin/main` with a clean merge tree.

## Tests

```bash
go test ./internal/exporter/... -count=1 -v   # PASS
go test ./cmd/... -run 'MkDocs|Export' -count=1 -v   # PASS
```
