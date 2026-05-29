# Schema — Runbook

Operational knowledge for on-call and platform teams.

## Directory Structure

    incidents/           One file per incident, from template
    procedures/          Reusable operational procedures
    postmortems/         Post-incident reviews
    index.md             Table of contents and severity guide
    SCHEMA.md            This file — structure and conventions

## Frontmatter Fields

Every `.md` file should have YAML frontmatter. Required fields marked *.

### Incidents (`incidents/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Short incident description                  |
| date            | date       | *        | ISO 8601 date of the incident               |
| severity        | string     | *        | `P1` · `P2` · `P3` · `P4`                  |
| status          | string     | *        | `active` · `mitigated` · `resolved`         |
| on-call         | string     |          | On-call person who responded                |
| related-alert   | string     |          | Alert name or ID that triggered this        |
| tags            | string[]   |          | Service and area tags                       |
| postmortem      | string     |          | Path to linked postmortem                   |

### Procedures (`procedures/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Procedure name                              |
| tags            | string[]   |          | Service and area tags                       |
| status          | string     |          | `active` · `draft` · `deprecated`           |
| last-reviewed   | date       |          | ISO 8601 date of last review                |

### Postmortems (`postmortems/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Postmortem title                            |
| date            | date       | *        | Date of the postmortem                      |
| incident        | string     |          | Path to linked incident                     |
| severity        | string     |          | Incident severity for cross-referencing     |
| authors         | string[]   |          | Who wrote this postmortem                   |
| tags            | string[]   |          | Service and area tags                       |

## Severity Levels

| Level | Criteria | Response Time |
|-------|----------|---------------|
| **P1** | Full outage, all users affected | Immediate, all hands |
| **P2** | Major feature broken, many users | < 30 minutes |
| **P3** | Minor feature broken, workaround exists | < 4 hours |
| **P4** | Cosmetic or low-impact | Next business day |

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.

## Conventions

- Every incident gets a file `incidents/YYYY-MM-DD-<slug>.md` copied from
  `incidents/template.md`.
- Every procedure lives in `procedures/` and is linkable from
  on-call playbooks via `[[procedure-name]]`.
- Postmortems live in `postmortems/` and link back to the incident.
- Procedures include exact commands with expected output — no vague descriptions.
- Every incident should have a linked postmortem, even if brief.
- Runbooks should be tested during game days and reviewed after incidents.
