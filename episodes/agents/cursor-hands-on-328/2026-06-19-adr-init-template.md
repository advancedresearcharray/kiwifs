---
memory_kind: episodic
episode_id: cursor-hands-on-328-2026-06-19
title: Issue #328 ADR init template delivery
tags: [kiwifs, workspace, adr, issue-328, hands-on, uc-adr]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#328 — feat(workspace): ship ADR init template with workflow and schema

## Actions

1. Searched Kiwi depot (`/api/kiwi/search?q=adr+init+template+328`) — no prior fix doc.
2. Reviewed issue #328, UC-7 wiki, research template pattern, and wiki ADR examples.
3. Created `internal/workspace/templates/adr/` with SCHEMA, playbook, workflow,
   schema, MADR template, example ADR-001, and index.
4. Registered `adr` in `workspace.Init` and `cmd/init.go`.
5. Added `adr_template_test.go` regression tests; updated `init_test.go`.
6. Fixed lint: wikilink pipe syntax in index; removed null `supersedes` keys from example.

## Test output

```
go test ./internal/workspace/... -count=1 -run 'ADR|InitADR|ListInit'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.007s

go test ./cmd/... -count=1 -run 'Init'
ok  github.com/kiwifs/kiwifs/cmd  0.372s
```

Manual: `kiwifs init --root /tmp/adr-init-test --template adr` — scaffold verified.

## Notes

- Kiwi MCP server unavailable in session; fix doc written to repo `pages/fixes/`.
- Fleet agent to commit, push, and open PR closing #328.
