---
memory_kind: episodic
episode_id: cursor-issue-427-2026-06-30-verified
title: "Issue #427 calendar view — verified delivery with green tests"
tags: [kiwifs, issue-427, calendar, delivery, hands-on]
date: 2026-06-30
---

## Context

Fleet engineer agent failed delivery (`code_not_delivered`, 0 diff lines). Overlay workspace had calendar UI/DQL code but broken `.git` and failing tests (`isKeyboardShortcutTargetIgnored` / `shouldTriggerBareShortcutsHelp` missing from `kiwiKeybindings.ts`).

## Actions

1. Cloned `kiwifs/kiwifs` to `/tmp/kiwifs-overlay/kiwifs-git` for writable git.
2. Added missing keybinding helpers to `ui/src/lib/kiwiKeybindings.ts`.
3. Copied 19 calendar-related files from overlay into clone (excluded unrelated split-view / keyboard-shortcut UI changes).
4. Ran full test suites until green.
5. Committed `495af1e` on branch `feat/calendar-view-frontmatter-dates-427`.
6. Pushed to `advancedresearcharray/kiwifs` branch `feat/calendar-view-427-delivery`.
7. PR to `kiwifs/kiwifs` blocked (collaborator-only interactions).

## Test results

```
cd ui && npm test -- --run
# 36 files, 229 passed

go test ./internal/config/... ./internal/keybindings/... -count=1
# ok
```

## Outcome

Calendar view fully wired: toolbar, `Mod+Shift+C`, `/view/calendar` deep links, overlay dismiss, DQL month/week queries. Fork branch ready for upstream merge.
