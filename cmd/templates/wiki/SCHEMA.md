# Schema — Team Wiki

Flat structure for a team replacing Confluence/Notion. Every top-level
folder is a functional area; every `index.md` is the landing page
for its folder.

## Directory Structure

    engineering/         Engineering docs, architecture, runbooks
    product/             Product specs, roadmaps
    onboarding/          New hire guides
    index.md             Wiki-wide table of contents
    SCHEMA.md            This file — structure and conventions

## Frontmatter Fields

Every `.md` file should have YAML frontmatter. Required fields marked *.

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Human-readable page title                   |
| owner           | string     |          | Team or person responsible                  |
| status          | string     |          | `active` · `draft` · `review` · `deprecated` |
| tags            | string[]   | *        | Topic tags, lowercase, hyphenated           |
| last-reviewed   | date       |          | ISO 8601 date of last quality review        |

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.

## Conventions

- Link between pages with `[[wiki links]]`.
- Keep pages short. Split when a page exceeds ~300 lines.
- Every folder has an `index.md` landing page.
- Runbooks live under `engineering/runbooks/`; specs under
  `product/specs/`.
