---
memory_kind: episodic
episode_id: cursor-agent-2026-06-30-issue-426-split-view-redelivery
title: "Issue #426 split view — branch feat/issue-426-split-view"
tags: [kiwifs, ui, split-view, issue-426, fleet-delivery]
date: 2026-06-30
---

# Issue #426 split view redelivery

## Task

Re-deliver kiwifs/kiwifs#426 on clean branch from main for fleet publish.

## Actions

1. Checked out `main`, created `feat/issue-426-split-view`.
2. Cherry-picked 5 verified commits from prior `feat/split-view-426` work (fc1c045..3f794a5).
3. Ran full regression suite.

## Verification

- `cd ui && npm test` → 206 passed (includes 14 splitView + keybinding tests)
- `go test ./internal/keybindings/... -count=1` → ok

## Handoff

Branch `feat/issue-426-split-view` ready for fleet push + PR closing #426. Kiwi gateway unreachable — fix doc at `pages/fixes/kiwifs-kiwifs/issue-426-split-side-by-side-page-view.md`.
