---
title: Getting Started
description: How this knowledge base is organized and how to use it
tags: [meta, onboarding]
status: active
---

# Getting Started

This knowledge base is maintained using the LLM Wiki pattern.

## Structure

- `pages/` — durable knowledge, one concept per page
- `episodes/` — session notes, consolidated into pages over time
- `index.md` — table of contents
- `log.md` — chronological record of all changes

## How It Works

An agent [[SCHEMA|follows the schema]] to ingest new information,
answer questions from existing pages, and periodically lint for
quality. See `.kiwi/playbook.md` for the full operation guide.

## Conventions

- Every page has YAML frontmatter with `title` and `tags`
- Pages link to each other with `[[wikilinks]]`
- One concept per page — split when a page exceeds 300 lines
