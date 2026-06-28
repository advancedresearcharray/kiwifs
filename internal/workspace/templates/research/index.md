---
title: Research
owner: researcher
status: active
tags: [meta, research]
---

# Research

Literature notes, experiment logs, and synthesis for your research.

## Research Questions

Open questions driving the research agenda. Each links to experiments
attempting to answer it.

| ID | Question | Status | Experiments |
|----|----------|--------|-------------|
| Q1 | _What is the baseline performance?_ | answered | [[experiments/exp-001-baseline]] |
| Q2 | _Does approach X improve over baseline?_ | unanswered | — |

<!-- Add new questions here. Status: unanswered · investigating · answered · abandoned -->

## Experiments

Chronological experiment logs, each prefixed `exp-NNN-<slug>.md`.

| Experiment | Status | Result | Question |
|------------|--------|--------|----------|
| [[experiments/exp-001-baseline]] | completed | positive | Q1 |

## Literature

One note per paper or source, named by author/topic slug.

| Paper | Year | Relevance | Status |
|-------|------|-----------|--------|
| [[literature/example-paper]] | 2025 | 4 | read |

## Notes & Synthesis

Free-form working notes connecting insights across experiments
and literature.

| Note | Type | Topic |
|------|------|-------|
| [[notes/synthesis-template]] | synthesis | Template for cross-source synthesis |

## Workflow

1. **Ask** a question → add it to the Research Questions table above
2. **Read** a paper → create a note in `literature/` with DOI/URL
3. **Design** an experiment → create in `experiments/` with status `planned`, link to question
4. **Run** the experiment → update status to `running`, record observations
5. **Analyze** → update status to `completed`, record results and conclusions
6. **Synthesize** → connect findings across sources in `notes/`
7. **Answer** → update question status, link to concluding experiment/note
