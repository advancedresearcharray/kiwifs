# Schema — Architecture Decision Records

_Template version: 1.0 (UC-7)_

Numbered markdown ADRs in `decisions/` following the
[MADR](https://adr.github.io/madr/) format. Status lifecycle is enforced by
`.kiwi/workflows/adr.json`; frontmatter is validated by `.kiwi/schemas/adr.json`.

## Directory Structure

    decisions/           One file per ADR (ADR-NNN-slug.md)
    index.md             Decision log with DQL tables
    SCHEMA.md            This file — structure and conventions
    .kiwi/
      workflows/adr.json       Status state machine
      schemas/adr.json       ADR frontmatter validation
      templates/adr.md       Blank MADR template

## Status Lifecycle

ADRs use `workflow: adr` and `state` frontmatter for the workflow engine.
Keep `status` in sync with `state` — both use the same values:

| State        | Meaning                                           |
|--------------|---------------------------------------------------|
| `proposed`   | Draft under review; not yet binding                 |
| `accepted`   | Active decision the team follows                    |
| `deprecated` | No longer recommended but not explicitly replaced   |
| `superseded` | Replaced by a newer ADR (terminal state)          |

Valid transitions (enforced by `.kiwi/workflows/adr.json`):

- `proposed → accepted` · `proposed → deprecated`
- `accepted → deprecated` · `accepted → superseded`
- `deprecated → superseded`

Rejected transitions include backward steps (`accepted → proposed`), skipping
states (`proposed → superseded`), and any exit from the terminal `superseded`
state.

Advance with `kiwi_workflow_advance` or by updating `state` through the
workflow (invalid transitions are rejected).

## Frontmatter Fields

Every ADR should have YAML frontmatter. Required fields marked *.

| Field              | Type       | Required | Values / Notes                           |
|--------------------|------------|----------|------------------------------------------|
| type               | string     | *        | Always `adr`                             |
| title              | string     | *        | Human-readable title (include ADR number) |
| status             | string     | *        | `proposed` · `accepted` · `deprecated` · `superseded` |
| date               | date       | *        | ISO 8601 date the decision was made      |
| deciders           | string[]   | *        | People or teams who made the decision    |
| workflow           | string     |          | `adr` (required for workflow advance)    |
| state              | string     |          | Same enum as `status`                    |
| adr_number         | integer    |          | Auto-assigned on write to `decisions/`   |
| domain             | string     |          | Scope area (e.g. `auth`, `storage`)      |
| decision           | string     |          | One-line summary                         |
| decision-drivers   | string[]   |          | What motivated this decision             |
| tags               | string[]   |          | Topic tags, lowercase                    |
| supersedes         | string     |          | Path to the ADR this one replaces       |
| superseded_by      | string     |          | Path to the ADR that replaced this one    |
| review-by          | date       |          | Scheduled reassessment date              |

Validated by `.kiwi/schemas/adr.json`.

## MADR Body Sections

Each ADR file should include these sections (see `.kiwi/templates/adr.md`):

1. **Context and Problem Statement** — why now?
2. **Decision Drivers** — forces at play
3. **Considered Options** — alternatives with pros/cons
4. **Decision Outcome** — what was chosen and why
5. **Consequences** — positive, negative, and neutral effects

## ADR Governance

- **Immutability.** Never edit the body of an accepted ADR. Supersede with a
  new file and link via `supersedes` / `superseded_by`.
- **Sequential numbering.** `adr_number` is auto-assigned by the pipeline when
  writing to `decisions/` without an existing number.
- **PR-based review.** Propose ADRs with `status: proposed`, discuss, then
  advance to `accepted` after approval.

## DQL Examples

All accepted ADRs:

```dql
TABLE adr_number, title, domain, date
FROM "decisions/"
WHERE type = "adr" AND status = "accepted"
SORT adr_number ASC
```

ADRs by domain:

```dql
TABLE adr_number, title, status
FROM "decisions/"
WHERE type = "adr" AND domain = "auth"
SORT adr_number DESC
```

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.

### Propose
Copy `.kiwi/templates/adr.md` to `decisions/`, fill MADR sections, set
`status: proposed` and `state: proposed`.

### Accept
After review, advance: `kiwi_workflow_advance(path, workflow: "adr", target_state: "accepted")`.
Update `status` to match.

### Supersede
Create new ADR with `supersedes: decisions/ADR-NNN-old.md`. Advance the old
ADR to `superseded` and set `superseded_by` on the old file.

### Query
Use `kiwi_search`, `kiwi_query`, and `kiwi_backlinks` to find decisions
constraining a design area before writing code.
