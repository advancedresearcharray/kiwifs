---
memory_kind: episodic
episode_id: cursor-fleet-2026-06-30-issue-421-v3
title: "Wiki-link hover preview (#421) — v3 delivery"
tags: [kiwifs, issue-421, wiki-links, hover-card, ui]
date: 2026-06-30
---

## Task

Implement inline page preview on `[[wiki-link]]` hover for kiwifs/kiwifs#421 on branch `feat/issue-421-wiki-link-hover-preview-v3` from `main`.

## Before

- Kiwi knowledge depot (`http://192.168.167.240:3333`) unreachable — `kiwi_search` skipped; local fix doc at `pages/fixes/kiwifs-kiwifs/issue-421-wiki-link-hover-preview.md` from prior fleet run.
- `GET /api/kiwi/peek` existed server-side but UI had no client or hover integration on `main`.

## Work done

1. Added `@radix-ui/react-hover-card` + shadcn `hover-card` component.
2. Implemented `api.peek`, `wikiLinkPeek` cache/dedup layer, and `WikiLinkPreview` HoverCard wrapper.
3. Wired `renderWikiLinkAnchor()` into `KiwiPage` prose and local-notes markdown renderers (read mode only).
4. Added 10 regression tests in `wikiLinkPeek.test.ts`.

## Test results

```
cd ui && npm test -- src/lib/wikiLinkPeek.test.ts
# 10 passed (2026-06-30)
```

## Notes

- `node_modules` not writable in overlay workspace; `npm install` deferred to fleet CI.
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-421-wiki-link-hover-preview.md`
