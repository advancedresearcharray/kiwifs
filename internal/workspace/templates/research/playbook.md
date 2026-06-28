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
   doi: "10.xxxx/xxxxx"
   url: https://...
   methodology: quantitative
   relevance: 4
   tags: [topic, method]
   status: read | skimmed | to-read
   cited-by: []
   ---
   ```
3. Fill in sections: Abstract/Summary, Key Findings, Methods,
   Relevance, Strengths/Limitations, Questions/Gaps.
4. Cross-link to related papers with `[[wikilinks]]`.
5. Update `index.md` literature table.
6. If this informs an existing experiment, update that experiment's
   `references` field and this paper's `cited-by` field.

## Log Experiment

When running an experiment:

1. `kiwi_search` for related past experiments.
2. Determine next experiment number from `index.md`.
3. `kiwi_write` to `experiments/exp-NNN-<slug>.md` with:
   ```yaml
   ---
   title: "Experiment Title"
   date: YYYY-MM-DD
   hypothesis: "What you expect to find"
   research-question: "Q1"
   status: planned | running | completed | failed | abandoned
   result: positive | negative | inconclusive | mixed
   protocol: "Description or link to methodology"
   environment: "OS, hardware, software versions"
   duration: "24 hours"
   raw-data: "data/exp-NNN/"
   sample-size: "N trials"
   tags: [topic, method]
   references: [literature/relevant-paper.md]
   ---
   ```
4. Structure the body with:
   - **Hypothesis** — what you expect
   - **Variables** — independent, dependent, controlled
   - **Environment** — exact setup for reproducibility
   - **Protocol** — step-by-step methodology
   - **Observations** — notes during the run
   - **Results** — data tables, measurements
   - **Conclusions** — what you learned
   - **Reproduction Steps** — how to re-run
5. Link to relevant `[[literature/<paper>]]` that informed this experiment.
6. Update `index.md` experiments table.
7. Update the Research Questions table if this answers a question.

## Synthesize Findings

When connecting insights across experiments and literature:

1. `kiwi_search` for all related experiments and papers.
2. `kiwi_read` each relevant file.
3. `kiwi_write` a synthesis note in `notes/<slug>.md` with:
   ```yaml
   ---
   title: "Synthesis Title"
   type: synthesis
   date: YYYY-MM-DD
   status: active
   tags: [topic]
   related: [experiments/exp-001-baseline.md, literature/example-paper.md]
   ---
   ```
4. Structure with: Question, Sources, Findings, Implications, Open Questions.
5. Cross-link to all sources with `[[wikilinks]]`.
6. Use `kiwi_query_meta` to filter experiments by status or result
   for systematic reviews.

## Manage Research Questions

Track the questions driving your research:

1. Open `index.md` to see the Research Questions table.
2. Add new questions with a unique ID and `unanswered` status.
3. When starting an experiment on a question: set status to `investigating`.
4. When you have a conclusion: set status to `answered` and link the
   concluding experiment or synthesis note.
5. Questions that are no longer relevant: set status to `abandoned`.

## Maintain

Run periodically:

1. `kiwi_lint` with `path` — check individual files for structural issues.
2. `kiwi_analytics` — find orphans and stale notes.
3. Find unread papers:
   ```
   kiwi_query("TABLE _path, title, relevance WHERE status = 'to-read' SORT relevance DESC")
   ```
4. Find stalled experiments:
   ```
   kiwi_query("TABLE _path, title, date WHERE status = 'planned' OR status = 'running' SORT date ASC")
   ```
5. Find unanswered research questions and check for relevant new experiments.
6. Update `cited-by` on literature when new experiments reference them.
7. Update `last-reviewed` on notes that are still accurate.

**Best practice:** After every `kiwi_write`, call `kiwi_lint` on the same path.
The server auto-formats cosmetic issues; `kiwi_lint` only reports semantic fixes.

## Quality Rules

- **One paper per file** in `literature/`, named after the paper's slug.
- **One experiment per file** in `experiments/`, prefixed `exp-NNN-`.
- **Include DOI/URL** on every literature entry where available.
- **Frontmatter required.** At least `title` and `tags`.
- **Link to sources.** Every experiment should reference the literature.
- **Reproducibility.** Every experiment should have environment + protocol + reproduction steps.
- **Track questions.** Every experiment should link to a research question.
- **No orphans.** All files reachable from `index.md`.
- **Bidirectional links.** Literature `cited-by` ↔ Experiment `references`.
