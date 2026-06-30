---
memory_kind: episodic
episode_id: cursor-issue-427-2026-06-30-app-integration
title: "Issue #427 calendar view — App.tsx shell integration"
tags: [kiwifs, issue-427, calendar, app-shell, delivery]
date: 2026-06-30
---

## Context

Issue #427 required a calendar view for frontmatter dates. `KiwiCalendar.tsx`, `calendarView.ts`, feature flags, and keybindings already existed, but `App.tsx` never wired the view — toolbar button, keyboard shortcut, overlay dismiss, and `/view/calendar` deep-link were non-functional.

## Actions

1. Searched Kiwi depot at `192.168.167.240:3333` — unreachable; read local fix doc `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view-frontmatter-dates.md`.
2. Added `appViewRoutes.ts` helper for `/view/*` deep-link resolution and pathname preservation.
3. Wired calendar into `App.tsx`: state, `openBuiltinView`, toolbar toggle, `toggle_calendar` keybinding, Escape dismiss, render branch, demo `initialView: "calendar"`.
4. Extended overlay dismiss regression tests for calendar priority.
5. Ran full test suites.

## Test results

```
cd ui && npm test -- --run
# 36 files, 229 passed

go test ./internal/config/... ./internal/keybindings/... -count=1
# ok
```

## Outcome

Calendar view is now accessible from the toolbar, `Mod+Shift+C`, and `/view/calendar` deep links. Fleet agent can commit and open PR.
