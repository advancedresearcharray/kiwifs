---
memory_kind: episodic
episode_id: cursor-issue-426-2026-06-30-hands-on-delivery
title: "Issue #426 split view — verified hands-on delivery"
tags: [kiwifs, issue-426, split-view, ui, delivery]
date: 2026-06-30
---

## Task

Hands-on takeover for kiwifs/kiwifs#426 after fleet delivery check failed (no_committed_diff).

## Work performed

1. Verified full split-view implementation on `feat/split-side-by-side-426`.
2. Extracted `resolveInitialPanelSizes` from `resizable.tsx` with regression tests for persisted pane width restore (#426).
3. Added durable fix doc at `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`.
4. Ran vitest + go keybindings; all green.
5. Pushed branch; fork PR #24 open.

## Tests

```
cd ui && npm test -- --run splitView kiwiKeybindings resizable  — 26/26 PASS
go test ./internal/keybindings/... -count=1                        — PASS
```

## Branch / PR

- Branch: `feat/split-side-by-side-426`
- Fork PR: https://github.com/advancedresearcharray/kiwifs/pull/24
- Closes kiwifs/kiwifs#426
