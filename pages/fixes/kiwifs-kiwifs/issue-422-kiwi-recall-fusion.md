---
memory_kind: semantic
doc_id: kiwifs-kiwifs-issue-422-kiwi-recall-fusion
title: "kiwi_recall fusion retrieval (FTS + vector + graph RRF)"
tags: [kiwifs, search, recall, rrf, mcp, issue-422, fusion]
repo: kiwifs/kiwifs
issue_number: 422
languages: [go]
status: fixed
date: 2026-06-30
---

## Problem

Agents had to call `kiwi_search` (FTS5), `kiwi_search_semantic` (vector), and `kiwi_backlinks` (graph) separately for memory retrieval. There was no single endpoint that fused all three signals with reciprocal rank fusion (RRF), which is standard practice in RAG systems.

## Root cause

Search primitives existed independently (`internal/search` FTS, `internal/vectorstore` semantic search, `links.Linker` backlinks) but no orchestration layer merged ranked lists or exposed a unified recall API/MCP tool.

## Solution

1. **`internal/search/recall.go`** — `Recaller` runs FTS and vector retrieval in parallel via `errgroup`, then expands graph signal from top FTS/vector seeds via 1-hop backlinks. Results merge with `FuseRRF` (`score = Σ 1/(k + rank)`, default `k=60`). Optional `boost_verified` multiplies fused score by frontmatter `confidence` (0.0–1.0).
2. **`POST /api/kiwi/recall`** — JSON body: `query`, `limit`, `sources` (`fts`/`vector`/`graph`), `scope`, `boost_verified`, `k`, `path_prefix`. Returns unified ranked list with per-source provenance (`fts_rank`, `vector_rank`, `graph_rank`).
3. **`kiwi_recall` MCP tool** — Same parameters; registered in `mcpserver.go` with `handleRecall`.
4. **Graceful degradation** — When `vectorstore.Service` is nil/disabled, vector source is skipped silently; FTS-only fusion still works.
5. **Adapters** — `buildRecaller` in API and `LocalBackend.Recall` wire `Searcher`, `VectorSearcher`, `BacklinkFinder`, and `MetaReader` (SQLite `FrontmatterForPaths`).

## Files changed

| File | Change |
|------|--------|
| `internal/search/recall.go` | RRF fusion, parallel retrieval, graph backlink expansion, scope/confidence helpers |
| `internal/search/recall_test.go` | Unit tests for RRF, FTS-only fallback, boost_verified, graph seeds, empty query |
| `internal/api/handlers_recall.go` | `POST /api/kiwi/recall` handler and adapter types |
| `internal/api/handlers_recall_test.go` | Integration tests for FTS fusion, query validation, vector-disabled fallback |
| `internal/api/server.go` | Route registration |
| `internal/mcpserver/mcpserver.go` | `kiwi_recall` tool registration and `handleRecall` |
| `internal/mcpserver/backend.go` | `RecallParams`, `RecallResult`, `Backend.Recall` interface |
| `internal/mcpserver/local.go` | `LocalBackend.Recall` orchestration |
| `internal/mcpserver/client.go` | `RemoteBackend.Recall` HTTP client |
| `internal/mcpserver/recall_tools_test.go` | MCP handler tests |
| `internal/search/sqlite.go` | `FrontmatterForPaths` for title/confidence boosts |

## Tests

```bash
go test ./internal/search/... -run 'Recall|FuseRRF|Recaller' -count=1
go test ./internal/api/... -run Recall -count=1
go test ./internal/mcpserver/... -run Recall -count=1
go vet ./internal/search/... ./internal/api/... ./internal/mcpserver/...
```

All pass. Full `go test ./...` passes except unrelated workspace template lint (broken example links in runbook/research templates).

## Peer review notes

- Scope filter in FTS fallback path must not self-assign `err` — use `if err == nil { results = filterResultsByScope(...) }` (see `issue-422-recall-go-vet-self-assignment.md`).
- Redundant `searchOpts.Scope != ""` guard removed; only `opts.Scope != ""` triggers scoped search path.
- Graph signal ranks 1-hop backlinks of top FTS/vector seeds by inbound link count from multiple seeds.
- MCP `limit` capped at 50 (consistent with `kiwi_search`).
- Hands-on delivery: branch must be rebased onto current `origin/main` (not stale fork branch); exclude unrelated template/mkdocs commits from the PR.

## Reuse guide

- **Agent memory retrieval**: prefer `kiwi_recall` over separate `kiwi_search` + `kiwi_search_semantic` + `kiwi_backlinks` calls.
- **Sources**: omit `sources` to fuse all three; pass `["fts"]` for keyword-only; `["fts","graph"]` when vectors disabled.
- **Scope**: set `scope` to match frontmatter `scope` field (e.g. `semantic`).
- **Verified boost**: `boost_verified: true` applies multiplicative `confidence` from frontmatter.
- **RRF constant**: override with `k` (default 60 per Cormack et al.).
- **REST**: `POST /api/kiwi/recall` with JSON body; **MCP**: `kiwi_recall` tool with same fields.
