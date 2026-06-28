# Schema — Team Wiki

_Template version: 2.0_

Flat, workflow-oriented wiki for teams replacing Confluence or Notion.
Organized around how work gets done, not org charts. Max 3 levels deep.

## Directory Structure

    welcome.md           Orientation — what this wiki is, how to contribute
    how-we-work.md       Communication norms, meetings, decision-making
    architecture.md      System overview, components, data flow
    onboarding/          New member checklist and resources
      index.md           Landing page with week-1 and week-2 checklists
    decisions/           Architecture Decision Records (MADR format)
      index.md           Decision log and ADR format guide
      adr-NNN-slug.md    Individual decisions
    processes/           Standard operating procedures
      index.md           Process index
      deployment.md      How to deploy
      dev-setup.md       Local environment setup
      incident-response.md  What to do when production breaks
    reference/           Shared resources and terminology
      index.md           Reference landing page
      glossary.md        Team vocabulary
      faq.md             Frequently asked questions
    index.md             Wiki-wide table of contents
    SCHEMA.md            This file — structure and conventions

## Frontmatter Fields

Every `.md` file should have YAML frontmatter. Required fields marked *.

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Human-readable page title                   |
| owner           | string     | *        | Team or person responsible for this page     |
| status          | string     | *        | `draft` · `active` · `review` · `deprecated` |
| tags            | string[]   | *        | Topic tags, lowercase, hyphenated           |
| last-reviewed   | date       |          | ISO 8601 date of last quality review        |
| freshness-days  | integer    |          | Days before considered stale (default: 90)  |

### Additional fields for ADRs (`decisions/adr-*.md`)

ADRs follow the [MADR](https://adr.github.io/madr/) format adapted for KiwiFS.

| Field              | Type       | Required | Values / Notes                           |
|--------------------|------------|----------|------------------------------------------|
| type               | string     | *        | Always `decision`                        |
| date               | date       | *        | Date the decision was made               |
| status             | string     | *        | `proposed` · `accepted` · `superseded` · `deprecated` |
| decision           | string     | *        | One-line summary of the decision         |
| decision-drivers   | string[]   |          | What motivated this decision             |
| alternatives       | object[]   |          | Each: `option`, `pros`, `cons`           |
| impact             | string     |          | What this decision affects               |
| reversal-conditions| string     |          | When to revisit this decision            |
| review-by          | date       |          | Scheduled date to reassess (for high-stakes decisions) |
| superseded-by      | string     |          | Path to the ADR that replaced this one   |
| supersedes         | string     |          | Path to the ADR this one replaces        |
| linked-docs        | string[]   |          | Related wiki pages                       |

### Additional fields for SOPs (`processes/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| type            | string     |          | `sop`                                       |
| scope           | string     |          | Who this applies to (team, role, everyone)  |
| frequency       | string     |          | `on-demand` · `daily` · `weekly` · `per-release` · `per-incident` |
| estimated-time  | string     |          | How long the process typically takes        |
| last-tested     | date       |          | When this procedure was last validated      |

## ADR Governance

- **Immutability rule.** Never edit an accepted ADR. If the decision changes,
  create a new ADR with `supersedes: decisions/adr-NNN-old.md` and update the
  old ADR's status to `superseded` with `superseded-by: decisions/adr-NNN-new.md`.
- **Review triggers.** High-stakes decisions should set `review-by`. When that
  date passes, the ADR appears in the quarterly review.
- **PR-based workflow.** ADRs should be reviewed like code — propose via draft,
  discuss, then accept. The `status` field tracks this lifecycle.

## Quarterly Review Protocol

Every 90 days, run a maintenance pass:

1. Query pages past their freshness window:
   `kiwi_query("TABLE _path, title, owner, last-reviewed WHERE last-reviewed < date_sub(now(), 90) SORT last-reviewed ASC")`
2. Query ADRs with passed `review-by` dates.
3. Notify page owners of stale content.
4. Archive pages with `status: deprecated` and no inbound links for 180+ days.

Integrate this review into an existing ritual (sprint retro, team sync).

## Conventions

- **Workflow-first structure.** Sections map to how work gets done.
- **Max 3 levels deep.** If it needs more nesting, split into a new section.
- **Every folder has `index.md`.** The landing page for that area.
- **Link between pages** with `[[wiki links]]`.
- **Keep pages short.** Split when a page exceeds ~300 lines.
- **Descriptive titles.** "Deployment Process" not "Ship It Guide".
- **Tags are lowercase, hyphenated.** `ci-cd`, `team-norms`, `api-design`.
- **Owner accountability.** Every page has an owner who reviews it periodically.
- **Contribution norms.** Anyone can edit; critical pages (processes, ADRs) need review.

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.
