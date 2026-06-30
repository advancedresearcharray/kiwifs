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

## Deliverable

- Branch: `feat/issue-352-theme-presets-config-clean`
- Commit: `553ef59`
- Closes #352
- Fleet agent to push and open PR (local only per fleet policy)
