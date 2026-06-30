---
memory_kind: episodic
episode_id: cursor-hands-on-426-2026-06-30
title: "Issue #426 split view — hands-on verification and PR"
tags: [kiwifs, ui, split-view, issue-426, hands-on, delivery]
date: 2026-06-30
---

# Issue #426 — split / side-by-side page view (verified)

## Pre-search

- Kiwi gateway `192.168.167.240:3333` unreachable from overlay; local fix doc at `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`.

## Actions

1. Linked workspace git via `GIT_DIR=.git.writable`; restored corrupted `internal/exporter/mkdocs.go` from HEAD.
2. Checked out `fork/feat/split-side-by-side-426` (6 commits ahead of main) with `SplitViewProvider`, `KiwiSplitView`, shadcn `ui/resizable.tsx`.
3. Ran regression tests (all green).

## Test output

```
cd ui && npm test -- --run src/lib/splitView.test.ts src/lib/kiwiKeybindings.test.ts src/components/ui/resizable.test.ts
→ 26 passed

go test ./internal/keybindings/... -count=1
→ ok
```

## Acceptance criteria

| Criterion | Status |
|-----------|--------|
| Tree right-click → Open in Split View | PASS |
| Wiki-link context menu → Open in Split View | PASS |
| Mod+\ toggles split view | PASS |
| Independent pane navigation | PASS |
| Resizable divider | PASS |
| Close secondary pane (×) | PASS |
| History → Compare with current | PASS |
| Mobile blocked below 768px | PASS |

## PR

Opened upstream PR from `advancedresearcharray:feat/split-side-by-side-426` → `kiwifs/kiwifs` main.
