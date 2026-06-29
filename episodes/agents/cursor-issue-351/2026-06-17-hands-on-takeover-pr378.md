---
memory_kind: episodic
episode_id: cursor-issue-351-hands-on-pr378-2026-06-17
title: "PR #378 hands-on takeover — slash commands peer-review fix"
tags: [kiwifs, issue-351, pr-378, slash-commands, takeover, peer-review]
date: 2026-06-17
---

## Task

Merge-first on [PR #378](https://github.com/kiwifs/kiwifs/pull/378) — configurable editor slash commands. Remote CI was green; applied pending peer-review hardening before fleet push.

## Actions

1. Verified upstream CI **pass** (test job 8m57s — UI tests, build, go vet, go test).
2. Applied peer-review fix: reject slash command IDs not matching `^[\w-]+$` (CodeMirror `validFor` compatibility).
3. Fixed OpenAPI tag on `GetEditorSlashCommands` from `theme` → `editor`.
4. Added `TestGetEditorSlashCommands_SkipsInvalidID` regression test.
5. Deduped `.git-writable/` entry in `.gitignore`.
6. Wrote episodic + fix docs to KiwiFS cluster memory.

## Tests

```bash
go test ./internal/api/... -run TestGetEditorSlashCommands -count=1   # PASS (4 tests)
go test ./internal/config/... -run TestUIConfigEditorSlashCommands -count=1  # PASS
cd ui && npm test -- editorSlashCommands markdownSlashCommands --run  # 12/12 PASS
cd ui && npm test -- --run  # 152/152 PASS
```

## Result

Local commit with peer-review fix ready for fleet push; PR #378 CI green, no review comments.
