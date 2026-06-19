---
title: Architecture Decision Records
kiwi-view: true
query: "TABLE adr_number AS \"#\", title AS Title, status AS Status, date AS Date, deciders AS Deciders WHERE type = \"adr\" SORT adr_number ASC"
---

# Architecture Decision Records

Significant technical and process decisions, recorded in [MADR](https://adr.github.io/madr/)
format with a enforced status lifecycle.

## Active Decisions

Accepted ADRs currently in effect:

```dql
TABLE adr_number AS "#", title AS Title, domain AS Domain, date AS Date
WHERE type = "adr" AND status = "accepted"
SORT adr_number ASC
```

## Proposed

Decisions awaiting review:

```dql
TABLE adr_number AS "#", title AS Title, deciders AS Deciders, date AS Date
WHERE type = "adr" AND status = "proposed"
SORT date DESC
```

## Decision Log

| # | Date | Decision | Status |
|---|------|----------|--------|
| 1 | 2026-06-19 | [[decisions/ADR-001-use-markdown-for-adrs|Use Markdown for ADRs]] | accepted |

_Add new rows above this line when creating ADRs. Use `kiwi_workflow_advance` to
move decisions through `proposed → accepted → deprecated → superseded`._

## Workflow

1. **Propose** — copy `.kiwi/templates/adr.md` to `decisions/ADR-NNN-slug.md` with `status: proposed`
2. **Review** — discuss in PR or meeting; record deciders in frontmatter
3. **Accept** — `kiwi_workflow_advance(path, workflow: "adr", target_state: "accepted")`
4. **Supersede** — create a new ADR with `supersedes:` pointing to the old one; advance the old ADR to `superseded`

See `.kiwi/playbook.md` for MCP tool sequences.
