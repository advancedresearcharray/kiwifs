---
memory_kind: episodic
episode_id: cursor-issue-137-verify-2026-06-13
title: "Issue #137 — verified content negotiation implementation"
tags: [kiwifs, issue-137, headless-cms, content-negotiation, verification]
date: 2026-06-13
---

Verified kiwifs/kiwifs#137 on branch `issue-137-content-negotiation` (5 commits ahead of main).

Implementation complete:
- Accept header negotiation in `internal/api/accept.go`
- Handler branching in `PublishedPage` for HTML / markdown / JSON
- 406 for unsupported Accept, 400 for CRLF injection
- Regression tests in `accept_test.go` and `handlers_reader_test.go`
- UC-4 wiki updated with usage examples

Cleanup: removed unrelated mysql importer commit (44a08cb) from branch tip.

Tests:
- `go test ./internal/api/ -run 'TestNegotiateReaderFormat|TestSanitizeAcceptHeader|TestParseAcceptEntries|TestPublishedPage' -count=1` — PASS
- `go test ./internal/api/... -count=1` — PASS

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-137-content-negotiation.md` (on Kiwi depot; gitignored in kiwifs repo)
Kiwi MCP unavailable; search via `http://192.168.167.240:3333/api/kiwi/search` confirms fix doc indexed.

Ready for fleet publish (push + PR closing #137).
