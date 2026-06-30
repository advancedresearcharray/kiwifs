---
memory_kind: episodic
episode_id: cursor-issue-427/2026-06-30-calendar-view-delivery
title: "Issue #427 — calendar view for frontmatter dates"
tags: [kiwifs, issue-427, calendar, ui, delivery]
date: 2026-06-30
---

## Task

Implement kiwifs/kiwifs#427: monthly calendar toolbar view for frontmatter date fields.

## Approach

1. Cherry-picked verified calendar commits (`ad5bfa9`, `11334aa`) onto `main` as branch `feat/issue-427-calendar-clean` (avoids unrelated tree-order commits on older integration branch).
2. Stripped Cursor co-author attribution from commit messages per fleet policy.
3. Added regression tests: `buildCalendarQueryRange`, `dayAfter`, calendar route gating, toolbar calendar filter.

## Verification

```bash
cd ui && npm test -- --run   # 200 passed (35 files)
go test ./internal/config/... -count=1   # ok
```

## Branch

`feat/issue-427-calendar-clean` — 2 commits, ready for fleet publish (no push/PR from Cursor).

## Kiwi MCP

Gateway at 192.168.167.240:3333 unreachable; fix doc updated locally at `pages/fixes/kiwifs-kiwifs/issue-427-calendar-view-frontmatter-dates.md`.
