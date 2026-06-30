---
memory_kind: episodic
episode_id: cursor-issue-426-2026-06-30-v2
title: "Issue #426 split view — verified commit delivery"
tags: [kiwifs, issue-426, split-view, ui]
date: 2026-06-30
---

## Task

Deliver kiwifs/kiwifs#426 — split / side-by-side page view with commit and green tests.

## Work performed

1. Recovered prior uncommitted implementation from overlay stash.
2. Created clean branch `feat/split-side-by-side-426` from `origin/main`.
3. Resolved `App.tsx` merge conflicts (dropped calendar-view hunks; kept split-only changes).
4. Ran regression tests — 19/19 vitest + Go keybindings pass.
5. Committed split-view files and opened PR.

## Tests

```
go test ./internal/keybindings/...  — PASS
vitest splitView + kiwiKeybindings  — 19/19 PASS
```

## Notes

- Inline `resizable.tsx` used instead of shadcn npm install (overlay permissions).
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`
