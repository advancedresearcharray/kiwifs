---
memory_kind: episodic
episode_id: cursor-issue-327-2026-06-16
title: "Issue #327 — verify PATCH merge=frontmatter and publish PR"
tags: [kiwifs, api, frontmatter, issue-327, verification]
date: 2026-06-16
---

## Task

Hands-on takeover for kiwifs/kiwifs#327 after fleet agent failed delivery check (no push, no PR).

## Verification

Re-ran regression and full API suite on branch `issue-327-frontmatter-patch` (commit `dadec24`):

```
go test ./internal/api/... -run 'Patch(File|Frontmatter)' -count=1 -v  # 9 tests PASS
go test ./internal/api/... -count=1                                      # full suite PASS (7.664s)
```

## Delivery

- Pushed branch to `fork/issue-327-frontmatter-patch`
- Opened PR against kiwifs/kiwifs main
- Updated Kiwi fix doc with peer review + test output

## Acceptance criteria

All met: merge=frontmatter PATCH, body byte preservation, If-Match 409/200, git commit, 404 missing file, add/update field tests.
