# Schema — Knowledge Base

Agent-maintained knowledge base following the LLM Wiki pattern:
raw sources in, compiled wiki out, agent maintains it over time.
Includes episodic memory with consolidation into durable pages.

## Directory Structure

    pages/           Durable knowledge — one page per concept, entity, or topic
    episodes/        Per-session episodic notes (transient, consolidate later)
    index.md         Auto-maintained table of contents
    log.md           Append-only chronological record of all operations
    SCHEMA.md        This file — structure and conventions

## Frontmatter Fields

Every `.md` file should have YAML frontmatter. Required fields marked *.

### Pages (`pages/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Human-readable page title                   |
| description     | string     |          | One-line summary for search results         |
| tags            | string[]   | *        | Topic tags, lowercase, hyphenated           |
| status          | string     |          | `active` · `draft` · `review` · `deprecated` |
| last-reviewed   | date       |          | ISO 8601 date of last quality review        |
| derived-from    | object[]   |          | Provenance: `type`, `id`, `date`, `actor`   |
| merged-from     | object[]   |          | Episode paths this page was consolidated from. Each entry: `type`, `id`, optional `path`, `date` |
| confidence      | float      |          | 0.0–1.0, certainty level of this knowledge  |

### Episodes (`episodes/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| memory_kind     | string     | *        | Always `episodic`                           |
| episode_id      | string     | *        | Unique session/episode identifier           |
| session_id      | string     |          | Groups episodes from the same session       |
| confidence      | float      |          | 0.0–1.0, how certain is this observation    |
| tags            | string[]   |          | Topic tags                                  |
| consolidated    | boolean    |          | `true` when merged into a page              |
| merged-into     | string[]   |          | Paths of pages this was merged into         |

## Operations

See `.kiwi/playbook.md` for step-by-step MCP tool sequences.

### Ingest
Read a raw source → create/update pages in `pages/` →
update `index.md` and `log.md`. Always deduplicate first.

### Query
Search the wiki to answer questions. Use `kiwi_search` +
`kiwi_read` + `kiwi_backlinks`.

### Lint
Audit for orphans, broken links, stale content, missing
frontmatter, and coverage gaps. Use `kiwi_analytics`.

### Remember
Write a new episodic note under `episodes/` during a session.
Include `memory_kind: episodic` and a unique `episode_id`.
Append a summary to `log.md`.

### Consolidate
Merge related `episodes/` notes into durable `pages/` entries.
Set `merged-from` on the page, `consolidated: true` on the episode.
Run `kiwi_memory_report` to find unconsolidated episodes.

### Recall
Search memory for past observations. Use `kiwi_search` for keyword
recall or `kiwi_search_semantic` for conceptual recall. Prefer
durable pages over raw episodes when both exist.

## Conventions

- Link between pages with `[[wiki links]]`
- Keep pages focused — one concept per page
- Split pages over 300 lines
- Use YAML frontmatter for all structured metadata
- Append to `log.md` after every write operation
- Every page reachable from `index.md` within 2 hops
