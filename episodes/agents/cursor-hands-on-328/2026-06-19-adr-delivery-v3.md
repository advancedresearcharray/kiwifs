---
memory_kind: episodic
episode_id: cursor-hands-on-328-2026-06-19-delivery-v3
title: Issue #328 ADR init template — verified delivery with cmd tests
tags: [kiwifs, workspace, adr, issue-328, hands-on, uc-adr, takeover]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#328 — feat(workspace): ship ADR init template with workflow and schema

## Actions

1. Kiwi search (`/api/kiwi/search?q=adr+init+template+328`) — no prior fix doc in depot.
2. Verified feature commit `90b9fae` and template scaffold on branch `feat/issue-328-adr-init-template`.
3. Added `TestADRTemplateEmbedded` and `TestADRTemplateInit` in `cmd/init_test.go` (CLI-layer regression).
4. Rebuilt git index via `GIT_INDEX_FILE=.git/index.new` after overlay stale-handle failure.
5. Committed `ae2a445`, pushed to fork; PR #406 updated.
6. Wrote fix doc to Kiwi depot.

## Test output

```
go test ./cmd/... -count=1 -run 'ADR|Init' -v
--- PASS: TestADRTemplateEmbedded (0.00s)
--- PASS: TestADRTemplateInit (0.00s)
PASS ok github.com/kiwifs/kiwifs/cmd 0.031s

go test ./internal/workspace/... -count=1 -run 'ADR|InitADR|ListInit' -v
PASS ok github.com/kiwifs/kiwifs/internal/workspace 0.008s
```

## Deliverables

- Feature: `90b9fae` — ADR template, workflow, schema, workspace tests
- Tests: `ae2a445` — cmd init regression tests
- PR: https://github.com/kiwifs/kiwifs/pull/406
