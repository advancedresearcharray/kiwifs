---
memory_kind: episodic
episode_id: cursor-issue-347-2026-06-16-peer-review-takeover
title: "PR #361 — strengthen path traversal regression test"
tags: [kiwifs, issue-347, pr-361, custom-css, path-traversal, peer-review, hands-on-takeover]
date: 2026-06-16
---

## Context

Hands-on takeover after fleet engineer `peer_review_blocked`. Prior agent ran unrelated `go test ./internal/exporter/... -run MkDocs` and did not verify the custom CSS regression test. PR #361 adds `TestGetCustomCSS_RejectsPathTraversal` on top of main (feature merged via #357).

## Actions

1. Searched Kiwi depot for existing fix docs (`custom css path traversal 347`).
2. Verified guard in `customCSSRelPath()` rejects `..` and absolute paths; test fails without guard.
3. Strengthened `TestGetCustomCSS_RejectsPathTraversal`:
   - Place sensitive file in parent of workspace temp dir (truly outside root)
   - Table-driven subtests: parent traversal, nested traversal, absolute path
   - Negative assertion that outside content is not leaked
4. Ran full custom CSS test matrix; committed and pushed to `fork/feat/custom-css-347-clean`.

## Test results

```
go test ./internal/api/ -run 'TestGetCustomCSS|TestSanitizeCustomCSS' -count=1 -v — PASS (8 subtests)
go test ./internal/config/ -run TestUIConfigCustomCSS -count=1 -v — PASS
cd ui && npm test -- --run kiwiCustomCss — PASS (2 tests)
```

## Outcome

PR #361 ready for merge; closes #347 with verified path traversal regression coverage.
