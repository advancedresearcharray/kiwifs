---
memory_kind: episodic
episode_id: cursor-issue-325-2026-06-30
title: "Issue #325 — runbook init template embed filter hardening"
tags: [kiwifs, runbooks, issue-325, init-template, embed-filter, uc-6]
date: 2026-06-30
---

# Issue #325 — runbook init template embed filter hardening

## Context

Hands-on delivery for kiwifs/kiwifs#325 on branch `feat/issue-325-runbook-init-template-clean`.
UC-6 runbook scaffold already shipped on main (PR #459 monorepo import) but issue
remains open. Added embed-filter hardening and service stubs so `kiwifs init --template runbook`
stays lint-clean when legacy dev paths exist on disk.

## Pre-search

- Local fix docs: `pages/fixes/kiwifs-kiwifs/issue-325-runbook-init-template.md`,
  `pages/fixes/kiwifs-kiwifs/issue-325-runbook-embed-filter.md`
- Kiwi MCP gateway unavailable; cluster HTTP search failed (connection refused)

## Root cause

`//go:embed all:templates` can ship local dev scaffolds (legacy runbook `incidents/`,
`postmortems/`, `procedures/`, superseded `knowledge/`, dev-only research paths) that
contain placeholder wiki links. Those break `schema.Lint` and fail the acceptance
criterion that `kiwifs check` passes on generated workspaces.

## Changes

1. `embed_filter.go` — wrap embedded FS to exclude dev-only template paths
2. `embed_filter_test.go` — path filter unit tests
3. `init.go` — use filtered FS for template copy
4. Service stubs `services/api-service.md`, `services/monitoring.md` for example runbook links
5. Regression tests updated for scaffold paths and embed visibility

## Verification

```bash
go test ./internal/workspace/... ./cmd/... -run 'Runbook|runbook|Embed|Filtered' -count=1 -p 1  # PASS
go test ./internal/workspace/... ./cmd/... -run 'InitRunbook|InitCheck' -count=1 -p 1           # PASS
```

## Acceptance criteria

| Criterion | Status |
|-----------|--------|
| `kiwifs init --template runbook` scaffolds workspace | PASS |
| Example runbook has 7 sections + fenced code blocks | PASS |
| JSON Schema validates required frontmatter | PASS |
| `kiwifs check` passes on generated scaffold | PASS |

## Outcome

Branch ready for fleet publish (push + PR closing #325). No Cursor attribution in commits.
