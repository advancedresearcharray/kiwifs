---
title: Research Library
kiwi-view: true
query: "TABLE _path AS Path, title AS Title, state AS State, year AS Year, venue AS Venue WHERE type = \"paper\" SORT state ASC, year DESC"
---

# Research Library

Track papers, reading progress, synthesis notes, and literature review drafts.

## Reading Queue

Papers not yet finished (`unread` or `reading`):

```dql
TABLE title AS Title, authors AS Authors, state AS State, year AS Year
WHERE type = "paper" AND (state = "unread" OR state = "reading")
SORT year DESC
```

## All Papers

| Paper | Year | Venue | State |
|-------|------|-------|-------|
| [[papers/example-paper]] | 2017 | NeurIPS | incorporated |
| [[papers/transformer-survey]] | 2021 | ACM Computing Surveys | summarized |

## Notes & Synthesis

Cross-paper insights live in `notes/`.

| Note | Type | Status |
|------|------|--------|
| [[notes/synthesis-example]] | synthesis | draft |

## Literature Reviews

Draft and published reviews in `reviews/`.

| Review | Status |
|--------|--------|
| [[reviews/literature-review-draft]] | draft |

## Workflow

1. **Add** a paper → create `papers/<slug>.md` with `state: unread`
2. **Read** → advance to `reading`, then `annotated` as you take notes
3. **Summarize** → capture key findings at `summarized`
4. **Incorporate** → link insights into `notes/` or `reviews/`, set `incorporated`
5. **Synthesize** → write cross-paper notes in `notes/`
6. **Review** → assemble literature review drafts in `reviews/`
