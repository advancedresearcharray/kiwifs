---
memory_kind: episodic
episode_id: cursor-hands-on-427-2026-06-29
title: "Hands-on delivery — issue #427 calendar view PR"
tags: [kiwifs, ui, calendar, issue-427, hands-on-takeover]
date: 2026-06-29
---

## Task

Hands-on takeover for kiwifs/kiwifs#427 after fleet agent failed peer review. Prior branch mixed spam-filter removal and split-view changes (~20k lines); peer review flagged unrelated spam filter deletion.

## Actions

1. Created worktree at `/tmp/kiwifs-427-pr` on `feat/issue-427-calendar-view-v2` (focused 17-file diff vs main).
2. Cherry-picked hardening commit `01e7b9e` (mobile week anchor + feature-flag URL guards); resolved App.tsx conflict using `features.calendar` from UI config store.
3. Ran tests: UI 200 passed, Go config ok.
4. Wrote fix doc at `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view.md`.
5. Pushed branch and opened PR against kiwifs/kiwifs main.

## Kiwi MCP

Gateway at 192.168.167.240:3333 unreachable. Fix doc committed locally for fleet sync.

## Outcome

Focused PR closes #427 with verified tests and no unrelated diffs.
