---
memory_kind: episodic
episode_id: cursor-hands-on-328-2026-06-19-peer-review
title: Issue #328 ADR init template — peer-review hardening
tags: [kiwifs, workspace, adr, issue-328, hands-on, peer-review, takeover]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#328 — feat(workspace): ship ADR init template with workflow and schema
PR: https://github.com/kiwifs/kiwifs/pull/406

## Takeover context

Fleet engineer `peer_review_blocked`. Prior agent ran MkDocs exporter tests only;
feature code was present but lacked prompt-library-style peer-review coverage.

## Actions

1. Kiwi search — no prior fix doc for issue #328.
2. Hardened ADR template peer-review coverage:
   - Auth guidance in `templates/adr/.kiwi/config.toml`
   - Blank template `deciders` placeholder fix
   - SCHEMA.md backward/terminal transition documentation
   - Workspace tests: empty parent init, no-overwrite, schema rejection matrix, config auth, blank template
   - Cmd test: `TestADRTemplateInitBlankRoot`
3. All ADR regression tests green locally; remote CI already green on prior push.

## Test output

```
go test ./internal/workspace/... -count=1 -run 'ADR|InitADR'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.009s

go test ./cmd/... -count=1 -run 'ADR|Init'
ok  github.com/kiwifs/kiwifs/cmd  0.030s
```
