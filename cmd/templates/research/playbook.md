# Agent Playbook — Research

Literature notes, experiment logs, and analysis for researchers.
When connected via MCP, use these operations to maintain the
research knowledge base.

## Quick Start

1. Call `kiwi_context` to get this playbook + schema + index in one call
2. Call `kiwi_tree` to see the current structure
3. Use the operations below to add literature, log experiments, and synthesize

## Add Literature Note

When reading a new paper or source:

1. `kiwi_search` to check if this paper is already noted.
2. `kiwi_write` to `literature/<slug>.md` with:
   ```yaml
   ---
   title: "Paper Title"
   authors: [Author One, Author Two]
   year: 2025
   tags: [topic, method]
   status: read | skimmed | to-read
   ---
   ```
3. Summarize key findings, methods, and relevance.
4. Cross-link to related papers with `[[wikilinks]]`.
5. Update `index.md`.

## Log Experiment

When running an experiment:

1. `kiwi_search` for related past experiments.
2. `kiwi_write` to `experiments/exp-NNN-<slug>.md` with:
   ```yaml
   ---
   title: "Experiment Title"
   date: YYYY-MM-DD
   hypothesis: "What you expect to find"
   status: planned | running | completed | failed
   result: positive | negative | inconclusive
   tags: [topic, method]
   ---
   ```
3. Document setup, observations, and results.
4. Link to relevant `[[literature/<paper>]]` that informed this experiment.
5. Update `index.md`.

## Synthesize Findings

When connecting insights across experiments and literature:

1. `kiwi_search` for all related experiments and papers.
2. `kiwi_read` each relevant file.
3. `kiwi_write` a synthesis note in `notes/<slug>.md`.
4. Cross-link to all sources with `[[wikilinks]]`.
5. Use `kiwi_query_meta` to filter experiments by status or result
   for systematic reviews.

## Maintain

Run periodically:

1. `kiwi_lint` with `path` — check individual files for structural issues.
2. `kiwi_analytics` — find orphans and stale notes.
3. `kiwi_query_meta` with `$.status=to-read` to find unread papers.
4. `kiwi_query_meta` with `$.status=planned` to find unstarted experiments.
5. Update `last-reviewed` on notes that are still accurate.

**Best practice:** After every `kiwi_write`, call `kiwi_lint` on the same path.
The server auto-formats cosmetic issues; `kiwi_lint` only reports semantic fixes.

## Quality Rules

- **One paper per file** in `literature/`, named after the paper's slug.
- **One experiment per file** in `experiments/`, prefixed `exp-NNN-`.
- **Frontmatter required.** At least `title` and `tags`.
- **Link to sources.** Every experiment should reference the literature.
- **No orphans.** All files reachable from `index.md`.
