---
memory_kind: episodic
episode_id: cursor-fleet-2026-06-30-issue-421-v3-verify
title: "Wiki-link hover preview (#421) — verification + error retry test"
tags: [kiwifs, issue-421, wiki-links, hover-card, ui, regression-test]
date: 2026-06-30
---

## Task

Verify and finalize kiwifs/kiwifs#421 on branch `feat/issue-421-wiki-link-hover-preview-v3` (ClawWork fleet handoff).

## Before

- Feature already implemented in 4 commits vs `main` (HoverCard preview, peek client, KiwiPage wiring).
- Kiwi depot at `192.168.167.240:3333` unreachable; MCP gateway had no servers registered.
- `npm install` blocked in overlay (`node_modules` owned by `nobody`); unit tests run without Radix package on disk.

## Work done

1. Re-read issue acceptance criteria — all met on branch (300ms/100ms delays, lazy `/api/kiwi/peek`, broken-link UI, read-mode only via `KiwiPage`, no keyboard popover).
2. Re-ran full UI suite: 204 passed (35 files).
3. Ran Go peek tests: `go test ./internal/mcpserver/ -run TestPeek` — pass.
4. Added regression test: transient fetch errors are **not** cached so hover can retry (`wikiLinkPeek.test.ts`).
5. Updated fix doc test count (16 tests) and retry behavior note.

## Test results

```
cd ui && npm test -- src/lib/wikiLinkPeek.test.ts src/lib/wikiLinkAnchor.test.ts
# 16 passed

cd ui && npm test
# 204 passed (35 files)

go test ./internal/mcpserver/ -run TestPeek
# ok
```

## Notes

- Local commit only (fleet publishes PR). Branch tracks `fork/feat/issue-421-wiki-link-hover-preview-v3`.
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-421-wiki-link-hover-preview.md`
