---
memory_kind: episodic
episode_id: cursor-issue-428-2026-06-30-delivery
title: "Issue #428 keyboard shortcut cheat sheet — delivery"
tags: [kiwifs, issue-428, ui, keybindings, cmdk]
date: 2026-06-30
---

## Run log

1. Searched Kiwi fix docs — prior fix doc at `pages/fixes/kiwifs-kiwifs/issue-428-keyboard-shortcut-cheat-sheet.md`; workspace had partial implementation (keybinding helpers + tests) but static Dialog overlay.
2. Completed
Errata: `CommandDialog` searchable overlay, bare `?` handler in `App.tsx`, `HelpCircle` toolbar button, `[cmdk-input]` in focus guard.
3. Applied focused diff across 5 files; added alt-key regression test.
4. Ran `cd ui && npm test -- --run` — **36 files, 230 passed**.

## Outcome

Searchable keyboard shortcut cheat sheet overlay complete: `?` and `mod+/` triggers, grouped categories, platform-aware `<kbd>` labels, custom bindings from API, Esc/click-outside dismiss, toolbar help button.
