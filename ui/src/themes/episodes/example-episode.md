---
memory_kind: episodic
episode_id: example-001
session_id: example
confidence: 0.8
tags: [onboarding]
---
# Example Episode

This is a sample episodic note. Replace or delete this file.

Each agent session can create files here with `memory_kind: episodic`
and a unique `episode_id`. A consolidation step later merges related
episodes into durable pages under `pages/` and records the link via
`merged-from` in frontmatter.

Run `kiwifs memory report` to see which episodes haven't been
consolidated yet.
