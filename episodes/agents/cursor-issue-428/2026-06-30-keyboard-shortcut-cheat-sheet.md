---
memory_kind: episodic
episode_id: cursor-issue-428-2026-06-30
title: "Issue #428 — keyboard shortcut cheat sheet overlay"
tags: [kiwifs, issue-428, keybindings, ui, cmdk, overlay]
date: 2026-06-30
---

## Run log

1. Searched Kiwi fix docs — found prior verified implementation on `feat/issue-428-keyboard-shortcut-cheat-sheet` (commit `dc3ee20`).
2. Cherry-picked UI-only commit onto current branch `feat/issue-425-image-paste`; resolved import conflict in `kiwiKeybindings.test.ts` (kept both `normalizeKeyPart` and `shouldTriggerBareShortcutsHelp`).
3. Verified acceptance criteria: searchable `CommandDialog`, bare `?` / `mod+/` triggers, grouped sections, platform kbd labels, live bindings, HelpCircle toolbar button, Esc dismiss via existing `close_overlay`.
4. Kiwi MCP depot at `192.168.167.240:3333` unreachable from this host; wrote docs locally.

## Verification

```bash
cd ui && npm test -- --run src/lib/keyboardShortcutsOverlay.test.ts src/lib/kiwiKeybindings.test.ts src/lib/overlayDismiss.test.ts
# 3 files, 59 passed

cd ui && npm test -- --run
# 37 files, 260 passed
```

## Fleet handoff

Local commit: `270748e` on `feat/issue-425-image-paste`. Extract UI-only diff for PR closing kiwifs/kiwifs#428. Do not include unrelated knowledge template files from older delivery attempts.

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-428-keyboard-shortcut-cheat-sheet.md`
