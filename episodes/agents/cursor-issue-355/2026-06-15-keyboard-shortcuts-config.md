---
memory_kind: episodic
episode_id: cursor-issue-355-2026-06-15
title: "Issue #355 — keyboard shortcuts config implementation"
tags: [kiwifs, issue-355, keybindings, ui, customization]
date: 2026-06-15
---

## Run log

1. Searched repo for existing keybinding/customization patterns; found hardcoded shortcuts in `App.tsx` and static list in `KeyboardShortcuts.tsx`.
2. Implemented `internal/keybindings` package with defaults, TOML/file merge, chord normalization, and conflict detection.
3. Added `GET /api/kiwi/keybindings` and `[ui.keybindings]` / `keybindings_file` config fields.
4. Built `kiwiKeybindings.ts` central manager + `useKeybindings` hook; refactored `App.tsx` to dispatch by action ID.
5. Updated shortcuts reference panel to show live bindings and conflict warnings.
6. Added Go + Vitest regression tests; all pass.

## Verification

```
go test ./internal/keybindings/... -count=1          # PASS
go test ./internal/api/... -run Keybindings -count=1 # PASS
go test ./internal/config/... -run UIConfigKeybindings -count=1 # PASS
cd ui && npm test -- --run kiwiKeybindings         # PASS (7 tests)
```

## Fleet handoff

Branch: `feat/keybindings-355-clean` (cherry-picked onto `origin/main`, conflicts resolved to exclude unrelated custom CSS). Push and open PR closing kiwifs/kiwifs#355.

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-355-keyboard-shortcuts-config.md`

## Takeover verification (2026-06-15)

Hands-on takeover after fleet publish failure. Rebased keybindings commit onto `origin/main` (4 conflict files resolved). All regression tests green. Kiwi fix docs written locally (gitignored); attempted depot write via REST (401 — no valid API key in env).
