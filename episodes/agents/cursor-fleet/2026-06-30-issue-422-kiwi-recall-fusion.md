---
memory_kind: episodic
episode_id: cursor-fleet-2026-06-30-issue-422
title: "Deliver kiwi_recall fusion retrieval (#422)"
tags: [kiwifs, recall, rrf, issue-422, mcp, search]
date: 2026-06-30
---

## Task

Implement `kiwi_recall` fusion retrieval for kiwifs/kiwifs#422 — RRF merge of FTS, vector, and graph (backlink) signals with REST and MCP exposure.

## Before

- `kiwi_search` via depot API: cluster `http://192.168.167.240:3333` unreachable (connection refused).
- Local fix doc existed only for go-vet self-assignment sub-fix (`pages/fixes/kiwifs-kiwifs/issue-422-recall-go-vet-self-assignment.md`).
- Feature branch `fork/feat/kiwi-recall-fusion-422` had core implementation; `fork/feat/kiwi-recall-422` had MCP tests and empty-query test.

## Work done

1. Checked out `feat/kiwi-recall-422` from `fork/feat/kiwi-recall-fusion-422`.
2. Cherry-picked MCP integration tests (`eb64a82`) and scope/empty-query hardening (`5cefaaf`) from `fork/feat/kiwi-recall-422`.
3. Verified all recall tests pass across `internal/search`, `internal/api`, `internal/mcpserver`.
4. Wrote durable fix doc at `pages/fixes/kiwifs-kiwifs/issue-422-kiwi-recall-fusion.md`.

## Test results

```
go test ./internal/search/... -run 'Recall|FuseRRF|Recaller' -count=1  → ok
go test ./internal/api/... -run Recall -count=1                        → ok
go test ./internal/mcpserver/... -run Recall -count=1                  → ok
```

## Hands-on delivery (2026-06-30, cursor-fleet)

Prior fleet publish failed (`no_committed_diff`, `peer_review_not_passed`) because branch diverged from `origin/main` and included unrelated mkdocs template commit.

1. Rebased recall work onto `origin/main` via cherry-pick (`61bcd26`..`329be8b`).
2. Dropped unrelated sprout-idle-nudge episode files from scope-fix commit.
3. All recall tests and `go vet` green on clean branch.
4. Pushed `feat/kiwi-recall-422` and opened PR closing #422.

## Hands-on delivery (2026-06-30, cursor implementation cycle)

1. Verified `feat/kiwi-recall-422` implements all #422 acceptance criteria (RRF fusion, REST + MCP, provenance, vector-disabled fallback).
2. Removed 14 unrelated `internal/workspace/templates/{knowledge,research,runbook}/` files accidentally bundled in commit `20e2571` (blocked fleet peer review).
3. Added regression tests: `TestFuseRRFConfigurableK`, `TestRecallerScopeFilterSkippedOnSearchError` (locks go-vet self-assignment fix).
4. Kiwi depot cluster (`192.168.167.240:3333`) unreachable; fix doc and episode written to local `pages/` and `episodes/` paths.

## Test results (final)

```
go test ./internal/search/... -run 'Recall|FuseRRF|Recaller' -count=1  → ok
go test ./internal/api/... -run Recall -count=1                        → ok
go test ./internal/mcpserver/... -run Recall -count=1                  → ok
go vet ./internal/search/... ./internal/api/... ./internal/mcpserver/... → ok
```

## Outcome

`kiwi_recall` fusion retrieval shipped: `POST /api/kiwi/recall` + MCP `kiwi_recall` with RRF across FTS, vector, and graph sources.

## Cursor implementation cycle (2026-06-30, feat/issue-422-kiwi-recall-fusion)

1. Created clean branch `feat/issue-422-kiwi-recall-fusion` from `main` (excludes unrelated #421 UI commits on `feat/kiwi-recall-422`).
2. Checked out scoped recall files from `feat/kiwi-recall-422`: search fusion core, REST handler, MCP tool, SQLite `FrontmatterForPaths`, fix doc.
3. Verified all acceptance criteria: RRF fusion (k=60 default), parallel FTS+vector, graph backlink expansion, provenance fields, vector-disabled fallback, `boost_verified` confidence multiplier.
4. Kiwi MCP gateway and cluster depot (`192.168.167.240:3333`) unreachable; fix doc and episode written to local repo paths for fleet publish.

### Test results

```
go test ./internal/search/... ./internal/api/... ./internal/mcpserver/... -count=1  → ok (38s search, 17s api, 16s mcpserver)
go vet ./internal/search/... ./internal/api/... ./internal/mcpserver/...            → ok
```

## Hands-on takeover (2026-06-30)

Fleet delivery check failed (`no_committed_diff`, `peer_review_not_passed`) due to unrelated template commit `3eaf944` on branch.

1. Reset branch to scoped commit `c52affa` (13 files, +1373 lines — recall feature only).
2. Re-ran full recall test suite and `go vet` — all green.
3. Pushed `feat/issue-422-kiwi-recall-fusion` to `fork`.
4. Opened PR https://github.com/advancedresearcharray/kiwifs/pull/55 (upstream `kiwifs/kiwifs` restricts PR creation to collaborators).
5. Kiwi cluster depot (`192.168.167.240:3333`) unreachable; fix doc at `pages/fixes/kiwifs-kiwifs/issue-422-kiwi-recall-fusion.md`.

### Test results (takeover)

```
go test ./internal/search/... ./internal/api/... ./internal/mcpserver/... -count=1  → ok (39s search, 17s api, 17s mcpserver)
go vet ./internal/search/... ./internal/api/... ./internal/mcpserver/...            → ok
```
