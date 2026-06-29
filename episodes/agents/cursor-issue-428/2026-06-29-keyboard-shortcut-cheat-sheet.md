---
memory_kind: episodic
episode_id: cursor-issue-428-2026-06-29
title: "Issue #428 — keyboard shortcut cheat sheet overlay"
tags: [kiwifs, issue-428, keybindings, ui, cheat-sheet, command-dialog]
date: 2026-06-29
---

## Run log

1. Read kiwifs/kiwifs#428 acceptance criteria; searched local `pages/fixes/` (no prior #428 doc; #355 episodic note covered keybindings config foundation).
2. Verified working tree already had partial implementation: `KeyboardShortcuts` CommandDialog, `useKeybindings`, `kiwiKeybindings`, API route — many files untracked.
3. Confirmed `App.tsx` wires `?`, `Cmd+/`, HelpCircle toolbar button, and Esc dismiss via `resolveOverlayDismiss`.
4. Added Vitest regressions: platform kbd labels, custom binding section, text-input guard, plain `?` vs mod chord.
5. Ran Go + Vitest suites for keybindings paths — all green.

## Verification

```
go test ./internal/keybindings/... -count=1                    # PASS (6 tests)
go test ./internal/api/ -run Keybinding -count=1               # PASS (4 tests)
cd ui && npm test -- --run src/lib/kiwiKeybindings.test.ts \
  src/lib/overlayDismiss.test.ts                               # PASS (17 tests)
```

## Fleet handoff

Commit locally on branch; fleet publishes PR closing kiwifs/kiwifs#428.

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-428-keyboard-shortcut-cheat-sheet.md`
