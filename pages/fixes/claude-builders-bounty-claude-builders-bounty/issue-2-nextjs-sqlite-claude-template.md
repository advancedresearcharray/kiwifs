---
memory_kind: semantic
doc_id: claude-builders-bounty-issue-2-nextjs-sqlite-claude-template
title: Next.js 15 + SQLite SaaS CLAUDE.md template (issue #2 / PR #2529)
tags: [claude-builders-bounty, nextjs, sqlite, claude-md, template, bounty, opire, pr-2529]
repo: claude-builders-bounty/claude-builders-bounty
issue_number: 2
languages: [markdown, javascript, typescript, shell]
status: verified
date: 2026-06-21
derived-from:
  - type: cursor-auto
    id: pr2529-sprout-idle-nudge-2026-06-21
    date: "2026-06-21T09:02:00Z"
    actor: cursor-auto
---

# Next.js 15 + SQLite SaaS CLAUDE.md template

## Problem

Bounty #2 requires an opinionated, production-ready `CLAUDE.md` for greenfield Next.js 15 App Router
SaaS projects backed by SQLite (`better-sqlite3` locally, Turso in production). The template must cover
project structure, naming conventions, DB migration rules, dev commands, patterns, anti-patterns (with
reasons), auth, route handlers, and testing — usable without editing.

## Root cause

No canonical template existed in the repo. Competing submissions lacked structural validation,
greenfield smoke tests, or fleet delivery gates to prevent scope drift and attribution violations.

## Solution

Ship `templates/nextjs-sqlite-saas/` with:

- **788-line `CLAUDE.md`** — stack table, annotated folder tree, SQL/migration conventions, component
  patterns, 16 anti-pattern rows (each with reason), auth/authorization, webhook-only API routes,
  testing rules, greenfield smoke-test checklist
- **Reference app** (`reference/`) — org-scoped invoices, server actions (create/void), migrations,
  Vitest suites (20 tests), seed script with production guard
- **Structural validator** (`validate-template.mjs`) — 63 explicit reasons, 5 smoke-test steps
- **Greenfield smoke test** (`tests/smoke-greenfield.sh`) — scaffolds fresh `create-next-app@15`,
  copies CLAUDE.md, asserts all bounty sections
- **7-stage CI runner** (`tests/run-ci-checks.sh` + `ci-step-*.sh`) — security audits, structural,
  reference, delivery, peer-review, attribution
- **Fleet delivery gates** — `no_committed_diff` and `peer_review_not_passed` in
  `delivery-verification.test.mjs`; branch diff excludes stray artifacts (`.cache`, root `package-lock.json`)
- **GitHub Actions** (`.github/workflows/test-nextjs-sqlite-template.yml`) — `pull_request_target` for
  fork PRs, read-only permissions

Branch: `bounty-2-nextjs-sqlite-claude-template`. PR #2529 closes issue #2.

## Files changed

Key paths (65 files, +7392 lines vs `main`):

- `templates/nextjs-sqlite-saas/CLAUDE.md` — primary deliverable
- `templates/nextjs-sqlite-saas/README.md` — 3-step usage guide
- `templates/nextjs-sqlite-saas/validate-template.mjs` — structural validator
- `templates/nextjs-sqlite-saas/reference/` — reference app + tests
- `templates/nextjs-sqlite-saas/tests/` — smoke, delivery, peer-review, security audits
- `scripts/run-ci-checks.sh`, `scripts/env-check.mjs` — root CI entry points
- `package.json`, `Makefile` — `npm test`, `test:delivery`, `test:peer-review`
- `.github/workflows/test-nextjs-sqlite-template.yml` — CI workflow

## Tests

```bash
cd /workspace/claude-builders-bounty/claude-builders-bounty
git checkout bounty-2-nextjs-sqlite-claude-template

npm test                    # 210/210 — full CI (verified @ 1bcdbf4a)
npm run test:delivery       # 30/30 — fleet delivery gates
npm run test:peer-review    # 161 asserts — peer review APPROVED
npm run test:claude-md      # 15 section checks

# Isolated template checks
node templates/nextjs-sqlite-saas/validate-template.mjs
bash templates/nextjs-sqlite-saas/tests/smoke-greenfield.sh
bash templates/nextjs-sqlite-saas/tests/run-ci-checks.sh
```

## Peer review notes

- Peer review status: **APPROVED** (locked in delivery gates)
- Common blockers already addressed in branch history:
  - Forbidden `Co-authored-by: Cursor` trailers (attribution guard)
  - Stray `.cache` submodule and root `package-lock.json` (delivery gate failures)
  - `env-check.mjs` hardened (`spawnSync`, `shell: false`, argv whitelist)
  - Folder READMEs and CLAUDE.md Table of Contents (peer-review feedback in `d96d33b2`)
- Upstream PR shows `mergeable: MERGEABLE`, `mergeStateStatus: CLEAN`; no blocking reviews

## Reuse guide

When nurturing PR #2529 or re-verifying bounty #2:

1. Checkout `bounty-2-nextjs-sqlite-claude-template` (not `main` or unrelated bounty branches).
2. Run `npm test` — expect 210 pass; failures usually mean stray untracked files
   (`package-lock.json` at repo root) or attribution trailer on HEAD commit.
3. If delivery gates fail on `no_committed_diff`: ensure `templates/nextjs-sqlite-saas/CLAUDE.md`
   diff vs `main` is committed and branch excludes non-template artifacts.
4. Do **not** rewrite working template code when CI/reviews are clean — post merge-ready signal only.
5. Fleet agent handles push, PR comment, and `/opire try` on issue #2.
