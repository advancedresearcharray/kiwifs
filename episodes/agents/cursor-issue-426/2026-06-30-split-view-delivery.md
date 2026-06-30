---
memory_kind: episodic
episode_id: cursor-issue-426-delivery
title: "Split view delivery — peer review fixes + commit"
tags: [ui, split-view, issue-426, peer-review]
date: 2026-06-30
---

## Work item

kiwifs/kiwifs#426 — feat(ui): split / side-by-side page view

## Actions

1. Verified prior implementation (15 files, 905 lines vs main)
2. Ran tests: `go test ./internal/keybindings/...` PASS; `npm test` 196/196 PASS
3. Peer review (bugbot) found navigation/keybinding gaps
4. Fixed in `App.tsx` + `SplitPageView.tsx`:
   - Set `activePath` when opening split without primary page
   - Left pane uses `navigate()` for URL/tree sync
   - Close split on primary page delete
   - Mod+E/S/mode wired to left-pane editor via `SplitPageViewHandle`
   - Removed broken edit button on secondary pane
5. Committed `72286d8` on `feat/issue-426-split-view`, pushed to fork
6. PR: https://github.com/advancedresearcharray/kiwifs/pull/37

## Kiwi MCP

Gateway unreachable from overlay; fix doc updated locally at `pages/fixes/kiwifs-kiwifs/issue-426-split-page-view.md`.
