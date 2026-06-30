---
memory_kind: episodic
episode_id: sprout-idle-nudge-2026-06-30-issue-325
title: "Issue #325 — runbook embed filter fixes kiwifs check regression"
tags: [kiwifs, runbooks, issue-325, embed-filter, sprout-idle-nudge, uc-6]
date: 2026-06-30
---

# Issue #325 — runbook embed filter

## Context

Autonomous pickup of kiwifs/kiwifs#325. UC-6 runbook scaffold existed on `main` but
`TestRunbookTemplateLintClean` and `TestRunbookInitCheckPasses` failed because legacy
`incidents/`, `postmortems/`, and `procedures/` paths were still embedded via
`go:embed all:templates` and copied into generated workspaces. Placeholder wiki links
in `postmortems/template.md` triggered broken-link errors.

## Pre-search

- Kiwi depot at `192.168.167.240:3333` unreachable from this host.
- Read local fix docs: `pages/fixes/kiwifs-kiwifs/issue-325-runbook-init-template.md`,
  `pages/fixes/kiwifs-kiwifs/issue-325-runbook-embed-filter.md`.

## Changes

1. Added `internal/workspace/embed_filter.go` — `filteredTemplatesFS` hides legacy runbook
   and dev-only embed paths from `Open`, `ReadFile`, and `ReadDir`.
2. Updated `internal/workspace/init.go` — wrap `templatesRaw` with filtered FS.
3. Added `internal/workspace/embed_filter_test.go` — path filter unit tests.
4. Added `TestRunbookEmbedUsesUC6ScaffoldOnly` in `runbook_template_test.go`.
5. Added service stubs `templates/runbook/services/api-service.md` and `monitoring.md`.
6. Dropped erroneous commit that re-added legacy template dirs to git; extended filter
   for dev-only `knowledge/`, `research/experiments/`, `research/literature/` paths.

## Verification

```bash
go test ./internal/workspace/... -run 'Runbook|Embed|Filtered' -count=1  # PASS
go test ./cmd/... -run 'Runbook|InitCheck' -count=1                      # PASS
```

`TestRunbookInitCheckPasses` now exits 0 (info-level orphans only, no errors).

## Outcome

Branch `feat/issue-325-runbook-embed-filter` ready for fleet publish. Closes #325.
