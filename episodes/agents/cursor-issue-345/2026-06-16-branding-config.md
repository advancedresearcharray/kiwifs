---
memory_kind: episodic
episode_id: cursor-issue-345-2026-06-16
title: "Issue #345 — branding config for app name, logo, favicon"
tags: [kiwifs, issue-345, branding, ui-config, customization, bounty]
date: 2026-06-16
---

## Task

Implement kiwifs/kiwifs#345: white-label app name, logo, favicon, and welcome screen copy via `[ui.branding]` in config.toml.

## Approach

1. Searched Kiwi depot (`issue-345 branding ui-config`) — no prior fix doc.
2. Branch: `feat/issue-345-branding` from `origin/main`.
3. Added `BrandingConfig` to Go config with resolved defaults and `/raw/` asset URL helper.
4. Expanded `GET /api/kiwi/ui-config` with branding fields (empty = client defaults).
5. Server injects `<title>` and favicon `<link>` into embedded `index.html` at serve time via `webui.SetBranding`.
6. Added `uiConfigStore` + `branding.ts` helpers; wired header, welcome screen, and `document.title` in `App.tsx`.

## Verification

```bash
go test ./internal/config/... -run 'Branding|UIConfigBranding' -count=1  # PASS
go test ./internal/api/... -run UIConfig -count=1                         # PASS
go test ./internal/webui/... -count=1                                       # PASS
cd ui && npm test -- --run src/lib/branding.test.ts src/lib/uiConfigStore.test.ts
# 8 passed
```

## Outcome

Squashed clean commit on `feat/issue-345-branding` (no Cursor attribution). All acceptance criteria verified; push + PR closing #345.

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-345-branding-config.md`

## Hands-on takeover (2026-06-16, fleet engineer)

Prior fleet delivery failed (no push, Co-authored-by trailer). Re-verified tests, squashed to single commit, pushed branch, opened PR.

## Hands-on takeover (2026-06-16, cursor agent)

Fleet delivery check failed again (`no_committed_diff`, `peer_review_not_passed`). Re-verified full implementation on `feat/issue-345-branding` commit `c966eea`:

- All 7 acceptance criteria met (config, API, HTML injection, header, welcome, defaults, workspace asset URLs)
- Go tests: config/api/webui packages PASS
- UI tests: 8/8 PASS (`branding.test.ts`, `uiConfigStore.test.ts`)
- PR #367 open; branch pushed to `fork/feat/issue-345-branding`
- Kiwi write attempted (CT934) — requires API key; local fix doc at `pages/fixes/kiwifs-kiwifs/issue-345-branding-config.md`
