# Schema — Research

_Template version: 2.0_

Literature notes, experiment logs, and analysis for researchers.
Follows FAIR principles (Findable, Accessible, Interoperable, Reusable)
for experiment documentation.

## Directory Structure

    literature/          One file per paper or source
    experiments/         One file per experiment, prefixed exp-NNN-
    notes/               Free-form working notes and synthesis
    questions.md         Open research questions registry
    index.md             Table of contents
    SCHEMA.md            This file — structure and conventions

## Frontmatter Fields

Every `.md` file should have YAML frontmatter. Required fields marked *.

### Literature (`literature/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Paper or source title                       |
| authors         | string[]   |          | List of authors                             |
| year            | number     |          | Publication year                            |
| doi             | string     |          | DOI for citation and retrieval              |
| url             | string     |          | Direct link to paper/source                 |
| methodology     | string     |          | `qualitative` · `quantitative` · `meta-analysis` · `review` · `theoretical` |
| relevance       | integer    |          | 1–5, how relevant to current research       |
| tags            | string[]   | *        | Topic and method tags                       |
| status          | string     |          | `read` · `skimmed` · `to-read`              |
| cited-by        | string[]   |          | Paths to experiments that reference this    |

### Experiments (`experiments/*.md`)

| Field              | Type       | Required | Values / Notes                              |
|--------------------|------------|----------|---------------------------------------------|
| title              | string     | *        | Experiment title                            |
| date               | date       | *        | ISO 8601 date                               |
| hypothesis         | string     |          | What you expect to find                     |
| research-question  | string     |          | Path to the question in `questions.md` this addresses |
| status             | string     | *        | `planned` · `running` · `completed` · `failed` · `abandoned` |
| result             | string     |          | `positive` · `negative` · `inconclusive` · `mixed` |
| protocol           | string     |          | Link to methodology/procedure used          |
| environment        | string     |          | Runtime environment, versions, hardware     |
| duration           | string     |          | How long the experiment ran                 |
| raw-data           | string     |          | Path or URI to raw data location            |
| sample-size        | string     |          | Number of samples/trials/runs               |
| tags               | string[]   |          | Topic and method tags                       |
| references         | string[]   |          | Paths to literature that informed this      |

### Notes (`notes/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Note title                                  |
| type            | string     |          | `synthesis` · `brainstorm` · `literature-review` · `methodology` · `observation` |
| date            | date       |          | Date written                                |
| status          | string     |          | `draft` · `active` · `archived`             |
| tags            | string[]   |          | Topic tags                                  |
| related         | string[]   |          | Paths to related experiments or literature  |

## Research Questions Registry

Maintain a `questions.md` file tracking open research questions.
Each question links to experiments attempting to answer it.

Format:
```markdown
## Open Questions

| ID | Question | Status | Experiments |
|----|----------|--------|-------------|
| Q1 | Does X cause Y? | investigating | [[experiments/exp-001-baseline]] |
| Q2 | Is A better than B? | unanswered | — |
```

## Experiment Reproducibility

Every experiment should be reproducible. Include:

1. **Environment** — exact software versions, hardware specs, OS
2. **Protocol** — step-by-step methodology (link to a procedure or inline)
3. **Variables** — independent, dependent, and controlled
4. **Raw data** — path or URI to unprocessed results
5. **Reproduction steps** — how another researcher would re-run this

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.

## Conventions

- One paper per file in `literature/`, named after the paper's slug.
- One experiment per file in `experiments/`, prefixed `exp-NNN-`
  with a zero-padded sequence.
- Free-form working notes live in `notes/` with a `type` field.
- Always link experiments to the literature that informed them via `references`.
- Always link literature to experiments that cite it via `cited-by`.
- Include DOI or URL for every literature entry where available.
- Set `relevance` score on literature to prioritize reading.
- Track open research questions in `questions.md`.
