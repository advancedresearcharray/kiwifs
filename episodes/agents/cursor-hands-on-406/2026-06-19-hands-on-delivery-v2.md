---
memory_kind: episodic
episode_id: cursor-hands-on-406-2026-06-19-delivery-v2
title: PR #406 ADR init template — hands-on delivery v2 (index fix + publish)
tags: [kiwifs, workspace, adr, issue-328, issue-406, hands-on, peer-review, takeover, overlay-fs]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#328 / PR #406 — feat(workspace): ship ADR init template with workflow and schema

## Takeover context

Fleet engineer delivery failed: `not_committed`, `no_committed_diff`, `peer_review_not_passed`.
Overlay FS left `.git/index` with a stale file handle; default index staged a partial revert of
peer-review hardening (`685f496`) while the working tree matched HEAD.

## Actions

1. Kiwi search — fix doc `pages/fixes/kiwifs-kiwifs/issue-328-adr-init-template.md` present.
2. Rebuilt git index via `GIT_INDEX_FILE=/tmp/kiwifs-index-commit git read-tree HEAD` and
   replaced stale `.git/index` copy.
3. Verified peer-review hardening intact at HEAD (`685f496`):
   - `TestInitADRIntoEmptyParent`, `TestInitADRDoesNotOverwriteExisting`
   - `TestADRSchemaRejectsInvalidFrontmatter`, `TestADRConfigHasAuthGuidance`
   - `TestBlankADRTemplateHasPlaceholderDeciders`, `TestADRTemplateInitBlankRoot`
   - Auth guidance in `templates/adr/.kiwi/config.toml`
   - `deciders: [team-or-person]` placeholder in blank template
   - SCHEMA.md documents rejected backward/skip/terminal transitions
4. Ran full workspace + cmd test suites — all green.
5. Updated fix doc with overlay FS index workaround; committed and pushed branch.

## Test output

```
go test ./internal/workspace/... ./cmd/... -count=1 -run 'ADR|InitADR|Init'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.022s
ok  github.com/kiwifs/kiwifs/cmd  0.029s

go test ./internal/workspace/... ./cmd/... -count=1
ok  github.com/kiwifs/kiwifs/internal/workspace  0.033s
ok  github.com/kiwifs/kiwifs/cmd  0.160s
```

## Outcome

Peer-review hardening verified in committed tree. Git index corruption resolved. Branch pushed;
PR #406 CI green, merge-ready.
