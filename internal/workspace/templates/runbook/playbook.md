# Agent Playbook — Runbooks

Operational runbooks for on-call and incident response. When connected via MCP,
use these operations to read, execute, and maintain runbooks in the DevHelm
7-section format.

## Quick Start

1. Call `kiwi_context` to load this playbook, SCHEMA.md, and index.md
2. Call `kiwi_tree` to list runbook files
3. Use the operations below during incidents

## Execute a Runbook

When an alert fires:

1. `kiwi_search` for the alert name or symptom (e.g. `high cpu api-service`)
2. `kiwi_read` the matching runbook (start with [[example-high-cpu]])
3. Work through sections in order:
   - **Diagnosis** — run commands, compare to expected output
   - **Mitigation** — execute idempotent steps; re-run is safe
   - **Verification** — confirm all checkboxes and health commands pass
4. `kiwi_append` execution notes under a dated heading
5. `kiwi_write` frontmatter updates (`last_executed`, `last_outcome`, `execution_count`)

Example frontmatter patch after a successful run:

```yaml
last_executed: 2026-06-20
last_outcome: success
execution_count: 4
```

Prefer `PATCH /api/kiwi/file?merge=frontmatter` during live incidents to avoid
body edit conflicts.

## Create a Runbook

1. `kiwi_read` `.kiwi/templates/runbook.md` for the blank 7-section scaffold
2. `kiwi_write` to `<slug>.md` with required frontmatter:

```yaml
---
type: runbook
title: "Runbook Title"
trigger: "Alert or condition"
severity: P3
owner: team-or-oncall
services:
  - "[[service-name]]"
status: draft
tags: [operations]
---
```

3. Fill all seven sections with fenced code blocks and expected output
4. `kiwi_lint` on the new path — fix schema or link issues before marking `status: active`
5. Add a row to [[index]]

## Maintain

Run periodically or after incidents:

1. `kiwi_lint` — schema validation via `type: runbook`
2. `kiwi_query` for stale or failed runs:
   ```
   TABLE _path, title, last_executed, last_outcome
   WHERE type = "runbook" AND (last_outcome = "failure" OR last_executed < date_sub(now(), 90))
   ```
3. `kiwi_query` for runbooks by service:
   ```
   TABLE title, severity FROM "." WHERE type = "runbook" AND services CONTAINS "[[api-service]]"
   ```
4. Update `reviewed` and `next-review` after accuracy reviews
5. Run game days; set `last_executed` and `last_outcome` from drill results

## Quality Rules

- **Seven sections required** — trigger through escalation
- **Commands are executable** — fenced blocks with expected output, not prose
- **Mitigation is idempotent** — safe for agents to re-run steps
- **Frontmatter required** — `trigger`, `severity`, `owner`, `services` validated by schema
- **Services are wiki-links** — enables "runbooks for this service" queries
- **Execution metadata updated** — every live run updates `last_executed` / `last_outcome`
