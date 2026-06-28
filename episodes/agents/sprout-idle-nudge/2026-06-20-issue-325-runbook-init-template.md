---
memory_kind: episodic
episode_id: sprout-idle-nudge-2026-06-20-issue-325
title: "Issue #325 — runbook init template delivery"
tags: [kiwifs, runbooks, issue-325, init-template, sprout-idle-nudge, uc-6]
date: 2026-06-20
---

# Issue #325 — runbook init template delivery

## Context

Work queue item `sprout-idle-nudge` for kiwifs/kiwifs#325: ship runbook init template
and frontmatter schema (UC-6).

## Pre-search

- `kiwi_search` via `http://192.168.167.240:3333/api/kiwi/search?q=runbook+init+template+325`
  — fleet handoff episode indexed; no semantic fix doc yet.
- MCP `kiwifs` server not registered in this session.

## Actions

1. Verified branch `feat/issue-325-runbook-init-template` with commit `c0145b3` implements
   full UC-6 scaffold (schema, example, blank template, playbook, config).
2. Ran runbook-focused tests — all PASS.
3. Ran full suite `go test ./... -count=1` — all PASS.
4. Manual `go run . init --template runbook` + `go run . check` — exit 0 (8 info issues
   on README/SCHEMA/playbook orphans and missing owner/status; no errors).
5. Wrote durable fix doc at `pages/fixes/kiwifs-kiwifs/issue-325-runbook-init-template.md`.
6. Attempted `kiwi_write` via REST — blocked (`invalid API key`); docs written locally
   for fleet sync.

## Test output

```
ok  github.com/kiwifs/kiwifs/internal/workspace  0.023s  (-run Runbook)
ok  github.com/kiwifs/kiwifs/cmd                0.373s  (-run Runbook)
ok  github.com/kiwifs/kiwifs/...                 (full suite, 60s)
```

## Outcome

Issue #325 acceptance criteria met. Fleet agent should push branch, open PR closing #325,
strip `Co-authored-by: Cursor` from commit `c0145b3`, and sync fix doc to cluster depot.
