# Schema — Research Library

_Template version: 3.0 (UC-9)_

Paper tracking, citation linking, and a reading workflow for literature
reviews. One file per paper in `papers/`, synthesis in `notes/`, and
literature review drafts in `reviews/`.

## Directory Structure

    papers/              One file per paper (metadata + reading notes)
    notes/               Synthesis notes linking multiple papers
    reviews/             Literature review drafts
    index.md             Library overview with DQL tables
    SCHEMA.md            This file — structure and conventions
    .kiwi/
      workflows/reading.json   Reading state machine
      schemas/paper.json       Paper frontmatter validation

## Reading Workflow

Papers use `workflow: reading` and `state` frontmatter. Valid states:

| State          | Meaning                                      |
|----------------|----------------------------------------------|
| `unread`       | Discovered but not started                   |
| `reading`      | Actively reading                             |
| `annotated`    | Marginal notes and highlights captured       |
| `summarized`   | Key findings written up                      |
| `incorporated` | Insights merged into notes or reviews        |

Transitions are enforced by `.kiwi/workflows/reading.json`:
`unread → reading → annotated → summarized → incorporated`.

Advance with `kiwi_workflow_advance` or by updating `state` through
the pipeline (invalid transitions are rejected).

## Frontmatter Fields

Every `.md` file should have YAML frontmatter. Required fields marked *.

### Papers (`papers/*.md`)

| Field        | Type       | Required | Values / Notes                              |
|--------------|------------|----------|---------------------------------------------|
| type         | string     | *        | `paper`                                     |
| title        | string     | *        | Paper title                                 |
| authors      | string[]   | *        | List of author names                        |
| year         | number     | *        | Publication year                            |
| venue        | string     | *        | Journal, conference, or preprint server     |
| doi          | string     |          | DOI for citation and retrieval              |
| bibtex_key   | string     |          | BibTeX citation key                         |
| abstract     | string     |          | Short abstract for quick recall             |
| tags         | string[]   |          | Topic and method tags                       |
| cites        | string[]   |          | Wikilinks to related papers, e.g. `[[other-paper]]` |
| workflow     | string     | *        | `reading`                                   |
| state        | string     | *        | `unread` · `reading` · `annotated` · `summarized` · `incorporated` |

Validated by `.kiwi/schemas/paper.json`.

### Notes (`notes/*.md`)

| Field   | Type       | Required | Values / Notes                              |
|---------|------------|----------|---------------------------------------------|
| title   | string     | *        | Note title                                  |
| type    | string     |          | `synthesis` · `brainstorm` · `methodology`  |
| date    | date       |          | ISO 8601 date                               |
| status  | string     |          | `draft` · `active` · `archived`             |
| tags    | string[]   |          | Topic tags                                  |
| related | string[]   |          | Paths or wikilinks to related papers        |

### Reviews (`reviews/*.md`)

| Field   | Type       | Required | Values / Notes                              |
|---------|------------|----------|---------------------------------------------|
| title   | string     | *        | Review title                                |
| status  | string     |          | `draft` · `in_review` · `published`         |
| scope   | string     |          | Brief description of review scope           |
| tags    | string[]   |          | Topic tags                                  |
| papers  | string[]   |          | Papers included in this review              |

## Citation Conventions

- **One paper per file** in `papers/`, named `<author-or-topic-slug>.md`.
- **Cross-cite with `cites`** — use wikilink syntax: `cites: ["[[transformer-survey]]"]`.
  KiwiFS indexes `cites` as typed backlinks when configured in `.kiwi/config.toml`.
- **Include DOI and bibtex_key** wherever available for export and bibliography tools.
- **Link synthesis to sources** — notes should reference papers with `[[wikilinks]]`
  and list them in `related`.
- **Reviews aggregate papers** — use the `papers` frontmatter field and body sections
  per theme or research question.

## DQL Examples

Unread papers by year:

```dql
TABLE _path AS Path, title AS Title, year AS Year, venue AS Venue
WHERE type = "paper" AND state = "unread"
SORT year DESC
```

Papers currently being read:

```dql
TABLE title AS Title, authors AS Authors, tags AS Tags
WHERE type = "paper" AND state = "reading"
SORT _updated DESC
```

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.

## Conventions

- Start new papers at `state: unread`.
- Advance through the reading workflow as you annotate and summarize.
- Write synthesis in `notes/` once multiple papers are `summarized` or `incorporated`.
- Draft literature reviews in `reviews/` when ready to publish findings.
- Keep `index.md` tables in sync or use `kiwi-view: true` with embedded DQL queries.
