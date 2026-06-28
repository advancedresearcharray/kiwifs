---
type: adr
title: "ADR-001: Use Markdown for Architecture Decision Records"
adr_number: 1
status: accepted
date: 2026-06-19
deciders: [engineering-team]
workflow: adr
state: accepted
domain: documentation
decision: Store architecture decisions as numbered markdown files with YAML frontmatter
decision-drivers:
  - Decisions must live in version control next to the code they affect
  - Agents and humans need searchable, machine-readable metadata
  - Industry practice (MADR, Nygard) uses plain markdown ADRs
tags: [adr, documentation, meta]
review-by: 2027-06-19
---

# ADR-001: Use Markdown for Architecture Decision Records

> **Self-referential example.** This ADR documents why this workspace uses
> markdown ADRs. Replace or supersede it when your team adopts different conventions.

## Status

Accepted — 2026-06-19

## Context and Problem Statement

We need a durable record of significant technical decisions — the alternatives
considered, the rationale, and the consequences. Code shows *what* was built;
ADRs capture *why*.

Plain text in wikis or chat threads is hard to query, lacks lifecycle tracking,
and drifts out of date. We want decisions indexed, validated, and navigable
through the same KiwiFS tools we use for the rest of the knowledge base.

## Decision Drivers

- Version control: every decision change is auditable via git history
- Agent-queryable: MCP tools can search decisions before proposing architecture
- Low ceremony: markdown + frontmatter, no proprietary formats
- Workflow enforcement: status lifecycle (`proposed → accepted → deprecated → superseded`)

## Considered Options

| Option | Pros | Cons |
|--------|------|------|
| Markdown ADRs in-repo (MADR) | Git-native, searchable, agent-friendly | Requires discipline to supersede rather than edit |
| adr-tools CLI | Sequential numbering, link management | Extra tooling, not integrated with KiwiFS |
| Confluence / Notion pages | Rich editing, comments | Off-repo, poor agent access, export friction |
| Inline code comments only | Zero overhead | Not discoverable, no structured lifecycle |

## Decision Outcome

Chosen option: **Markdown ADRs in-repo (MADR)**, because KiwiFS already
indexes frontmatter, validates schemas, enforces workflow transitions, and
exposes decisions to agents via MCP.

**We will** store ADRs as numbered markdown files under `decisions/` with
YAML frontmatter (`status`, `date`, `deciders`) validated by `.kiwi/schemas/adr.json`.

## Consequences

### Positive

- Decisions are searchable via `kiwi_search`, DQL, and graph backlinks
- Status workflow prevents silent edits to accepted decisions
- Auto-sequence assigns `adr_number` on write to `decisions/`
- New team members can read the decision log from `index.md`

### Negative

- Authors must learn frontmatter conventions and MADR section structure
- Supersession requires creating a new ADR instead of editing in place
- Sequential numbering depends on pipeline auto-sequence being enabled

### Neutral

- We adopt MADR section headings (Context, Decision Drivers, Considered Options,
  Decision Outcome, Consequences) as the team standard

## Related

- [[index|Decision log]]
- `.kiwi/templates/adr.md` — blank ADR template
- [UC-7: Architecture Decision Records](https://github.com/kiwifs/kiwifs/wiki/UC-7-Architecture-Decision-Records)
