---
memory_kind: episodic
episode_id: cursor-hands-on-406-2026-06-19-merge-nurture
title: PR #406 ADR init template — merge nurture CI green
tags: [kiwifs, workspace, adr, issue-328, issue-406, hands-on, merge-nurture, ci]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#328 / PR #406 — feat(workspace): ship ADR init template with workflow and schema

## Context

Merge-first queue item. CI was IN_PROGRESS on arrival. Overlay FS `.git/index` had stale file handle showing staged partial revert of peer-review hardening (`685f496`) while working tree matched HEAD.

## Actions

1. Kiwi search — fix doc `pages/fixes/kiwifs-kiwifs/issue-328-adr-init-template.md` present.
2. Rebuilt git index via `GIT_INDEX_FILE=/tmp/kiwifs-index-fresh git read-tree HEAD` (stale handle on `.git/index` prevents `mv`).
3. Verified peer-review hardening intact at HEAD (`685f496`): auth guidance, deciders placeholder, SCHEMA rejected transitions, workspace + cmd regression tests.
4. Ran local ADR regression tests — all green.
5. Monitored CI run 27851365303 — test job SUCCESS; PR checks green.

## Test output

```
go test ./internal/workspace/... -count=1 -run 'ADR|InitADR|ListInit'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.015s

go test ./cmd/... -count=1 -run 'ADR|Init'
ok  github.com/kiwifs/kiwifs/cmd  0.046s
```

## Outcome

No code changes required. PR #406 CI green, merge-ready. Fleet agent may publish if index rebuild needed on push host.
