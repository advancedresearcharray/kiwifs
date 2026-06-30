---
memory_kind: episodic
episode_id: cursor-issue-428-2026-06-30
title: Issue #428 keyboard shortcut cheat sheet overlay — UI completion
tags: [kiwifs, ui, issue-428, keybindings]
date: 2026-06-30
---

Completed frontend gaps for kiwifs/kiwifs#428 on branch `feature-mcp-2026-07-28-spec` workspace overlay.

## Problem

Backend keybindings API and partial UI existed, but `KeyboardShortcuts.tsx` still used a static `Dialog` without search, bare `?` trigger, typing-target guard, or toolbar `HelpCircle` button.

## Actions

1. Upgraded `KeyboardShortcuts.tsx` to searchable `CommandDialog` with grouped sections and `(custom)` labels.
2. Added `isTypingTarget`, `shouldOpenShortcutsHelp`, `isCustomBinding` helpers in `kiwiKeybindings.ts`.
3. Wired `App.tsx`: plain `?` opens overlay outside inputs; global shortcuts skip editable targets; toolbar `HelpCircle` button.
4. Added `title` prop to `CommandDialog` for screen-reader accessibility.
5. Extended `useKeybindings` to return fully merged `defaults` record.

## Tests (green)

```
go test ./internal/keybindings/... -count=1          ok (6 tests)
go test ./internal/api/ -run GetKeybindings -count=1 ok (4 tests)
cd ui && npm test -- --run kiwiKeybindings overlayDismiss  # 17 passed
```

## Files changed

- `ui/src/lib/kiwiKeybindings.ts`, `kiwiKeybindings.test.ts`
- `ui/src/hooks/useKeybindings.ts`
- `ui/src/components/KeyboardShortcuts.tsx`
- `ui/src/components/ui/command.tsx`
- `ui/src/App.tsx`

## Verification (2026-06-30 hands-on takeover)

Re-ran full test suite after fleet delivery rejection (broken `.git` overlay; use `GIT_DIR=.git.writable`):

```
go test ./internal/keybindings/... -count=1          ok
go test ./internal/api/ -run GetKeybindings -count=1 ok
cd ui && npm test -- --run kiwiKeybindings overlayDismiss  # 17 passed
```

Applied peer-review fix: swagger `@Tags ui` on GetKeybindings (was `theme`).
Branch `feat/428-keyboard-shortcut-cheat-sheet` pushed to fork; PR #36 on advancedresearcharray/kiwifs.

## Hands-on takeover (2026-06-30)

- Verified implementation in workspace overlay; `.git` mount is empty — use `GIT_DIR=.git.writable`.
- Re-ran tests: Go keybindings + GetKeybindings API + UI kiwiKeybindings/overlayDismiss (17 pass).
- Commits on branch: `6ad309a` (UI overlay), `9117668` (swagger tag peer-review fix).
- Kiwi depot at 192.168.167.240:3333 unreachable; fix doc at `pages/fixes/kiwifs-kiwifs/issue-428-keyboard-shortcut-cheat-sheet.md`.
