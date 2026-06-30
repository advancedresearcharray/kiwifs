---
memory_kind: semantic
doc_id: kiwifs-kiwifs-issue-428-keyboard-shortcut-cheat-sheet
title: Searchable keyboard shortcut cheat sheet overlay
tags: [kiwifs, issue-428, keybindings, ui, shortcuts, command-dialog]
repo: kiwifs/kiwifs
issue_number: 428
languages: [typescript, react]
status: verified
peer_review: pass
date: 2026-06-30
verified: 2026-06-30T03:37:00Z
delivery_commit: 9b78ea7
---

## Problem

Issue #428 requested a keyboard shortcut cheat sheet overlay: searchable, grouped by category, platform-correct modifier labels, custom keybinding overrides from config, and triggers via `?` and `Cmd+/` outside editable targets. The prior static `Dialog` list did not support filtering and lacked standalone `?` / toolbar entry.

## Root cause

#355 added keybinding infrastructure (`kiwiKeybindings.ts`, `useKeybindings`, `KeyboardShortcuts` dialog) but the cheat sheet UI was a non-searchable static list. `App.tsx` only handled `shortcuts_help` via `matchBoundAction` (mod+/), not standalone `?`, and had no focus guard for inputs/editors.

## Solution

1. **`KeyboardShortcuts.tsx`** — Replaced `Dialog` with shadcn `CommandDialog` (`CommandInput`, grouped `CommandItem`s, styled `<kbd>` chips). Sections built via `buildShortcutSectionsForDisplay`.
2. **`kiwiKeybindings.ts`** — Added `isKeyboardShortcutIgnoredTarget`, `isQuestionMarkShortcut`, `getCustomKeybindingItems`, `buildShortcutSectionsForDisplay` (appends Custom section when bindings differ from defaults).
3. **`App.tsx`** — Standalone `?` handler before `matchBoundAction`; focus guard on `shortcuts_help`; `HelpCircle` toolbar button; pass `defaults` from `useKeybindings` to overlay.
4. **`command.tsx`** — Optional `title` prop on `CommandDialog` for a11y.

## Files changed

- `ui/src/components/KeyboardShortcuts.tsx`
- `ui/src/lib/kiwiKeybindings.ts`
- `ui/src/lib/kiwiKeybindings.test.ts`
- `ui/src/App.tsx`
- `ui/src/components/ui/command.tsx`

## Tests

```bash
cd ui && npm test -- --run kiwiKeybindings overlayDismiss
# Test Files  2 passed (2)
# Tests  20 passed (20)
```

New regressions cover ignored targets, `?` vs mod+/, Custom section for config overrides, and default-only bindings (no Custom section).

## Peer review notes

- `?` opens overlay only when `!isKeyboardShortcutIgnoredTarget` (inputs, contenteditable, CodeMirror, role=textbox).
- `Cmd+/` toggles with same guard; Esc dismisses via existing `resolveOverlayDismiss` priority (shortcuts first).
- Custom section compares merged bindings to server `defaults` from `GET /api/kiwi/keybindings`.
- Search uses cmdk fuzzy filter on label + chord + section.

## Reuse guide

When adding a new shortcut:

1. Add action to `KeybindingAction` and `DEFAULT_KEYBINDINGS` in `kiwiKeybindings.ts`.
2. Add handler in `App.tsx` `switch (action)`.
3. Add entry to `SHORTCUT_SECTIONS` for cheat sheet display.
4. Go side: register in `internal/keybindings` defaults if needed.

To change overlay behavior (e.g. new trigger key), extend `isQuestionMarkShortcut` or add parallel guard in `App.tsx` keydown effect.
