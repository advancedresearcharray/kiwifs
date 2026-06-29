---
memory_kind: episodic
episode_id: cursor-issue-426-2026-06-29
title: Issue #426 split side-by-side page view
tags: [kiwifs, ui, split-view, issue-426]
date: 2026-06-29
---

## Task

Implement kiwifs/kiwifs#426 — split / side-by-side page view with resizable panes, context menus, keyboard shortcut, history compare, session persistence, mobile guard.

## Approach

1. Searched repo — no prior split-view implementation (`LAYOUT.md` noted single-pane only).
2. Kiwi MCP gateway at 192.168.167.240:3333 unreachable from workspace; wrote fix doc locally under `pages/fixes/`.
3. Built `SplitViewProvider` + sessionStorage persistence, inline resizable panels (npm install blocked on read-only `node_modules`).
4. Wired App main area, tree/wiki-link menus, `Mod+\` keybinding, and KiwiHistory compare.

## Verification

- `go test ./internal/keybindings/...` — pass
- `npx vitest run src/lib/splitView.test.ts src/lib/kiwiKeybindings.test.ts` — 17 pass

## Deliverable

Local commit only (fleet publishes PR). Closes #426 when merged.
