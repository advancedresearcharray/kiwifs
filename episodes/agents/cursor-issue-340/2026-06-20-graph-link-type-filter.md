---
memory_kind: episodic
episode_id: cursor-issue-340-2026-06-20
title: "Issue #340 — graph link-type filter delivery"
tags: [kiwifs, graph, ui, issue-340, typed-links, cursor-issue-340]
date: 2026-06-20
---

## Task

Implement [kiwifs/kiwifs#340](https://github.com/kiwifs/kiwifs/issues/340): link-type filter controls in the knowledge graph view.

## Pre-implementation search

- `kiwi_search` on cluster depot (`graph link type filter 340`) → found existing fix doc at `pages/fixes/kiwifs-kiwifs/issue-340-graph-link-type-filter.md` and prior fleet episodes.
- Root cause confirmed: API already returns `relation` on edges (#323); UI deduplicated edges without relation metadata and had no filter UI.

## Work done

1. Checked out prior implementation on `feat/graph-link-type-filter-340` (commit `97e2f47`).
2. Rebased cleanly onto `origin/main` as branch `feat/graph-link-type-filter-340-clean` (cherry-pick only the feature commit).
3. Restored accidental removal of `toolbarViews?: string[] | null` from `ui-config` return type in `api.ts`.

## Tests

```bash
cd ui && npm test -- src/lib/kiwiGraphFilters.test.ts   # 14 passed
cd ui && npm test -- src/lib/uiConfigStore.test.ts       # 4 passed
```

## Branch

- `feat/graph-link-type-filter-340-clean` @ HEAD (local, not pushed — fleet publishes PR)

## Acceptance criteria

| Criterion | Status |
| --- | --- |
| Filter controls visible in graph view | ✅ Badge chips in analytics bar |
| Selecting a link type shows only edges of that type | ✅ `linkVisible` + `edgeMatchesRelationFilter` |
| "All" option shows all links (default) | ✅ Empty `Set` = no filter |
| Multiple types can be selected simultaneously | ✅ Multi-select chip toggles |
| Nodes without matching edges hidden or dimmed | ✅ Dimmed via `nodeColor` (`#243042`) |
| Filter state persists during session | ✅ `sessionStorage` key `kiwifs-graph-relation-filter` |
