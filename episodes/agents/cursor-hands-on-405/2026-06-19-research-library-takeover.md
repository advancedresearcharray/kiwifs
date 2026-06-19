---
memory_kind: episodic
episode_id: cursor-hands-on-405-2026-06-19-takeover-v2
title: PR #405 hands-on takeover — peer review and delivery commit
tags: [kiwifs, workspace, research, issue-334, pr-405, hands-on, delivery, peer-review]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#405 — feat(workspace): ship research library init template with reading workflow (closes #334)

## Problem

Fleet delivery check failed (`not_committed`, `peer_review_not_passed`). Overlay
git index was unwritable; a prior agent left staged changes that would revert the
research template (delete `.kiwi/workflows/reading.json`, restore legacy
`literature/` layout).

## Peer review findings

1. `TestInitResearchTemplateIncludesReadingWorkflow` did not assert
   `papers/transformer-survey.md` — the second cross-cited example paper.
2. Workflow tests only covered forward transitions; backward transitions in
   `reading.json` were untested.
3. Schema tests did not reject invalid `state` enum or wrong `type` const.
4. `SCHEMA.md` / `playbook.md` implied strictly linear transitions; workflow
   allows backward moves when revisiting a paper.

## Fix

- Extended init and schema/workflow regression tests.
- Documented backward transitions in SCHEMA and playbook.

## Test output

```
go test ./internal/workspace/... -count=1 -run 'Research'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.009s
```

## Commit

`fix(workspace): peer-review hardening for research template tests`

## PR

https://github.com/kiwifs/kiwifs/pull/405
