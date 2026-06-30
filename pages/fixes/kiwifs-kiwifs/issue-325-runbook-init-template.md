---
memory_kind: semantic
doc_id: kiwifs-kiwifs-issue-325-runbook-init-template
title: Runbook init template with 7-section schema
tags: [kiwifs, workspace, runbooks, issue-325, init-template, uc-6, devhelm, embed-filter]
repo: kiwifs/kiwifs
issue_number: 325
languages: [go, markdown, json]
status: verified
peer_review: pass
date: 2026-06-30
verified: 2026-06-30T17:45:00Z
---

## Problem

UC-6 (Runbooks) required `kiwifs init --template runbook` to scaffold an operational
runbook workspace. Users needed a JSON Schema for frontmatter (`trigger`, `severity`,
`owner`, `services`), a worked example in the DevHelm 7-section format, a blank template,
and regression tests so `kiwifs check` passes on the generated scaffold.

Additionally, `//go:embed all:templates` could ship local dev scaffolds (legacy
`incidents/`, `postmortems/`, `procedures/`, superseded `knowledge/`) containing
placeholder wiki links that fail lint on generated workspaces.

## Root cause

1. The embedded runbook template predated UC-6 conventions — missing schema, 7-section
   example, and CLI registration.
2. Go embed includes all files on disk under `templates/`, not just git-tracked paths.
   Dev-only directories with broken placeholder links could leak into init scaffolds.

## Solution

1. **UC-6 template** under `internal/workspace/templates/runbook/`:
   - `SCHEMA.md`, `index.md`, `example-high-cpu.md` (7 sections + fenced bash blocks)
   - `.kiwi/schemas/runbook.json` — validates `type`, `title`, `trigger`, `severity`, `owner`, `services`
   - `.kiwi/templates/runbook.md` — blank 7-section scaffold
   - `.kiwi/config.toml` — execution staleness janitor, typed `services` links
   - `services/api-service.md`, `services/monitoring.md` — wiki-link targets for example runbook

2. **Registration** — `runbook` in `internal/workspace/init.go` switch; `cmd/init.go` flag help and example.

3. **Embed filter** — `filteredTemplatesFS` in `embed_filter.go` excludes dev-only paths
   from init copy while keeping UC-6 scaffold files visible.

## Files changed

- `internal/workspace/templates/runbook/**` — UC-6 scaffold + service stubs
- `internal/workspace/embed_filter.go` — filtered embedded FS
- `internal/workspace/embed_filter_test.go` — path exclusion tests
- `internal/workspace/init.go` — wrap embed FS with filter
- `internal/workspace/runbook_template_test.go` — schema, scaffold, lint, embed tests
- `internal/workspace/init_test.go` — runbook in `ListInitTemplates`, embedded paths
- `cmd/init.go` — flag help + example for `--template runbook`
- `cmd/init_test.go` — embedded, init, CLI help, blank-root tests
- `cmd/check_test.go` — `TestRunbookInitCheckPasses`

## Tests

```bash
go test ./internal/workspace/... -count=1 -run 'Runbook|Embed|Filtered'
go test ./cmd/... -count=1 -run 'Runbook|InitCheck'
```

Manual verification (requires `ui/dist` for full binary build):

```bash
TMP=$(mktemp -d)
go run . init --root "$TMP/runbooks" --template runbook
go run . check --root "$TMP/runbooks"   # exit 0 (info-level orphans only)
```

## Peer review notes

- `runbook.json` requires `type`, `title`, `trigger`, `severity`, `owner`, `services`
- `example-high-cpu.md` has all 7 UC-6 sections with fenced bash blocks and expected output
- `TestRunbookInitCheckPasses` confirms `kiwifs check` exit 0 on scaffold
- `TestRunbookEmbedUsesUC6ScaffoldOnly` guards against legacy path regression
- Embed filter is transparent to tracked UC-6 files; dev-only paths remain on disk but do not ship

## Reuse guide

When adding or updating the runbook init template:

1. Keep all seven body sections in example and blank template.
2. Update `runbook.json` required fields if UC-6 schema changes.
3. Run `TestRunbookSchemaValidatesExample` and `TestRunbookSchemaRejectsInvalidFrontmatter` after schema edits.
4. Add dev-only template dirs to `excludedEmbedDirs` in `embed_filter.go` if they contain placeholder links.
5. Run `TestFilteredTemplatesFSHidesLegacyRunbookPaths` after embed filter changes.
