---
memory_kind: episodic
episode_id: cursor-issue-427/2026-06-30-calendar-view
title: "Issue #427 — calendar view for frontmatter dates"
tags: [kiwifs, issue-427, calendar, ui, feat]
date: 2026-06-30
---

## Task

Implement feat(ui): calendar view for frontmatter dates (kiwifs/kiwifs#427).

## Before

- Searched local fix doc at `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view-frontmatter-dates.md` (prior partial delivery noted).
- Kiwi MCP gateway unavailable; cluster at `192.168.167.240:3333` connection refused.
- Prior implementation existed on branch `feat/issue-427-calendar-view` (commits `c17a255`, `2c00e7d`) but was not on current `main` / active work branches.

## Work

1. Created branch `feat/issue-427-calendar-view` from current HEAD.
2. Restored 19 calendar-related files from prior branch (UI component, DQL helpers, App shell wiring, feature flags, keybindings, tests).
3. Verified acceptance criteria: monthly grid, week view on mobile, date-field selector, popover page list, prev/next/today/month picker, tag/state dot colors, `/view/calendar` deep links, `[ui.features] calendar`.

## Tests

```bash
cd ui && npm test -- --run
# 35 files, 222 passed

cd ui && npm test -- --run calendarView appViewRoutes
# 2 files, 26 passed

go test ./internal/config/... ./internal/keybindings/... -count=1
# ok
```

## Peer review fix (hands-on takeover)

Prior branch mixed MCP #424 commits (`spec2026Handler`, `validateRoutingHeaders`) with calendar work, causing false security/correctness review findings on a 2187-line diff.

1. Reset `feat/issue-427-calendar-view` to `main`.
2. Cherry-picked only calendar commit (`3abee7e` → `f2ae395`).
3. Verified diff is 20 files / ~1151 lines, calendar-only.
4. Added durable fix doc at `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view-frontmatter-dates.md`.

## Outcome

Calendar view ready for PR. Closes #427 when merged.
