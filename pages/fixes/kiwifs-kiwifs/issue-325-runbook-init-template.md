---
memory_kind: semantic
doc_id: kiwifs-kiwifs-issue-325-runbook-init-template
title: Runbook init template with 7-section schema
tags: [kiwifs, workspace, runbooks, issue-325, init-template, uc-6, devhelm]
repo: kiwifs/kiwifs
issue_number: 325
languages: [go, markdown, json]
status: verified
peer_review: pass
date: 2026-06-21
verified: 2026-06-21T15:42:00Z
delivery_commit: 2724b00e162490a9b7440546a7f5a950c8f5cf33
ci_run: 27909055535
---

## Problem

UC-6 (Runbooks) required `kiwifs init --template runbook` to scaffold an operational
runbook workspace. Users needed a JSON Schema for frontmatter (`trigger`, `severity`,
`owner`, `services`), a worked example in the DevHelm 7-section format, a blank template,
and regression tests so `kiwifs check` passes on the generated scaffold.

The prior runbook template used a different layout (`incidents/`, `procedures/`,
`postmortems/`) without schema validation or the UC-6 section structure.

## Root cause

The embedded runbook template predated UC-6 conventions. It lacked:

1. `.kiwi/schemas/runbook.json` with required frontmatter fields
2. A single reference runbook (`example-high-cpu.md`) demonstrating all seven sections
3. Registration in `cmd/init.go` help text and examples
4. Regression tests for schema validation, scaffold paths, and lint cleanliness

## Solution

Replace the legacy runbook scaffold with UC-6 DevHelm format:

1. **Template files** under `internal/workspace/templates/runbook/`:
   - `SCHEMA.md` — structure, frontmatter table, severity guide, execution staleness
   - `index.md` — table of contents with DQL query for active runbooks
   - `example-high-cpu.md` — full 7-section example with fenced bash blocks and expected output
   - `.kiwi/schemas/runbook.json` — validates `type`, `title`, `trigger`, `severity`, `owner`, `services`
   - `.kiwi/templates/runbook.md` — blank 7-section scaffold
   - `.kiwi/config.toml` — execution staleness janitor, typed `services` links, auth guidance
   - `playbook.md` — MCP agent operations for execute/create/maintain

2. **Registration** — `runbook` already in `internal/workspace/init.go` switch; added to
   `cmd/init.go` flag help and example.

3. **Removed legacy paths** — `incidents/`, `postmortems/`, `procedures/` subdirs replaced
   by flat runbook files and `.kiwi/` schema/template.

## Files changed

- `internal/workspace/templates/runbook/**` — new UC-6 scaffold
- `internal/workspace/runbook_template_test.go` — schema, scaffold, lint, metadata tests
- `internal/workspace/init_test.go` — include `runbook` in `ListInitTemplates` assertion
- `cmd/init.go` — flag help + example for `--template runbook`
- `cmd/init_test.go` — `TestRunbookTemplateEmbedded`, `TestRunbookTemplateInit`, `TestInitCmdDocumentsRunbookTemplate`, `TestRunbookTemplateInitBlankRoot`
- `cmd/check_test.go` — `TestRunbookInitCheckPasses` (acceptance: `kiwifs check` on scaffold)

## Tests

```bash
go test ./internal/workspace/... -count=1 -run 'Runbook|runbook'
go test ./cmd/... -count=1 -run 'Runbook|runbook'
go test ./... -count=1
```

Manual verification:

```bash
TMP=$(mktemp -d)
go run . init --root "$TMP/runbooks" --template runbook
go run . check --root "$TMP/runbooks"   # exit 0 (info-level orphans only)
```

## Peer review notes

**Status: pass** (2026-06-21 hands-on delivery v11, PR #418, CI run 27909055535 SUCCESS)

Verified template scaffold, schema, registration, and tests:

- `runbook.json` requires `type`, `title`, `trigger`, `severity`, `owner`, `services`
- `example-high-cpu.md` has all 7 UC-6 sections with fenced bash blocks and expected output
- `cmd/init.go` lists `runbook` in help and example; `internal/workspace/init.go` registers template
- `TestRunbookInitCheckPasses` confirms `kiwifs check` exit 0 on scaffold (info-level orphans only)
- `TestInitCmdDocumentsRunbookTemplate` guards `--template runbook` in flag help and CLI example
- `TestRunbookTemplateInitBlankRoot` hardens blank-parent init path (matches ADR/prompt peer-review pattern)
- UC-6 wiki updated: runbook init template removed from "What's Missing", milestone 1 marked shipped
- No code defects found; implementation complete

- Example runbook frontmatter must pass `runbook.json` — include at least one service in
  `services` array with wiki-link syntax.
- `TestRunbookTemplateLintClean` rejects broken-link, orphan, and empty-file lint issues
  on the scaffold; README/playbook orphans are info-only in `kiwifs check`.
- Execution staleness janitor config ships in template `.kiwi/config.toml`; pairs with
  issue #326 janitor rule implementation.
- Follow ADR/prompt template patterns for future UC init templates: SCHEMA + playbook +
  `.kiwi/schemas/*.json` + `*_template_test.go`.

## Reuse guide

When adding or updating the runbook init template:

1. Keep all seven body sections in example and blank template.
2. Update `runbook.json` required fields if UC-6 schema changes.
3. Run `TestRunbookSchemaValidatesExample` and `TestRunbookSchemaRejectsInvalidFrontmatter`
   after schema edits.
4. Ensure diagnosis/verification sections include fenced code blocks with expected output.
5. Register new optional frontmatter in both `SCHEMA.md` and `runbook.json`.
