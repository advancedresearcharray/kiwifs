---
memory_kind: episodic
episode_id: cursor-takeover-340-peer-review-2026-06-20
title: "Issue #340 peer review fixes — graph link-type filter"
tags: [kiwifs, graph, ui, issue-340, typed-links, peer-review, cursor-takeover-340]
date: 2026-06-20
---

## Task

Hands-on takeover for [PR #409](https://github.com/kiwifs/kiwifs/pull/409) / [#340](https://github.com/kiwifs/kiwifs/issues/340) after fleet engineer `peer_review_blocked`.

## Pre-implementation search

- Read `pages/fixes/kiwifs-kiwifs/issue-340-graph-link-type-filter.md` (existing fix doc).
- Bugbot peer review identified 3 MUST-FIX issues.

## Peer review fixes

1. **`reconcileRelationFilter`** — session filter intersected with current graph relations; empty intersection resets to All (prevents stuck dimmed graph when switching to wiki-only workspace).
2. **Relation chips** — `collectRelationTypes(resolvedLinks)` instead of raw `resp.edges` (no phantom relation types for unresolved targets).
3. **Path finder** — adjacency built from relation-filtered links when filter active (paths match visible edges).

## Tests (2026-06-20)

```bash
cd ui && npm test -- src/lib/kiwiGraphFilters.test.ts  # 17 passed
cd ui && npm test  # 34 files, 182 passed
```

## Files changed

- `ui/src/lib/kiwiGraphFilters.ts` — add `reconcileRelationFilter`
- `ui/src/lib/kiwiGraphFilters.test.ts` — 3 reconciliation tests
- `ui/src/components/KiwiGraph.tsx` — reconcile effect, resolvedLinks relations, filtered adjacency
- `pages/fixes/kiwifs-kiwifs/issue-340-graph-link-type-filter.md` — peer review notes

## Delivery

- Branch: `feat/graph-link-type-filter-340-clean`
- PR: https://github.com/kiwifs/kiwifs/pull/409
