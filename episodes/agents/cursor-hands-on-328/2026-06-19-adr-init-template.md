---
memory_kind: episodic
episode_id: cursor-hands-on-328-2026-06-19-takeover
title: Issue #328 ADR init template — hands-on takeover delivery
tags: [kiwifs, workspace, adr, issue-328, hands-on, uc-adr, takeover]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#328 — feat(workspace): ship ADR init template with workflow and schema

## Takeover context

Prior fleet delivery failed: `not_committed`, `peer_review_not_passed`. Working tree
had staged ADR deletions on wrong branch (`feat/issue-334-research-library-template`).
Stale git index on overlay FS required `GIT_INDEX_FILE=.git/index.rebuilt`.

## Actions

1. Searched Kiwi depot (`/api/kiwi/search?q=adr+init+template+328`) — no prior fix doc indexed.
2. Rebuilt clean branch `feat/issue-328-adr-init-template` from `origin/main`.
3. Cherry-picked ADR commit; resolved `init_test.go` conflicts (ADR only, no research paths).
4. Ran regression tests — green.
5. Pushed to fork and opened PR #406 closing #328.
6. Wrote fix doc and episode to Kiwi depot via REST API.

## Test output

```
go test ./internal/workspace/... -count=1 -run 'ADR|InitADR|ListInit'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.007s

go test ./cmd/... -count=1 -run 'Init'
ok  github.com/kiwifs/kiwifs/cmd  0.030s
```

## Deliverables

- Commit: `90b9fae` on `feat/issue-328-adr-init-template`
- PR: https://github.com/kiwifs/kiwifs/pull/406
