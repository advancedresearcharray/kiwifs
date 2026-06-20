---
memory_kind: episodic
episode_id: cursor-hands-on-328-2026-06-20-workflow-advance
title: Issue #328 ADR workflow advance status sync fix
tags: [kiwifs, workspace, adr, issue-328, workflow, bugfix]
date: 2026-06-20
---

## Work item

kiwifs/kiwifs#328 — feat(workspace): ship ADR init template with workflow and schema

Issue still OPEN after PR #406 merged to main. Acceptance criterion "Workflow transitions
are enforced via `kiwi_workflow_advance`" failed in practice: advancing an ADR left
`status` stale and could wipe frontmatter.

## Root cause

1. `LocalBackend.WorkflowAdvance` rebuilt frontmatter with `yamlMarshal` (JSON-to-YAML
   shim). Arrays and complex fields produced YAML the pipeline could not round-trip.
2. On write, `auto_sequence` saw missing/corrupt frontmatter and replaced it with only
   `adr_number`, destroying `type`, `status`, `state`, `deciders`, etc.
3. Neither MCP nor REST workflow advance synced `status` with `state`, so DQL queries on
   `status = "accepted"` missed advanced ADRs.

## Fix

- Replace `yamlMarshal` rebuild in `WorkflowAdvance` with `markdown.SetFrontmatterField`.
- Add `workflow.SyncStatusOnAdvance` and call from MCP + REST advance handlers.
- Add regression tests: `internal/mcpserver/adr_workflow_test.go`, `TestSyncStatusOnAdvance`.

## Tests

```
go test ./internal/workflow/... ./internal/mcpserver/... ./internal/workspace/... ./cmd/... -count=1 -run 'SyncStatus|ADR'
go test ./... -count=1
```

All green.

## Files changed

- `internal/mcpserver/local.go`
- `internal/mcpserver/adr_workflow_test.go` (new)
- `internal/api/handlers_workflow.go`
- `internal/workflow/workflow.go`
- `internal/workflow/workflow_test.go`
- `pages/fixes/kiwifs-kiwifs/issue-328-adr-init-template.md`
