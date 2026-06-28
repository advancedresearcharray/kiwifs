---
memory_kind: episodic
episode_id: cursor-hands-on-325-2026-06-21-verification
title: "Issue #325 — runbook init template final verification"
tags: [kiwifs, runbooks, issue-325, init-template, uc-6, verification]
date: 2026-06-21
---

# Issue #325 — runbook init template final verification

## Context

Autonomous pickup of kiwifs/kiwifs#325 on branch `feat/issue-325-runbook-init-template`.
Implementation landed in commits `79c770e` (scaffold), `7e124e9` (fix doc), `72c174f`
(check regression test). This session re-verified all acceptance criteria before fleet publish.

## Pre-search

- `kiwi_search` via `http://192.168.167.240:3333/api/kiwi/search?q=runbook+init+template+325`
  — semantic fix doc and fleet episodes indexed.
- Read `pages/fixes/kiwifs-kiwifs/issue-325-runbook-init-template.md`.

## Verification

1. Confirmed UC-6 scaffold under `internal/workspace/templates/runbook/`:
   `SCHEMA.md`, `index.md`, `example-high-cpu.md`, `.kiwi/schemas/runbook.json`,
   `.kiwi/templates/runbook.md`, `.kiwi/config.toml`, `playbook.md`.
2. `runbook` registered in `cmd/init.go` flag help and `internal/workspace/init.go` switch.
3. `go test ./internal/workspace/... ./cmd/... -run 'Runbook|runbook' -count=1` — PASS.
4. `go test ./... -count=1` — PASS (~61s).
5. Manual `go run . init --root /tmp/runbooks --template runbook` + `go run . check` — exit 0
   (8 info-level issues on README/SCHEMA/playbook; no errors).

## Acceptance criteria

| Criterion | Status |
|-----------|--------|
| `kiwifs init --template runbook` scaffolds workspace | PASS |
| Example runbook has 7 sections + fenced code blocks | PASS |
| JSON Schema validates required frontmatter | PASS |
| `kiwifs check` passes on generated scaffold | PASS |

## Kiwi sync

- `kiwi_write` via REST blocked (`invalid API key`); fix doc and episode written locally for
  fleet sync (fix doc already indexed on cluster from prior session).

## Outcome

Issue #325 complete. Fleet agent should push `feat/issue-325-runbook-init-template` and open
PR closing #325.
