# Schema — Research

Literature notes, experiment logs, and analysis for researchers.

## Directory Structure

    literature/          One file per paper or source
    experiments/         One file per experiment, prefixed exp-NNN-
    notes/               Free-form working notes and synthesis
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
| tags            | string[]   | *        | Topic and method tags                       |
| status          | string     |          | `read` · `skimmed` · `to-read`              |

### Experiments (`experiments/*.md`)

| Field           | Type       | Required | Values / Notes                              |
|-----------------|------------|----------|---------------------------------------------|
| title           | string     | *        | Experiment title                            |
| date            | date       | *        | ISO 8601 date                               |
| hypothesis      | string     |          | What you expect to find                     |
| status          | string     | *        | `planned` · `running` · `completed` · `failed` |
| result          | string     |          | `positive` · `negative` · `inconclusive`    |
| tags            | string[]   |          | Topic and method tags                       |

## Operations

See `.kiwi/playbook.md` for MCP tool sequences.

## Conventions

- One paper per file in `literature/`, named after the paper's slug.
- One experiment per file in `experiments/`, prefixed `exp-NNN-`
  with a zero-padded sequence.
- Free-form working notes live in `notes/`.
