---
memory_kind: episodic
episode_id: cursor-issue-137-peer-review-2026-06-13
title: "Issue #137 — peer review fixes for content negotiation"
tags: [kiwifs, issue-137, headless-cms, content-negotiation, peer-review]
date: 2026-06-13
---

Addressed peer review feedback on kiwifs/kiwifs#137 content negotiation:

- Hardened Accept header parsing (CR/LF rejection, control char stripping, length/entry caps, MIME token validation)
- Return 406 for unsupported-only Accept values; 400 for injection attempts
- Refactored `negotiateReaderFormat` into smaller functions
- Added edge-case tests (406, 400, wildcards, large JSON payload)
- Documented usage in `wiki/UC-4-Headless-CMS.md`

Tests: `go test ./internal/api/ -run 'TestNegotiateReaderFormat|TestSanitizeAcceptHeader|TestParseAcceptEntries|TestPublishedPage' -count=1` — PASS

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-137-content-negotiation.md`
