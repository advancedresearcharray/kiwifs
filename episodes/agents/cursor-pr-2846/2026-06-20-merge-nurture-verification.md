---
memory_kind: episodic
episode_id: cursor-pr-2846-2026-06-20-merge-nurture-verification
title: "PR #2846 — README bounty table links merge-nurture verification"
tags: [claude-builders-bounty, readme, issue-2746, pr-2846, opire, merge-nurture]
date: 2026-06-20
---

# PR #2846 — README bounty table links merge-nurture

## Context

Cursor hands-on takeover for idle queue item targeting [PR #2846](https://github.com/claude-builders-bounty/claude-builders-bounty/pull/2846) (fixes #2746). Opire bounty — merge-to-pay.

## Pre-search (KiwiFS)

- Found durable fix doc: `pages/fixes/claude-builders-bounty-claude-builders-bounty/issue-2746-readme-bounty-table-links.md`
- Prior fleet episodes confirm repeated green `npm test` runs; no new code defect identified.

## Verification

Branch: `fix-issue-2746-opire-try` @ `3ae7129d`

```bash
cd /workspace/claude-builders-bounty/claude-builders-bounty
npm test
```

Result: **23/23 pass** (3 readme + 11 peer-review + 9 delivery)

GitHub PR state: **MERGEABLE / CLEAN** (no checks reported on branch; no review blockers)

README diff vs `main`: fragile `../../issues/N` replaced with `/claude-builders-bounty/claude-builders-bounty/issues/N` for rows #1–#5; contributor note present.

## Outcome

No code changes required — deliverable already committed and green. Fleet may publish merge-ready comment and refresh `/opire try` on issue #2746.
