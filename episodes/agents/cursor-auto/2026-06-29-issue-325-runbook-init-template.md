---
memory_kind: episodic
episode_id: cursor-auto-2026-06-29-issue-325
title: Issue 325 runbook embed filter delivery
tags: [kiwifs, issue-325, runbook, workspace, embed-filter]
date: 2026-06-29
---

## Context

Hands-on takeover for kiwifs/kiwifs#325. Prior fleet agent left branch
`feat/issue-325-runbook-init-template` with a bad merge commit (`e2ae1a6`) that
broke janitor/api builds. UC-6 runbook template was already on main (PR #418).

## Root cause

Legacy runbook scaffold files under `templates/runbook/{incidents,postmortems,procedures}/`
still on disk get embedded via `//go:embed all:templates`. Placeholder wiki links in
those files cause `schema.Lint` / `kiwifs check` failures on `kiwifs init --template runbook`.

## Solution

Rebuilt branch from `origin/main` in clean worktree `/tmp/kiwifs-issue325`:

- Added `internal/workspace/embed_filter.go` (`filteredTemplatesFS`)
- Updated `init.go` to use filtered embed FS and skip legacy paths in `copyEmbedDir`
- Added `TestRunbookEmbedUsesUC6ScaffoldOnly` regression test

## Tests

```bash
go test ./internal/workspace/... -run 'Runbook|runbook' -count=1  # PASS
go test ./cmd/... -run TestRunbookInitCheckPasses -count=1        # PASS
go test ./internal/workspace/... -count=1                         # PASS
```

## Commit

`3013d4b` on `feat/issue-325-runbook-init-template-clean`, pushed to fork
`advancedresearcharray/kiwifs`. PR creation blocked (repo restricted to collaborators).
