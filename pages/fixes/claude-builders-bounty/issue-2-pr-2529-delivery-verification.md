---
memory_kind: semantic
doc_id: fix-claude-builders-bounty-issue-2-pr-2529
title: "PR #2529 bounty #2 — Next.js 15 SQLite CLAUDE.md template delivery verification"
tags: [claude-builders-bounty, issue-2, pr-2529, nextjs, sqlite, claude-md, delivery-gates, merge-nurture]
repo: claude-builders-bounty/claude-builders-bounty
issue_number: 2
languages: [javascript, typescript, bash, markdown]
status: verified
date: 2026-06-24
---

## Problem

Bounty #2 requires an opinionated `CLAUDE.md` template for greenfield Next.js 15 App Router SaaS projects
backed by SQLite. PR #2529 is long-lived merge-to-pay work that must stay merge-ready with fleet delivery
gates green.

## Root cause

N/A — this is a delivery verification guide, not a bug fix. Common blockers observed across nurture cycles:

1. **Fleet delivery gates** — `no_committed_diff` and `peer_review_not_passed` fail if CLAUDE.md diff vs `main` is unstaged or peer-review criteria regress.
2. **Vitest sequential CI flake** — running migration tests then full suite in separate `vitest run` invocations can hit ENOENT on shared `/tmp` Vite cache (fixed with project-local `cacheDir` + `pool: "forks"`).
3. **Stray branch artifacts** — unrelated files (e.g. `bounties/issue-2853/sessions/*.patch`, root `package-lock.json`) break peer-review branch scope gates.
4. **Attribution guard** — `Co-authored-by: Cursor` trailers in commits fail CI.

## Solution

Verify on branch `bounty-2-nextjs-sqlite-claude-template`:

```bash
cd /workspace/claude-builders-bounty/claude-builders-bounty
git checkout bounty-2-nextjs-sqlite-claude-template
npm test                  # full CI (236 tests as of 2026-06-24)
npm run test:delivery     # 36 delivery gates
npm run test:peer-review  # 172 peer-review asserts
```

Confirm PR state:

```bash
gh pr view 2529 --json mergeable,mergeStateStatus
# expect: MERGEABLE, CLEAN
```

If CI fails on reference Vitest ENOENT between sequential runs, confirm `reference/vitest.config.ts` has:

```ts
cacheDir: "node_modules/.vitest",
test: { pool: "forks" },
```

Optional hardening: set `TMPDIR` to project-local in `reference/scripts/run-tests.sh`.

## Files changed

No code changes required when gates are green (verified 2026-06-24 @ `6c37c203`).

Key deliverable paths:

- `templates/nextjs-sqlite-saas/CLAUDE.md` — 884-line opinionated template
- `templates/nextjs-sqlite-saas/reference/` — reference app with Vitest suites
- `templates/nextjs-sqlite-saas/tests/` — structural validation, greenfield smoke, delivery + peer-review gates
- `.github/workflows/test-nextjs-sqlite-template.yml` — CI workflow

## Tests

| Command | Expected (2026-06-24) |
|---------|----------------------|
| `npm test` | 236/236 pass |
| `npm run test:delivery` | 36/36 pass |
| `npm run test:peer-review` | 172 asserts pass |
| `node templates/nextjs-sqlite-saas/validate-template.mjs` | 87 explicit reasons, 5 smoke steps |
| `bash templates/nextjs-sqlite-saas/tests/smoke-greenfield.sh` | greenfield create-next-app@15 OK |

Fork CI: `advancedresearcharray/claude-builders-bounty` workflow `test-nextjs-sqlite-template` — green on `6c37c203`.

## Peer review notes

- Peer review status: **APPROVED** (locked in delivery gates).
- PR has no blocking human reviews; merge waits on upstream maintainer.
- Do not rewrite working template code when CI/reviews are clean — refresh verification timestamps only if fleet requires a new commit.

## Reuse guide

1. Search Kiwi for `pr-2529` or `issue-2` before re-verifying.
2. Checkout `bounty-2-nextjs-sqlite-claude-template`, not other bounty branches.
3. Run `npm test` from repo root (runs env-check → 7-stage CI runner).
4. If gates fail, fix only the blocking gate — do not open duplicate PRs.
5. Fleet agent publishes: push, merge-ready PR comment, `/opire try` on issue #2.
