---
memory_kind: episodic
episode_id: cursor-fleet-2026-06-30-issue-426-split-view
title: "Issue #426 split / side-by-side page view"
tags: [kiwifs, ui, split-view, issue-426, fleet-delivery]
date: 2026-06-30
---

# Issue #426 split view delivery

## Task

Implement kiwifs/kiwifs#426 — side-by-side page view with tree/wiki-link menus, `Mod+\` toggle, resizable panes, history compare, mobile guard, sessionStorage persistence, and regression tests.

## Actions

1. Branched `feat/split-view-426` from `main`.
2. Added `splitView.ts` pure state + 8 vitest cases; wired state in `App.tsx`.
3. Built `KiwiSplitView` + self-contained `ui/resizable.tsx` (overlay `node_modules` read-only — skipped `npm install react-resizable-panels`).
4. Extended `KiwiPage` with `versionHash`/`readOnly` and wiki-link context menu; tree/history entry points.
5. Registered `toggle_split_view` in Go + TS keybindings with backslash normalization.
6. Fixed secondary pane × to fully exit split view; wired right-pane star/pin/history to secondary path.
7. Added Go regression assertion for `toggle_split_view` default binding.

## Verification

- `npm test -- src/lib/splitView.test.ts src/lib/kiwiKeybindings.test.ts` → 17 passed
- `npm test` (full UI suite) → 198 passed
- `go test ./internal/keybindings/... -count=1` → ok (includes `toggle_split_view` default)
- Kiwi MCP gateway at `192.168.167.240:3333` unreachable — fix doc written locally under `pages/fixes/`

## Handoff

Branch `feat/split-view-426` ready for fleet publish (push + PR closing #426).
