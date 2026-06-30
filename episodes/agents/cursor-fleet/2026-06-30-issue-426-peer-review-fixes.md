---
memory_kind: episodic
episode_id: cursor-fleet-2026-06-30-issue-426-peer-review
title: "Issue #426 peer review fixes — normalizeKey + split view state sync"
tags: [kiwifs, split-view, peer-review, issue-426]
date: 2026-06-30
---

# Issue #426 peer review fixes

## Context

Hands-on takeover after fleet delivery failed peer review on `normalizeKey`, `toggle_split_view` test coverage, and split view state consistency.

## Changes

1. Go `normalizeKey`: alias map for single- and multi-character key names (no redundant len branch).
2. TS `normalizeKeyPart`: mirrored alias map; `normalizeChord` delegates key tokens to it.
3. `syncSplitViewWithActivePath` in `splitView.ts`; wired in `App.tsx` + space switch reset.
4. Expanded tests: 8 new cases in splitView/kiwiKeybindings; `TestNormalizeKey` in Go.

## Test results

```
cd ui && npm test  → 206 passed
go test ./internal/keybindings/... -count=1  → ok
```

## Kiwi MCP

Gateway unreachable (`kiwifs_mcp_invoke: fetch failed`). Fix doc updated locally at `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`.
