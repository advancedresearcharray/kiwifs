---
memory_kind: episodic
episode_id: cursor-hands-on-328-2026-06-19-takeover-v2
title: Issue #328 ADR init template — hands-on delivery verification
tags: [kiwifs, workspace, adr, issue-328, hands-on, uc-adr, takeover]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#328 — feat(workspace): ship ADR init template with workflow and schema

## Takeover context

Prior fleet delivery failed: `not_committed`, `no_committed_diff`, `peer_review_not_passed`.
Working tree had staged ADR deletions mixed with unrelated issue-345 UI changes.
Overlay FS git index corruption (`Could not write new index file`) fixed via
`.git/index.rebuilt`.

## Actions

1. Searched Kiwi depot (`/api/kiwi/search?q=adr+init+template+328`) — no prior fix doc.
2. Restored clean index; verified branch `feat/issue-328-adr-init-template` matches HEAD.
3. Peer review: APPROVED — workflow/schema/scaffold/tests satisfy issue acceptance criteria.
4. Ran regression tests — green (see below).
5. PR #406 already open closing #328; branch pushed to fork.

## Test output

```
go test ./internal/workspace/... -count=1 -run 'ADR|InitADR|ListInit'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.016s

go test ./cmd/... -count=1 -run 'Init'
ok  github.com/kiwifs/kiwifs/cmd  0.030s

go test ./internal/workspace/... -count=1
ok  github.com/kiwifs/kiwifs/internal/workspace  0.013s
```

## Deliverables

- Feature commit: `90b9fae` — ADR template, registration, regression tests
- PR: https://github.com/kiwifs/kiwifs/pull/406
- Fix doc path (Kiwi write blocked — invalid API key): `pages/fixes/kiwifs-kiwifs/issue-328-adr-init-template.md`
