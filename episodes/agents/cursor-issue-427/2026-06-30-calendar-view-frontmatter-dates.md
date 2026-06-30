---
memory_kind: episodic
episode_id: cursor-issue-427-2026-06-30
title: Calendar view for frontmatter dates (#427)
tags: [kiwifs, ui, calendar, dql, issue-427]
date: 2026-06-30
---

## Work

Hands-on delivery for kiwifs/kiwifs#427 after prior fleet attempt left calendar shell wiring without implementation modules on the keyboard-shortcuts branch.

## Root cause

Partial merge: `App.tsx` referenced `KiwiCalendar` and `/view/calendar` routes before `KiwiCalendar.tsx`, `calendarView.ts`, and `appViewRoutes.ts` existed on that branch.

## Solution

Dedicated branch `feat/calendar-view-frontmatter-dates-427` from `origin/main` with:

- `KiwiCalendar.tsx` — month grid, mobile week view, date-field selector, popover page lists
- `calendarView.ts` — DQL month/week queries, field discovery, dot colors
- `appViewRoutes.ts` — `/view/calendar` deep links
- Feature flag `[ui.features] calendar`, `toggle_calendar` keybinding (Mod+Shift+C)
- 16 regression tests in `calendarView.test.ts` plus routing/dismiss tests

## Tests

```
cd ui && npm test -- --run  → 35 files, 222 passed
go test ./internal/config/... ./internal/keybindings/...  → ok
```

## Branch

`feat/calendar-view-frontmatter-dates-427` — single commit, no keyboard-shortcut (#428) mixing.
