---
memory_kind: episodic
episode_id: cursor-hands-on-423-2026-06-30
title: "Issue #423 external link rot — hands-on delivery"
tags: [kiwifs, janitor, external-links, issue-423, link-rot, bounty]
date: 2026-06-30
---

# Issue #423 — external link rot janitor (hands-on)

## Task

Implement kiwifs/kiwifs#423: opt-in janitor rule `external_link_rot` that scans
markdown for external URLs, HTTP-probes with cache/rate limits, and reports broken
links in `GET /api/kiwi/janitor` under `external_links`.

## Approach

1. Searched Kiwi fix docs — prior fix at `pages/fixes/kiwifs-kiwifs/issue-423-external-link-rot-janitor.md`.
2. Cherry-picked verified commits from `feat/issue-423-external-link-rot-clean` onto
   `main` as branch `feat/issue-423-external-link-rot-v2`.
3. Dropped out-of-scope workspace template scaffolds accidentally reintroduced during
   cherry-pick conflict resolution (14 files, 1246 lines).
4. Overlay whiteouts applied for ghost lower-layer template dirs that broke
   `cmd.TestRunbookInitCheckPasses`.

## Tests (hands-on verification 2026-06-30)

```text
ok  github.com/kiwifs/kiwifs/internal/janitor  0.072s
ok  github.com/kiwifs/kiwifs/internal/config   0.010s
ok  github.com/kiwifs/kiwifs/cmd               0.600s
ok  github.com/kiwifs/kiwifs/internal/api      10.490s
```

Regression coverage includes: 404/500 flagging, SSRF blocks, whitelist, cache skip,
HEAD→GET fallback, max-checks cap, unreachable hosts, config TOML parsing.

## Outcome

Branch `feat/issue-423-external-link-rot-v2` verified green; pushed and PR opened.
Closes #423. Fix doc: `pages/fixes/kiwifs-kiwifs/issue-423-external-link-rot-janitor.md`
(Kiwi MCP gateway unreachable from this environment).
