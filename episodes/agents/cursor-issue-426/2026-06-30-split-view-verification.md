---
memory_kind: episodic
episode_id: cursor-issue-426-2026-06-30
title: "Issue #426 split view — verification and CI template fixes"
tags: [kiwifs, split-view, issue-426, regression-tests]
date: 2026-06-30
---

# Issue #426 split view verification

## Context

Redelivery on `feat/issue-426-split-view` after PR #432 was closed (bot nudge comments, not code quality). Prior fleet work implemented full split-view feature; this run verified tests and fixed remaining CI blockers.

## Pre-work

- Searched local fix doc: `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`
- Kiwi MCP gateway unavailable (empty MCP catalog)

## Verification

Branch `feat/issue-426-split-view` contains 7 commits on top of main implementing:
- `splitView.ts` state + sessionStorage
- `KiwiSplitView.tsx` + `ui/resizable.tsx`
- App wiring, tree/wiki-link/history entry points, `Mod+\` toggle
- Mobile guard (<768px)

## Fixes this run

1. **Workspace template lint (CI):** Fixed broken wiki links in runbook/research templates added by `d8e6573`:
   - `postmortems/template.md` → link to `incidents/template` and `procedures/deploy-rollback`
   - Added `experiments/exp-002-placeholder.md`; updated exp-001 baseline link
2. **Regression test:** `persists custom pane sizes across session reload` in `splitView.test.ts`

## Test results

```
cd ui && npm test -- src/lib/splitView.test.ts src/lib/kiwiKeybindings.test.ts  → 26 passed
go test ./internal/keybindings/... -count=1  → ok
go test ./internal/workspace/... -count=1  → ok
```

## Fleet handoff

- Branch: `feat/issue-426-split-view`
- Do not push from Cursor; fleet publishes PR closing #426
- Fix doc updated at `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`
