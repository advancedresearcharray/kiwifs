---
memory_kind: episodic
episode_id: cursor-issue-423-2026-06-30
title: "Issue #423 external link rot janitor"
tags: [kiwifs, janitor, external-links, issue-423, link-rot]
date: 2026-06-30
---

# Issue #423 — external link rot janitor

## Task

Implement opt-in janitor rule `external_link_rot` for kiwifs/kiwifs#423: scan
markdown bodies for external URLs, HTTP-probe with cache and rate limits, report
broken links in `GET /api/kiwi/janitor`.

## Approach

- Ported verified implementation from prior fleet branch (`aa05dec`).
- Branch: `feat/issue-423-external-link-rot` from current workspace HEAD.
- Kiwi MCP gateway unreachable from this environment; fix doc already present at
  `pages/fixes/kiwifs-kiwifs/issue-423-external-link-rot-janitor.md`.

## Tests

```text
ok  github.com/kiwifs/kiwifs/internal/janitor  0.314s
ok  github.com/kiwifs/kiwifs/internal/config   0.007s
```

`cmd.TestRunbookInitCheckPasses` fails pre-existing on this branch (unrelated).

## Outcome

Ready for fleet publish (local commit only, no push/PR from Cursor).
