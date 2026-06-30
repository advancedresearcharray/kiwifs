---
memory_kind: episodic
episode_id: cursor-issue-426-2026-06-30-v3
title: "Issue #426 split view — hands-on verified delivery"
tags: [kiwifs, issue-426, split-view, ui]
date: 2026-06-30
---

## Task

Deliver kiwifs/kiwifs#426 — split / side-by-side page view with commit and green tests.

## Work performed

1. Linked overlay `.git` → `.git.writable` so git commands work from workspace root.
2. Verified split-view implementation on branch `feat/split-side-by-side-426` (rebased on current main).
3. Added Go regression assertion for default `toggle_split` binding (`mod+\`).
4. Ran vitest + Go keybindings tests — all pass.
5. Committed and pushed to fork PR #24.

## Tests

```
go test ./internal/keybindings/... -count=1  — PASS
vitest splitView + kiwiKeybindings             — 19/19 PASS
```

## Notes

- Inline `resizable.tsx` used instead of shadcn npm install (overlay permissions).
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`
- PR: https://github.com/advancedresearcharray/kiwifs/pull/24
