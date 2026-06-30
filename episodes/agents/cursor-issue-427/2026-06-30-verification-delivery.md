---
memory_kind: episodic
episode_id: cursor-issue-427/2026-06-30-verification-delivery
title: "Issue #427 — calendar view verification and demo mock fix"
tags: [kiwifs, issue-427, calendar, ui, verification, delivery]
date: 2026-06-30
---

## Context

Autonomous delivery for kiwifs/kiwifs#427 on branch `feat/issue-427-calendar-clean`. Prior hands-on takeover landed core calendar UI; this run verified acceptance criteria, fixed demo mock TABLE-query routing, and confirmed regression tests.

## Kiwi search

- Cluster depot `http://192.168.167.240:3333` unreachable (curl exit 7).
- Local fix doc found: `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view-frontmatter-dates.md`.

## Root cause (demo gap)

`KiwiCalendar` issues `TABLE … WHERE striptime(field) >= DATE(…)` DQL, but `apiMock.ts` only returned `calendarRows` for legacy `CALENDAR` queries. Demo log template calendar view showed empty days despite `calendarRows` overrides.

## Fix

- Added `isCalendarTableQuery()` in `calendarView.ts` to detect month/range TABLE queries.
- Updated `apiMock.ts` to serve `calendarRows` for both `CALENDAR` and calendar TABLE DQL.
- Regression tests for query detection in `calendarView.test.ts`.

## Verification

```bash
GIT_DIR=.git.writable git diff main...HEAD --stat
cd ui && npm test -- --run                    # 206 passed (35 files)
go test ./internal/config/... ./internal/keybindings/...
```

## Deliverables

- Branch: `feat/issue-427-calendar-clean` (local commit; fleet publishes PR)
- Closes: kiwifs/kiwifs#427
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view-frontmatter-dates.md`
