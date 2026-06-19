---
memory_kind: episodic
episode_id: cursor-issue-345-2026-06-19
title: "Issue #345 — complete UI branding config"
tags: [kiwifs, issue-345, branding, ui-config, white-label]
date: 2026-06-19
---

# Issue #345 — complete UI branding config

## Target

[kiwifs/kiwifs#345](https://github.com/kiwifs/kiwifs/issues/345): `[ui.branding]` for app name, logo, favicon, and welcome copy.

## Investigation

1. Searched Kiwi depot (`branding config 345`) — found prior fix doc and fleet notes from 2026-06-18.
2. Confirmed PR #374/#376 landed config parsing, ui-config API, HTML injection, and React shell wiring.
3. Root cause for open issue: `document.title` not updated on navigation; Go regression tests for branding were removed during toolbar refactor on `feat/reader-workspace-theme-348`.

## Changes

- Added `ui/src/lib/pageTitle.ts` + tests — `formatDocumentTitle(activePath, branding.name)`.
- Wired `document.title` useEffect in `App.tsx`.
- Restored `TestLoadUIBranding`, `TestBrandingConfigResolved`, `TestResolveBrandingAssetURL` in `config_test.go`.
- Restored `TestUIConfig_BrandingFromConfig`, `TestUIConfig_BrandingDefaultsEmpty` in `handlers_ui_config_test.go`.

## Tests

```
go test ./internal/config/... -run 'UIBranding|BrandingConfig|ResolveBranding' -count=1  # PASS
go test ./internal/api/... -run 'UIConfig_Branding' -count=1                            # PASS
go test ./internal/webui/... -run 'InjectBranding' -count=1                               # PASS
cd ui && npm test -- --run src/lib/pageTitle.test.ts src/lib/branding.test.ts src/lib/uiConfigStore.test.ts  # PASS (11)
```

## Branch / PR

- Branch: `feat/issue-345-branding-config`
- Commit: `8dcf8ab`
- PR: https://github.com/kiwifs/kiwifs/pull/404 (Closes #345)
