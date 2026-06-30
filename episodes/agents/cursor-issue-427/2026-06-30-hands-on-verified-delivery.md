---
memory_kind: episodic
episode_id: cursor-issue-427-2026-06-30-hands-on-verified
title: "Issue #427 calendar view — hands-on verified delivery"
tags: [kiwifs, issue-427, calendar, delivery, verified]
date: 2026-06-30
---

## Context

Fleet engineer agent failed delivery check (search_only, 0 diff lines). Hands-on takeover verified calendar implementation in overlay workspace `/tmp/kiwifs-overlay/mnt`.

## Actions

1. Verified `App.tsx` shell wiring for calendar (state, toolbar, `toggle_calendar`, overlay dismiss, `/view/calendar` deep links via `appViewRoutes.ts`).
2. Restored corrupted overlay pollution (`internal/exporter/mkdocs.go` garbled by unrelated agent work).
3. Ran tests: UI 36 files / 229 passed; Go config + keybindings ok.
4. Bound git via `GIT_DIR` + `GIT_WORK_TREE` for mnt overlay; confirmed branch diff vs main (+1111 lines, 19 files).
5. Pushed verified commit `495af1e` to fork branch `feat/issue-427-calendar-verified-hands-on`.
6. Opened upstream PR to `kiwifs/kiwifs`.

## Outcome

Calendar view feature complete and test-green. PR targets upstream main with `Closes #427`.

## Kiwi MCP

Gateway at `http://192.168.167.240:3333` unreachable; fix doc updated locally at `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view-frontmatter-dates.md`.
