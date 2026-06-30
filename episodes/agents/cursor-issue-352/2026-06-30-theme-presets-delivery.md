---
memory_kind: episodic
episode_id: cursor-issue-352-2026-06-30
title: "Issue #352 — workspace theme presets config delivery"
tags: [kiwifs, issue-352, theme, ui, config, presets]
date: 2026-06-30
---

## Task

Implement [kiwifs/kiwifs#352](https://github.com/kiwifs/kiwifs/issues/352): workspace-defined theme presets via `[ui.theme]` config, API endpoint, and UI merge/filter.

## Approach

1. Searched Kiwi fix docs — found prior `pages/fixes/kiwifs-kiwifs/issue-352-theme-presets-config.md` from earlier fleet run.
2. Located clean implementation on `feat/issue-352-theme-presets-config` commit `b1c88c7` (16 files, no unrelated template bloat).
3. Created `feat/issue-352-theme-presets-config-clean` from `main` and cherry-picked `b1c88c7`.
4. Ran regression tests — all passed.

## Tests

```
go test ./internal/themepresets/... -count=1 -v          → 7 passed
go test ./internal/config/... -run TestUIConfigThemePresets → PASS
go test ./internal/api/... -run 'ThemePresets|UIConfig_Theme' → 2 passed
cd ui && npm test -- --run src/themes/index.test.ts      → 4 passed
```

## Verification (2026-06-30, cursor autonomous run)

- Kiwi depot search attempted (`192.168.167.240:3333`) — unreachable; used local `pages/fixes/kiwifs-kiwifs/issue-352-theme-presets-config.md`.
- All acceptance criteria verified on `feat/issue-352-theme-presets-config-clean`:
  - Workspace JSON presets load from configurable `presets_dir` (default `.kiwi/themes/`)
  - Presets merge with built-ins; built-in names take precedence on clash
  - `allowed_presets` filters header selector and theme editor
  - Invalid JSON reported in API `errors` and shown in `KiwiThemeEditor`
  - Default behavior unchanged when no `[ui.theme]` config
- Fixed misleading JSDoc on `mergePresets` (built-in wins, not workspace).

## Deliverable

- Branch: `feat/issue-352-theme-presets-config-clean`
- Feature commit: `553ef59` · Docs: `587e46c` · Verification: `0b1094e`
- Closes #352
- Pushed to `fork/feat/issue-352-theme-presets-config-clean` and PR opened against `kiwifs/kiwifs` main

## Hands-on delivery (2026-06-30)

Re-ran full regression suite before push — all green:

```
go test ./internal/themepresets/... -count=1 -v          → 7 passed
go test ./internal/config/... -run TestUIConfigThemePresets → PASS
go test ./internal/api/... -run 'ThemePresets|UIConfig_Theme' → 2 passed
cd ui && npm test -- --run src/themes/index.test.ts      → 4 passed
```

Kiwi cluster depot (`192.168.167.240:3333`) unreachable; durable fix doc at `pages/fixes/kiwifs-kiwifs/issue-352-theme-presets-config.md` (gitignored locally).
