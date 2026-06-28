# Research Library Playbook

Paper tracking, reading workflow, and literature review synthesis.
When connected via MCP, use these operations to maintain the library.

## Quick Start

1. Call `kiwi_context` to get this playbook + schema + index
2. Call `kiwi_workflow_board` with `workflow: reading` to see reading progress
3. Use the operations below to add papers and advance through states

## Add a Paper

When you discover a paper to read:

1. `kiwi_search` to check if it is already in the library.
2. `kiwi_write` to `papers/<slug>.md` with:
   ```yaml
   ---
   type: paper
   title: "Paper Title"
   authors: [Author One, Author Two]
   year: 2024
   venue: Conference or Journal Name
   doi: "10.xxxx/xxxxx"
   bibtex_key: author2024title
   abstract: "One-line abstract for quick recall."
   tags: [topic, method]
   cites: ["[[related-paper-slug]]"]
   workflow: reading
   state: unread
   ---
   ```
3. Fill in sections: Summary, Key Findings, Annotations, Relevance.
4. Cross-cite related papers with `cites` and `[[wikilinks]]`.
5. Update `index.md` or rely on embedded DQL views.

## Advance Reading State

Move papers through the reading workflow:

```
kiwi_workflow_advance(path: "papers/my-paper.md", workflow: "reading", target_state: "reading")
```

Valid progression: `unread → reading → annotated → summarized → incorporated`.
Backward transitions (e.g. returning a paper to `reading` for re-annotation) are
also allowed. Skipping states is rejected.

At each stage:
- **reading** — actively reading the source
- **annotated** — marginal notes and highlights captured in the file body
- **summarized** — key findings written in Summary/Key Findings sections
- **incorporated** — insights linked into `notes/` or `reviews/`

Invalid transitions are rejected by the workflow engine.

## Synthesize Across Papers

When connecting insights from multiple papers:

1. `kiwi_query` for papers at `summarized` or `incorporated`:
   ```
   kiwi_query("TABLE _path, title, state WHERE type = 'paper' AND state IN ('summarized', 'incorporated')")
   ```
2. `kiwi_read` each relevant paper.
3. `kiwi_write` a synthesis note in `notes/<slug>.md` with `type: synthesis`
   and `related` listing source papers.
4. Link sources with `[[wikilinks]]` in the body.

## Draft a Literature Review

When ready to publish findings:

1. `kiwi_write` to `reviews/<slug>.md` with `papers` listing included sources.
2. Structure with Introduction, Background, Related Work, Discussion, Conclusion.
3. Reference papers with `[[wikilinks]]` throughout.

## Query the Library

Unread papers by relevance:

```
kiwi_query("TABLE _path, title, year, venue WHERE type = 'paper' AND state = 'unread' SORT year DESC")
```

Reading queue:

```
kiwi_query("TABLE _path, title, state WHERE type = 'paper' AND state IN ('unread', 'reading')")
```

Papers by venue:

```
kiwi_query("TABLE title, authors, year WHERE type = 'paper' AND venue = 'NeurIPS'")
```

## Maintain

Run periodically:

1. `kiwi_lint` with `path` — validates frontmatter against `.kiwi/schemas/paper.json`.
2. `kiwi_workflow_board` for `reading` — spot stalled papers in `reading`.
3. Find papers missing DOI:
   ```
   kiwi_query("TABLE _path, title WHERE type = 'paper' AND doi IS NULL")
   ```
4. Ensure synthesis notes link back to incorporated papers.

**Best practice:** After every `kiwi_write`, call `kiwi_lint` on the same path.

## Quality Rules

- **Every paper has `type: paper`** — required for schema validation and DQL.
- **Include authors, year, venue** — required fields validated by JSON Schema.
- **Start at `state: unread`** — advance through the reading workflow deliberately.
- **Cross-cite with `cites`** — use wikilink syntax for typed backlinks.
- **One paper per file** in `papers/`, named by slug.
- **Synthesis in `notes/`** — connect multiple incorporated papers.
- **Reviews in `reviews/`** — aggregate papers into publishable drafts.
