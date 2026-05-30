# Schema — Runbook

_Template version: 2.0_

Operational knowledge for on-call and platform teams.

## Directory Structure

    incidents/           One file per incident, from template
    procedures/          Reusable operational procedures
    postmortems/         Post-incident reviews (blameless)
    index.md             Table of contents and severity guide
    SCHEMA.md            This file — structure and conventions

## Frontmatter Fields

Every `.md` file should have YAML frontmatter. Required fields marked *.

### Incidents (`incidents/*.md`)

| Field              | Type       | Required | Values / Notes                              |
|--------------------|------------|----------|---------------------------------------------|
| title              | string     | *        | Short incident description                  |
| date               | date       | *        | ISO 8601 date of the incident               |
| severity           | string     | *        | `P1` · `P2` · `P3` · `P4`                  |
| status             | string     | *        | `active` · `mitigated` · `resolved`         |
| on-call            | string     |          | On-call person who responded                |
| related-alert      | string     |          | Alert name or ID that triggered this        |
| detection-minutes  | integer    |          | Minutes from incident start to detection    |
| mitigation-minutes | integer    |          | Minutes from detection to mitigation        |
| resolution-minutes | integer    |          | Minutes from detection to full resolution   |
| users-affected     | string     |          | Count or percentage of users impacted       |
| error-budget-impact| string     |          | SLO/error budget consumed (e.g., "15% of monthly") |
| tags               | string[]   |          | Service and area tags                       |
| postmortem         | string     |          | Path to linked postmortem                   |

### Procedures (`procedures/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Procedure name                              |
| tags            | string[]   |          | Service and area tags                       |
| status          | string     |          | `active` · `draft` · `deprecated`           |
| last-reviewed   | date       |          | ISO 8601 date of last review                |
| last-tested     | date       |          | ISO 8601 date of last game day / drill      |
| test-cadence    | string     |          | How often to test: `monthly` · `quarterly` · `per-incident` |
| estimated-time  | string     |          | How long this procedure typically takes     |

### Postmortems (`postmortems/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Postmortem title                            |
| date            | date       | *        | Date of the postmortem                      |
| incident        | string     |          | Path to linked incident                     |
| severity        | string     |          | Incident severity for cross-referencing     |
| authors         | string[]   |          | Who wrote this postmortem                   |
| duration-minutes| integer    |          | Total incident duration                     |
| tags            | string[]   |          | Service and area tags                       |

## Severity Levels

| Level | Criteria | Response Time | Escalation |
|-------|----------|---------------|------------|
| **P1** | Full outage, all users affected | Immediate, all hands | VP/Director within 15 min |
| **P2** | Major feature broken, many users | < 30 minutes | Team lead within 30 min |
| **P3** | Minor feature broken, workaround exists | < 4 hours | During business hours |
| **P4** | Cosmetic or low-impact | Next business day | No escalation |

## Escalation Matrix

Define your team's escalation paths in `escalation-matrix.md`:

| Severity | Primary | Secondary | Executive |
|----------|---------|-----------|-----------|
| P1 | On-call engineer | Team lead → Engineering Manager | VP Engineering |
| P2 | On-call engineer | Team lead | Engineering Manager (if > 1hr) |
| P3 | Assigned engineer | Team lead (if > 4hr) | — |
| P4 | Backlog | — | — |

## Game Days

Procedures should be tested regularly via game days (chaos engineering drills):

- Every procedure should have a `last-tested` date
- P1/P2 procedures: test quarterly at minimum
- P3/P4 procedures: test semi-annually
- After any incident where a procedure failed, test the fixed version within 2 weeks

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.

## Conventions

- Every incident gets a file `incidents/YYYY-MM-DD-<slug>.md` copied from
  `incidents/template.md`.
- Every procedure lives in `procedures/` and is linkable from
  on-call playbooks via `[[procedure-name]]`.
- Postmortems are **blameless** — focus on systems, not individuals.
- Postmortems live in `postmortems/` and link back to the incident.
- Procedures include exact commands with expected output — no vague descriptions.
- Every incident should have a linked postmortem, even if brief.
- Runbooks should be tested during game days and reviewed after incidents.
- Action items in postmortems have a named person (not team), a deadline, and a ticket.
