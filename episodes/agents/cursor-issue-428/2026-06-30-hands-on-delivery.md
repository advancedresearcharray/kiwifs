---
memory_kind: episodic
episode_id: cursor-issue-428-hands-on-2026-06-30
title: Hands-on delivery for issue #428 keyboard shortcut overlay
tags: [frontend, keyboard-shortcuts, ui, issue-428, hands-on]
date: 2026-06-30
---

## Task

Deliver kiwifs/kiwifs#428 — searchable keyboard shortcut cheat sheet overlay with verified tests and PR.

## Approach

1. Rebased implementation onto `origin/main` (branch `feat/issue-428-keyboard-shortcuts-pr`) to avoid bundling calendar view (#427) changes.
2. Replaced static `Dialog` overlay with filterable `CommandDialog`.
3. Added plain `?` trigger with `isTextInputTarget` / `shouldOpenShortcutsHelp` guards.
4. Added HelpCircle toolbar button; custom bindings section via `getCustomShortcutItems`.

## Tests

```bash
cd ui && npm test -- --run kiwiKeybindings overlayDismiss
go test ./internal/keybindings/... -count=1
```

18 UI tests passed; Go keybindings ok.

## Branch

`feat/issue-428-keyboard-shortcuts-pr` — pushed for PR against main.
