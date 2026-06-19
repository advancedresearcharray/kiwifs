---
memory_kind: episodic
episode_id: cursor-issue-348-2026-06-19
title: Apply workspace theme to published reader pages (kiwifs#348)
tags: [kiwifs, issue-348, reader, theme, sprout-idle-nudge]
date: 2026-06-19
---

# Apply workspace theme to published reader pages (kiwifs#348)

## Task

Fix [kiwifs/kiwifs#348](https://github.com/kiwifs/kiwifs/issues/348): published `/p/*` reader HTML must use `.kiwi/theme.json` CSS tokens and `[ui.branding]` (favicon, title prefix, footer logo).

## Investigation

1. Searched Kiwi depot via `http://192.168.167.240:3333/api/kiwi/search?q=reader+theme` — prior fix doc and episodic notes found from 2026-06-16/18 attempts.
2. Confirmed root cause: `handlers_reader.go` hardcoded CSS vars and ignored theme/branding.
3. Cherry-picked proven implementation from commit `67c130de` (scoped readertheme package + handler integration).

## Implementation

- New `internal/readertheme/` — theme.json mtime cache, CSS builder (light/dark/system), branding resolver
- Updated `internal/api/handlers_reader.go` — inject `{{.ThemeCSS}}`, branded title/favicon/footer
- Regression tests in `handlers_reader_theme_test.go` and `readertheme/theme_test.go`

## Test results

```
go test ./internal/readertheme/... ./internal/api/... -run 'TestPublishedPage|TestBuildCSS|TestBranding|TestCache|TestApplyTheme' -count=1 -v
→ PASS (17 tests)
```

## Deliverables

- Branch: `feat/reader-workspace-theme-348`
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-348-reader-theme-published-pages.md`
- Local commit ready for fleet publish (no remote push per fleet policy)
