---
memory_kind: episodic
episode_id: cursor-hands-on-427/2026-06-30-delivery-takeover
title: "Issue #427 — hands-on delivery takeover (verified)"
tags: [kiwifs, issue-427, calendar, ui, takeover, delivery, peer-review]
date: 2026-06-30
---

## Context

Fleet engineer failed delivery check. Hands-on takeover on `feat/issue-427-calendar-clean` using `GIT_DIR=.git.writable`. Kiwi cluster depot unreachable; local fix doc at `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view-frontmatter-dates.md`.

## Kiwi search

- Searched local fix doc (gitignored overlay path); no cluster MCP available.

## Peer review fixes

1. **Refetch loop** — `KiwiCalendar` had `now` in `loadMonth` useCallback deps; removed, read `new Date()` inside callback.
2. **Navigate regression** — `App.navigate()` did not call `setCalendarOpen(false)`; sidebar navigation left calendar visible over page content.
3. **Demo mock** — `isCalendarTableQuery()` routes TABLE calendar DQL to `calendarRows` (legacy `CALENDAR` only matched before).

## Verification

```bash
GIT_DIR=.git.writable git diff main...HEAD --stat
cd ui && npm test -- --run                    # 205 passed (35 files)
go test ./internal/config/... ./internal/keybindings/...
```

## Deliverables

- Branch: `feat/issue-427-calendar-clean`
- PR: opens/updates on fork → kiwifs/kiwifs main
- Closes: kiwifs/kiwifs#427
