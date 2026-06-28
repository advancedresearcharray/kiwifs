---
memory_kind: episodic
episode_id: cursor-issue-392-2026-06-19
title: "Issue #392 — spam moderation log unlock fix"
tags: [kiwifs, issue-392, spam-filter, ci, github-actions, bugfix]
date: 2026-06-19
---

## Context

Work-queue bounty for kiwifs/kiwifs#392. Internal tracking issue receives spam filter log comments. Failed workflow run `27778167379` showed `HttpError: Unable to create comment because issue is locked` when posting to #392.

## Investigation

1. Searched Kiwi pages/fixes — no prior doc for issue #392.
2. Issue #392 timeline: locked `resolved` at creation, unlocked later same day.
3. PR #395 merged try/catch only; logging still failed silently when locked.

## Fix

- Extracted `.github/scripts/spam-filter.cjs` from inline workflow script.
- Added `ensureIssueUnlocked()` before `createComment` on #392.
- Skip spam filter when event targets #392.
- Added 9 regression tests; wired into CI infra path.

## Verification

```
node --test .github/scripts/spam-filter.test.mjs
# 9 pass, 0 fail
```

## Deliverables

- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-392-spam-moderation-log.md`
- Branch: `fix/issue-392-spam-moderation-log` (local commit, fleet publishes PR)
