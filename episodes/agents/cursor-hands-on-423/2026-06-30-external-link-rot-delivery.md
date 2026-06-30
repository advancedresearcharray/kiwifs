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
ok  github.com/kiwifs/kiwifs/internal/janitor  0.228s  (20 external-link tests)
ok  github.com/kiwifs/kiwifs/internal/config   0.023s
ok  github.com/kiwifs/kiwifs/cmd               0.653s
ok  github.com/kiwifs/kiwifs/internal/api      10.770s
ok  github.com/kiwifs/kiwifs/internal/bootstrap 1.061s
```

20 external-link regression tests pass (404/500, SSRF, cache, HEAD→GET fallback,
frontmatter skip, whitelist, max-checks cap, concurrent limit, User-Agent,
JSON serialization, unreachable host).

Added regression tests: concurrent probe limit (max 2 in-flight), User-Agent
header on outbound probes, ScanResult JSON `external_links` field shape.

## Outcome

Branch `feat/issue-423-external-link-rot-v2` verified green (7 commits from
`main`, 1533 lines). PR #435 on the older branch was closed; opening fresh PR
from v2 branch. Fix doc committed at
`pages/fixes/kiwifs-kiwifs/issue-423-external-link-rot-janitor.md`.
Kiwi MCP gateway at 192.168.167.240:3333 unreachable — fix doc stored in-repo.
