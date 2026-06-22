---
memory_kind: episodic
episode_id: sprout-idle-nudge-2026-06-19-issue-340
title: "Issue #340 graph link-type filter delivery"
tags: [kiwifs, graph, ui, issue-340, sprout-idle-nudge]
date: 2026-06-19
---

## Task

Implement kiwifs/kiwifs#340 — link-type filter controls in the knowledge graph view.

## Before

- `kiwi_search` on cluster depot found existing fix doc at `pages/fixes/kiwifs-kiwifs/issue-340-graph-link-type-filter.md` and prior fleet episodes.
- Root cause confirmed: API already returns `relation` on edges (#323); UI deduplicated edges without relation metadata.

## Work done

- Landed implementation from `feat/graph-link-type-filter-340` as commit `91e99e7`.
- Added `ui/src/lib/kiwiGraphFilters.ts` with pure filter helpers + session persistence.
- Updated `KiwiGraph.tsx` with multi-select Badge chips, relation-aware link visibility, node dimming.
- Extended `GraphEdge.relation` in `api.ts`; mock data includes typed edges.

## Tests

```bash
cd ui && npm test -- src/lib/kiwiGraphFilters.test.ts
# 14 passed
```

## Branch

`feat/graph-link-type-filter-340` @ `91e99e7` (local commit, not pushed — fleet publishes).
