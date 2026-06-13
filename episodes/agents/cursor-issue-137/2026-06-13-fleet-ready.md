---
memory_kind: episodic
episode_id: cursor-issue-137-fleet-ready-2026-06-13
title: "Issue #137 — content negotiation verified, ready for fleet PR"
tags: [kiwifs, issue-137, headless-cms, content-negotiation, fleet-ready]
date: 2026-06-13
---

Autonomous verification run for kiwifs/kiwifs#137 on branch `issue-137-content-negotiation` (4 commits ahead of main).

## Reproduction (pre-fix behavior on main)

`GET /p/{path}` on main always returned HTML via `readerTmpl.Execute` with hard-coded `Content-Type: text/html; charset=utf-8`. No Accept header parsing existed.

## Fix summary

Added Accept header content negotiation to the public reader endpoint:

| Accept | Response |
|--------|----------|
| (missing) / `text/html` | Server-rendered HTML (unchanged) |
| `text/markdown` | Raw markdown source with frontmatter |
| `application/json` | `{ frontmatter, html, markdown }` |
| unsupported only | 406 + `Accept: text/html, text/markdown, application/json` |
| CR/LF injection | 400 Bad Request |

## Files changed (branch vs main)

- `internal/api/accept.go` — negotiation helpers (new)
- `internal/api/accept_test.go` — unit tests (new)
- `internal/api/handlers_reader.go` — format branching in PublishedPage
- `internal/api/handlers_reader_test.go` — TestPublishedPageContentNegotiation
- `wiki/UC-4-Headless-CMS.md` — documentation

## Test results

```
go test ./internal/api/ -run 'TestNegotiateReaderFormat|TestSanitizeAcceptHeader|TestParseAcceptEntries|TestPublishedPageContentNegotiation' -count=1 — PASS
go test ./internal/api/ -count=1 — PASS (7.4s)
```

## Kiwi docs

- Searched Kiwi (`kiwi_search`: issue-137 content negotiation) — fix doc found at `pages/fixes/kiwifs-kiwifs/issue-137-content-negotiation.md` (status: verified)
- Kiwi write requires API key; fix doc already complete on cluster from prior run.

## Fleet handoff

Branch clean, all tests green. Fleet agent should push `issue-137-content-negotiation` and open PR closing #137.
