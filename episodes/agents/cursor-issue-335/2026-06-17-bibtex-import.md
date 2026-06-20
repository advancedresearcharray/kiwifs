---
memory_kind: episodic
episode_id: cursor-issue-335-2026-06-17
title: BibTeX importer for kiwifs#335
tags: [kiwifs, importer, bibtex, issue-335, fleet]
date: 2026-06-17
---

# BibTeX importer for kiwifs#335

## Task

Implement `kiwifs import --from bibtex --file refs.bib` per [issue #335](https://github.com/kiwifs/kiwifs/issues/335).

## Approach

1. Searched Kiwi depot — no prior bibtex import fix (MCP API key unavailable; checked in-repo `pages/fixes/`).
2. Added `BibTeXSource` with `github.com/nickng/bibtex` parser following CSV/YAML importer pattern.
3. Wired CLI, REST API, upload endpoint, builtin registry, and import wizard UI.
4. Regression tests for article/inproceedings/book, LaTeX unescape, full pipeline write.

## Test results

```
go test ./internal/importer/ -run 'BibTeX|UnescapeBibTeX|ParseBibAuthors|TestAirbyteBuiltinCheck' -count=1 -v  → PASS (6 tests)
```

Verified 2026-06-17: all BibTeX regression tests pass after adding missing `testcontainers-go/modules/*` deps required by `integrations_test.go` package compile.

## Deliverables

- Code ready for fleet PR (closes #335)
- Fix doc: `pages/fixes/kiwifs-kiwifs/issue-335-bibtex-import.md`
- Kiwi depot: fix doc + episode written via HTTP API
- Local commit on `feat/kiwi-cite-336` branch (fleet publishes PR)

## Notes

Importer package tests validate stream, pipeline write, LaTeX unescape, and builtin registry. Complements `kiwi_cite` MCP tool (#336) for bulk `.bib` library import.
