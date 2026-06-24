---
memory_kind: semantic
doc_id: kiwifs-kiwifs-issue-422-recall-go-vet-self-assignment
title: "PR #436 CI go vet self-assignment in recall scope filter"
tags: [kiwifs, search, recall, go-vet, ci, issue-422, pr-436]
repo: kiwifs/kiwifs
issue_number: 422
languages: [go]
status: fixed
date: 2026-06-24
---

## Problem

PR #436 (`feat/kiwi-recall-fusion-422`) and duplicate PR #437 failed CI on `go vet ./...` with:

```
internal/search/recall.go:131:6: self-assignment of err
```

## Root cause

In the FTS retrieval goroutine, when scope filtering fell back to the non-`OptionsSearcher` path, the code used a comma-assignment that re-assigned `err` to itself:

```go
results, err = r.Searcher.Search(...)
results, err = filterResultsByScope(...), err  // vet: self-assignment of err
```

`filterResultsByScope` returns `[]Result` only (no error). The trailing `, err` made the RHS evaluate to `err`, producing `err = err`.

## Solution

Apply scope filtering only after a successful search:

```go
results, err = r.Searcher.Search(ctx, opts.Query, fetchLimit, 0, opts.PathPrefix)
if err == nil {
    results = filterResultsByScope(ctx, r.Searcher, results, opts.Scope)
}
```

## Files changed

| File | Change |
|------|--------|
| `internal/search/recall.go` | Fix scope-filter assignment in FTS fallback branch (line ~131) |

## Tests

Verified after push (`f5d4ec7` on `feat/kiwi-recall-fusion-422`):

```bash
go vet ./internal/search/... ./internal/api/... ./internal/mcpserver/...
go test ./internal/search/... -run 'Recall|FuseRRF' -count=1
go test ./internal/api/... -run Recall -count=1
go test ./internal/mcpserver/... -count=1
```

All pass.

## Peer review notes

- Same bug present on both PR #436 and PR #437 branches; one-line fix unblocks both.
- No behavior change when scope is empty; filtering only runs when search succeeds.

## Reuse guide

When combining multi-value assignment with helper functions that return a single value, never append `, err` unless the helper returns an error. Run `go vet ./...` locally before pushing recall/search changes.
