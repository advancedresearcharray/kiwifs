---
memory_kind: episodic
episode_id: cursor-hands-on-436-2026-06-24-pr436-recall-vet-fix
title: "PR #436 hands-on: recall go vet self-assignment fix"
tags: [kiwifs, pr-436, recall, go-vet, ci, issue-422]
date: 2026-06-24
---

## Context

Fleet engineer failed delivery check (`not_committed`, `no_committed_diff`, `peer_review_not_passed`) on kiwifs/kiwifs#436. CI failed `go vet ./...` with self-assignment at `internal/search/recall.go:131`.

## Pre-search

- `http://192.168.167.240:3333/api/kiwi/search?q=recall+go+vet+self-assignment+422` — no indexed fix doc yet.
- Local fix doc existed at `pages/fixes/kiwifs-kiwifs/issue-422-recall-go-vet-self-assignment.md`.

## Actions

1. Cloned fork branch `feat/kiwi-recall-fusion-422` to `/tmp/kiwifs-pr436` (overlay `.git` read-only).
2. Reproduced: `go vet ./internal/search/...` → `self-assignment of err` at line 131.
3. Fixed FTS scope-filter fallback: guard with `if err == nil` before `filterResultsByScope`.
4. Ran tests — all green.
5. Committed `f5d4ec7`, pushed to `advancedresearcharray/kiwifs:feat/kiwi-recall-fusion-422`.
6. Commented on PR #436 with verification details.

## Tests

```bash
go vet ./internal/search/... ./internal/api/... ./internal/mcpserver/...
go test ./internal/search/... -run 'Recall|FuseRRF' -count=1
go test ./internal/api/... -run Recall -count=1
go test ./internal/mcpserver/... -count=1
```

All pass.

## Commits

| Repo | Commit | Branch |
|------|--------|--------|
| advancedresearcharray/kiwifs | f5d4ec7 | feat/kiwi-recall-fusion-422 |

## Kiwi depot

- `kiwi_write` via REST blocked (`401 invalid API key`); fix doc and episode written locally for fleet sync.
