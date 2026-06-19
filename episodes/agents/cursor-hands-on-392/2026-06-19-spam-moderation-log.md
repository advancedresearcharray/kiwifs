---
memory_kind: episodic
episode_id: cursor-hands-on-392-2026-06-19
title: "Issue #392 — spam moderation log unlock fix (hands-on delivery)"
tags: [kiwifs, issue-392, spam-filter, ci, github-actions, bugfix]
date: 2026-06-19
---

## Context

Hands-on takeover for kiwifs/kiwifs#392 after fleet agent delivered incorrect commit `791a933` (deleted `spam-filter.yml`). Correct fix recovered from `/tmp/kiwifs-work` commit `0aeb1fe`.

## Investigation

1. `kiwi_search` on cluster depot — no prior fix doc for issue #392.
2. Root cause: locked tracking issue #392 caused `403 Unable to create comment because issue is locked` on workflow run `27778167379`.
3. PR #395 try/catch kept moderation actions running but logging still failed silently.

## Fix

- Extracted `.github/scripts/spam-filter.cjs` with `ensureIssueUnlocked()` before `createComment`.
- Skip spam filter when event targets #392.
- Added 9 regression tests; CI runs `node --test .github/scripts/*.test.mjs` on infra changes.
- Added `.github/scripts/**` to CI path filter (commit `f71ae8c`).

## Verification

```
node --test .github/scripts/spam-filter.test.mjs
# 9 pass, 0 fail
```

## Deliverables

- PR: https://github.com/kiwifs/kiwifs/pull/397
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-392-spam-moderation-log.md`
- Branch: `fix/issue-392-spam-moderation-log` @ `f71ae8c`
