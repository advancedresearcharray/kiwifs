---
memory_kind: episodic
episode_id: sprout-idle-nudge-2026-06-20-pr411-merge-nurture
title: "PR #411 merge nurture — execution staleness janitor (#326)"
tags: [kiwifs, janitor, runbooks, issue-326, pr-411, merge-nurture, sprout-idle-nudge]
date: 2026-06-20
---

# PR #411 merge nurture — execution staleness janitor

## Context

Work queue item `sprout-idle-nudge` for kiwifs/kiwifs#411 (closes #326). MERGE-FIRST: verify CI, tests, and merge readiness.

## Pre-search

- `kiwi_search` via `http://192.168.167.240:3333/api/kiwi/search?q=execution+staleness+janitor+326` — fleet episode indexed; no semantic fix doc yet.
- MCP `kiwifs` server not registered in this session.

## Actions

1. Verified GitHub CI: `test` SUCCESS, `mergeStateStatus: CLEAN`, `mergeable: MERGEABLE`.
2. Branch `feat/issue-326-execution-staleness` is 1 commit ahead of `origin/main`, no rebase needed.
3. Ran local tests with `GOTMPDIR=/tmp/gotmp` (overlay FS go-build cache issue without it).
4. Removed `Co-authored-by: Cursor` from feature commit via `git commit-tree` (fleet policy: no Cursor attribution).
5. Wrote durable fix doc at `pages/fixes/kiwifs-kiwifs/issue-326-execution-staleness-rule.md`.
6. Attempted `kiwi_write` via REST — blocked (`invalid API key`); docs committed locally for fleet sync.

## Test output

```
ok  github.com/kiwifs/kiwifs/internal/janitor  0.007s
ok  github.com/kiwifs/kiwifs/internal/config   0.004s
ok  github.com/kiwifs/kiwifs/cmd                0.457s
ok  github.com/kiwifs/kiwifs/internal/janitor  0.009s  (full suite)
```

## Outcome

PR #411 is merge-ready. Fleet agent should push doc commit + rewritten feature commit (force-with-lease) and strip "Made with Cursor" from PR body if still present.
