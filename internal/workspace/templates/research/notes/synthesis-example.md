---
title: "Synthesis: Transformer Architectures"
type: synthesis
tags: [synthesis, transformers]
status: draft
date: 2026-01-01
related: [papers/example-paper.md, papers/transformer-survey.md]
---

# Synthesis: Transformer Architectures

_Connect insights across multiple papers in the reading library._

## Question

How have Transformer architectures evolved since the original attention paper,
and what efficiency improvements matter for practical deployment?

## Sources

| Source | Type | Key Insight |
|--------|------|-------------|
| [[papers/example-paper]] | Paper | Introduced self-attention without recurrence |
| [[papers/transformer-survey]] | Paper | Taxonomy of variants and applications |

## Findings

1. Self-attention replaced recurrence as the dominant sequence modeling primitive.
2. Efficiency variants (sparse, linear attention) trade expressivity for scale.
3. Multimodal and vision Transformers extend the same core mechanism.

## Implications

- For literature reviews: organize by architectural variant, not application domain.
- For reading queue: prioritize survey papers before diving into niche variants.

## Open Questions

- Which efficiency improvements survive at production scale?
- What gaps remain in the survey's coverage of recent architectures?
