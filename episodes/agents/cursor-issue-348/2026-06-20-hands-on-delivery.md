---
memory_kind: episodic
episode_id: cursor-issue-348-2026-06-20-hands-on
title: Hands-on delivery — reader workspace theme (kiwifs#348)
tags: [kiwifs, issue-348, reader, theme, hands-on-takeover]
date: 2026-06-20
---

# kiwifs#348 — hands-on delivery (2026-06-20)

## Task

Deliver verified code for [kiwifs/kiwifs#348](https://github.com/kiwifs/kiwifs/issues/348): apply workspace theme and branding to published `/p/*` reader pages.

## Prior failure

Fleet agent left `internal/exporter/mkdocs.go` accidentally emptied (402 lines deleted, uncommitted), breaking compilation. Branch also contained many unrelated commits vs `origin/main`.

## Fix

1. Restored `internal/exporter/mkdocs.go` via `git restore`.
2. Rebased feature onto `origin/main` as branch `feat/reader-workspace-theme-348-clean` (cherry-picked commits `86273d7`, `6daae2f`).
3. All tests green on clean branch.

## Test results

```
go test ./internal/readertheme/... ./internal/api/... ./internal/exporter/... -count=1
→ PASS (readertheme 0.008s, api 10.283s, exporter 0.520s)
```

Targeted regression: 17 tests PASS (`TestPublishedPage_*`, `TestBuildCSS_*`, `TestBrandingFromConfig_*`, `TestCache_Get`, `TestApplyTheme`, `TestPublishedPageContentNegotiation`).

## Deliverables

- Branch: `feat/reader-workspace-theme-348-clean`
- Commits: `c9c2112` (feat), `be6cbc0` (episodic), `e3dbb70` (hands-on), `21d59cf` (CSS key sanitization)
- PR: https://github.com/kiwifs/kiwifs/pull/407
