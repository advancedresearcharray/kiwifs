---
memory_kind: episodic
episode_id: cursor-fleet-337-2026-06-17-append-only
title: Issue 337 append_only frontmatter enforcement
tags: [kiwifs, pipeline, append-only, issue-337, uc-event-log]
date: 2026-06-17
---

# Run log — kiwifs#337 append_only enforcement

Implemented pipeline-level rejection of PUT/full overwrites when an existing file has `append_only: true` in YAML frontmatter.

## Work done
- Added `ErrAppendOnlyDenied` and `IsAppendOnly()` in `internal/pipeline/append_only.go`
- Hooked checks into `WriteWithOpts`, `WriteStream`, and `BulkWrite` (not `Append`)
- Centralized HTTP mapping in `internal/api/write_errors.go` (`pipelineWriteHTTPError`)
- Regression tests: pipeline, REST handlers, MCP `kiwi_write`

## Test results
All targeted tests passed:
`go test ./internal/pipeline/... ./internal/api/... ./internal/mcpserver/... -run 'AppendOnly|Put_Rejects|ToolHandlerWrite_Rejects'`

## Commit
`5765644` — `feat(pipeline): enforce append_only frontmatter on PUT overwrites` (local, not pushed)

## Kiwi depot note
Remote write to `http://192.168.167.240:3333` returned `invalid API key`; docs staged in-repo under `pages/fixes/` and `episodes/agents/` for fleet sync.
