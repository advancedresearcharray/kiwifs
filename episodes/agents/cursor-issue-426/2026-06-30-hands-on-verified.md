---
memory_kind: episodic
episode_id: cursor-issue-426-2026-06-30-verified
title: "Issue #426 split view — verification and regression tests"
tags: [kiwifs, issue-426, split-view, ui, verification]
date: 2026-06-30
---

## Task

Verify and complete kiwifs/kiwifs#426 (split / side-by-side page view) on branch `feat/split-side-by-side-426`.

## Work performed

1. Checked out `feat/split-side-by-side-426` (3 commits: feature, keybinding test, pane-size restore fix).
2. Verified all acceptance criteria against implementation (context menus, `mod+\\` toggle, independent panes, resizable divider, close button, history compare, mobile guard).
3. Added regression tests: `normalizeSizes` edge cases, `matchBoundAction` for `toggle_split`.
4. Wrote durable fix doc at `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`.
5. Kiwi MCP depot unreachable (`192.168.167.240:3333` connection refused); docs written to workspace filesystem.

## Tests

```
cd ui && npm test -- --run src/lib/splitView.test.ts src/lib/kiwiKeybindings.test.ts  — 22/22 PASS
go test ./internal/keybindings/... -count=1                                           — PASS
```

## Branch state

Ready for fleet publish (push + PR). Closes #426.
