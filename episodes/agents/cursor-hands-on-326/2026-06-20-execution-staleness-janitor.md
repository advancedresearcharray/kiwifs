---
memory_kind: episodic
episode_id: cursor-hands-on-326-2026-06-20-execution-staleness-janitor
title: "Issue #326 — janitor execution staleness rule for runbooks"
tags: [kiwifs, janitor, runbooks, issue-326, peer-review]
date: 2026-06-20
---

# Issue #326 — janitor execution staleness rule

## Context

Hands-on takeover after fleet engineer delivery failed peer review (missing docs + edge-case tests).

## Pre-search

- No prior fix doc on cluster for execution staleness.
- Prior commit `0fc9755` implemented core rule; mixed unrelated ADR test changes from #328.

## Peer review fixes

1. **Documentation:** config struct comments, template `config.toml` example, `kiwifs janitor` help text, `wiki/UC-6-Runbooks.md` configuration section.
2. **Tests:** 10 additional janitor tests (multiple flags, custom date field, missing/invalid dates, RFC3339, defaults, dual flags), config disabled-by-default test, check integration for stale custom date field.

## Verification

```bash
go test ./internal/janitor/... ./internal/config/... ./cmd/... -count=1 -run 'ExecutionStaleness|JanitorExecution|RunKnowledgeScan_Execution'
go test ./internal/janitor/... -count=1
```

## Outcome

Clean PR branch `feat/issue-326-execution-staleness` from `origin/main` with only #326 scope. Fix doc: `pages/fixes/kiwifs-kiwifs/issue-326-execution-staleness-rule.md`.
