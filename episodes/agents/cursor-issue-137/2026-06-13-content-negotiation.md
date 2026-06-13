---
memory_kind: episodic
episode_id: cursor-issue-137-2026-06-13
title: "Issue #137 — public reader content negotiation"
tags: [kiwifs, issue-137, headless-cms, content-negotiation]
date: 2026-06-13
---

Implemented kiwifs/kiwifs#137: `GET /p/{path}` now negotiates response format via the `Accept` header (`text/html` default, `text/markdown` raw source, `application/json` structured payload).

Tests passed:
- `go test ./internal/api/ -run 'TestNegotiateReaderFormat|TestPublishedPage' -count=1`

Fix doc: `pages/fixes/kiwifs-kiwifs/issue-137-content-negotiation.md`

Note: Kiwi MCP gateway unavailable; remote Kiwi write at CT934 returned `invalid API key`. Fix doc written to workspace `pages/` and `episodes/` trees directly.
