---
memory_kind: episodic
episode_id: cursor-issue-339-2026-06-15
title: "Issue #339 — implement DQL FLATTEN dot notation for nested arrays"
tags: [kiwifs, dql, flatten, event-log, issue-339, cursor]
date: 2026-06-15
---

## Task

Implement kiwifs/kiwifs#339: `FLATTEN <field>` with dot notation for querying nested JSON array objects (event log entries).

## Investigation

- Parser and basic `FLATTEN tags` already existed in `internal/dataview/`.
- Reproduced failure: `TABLE entries.event_type ... FLATTEN entries` returned `malformed JSON` when test data included files with missing or scalar `entries`.
- Root cause: compiler mapped only exact flatten field to `_flat.value`; subfields used frontmatter path; no array-type guard.

## Changes

- Added `flattenFieldSQL`, `usesFlattenSubfields`, `exprUsesFlattenSubfield` in `compiler.go`.
- Added array/object type guards in `writeWhere`.
- Added regression tests in `flatten_nested_test.go`.

## Verification

`go test ./internal/dataview/... -run Flatten` — all pass.
`go test ./internal/dataview/...` — full package pass.

## Notes

- Code edited in `/tmp/kiwifs-overlay/kiwifs-git` (writable checkout); git shared with overlay worktree at `mnt`.
- Kiwi MCP gateway unavailable in session; fix doc written to `pages/fixes/kiwifs-kiwifs/issue-339-dql-flatten-nested-arrays.md`.
