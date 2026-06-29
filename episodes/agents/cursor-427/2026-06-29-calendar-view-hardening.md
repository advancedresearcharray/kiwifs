---
memory_kind: episodic
episode_id: cursor-427-2026-06-29
title: "Issue #427 — calendar view hardening + mobile week fix"
tags: [kiwifs, ui, calendar, issue-427, bounty]
date: 2026-06-29
---

## Task

Finish kiwifs/kiwifs#427: calendar toolbar view for date frontmatter. Prior commit `c351c4d` delivered core feature; this run fixed mobile week anchor and feature-flag deep-link bugs.

## Reproduction

1. Open Calendar on viewport `< 768px` with a selected day mid-month (e.g. June 18).
2. Week grid showed June 1–7 (first week of month) instead of the week containing the 18th.
3. Navigate to `/view/calendar` with `[ui.features] calendar = false` — view still opened.

## Root cause

- `CalendarGrid` passed `startOfMonth(month)` to `buildWeekGrid()` on mobile.
- `App.tsx` popstate/deep-link did not check `isFeatureEnabled(uiFeatures, "calendar")`.

## Fix

- Added `weekGridAnchor()` and `defaultSelectedDateKey()` in `calendarView.ts`.
- `KiwiCalendar` syncs selection on month change; passes `weekAnchor` to grid.
- `App.tsx` closes calendar and clears URL when feature disabled.

## Tests

- 2 new regression tests in `calendarView.test.ts` (10 total).
- Full UI suite: 133 passed. Go config: ok.

## Kiwi MCP

Gateway at 192.168.167.240:3333 unreachable (curl exit 7). Fix doc updated locally at `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view.md`.

## Outcome

Local commit ready for fleet publish. Closes #427 acceptance criteria.
