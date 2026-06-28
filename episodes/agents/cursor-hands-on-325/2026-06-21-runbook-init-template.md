---
memory_kind: episodic
episode_id: cursor-hands-on-325-2026-06-21
title: "Issue #325 — runbook init template verification and check regression test"
tags: [kiwifs, runbooks, issue-325, init-template, uc-6, regression-test]
date: 2026-06-21
---

# Issue #325 — runbook init template

## Context

Autonomous pickup of kiwifs/kiwifs#325 on branch `feat/issue-325-runbook-init-template`.
UC-6 runbook init template was implemented in prior commits; this session verified
acceptance criteria and added a cmd-level regression test.

## Pre-search

- `kiwi_search` via `http://192.168.167.240:3333/api/kiwi/search?q=runbook+init+template+325`
  — fix doc and fleet episodes indexed.
- Read `pages/fixes/kiwifs-kiwifs/issue-325-runbook-init-template.md`.

## Changes

1. Added `TestRunbookInitCheckPasses` in `cmd/check_test.go` — initializes runbook
   scaffold via `workspace.Init`, runs `runCheckWithCode`, asserts exit 0 and no
   error-severity janitor issues.
2. Updated fix doc verified date to 2026-06-21.

## Verification

```bash
go test ./cmd/... -run TestRunbookInitCheckPasses -count=1   # PASS
go test ./internal/workspace/... ./cmd/... -run 'Runbook|runbook' -count=1  # PASS
go test ./... -count=1  # PASS (~62s)
go run . init --root /tmp/runbooks --template runbook && go run . check --root /tmp/runbooks  # exit 0
```

## Acceptance criteria

| Criterion | Status |
|-----------|--------|
| `kiwifs init --template runbook` scaffolds workspace | PASS |
| Example runbook has 7 sections + fenced code blocks | PASS |
| JSON Schema validates required frontmatter | PASS |
| `kiwifs check` passes on generated scaffold | PASS |

## Outcome

Ready for fleet publish: push branch, open PR closing #325.
