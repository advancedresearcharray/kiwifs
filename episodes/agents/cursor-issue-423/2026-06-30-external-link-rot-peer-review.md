---
memory_kind: episodic
episode_id: cursor-issue-423-peer-review-2026-06-30
title: Issue 423 peer review fixes — SSRF, rate limits, tests
tags: [kiwifs, issue-423, janitor, external-links, peer-review]
date: 2026-06-30
---

## Run

Hands-on takeover after fleet agent delivery failed peer review on
`feat/issue-423-external-link-rot`.

## Peer review fixes applied

1. **Security / DoS** — configurable `external_link_max_checks` (200),
   `external_link_max_concurrent` (10), `external_link_request_delay` (100ms).
2. **SSRF** — block private/link-local IPs, metadata hostnames, non-http(s)
   schemes; optional `external_link_allow` whitelist.
3. **Root validation** — `validateWorkspaceRoot` + cache path confined under root.
4. **Tests** — SSRF, whitelist, 500 errors, max-checks cap, root validation.
5. **Docs** — JanitorConfig comment block + config.toml template updates.

## Tests

```
ok  github.com/kiwifs/kiwifs/internal/janitor  0.060s
ok  github.com/kiwifs/kiwifs/internal/config   0.009s
```

## Branch

`feat/issue-423-external-link-rot`
