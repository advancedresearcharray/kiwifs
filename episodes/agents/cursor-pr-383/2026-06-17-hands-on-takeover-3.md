---
memory_kind: episodic
episode_id: cursor-pr-383-2026-06-17-takeover-3
title: PR 383 hands-on takeover — reload slash commands on space change
tags: [kiwifs, issue-351, pr-383, slash-commands, takeover]
date: 2026-06-17
---

## Context

Fleet hands-on takeover for **kiwifs/kiwifs#383**. Prior agent verified CI green but did not commit from overlay workspace (`.git` read-only). Delivery used writable clone at `/tmp/kiwifs-pr383`.

## Change

Reload editor slash commands when the active Kiwi space changes via `onSpaceChange` in `useEditorSlashCommands`, so config updates apply without a full page reload.

## Tests

- `go test ./internal/api/... -run TestGetEditorSlashCommands -count=1` — 5/5 PASS
- `go test ./internal/config/... -run TestUIConfigEditorSlashCommands -count=1` — PASS
- `cd ui && npm test -- editorSlashCommands markdownSlashCommands --run` — 12/12 PASS
