---
memory_kind: episodic
episode_id: cursor-auto-2026-06-29-issue-427
title: "Hands-on delivery: calendar view (#427)"
tags: [kiwifs, issue-427, calendar, ui, hands-on-takeover]
date: 2026-06-29
---

## Summary

Rebuilt issue #427 on current `main` after prior branch diverged (280-file destructive diff). Clean commit `2f3267d` adds 15 files, +870 lines.

## Actions

1. Restored missing `useUIFeatures.ts` deleted by fleet commit `80cc41c`.
2. Created clean branch from `main` via worktree; integrated calendar with `toolbarComposition` + `uiFeatures` (not standalone hook).
3. Ran tests: 198 passed (35 files, 8 calendarView tests). Go config tests ok.
4. Force-pushed `feat/issue-427-calendar-view` to fork.

## Verification

```bash
cd ui && npm test -- --run   # 198 passed
go test ./internal/config/... # ok
git diff main...HEAD --shortstat  # 15 files, +870/-4
```
