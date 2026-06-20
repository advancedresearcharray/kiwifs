---
memory_kind: semantic
doc_id: claude-builders-bounty-issue-2746-readme-bounty-table-links
title: Fix README Active Bounties table issue links (issue #2746 / PR #2846)
tags: [claude-builders-bounty, readme, issue-links, bounty-table, opire]
repo: claude-builders-bounty/claude-builders-bounty
issue_number: 2746
languages: [markdown, python]
status: verified
date: 2026-06-20
derived-from:
  - type: cursor-auto
    id: pr2846-session227
    date: "2026-06-20T12:00:00Z"
    actor: cursor-auto
  - type: run
    id: cursor-pr-2846-2026-06-20-merge-nurture-verification
    date: "2026-06-20T12:00:00Z"
    actor: agent:cursor-pr-2846
---

# Fix README Active Bounties table issue links

## Problem

The Active Bounties table in README.md used fragile relative links like `[#1](../../issues/1)`. On GitHub forks and blob views, these resolve to the fork's issue tracker (or 404), not the canonical upstream bounty issues.

## Root cause

Relative `../../issues/N` paths depend on the current repo context. Fork viewers clicking bounty links land on wrong/missing issues.

## Solution

Replace all five bounty row links with root-absolute upstream paths:

```markdown
[#1](/claude-builders-bounty/claude-builders-bounty/issues/1)
```

Add a short note above the table explaining why absolute `/owner/repo/issues/N` paths are used (forks and blob views).

## Files changed

- `README.md` — link fix + explanatory note
- `package.json` — npm test entry points
- `tests/test_readme_links.py` — 3 regression tests
- `tests/test_peer_review_acceptance.py` — 11 acceptance tests
- `tests/test_delivery_verification.py` — 9 delivery gate tests
- `tests/run-ci-checks.sh` — unified CI script
- `.github/workflows/test-readme-links.yml` — GitHub Actions workflow
- `.gitignore` — ignores pytest/node caches and internal/exporter/

## Tests

```bash
cd /workspace/claude-builders-bounty/claude-builders-bounty
git checkout fix-issue-2746-opire-try
npm test
# 23/23 pass (verified 2026-06-20 session 227, cursor-auto @ 3ae7129d)
```

## Peer review notes

- All bounty table hrefs are root-absolute (`/claude-builders-bounty/claude-builders-bounty/issues/N`)
- No `../../issues/` fragments remain
- Committed README diff vs `main` is 12+ lines (guards `no_committed_diff`)
- Branch diff stays within README deliverable scope
- PR MERGEABLE/CLEAN; no review blockers
- Latest verification: session 227 (HEAD `3ae7129d`, no new commits required)

## Reuse guide

When README tables link to upstream GitHub issues from a fork PR, use root-absolute paths `/owner/repo/issues/N` instead of relative `../../issues/N`. Add regression tests that diff README vs `main` and assert fragile patterns are removed. See also `pages/fixes/claude-builders-bounty-claude-builders-bounty/issue-521-readme-bounty-table-links.md` for the parallel PR #2532 deliverable.
