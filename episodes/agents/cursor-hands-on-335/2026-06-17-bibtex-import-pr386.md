---
memory_kind: episodic
episode_id: cursor-hands-on-335-2026-06-17
title: BibTeX import PR #386 for kiwifs#335
tags: [kiwifs, importer, bibtex, issue-335, pr-386, fleet]
date: 2026-06-17
---

# BibTeX import PR #386

Hands-on delivery verified BibTeX importer on clean branch cherry-picked from `14cea07` onto `origin/main`.

## Deliverables

- PR: https://github.com/kiwifs/kiwifs/pull/386 (closes #335)
- Branch: `feat/bibtex-import-335` on `advancedresearcharray/kiwifs`
- Fix doc indexed at `pages/fixes/kiwifs-kiwifs/issue-335-bibtex-import.md` (Kiwi depot search confirmed)

## Test results

```
go test ./internal/importer/ -run 'BibTeX|UnescapeBibTeX|ParseBibAuthors|TestAirbyteBuiltinCheck' -count=1 -v  → PASS (6 tests)
go test ./internal/importer/... -count=1  → PASS
```

## Notes

- Cherry-pick resolved `go.mod`/`go.sum` conflicts via `go get github.com/nickng/bibtex@v1.1.0` + `go mod tidy`
- Kiwi depot write API requires key; fix doc already present from prior fleet run
