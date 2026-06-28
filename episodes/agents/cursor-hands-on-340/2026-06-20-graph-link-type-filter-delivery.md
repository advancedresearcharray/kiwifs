---
memory_kind: episodic
episode_id: cursor-hands-on-340-2026-06-20
title: "Issue #340 hands-on delivery — graph link-type filter"
tags: [kiwifs, graph, ui, issue-340, typed-links, cursor-hands-on-340]
date: 2026-06-20
---

## Task

Hands-on takeover for [kiwifs/kiwifs#340](https://github.com/kiwifs/kiwifs/issues/340) after fleet agent failed delivery check (`no_committed_diff`, `peer_review_not_passed`).

## Pre-implementation search

- `kiwi_search` on cluster depot (`graph link type filter 340`) → found fix doc at `pages/fixes/kiwifs-kiwifs/issue-340-graph-link-type-filter.md`.
- Verified implementation already on branch `feat/graph-link-type-filter-340-clean` (3 commits ahead of `origin/main`).

## Verification

```bash
cd ui && npm test -- src/lib/kiwiGraphFilters.test.ts src/lib/uiConfigStore.test.ts  # 18 passed
cd ui && npm test  # 34 files, 179 passed
```

## Delivery

- Pushed branch to fork: `advancedresearcharray/kiwifs@feat/graph-link-type-filter-340-clean`
- Opened PR: https://github.com/kiwifs/kiwifs/pull/409 (closes #340)
- Updated cluster fix doc status to verified with PR reference.

## Acceptance criteria

| Criterion | Status |
| --- | --- |
| Filter controls visible in graph view | ✅ Badge chips when typed relations exist |
| Selecting a link type shows only edges of that type | ✅ `linkVisible` + `edgeMatchesRelationFilter` |
| "All" option shows all links (default) | ✅ Empty `Set` = no filter |
| Multiple types can be selected simultaneously | ✅ Multi-select chip toggles |
| Nodes without matching edges hidden or dimmed | ✅ Dimmed via `nodeColor` (`#243042`) |
| Filter state persists during session | ✅ `sessionStorage` key `kiwifs-graph-relation-filter` |
