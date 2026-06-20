---
memory_kind: episodic
episode_id: cursor-pr-399-2026-06-19-path-prefix
title: "PR #399 hands-on — MkDocs PathPrefix boundary fix"
tags: [kiwifs, pr-399, issue-103, mkdocs, exporter, path-prefix, hands-on]
date: 2026-06-19
---

## Context

Peer review on PR #399 flagged PathPrefix boundary bug in MkDocs exporter.

## Actions

1. Peer review (bugbot) flagged PathPrefix boundary bug: `strings.HasPrefix("pages-extra/foo", "pages")` returned true.
2. Implemented `pathUnderPrefix()` in `internal/exporter/mkdocs.go` with directory-boundary semantics.
3. Added `TestPathUnderPrefix` and `TestExportMkDocsPathPrefix`.
4. Ran tests:
   - `go test ./internal/exporter/... ./cmd/... -race -count=1 -run 'MkDocs|PathUnder|PathPrefix'`

## Outcome

Prefix `pages` no longer matches `pages-extra/foo.md`.
