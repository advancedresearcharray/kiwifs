---
memory_kind: episodic
episode_id: cursor-issue-427-2026-06-30-hands-on
title: "Issue #427 calendar view — hands-on verified delivery"
tags: [kiwifs, issue-427, calendar, delivery]
date: 2026-06-30
---

## Context

Fleet agent failed delivery check (`code_not_delivered`, overlay `.git` empty). Overlay workspace contained full calendar implementation but no verifiable commit.

## Actions

1. Searched Kiwi depot at `192.168.167.240:3333` — unreachable; read local fix doc.
2. Cloned `kiwifs/kiwifs` to `/tmp/kiwifs-git-work` (writable git).
3. Copied calendar feature files from overlay; stripped unrelated issue #428 keyboard-shortcut toolbar changes.
4. Ran full test suites until green.
5. Committed `4b3599a` on `feat/issue-427-calendar-view-frontmatter-dates`.
6. Pushed to `advancedresearcharray/kiwifs` fork; upstream PR blocked (collaborator-only).

## Test results

```
cd ui && npm test -- --run
# 35 files, 215 passed

go test ./internal/config/... ./internal/keybindings/... -count=1
# ok
```

## Outcome

Calendar view wired into App shell with `/view/calendar` deep links, toolbar toggle, `Mod+Shift+C`, and overlay dismiss. Fork PR ready for upstream merge.
