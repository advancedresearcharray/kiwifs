---
memory_kind: episodic
episode_id: cursor-hands-on-406-2026-06-19-delivery
title: PR #406 ADR init template — hands-on delivery verification
tags: [kiwifs, workspace, adr, issue-328, issue-406, hands-on, peer-review, takeover]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#328 / PR #406 — feat(workspace): ship ADR init template with workflow and schema

## Takeover context

Fleet engineer delivery failed: `not_committed`, `peer_review_not_passed`. Overlay FS left
`.git/index` with a stale file handle; staged index contained a partial revert of peer-review
hardening while the working tree matched HEAD (`685f496`).

## Actions

1. Kiwi search — fix doc `pages/fixes/kiwifs-kiwifs/issue-328-adr-init-template.md` already present.
2. Rebuilt git index at `/tmp/kiwifs-index-fresh` via `git read-tree HEAD` (bypass stale handle).
3. Verified peer-review parity with prompt-library template:
   - `TestInitADRIntoEmptyParent`, `TestInitADRDoesNotOverwriteExisting`
   - `TestADRSchemaRejectsInvalidFrontmatter`, `TestADRConfigHasAuthGuidance`
   - `TestBlankADRTemplateHasPlaceholderDeciders`, `TestADRTemplateInitBlankRoot`
   - Auth guidance in `templates/adr/.kiwi/config.toml`
   - `deciders: [team-or-person]` placeholder in blank template
   - SCHEMA.md documents rejected backward/skip/terminal transitions
4. All workspace + cmd tests green; committed fix doc + episode; pushed branch.

## Test output

```
go test ./internal/workspace/... ./cmd/... -count=1 -run 'ADR|InitADR|Init'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.043s
ok  github.com/kiwifs/kiwifs/cmd  0.173s

go test ./internal/workspace/... ./cmd/... -count=1
ok  github.com/kiwifs/kiwifs/internal/workspace  0.043s
ok  github.com/kiwifs/kiwifs/cmd  0.173s
```

## Deliverables

- Peer-review hardening: `685f496`
- Delivery verification commit: hands-on takeover episode + fix doc in repo
- PR: https://github.com/kiwifs/kiwifs/pull/406
