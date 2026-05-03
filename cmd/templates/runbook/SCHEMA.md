# Schema — Runbook

Operational knowledge for on-call and platform teams.

## Directory Structure

    incidents/           One file per incident, from template
    procedures/          Reusable operational procedures
    postmortems/         Post-incident reviews
    index.md             Table of contents
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
| tags            | string[]   |          | Service and area tags                       |

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.

## Conventions

- Every incident gets a file `incidents/YYYY-MM-DD-<slug>.md` copied from
  `incidents/template.md`.
- Every procedure lives in `procedures/` and is linkable from
  on-call playbooks via `[[procedure-name]]`.
- Postmortems live in `postmortems/` and link back to the incident.
