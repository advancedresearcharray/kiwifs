---
memory_kind: semantic
doc_id: kiwifs-kiwifs-issue-337-append-only-frontmatter
title: Enforce append_only frontmatter on PUT overwrites
tags: [kiwifs, pipeline, append_only, frontmatter, PUT, 409]
repo: kiwifs/kiwifs
issue_number: 337
languages: [go]
status: resolved
date: 2026-06-20
---

# Enforce `append_only` frontmatter on PUT overwrites

## Problem

Markdown files with `append_only: true` in YAML frontmatter (event logs, audit trails) could be fully overwritten via PUT, bulk write, WriteStream, or frontmatter PATCH. Only append should be allowed once the file exists.

## Root cause

The pipeline treated all writes as full replacements. Config-driven `[[validate_write]]` rules could block overwrites when configured, but enforcement was not built into the pipeline — workspaces without that config remained vulnerable.

## Solution

Add hardcoded pipeline guards that read existing file frontmatter before any PUT-class write:

1. **`internal/pipeline/append_only.go`** — `ErrAppendOnly`, `isAppendOnly()`, `rejectAppendOnlyOverwrite()`, `rejectAppendOnlyBulkOverwrite()` (also rejects duplicate paths in one bulk batch when the first entry is append-only).
2. **`WriteWithOpts` / `WriteStream` / `BulkWrite`** — call reject helpers under `writeMu` (TOCTOU-safe with concurrent writers).
3. **`Append`** — unchanged; appends still allowed.
4. **First write** — creating a new file with `append_only: true` is allowed (no existing file to protect).
5. **API** — map `ErrAppendOnly` to HTTP 409 Conflict in PUT, bulk, and frontmatter PATCH handlers.

Coexists with config `[[validate_write]]` rules; hardcoded check runs first and returns `ErrAppendOnly` (409). Config rules still handle other reject types (`body_change`, custom frontmatter).

## Files changed

| File | Change |
|------|--------|
| `internal/pipeline/append_only.go` | New — detection and rejection helpers |
| `internal/pipeline/pipeline.go` | Hooks in WriteWithOpts, WriteStream, BulkWrite |
| `internal/api/handlers_file.go` | ErrAppendOnly → 409 |
| `internal/pipeline/append_only_test.go` | Pipeline unit tests |
| `internal/api/handlers_append_only_test.go` | REST integration tests |
| `internal/mcpserver/mcpserver_test.go` | MCP kiwi_write rejection test |
| `internal/pipeline/validate_test.go` | Integration test expects ErrAppendOnly |

## Tests

```bash
go test ./internal/pipeline/... -run AppendOnly
go test ./internal/api/... -run AppendOnly
go test ./internal/mcpserver/... -run AppendOnly
go test ./internal/pipeline/... -run ValidateWrite
go test ./internal/...
```

## Peer review notes

- Rebased onto main preserving `ValidateWrite(ctx, path, content, WriteKind)` and `validate.go` — earlier PR branch had regressed that API.
- Bulk duplicate-path guard closes bypass where two entries for the same path could overwrite an append-only first entry.
- Checks run under `writeMu` to avoid TOCTOU with concurrent PUTs.
- `isAppendOnly` accepts bool `true`, string `"true"` (case-insensitive), and string `"1"` — pipeline tests cover string forms.
- Frontmatter PATCH routes through `WriteWithOpts`, so append-only files reject field updates with 409.
- MCP `kiwi_write` uses pipeline `Write` → guarded `WriteWithOpts`.
- **Overlay FS pitfall:** on `/tmp/kiwifs-overlay/mnt`, use `GIT_INDEX_FILE=/tmp/kiwifs-overlay/upper/.git/index` when index writes fail; remove bleed-through untracked files before testing. A stale-index merge deleted `append_only.go` (v12, 2026-06-20).

## Reuse guide

When adding a new full-file write path in the pipeline:

1. Acquire `writeMu` before reading existing content.
2. Call `rejectAppendOnlyOverwrite(ctx, path)` for single-file writes.
3. Call `rejectAppendOnlyBulkOverwrite(ctx, files)` for batch writes.
4. Map `errors.Is(err, pipeline.ErrAppendOnly)` to HTTP 409 in API handlers.

Detection accepts `append_only: true` (bool) or string `"true"` / `"1"`.
