---
type: paper
title: "Attention Is All You Need"
authors: [Vaswani, A., Shazeer, N., Parmar, N., Uszkoreit, J., Jones, L., Gomez, A. N., Kaiser, L., Polosukhin, I.]
year: 2017
venue: NeurIPS
doi: "10.48550/arXiv.1706.03762"
bibtex_key: vaswani2017attention
abstract: "The dominant sequence transduction models are based on complex recurrent or convolutional neural networks. We propose the Transformer, based solely on attention mechanisms."
tags: [transformers, attention, nlp]
cites: ["[[papers/transformer-survey]]"]
workflow: reading
state: incorporated
---

# Attention Is All You Need

> **Example paper.** Replace with your first real reading note, or delete
> once you've added your own library.

## Summary

Introduces the Transformer architecture using multi-head self-attention,
replacing recurrence and convolutions for sequence transduction.

## Key Findings

- Self-attention enables parallelization and captures long-range dependencies.
- Multi-head attention learns different representation subspaces.
- Achieves state-of-the-art on WMT 2014 EN-DE and EN-FR translation.

## Annotations

- _Section 3.1:_ Scaled dot-product attention — note the √d_k scaling factor.
- _Figure 1:_ Encoder-decoder stack layout.

## Relevance

Foundational for modern LLMs. Cited by [[papers/transformer-survey]].
Informs synthesis in [[notes/synthesis-example]].

## Quotes

> "Attention is all you need." (Abstract)
