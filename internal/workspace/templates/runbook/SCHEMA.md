# Schema — Runbooks

_Template version: 3.0 (UC-6)_

Operational runbooks in the DevHelm 7-section format. Frontmatter is validated
by `.kiwi/schemas/runbook.json`. Blank runbooks start from
`.kiwi/templates/runbook.md`.

## Directory Structure

    example-high-cpu.md   Reference runbook with all seven sections
    index.md              Table of contents and severity guide
    SCHEMA.md             This file — structure and conventions
    .kiwi/
      schemas/runbook.json    Runbook frontmatter validation
      templates/runbook.md    Blank 7-section runbook template
      config.toml             Workspace defaults (execution staleness, links)

## Runbook Body Sections

Every runbook file with `type: runbook` should include these sections (see
`example-high-cpu.md` and `.kiwi/templates/runbook.md`):

1. **Trigger / When to Use** — alerts, thresholds, or symptoms
2. **Diagnosis** — commands in fenced code blocks with expected output
3. **Mitigation** — ordered, idempotent steps (safe to re-run)
4. **Verification** — concrete success criteria and check commands
5. **Rollback** — how to undo mitigation if it worsens the incident
6. **RTO and Data Loss Expectations** — recovery targets in a table
7. **Escalation Path** — who to page and when

## Frontmatter Fields

Every runbook should have YAML frontmatter. Required fields marked *.
Validated by `.kiwi/schemas/runbook.json`.

| Field                 | Type       | Required | Values / Notes                              |
|-----------------------|------------|----------|---------------------------------------------|
| type                  | string     | *        | Always `runbook`                            |
| title                 | string     | *        | Human-readable runbook title                |
| trigger               | string     | *        | Alert or condition that invokes this runbook |
| severity              | string     | *        | `P1` · `P2` · `P3` · `P4`                  |
| owner                 | string     | *        | Team or on-call rotation responsible        |
| services              | string[]   | *        | Affected services (wiki-link strings)       |
| status                | string     |          | `draft` · `active` · `deprecated`           |
| tags                  | string[]   |          | Topic tags for search and DQL               |
| last_executed         | date       |          | ISO 8601 date of last drill or live run      |
| last_outcome          | string     |          | `success` · `failure` · `partial`           |
| execution_count       | integer    |          | Number of times executed                    |
| avg_resolution_time   | string     |          | Typical mitigation duration (e.g. `20m`)    |
| reviewed              | date       |          | Last accuracy review                        |
| next-review           | date       |          | Scheduled next review                       |

## Severity Levels

| Level | Criteria | Response Time | Escalation |
|-------|----------|---------------|------------|
| **P1** | Full outage, all users affected | Immediate, all hands | VP/Director within 15 min |
| **P2** | Major feature broken, many users | < 30 minutes | Team lead within 30 min |
| **P3** | Minor feature broken, workaround exists | < 4 hours | During business hours |
| **P4** | Cosmetic or low-impact | Next business day | No escalation |

## Execution Staleness

When `[janitor.execution_staleness]` is configured in `.kiwi/config.toml`,
runbooks with `last_executed` older than `max_age_days` or `last_outcome:
failure` are flagged by `kiwifs check` and `kiwifs janitor`.

After every live run or game day:

1. Append execution notes to the runbook body (or use `POST /api/kiwi/file/append`)
2. Update `last_executed`, `last_outcome`, and increment `execution_count`

## Conventions

- One runbook per file; name files by symptom or alert (e.g. `high-cpu.md`).
- Commands must be copy-pasteable in fenced code blocks with expected output.
- Mitigation steps are **ordered and idempotent** — agents can re-run safely.
- Link affected services in `services` using wiki-link syntax: `"[[api-service]]"`.
- Query runbooks by service:
  ```
  TABLE title, severity, last_outcome FROM "." WHERE type = "runbook" AND services CONTAINS "[[api-service]]"
  ```

## Operations

See `.kiwi/playbook.md` for MCP tool sequences during incidents.
