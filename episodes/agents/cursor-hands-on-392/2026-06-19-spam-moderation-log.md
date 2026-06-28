---
memory_kind: episodic
episode_id: cursor-hands-on-392-2026-06-19
title: "PR 397 hands-on delivery — spam moderation log unlock"
tags: [kiwifs, issue-392, spam-filter, ci, pr-397, hands-on-takeover]
date: 2026-06-19
---

# Hands-on takeover — kiwifs/kiwifs#397

## Context

Fleet engineer agent failed delivery check (`not_committed`, `no_committed_diff`). Overlay workspace `/tmp/kiwifs-overlay/mnt` diverged with erroneous commit `1c77224` that deleted spam-filter scripts/tests. Correct fix lived at `4ce2a25` on origin.

## Pre-work

- `kiwi_search` on cluster depot — no indexed fix doc yet for issue #392.
- PR #397 head `4ce2a25` already green on CI (run 27839043215).
- Recovered writable tree from `/tmp/kiwifs-overlay/upper` (overlay upper layer).

## Verification

```text
cd /tmp/kiwifs-overlay/upper
node --test .github/scripts/spam-filter.test.mjs  → 9 pass, 0 fail
```

## Delivery

- Committed durable fix doc: `pages/fixes/kiwifs-kiwifs/issue-392-spam-moderation-log.md`
- Synced overlay mnt git ref to match PR head; pushed branch to origin
- Wrote fix doc + episode to Kiwi cluster depot

## Outcome

Spam filter unlocks #392 before logging; regression tests and CI path filter verified. Closes #392.
