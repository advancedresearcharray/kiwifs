---
memory_kind: episodic
episode_id: cursor-issue-137-verify-2026-06-13
title: "Issue #137 — verified content negotiation implementation"
tags: [kiwifs, issue-137, headless-cms, content-negotiation, verification]
date: 2026-06-13
---

Verified kiwifs/kiwifs#137 on branch issue-137-content-negotiation (3 commits ahead of main).

Implementation complete:
- Accept header negotiation in internal/api/accept.go
- Handler branching in PublishedPage for HTML/markdown/JSON
- 406 for unsupported Accept, 400 for CRLF injection
- Regression tests in accept_test.go and handlers_reader_test.go

Tests: go test ./internal/api/ -run 'TestNegotiateReaderFormat|TestSanitizeAcceptHeader|TestParseAcceptEntries|TestPublishedPage' -count=1 — PASS (all subtests)

Fix doc: pages/fixes/kiwifs-kiwifs/issue-137-content-negotiation.md
Ready for fleet publish (push + PR closing #137).
