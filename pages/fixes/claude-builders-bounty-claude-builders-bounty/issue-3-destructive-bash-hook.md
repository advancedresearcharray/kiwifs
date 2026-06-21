---
memory_kind: semantic
doc_id: claude-builders-bounty-issue-3-destructive-bash-hook
title: PreToolUse hook blocking destructive bash commands (issue #3 / PR #2556)
tags: [claude-builders-bounty, hooks, pretooluse, bash, security, bounty, opire]
repo: claude-builders-bounty/claude-builders-bounty
issue_number: 3
languages: [python, bash, shell]
status: verified
date: 2026-06-21
derived-from:
  - type: cursor-auto
    id: pr2556-sprout-idle-nudge-2026-06-21
    date: "2026-06-21T01:10:00Z"
    actor: cursor-auto
---

# PreToolUse hook blocking destructive bash commands

## Problem

Claude Code can execute destructive bash commands (`rm -rf`, `git push --force`, SQL drops, etc.) via the Bash tool. Bounty #3 requires a PreToolUse hook that blocks these patterns before execution, logs attempts, and returns a clear denial using the official Claude Code hook schema.

## Root cause

No guard existed in the default Claude Code setup. Competing submissions often used legacy `decision: block` JSON instead of the official `hookSpecificOutput` / `permissionDecision: deny` schema, and branch pollution from parallel fleet jobs repeatedly broke delivery scope checks.

## Solution

Ship `hooks/block-destructive-commands/pre-tool-use.py` with:

- Official `hookSpecificOutput` deny responses and `permissionDecisionReason` messages
- Pattern checks for bounty minimums plus extended coverage (force-push bypasses, disk destruction, container exec unwrapping, Unicode/filler evasion normalization)
- Chained-command splitting on `;`, `&&`, `||` before matching
- Sanitized logging to `~/.claude/hooks/blocked.log`
- Two-command install via `install.sh` (merges into existing `~/.claude/settings.json`)

Delivery gates in `tests/test_delivery_verification.py` and `tests/test_peer_review_acceptance.py` enforce hook-only branch scope and committed source diffs vs `main`.

## Files changed

- `hooks/block-destructive-commands/pre-tool-use.py` — core hook logic
- `hooks/block-destructive-commands/install.sh` — install/uninstall
- `hooks/block-destructive-commands/settings.json` — sample hook registration
- `hooks/block-destructive-commands/README.md` — install, blocked patterns, operator tradeoffs
- `hooks/block-destructive-commands/tests/test_pre_tool_use.py` — 131 regression tests
- `hooks/block-destructive-commands/tests/test_bounty_acceptance.py` — bounty acceptance criteria
- `hooks/block-destructive-commands/tests/test_delivery_verification.py` — fleet delivery gates
- `hooks/block-destructive-commands/tests/test_peer_review_acceptance.py` — peer-review gates
- `hooks/block-destructive-commands/tests/run-ci-checks.sh` — unified CI script
- `.github/workflows/test-block-destructive-commands.yml` — GitHub Actions workflow
- `package.json`, `Makefile` — root `npm test` / `make ci-validate` entry points

## Tests

```bash
cd /workspace/claude-builders-bounty/claude-builders-bounty
git checkout feat/issue-3-destructive-command-hook-v2
npm test
# 131 unittest + 8 delivery + 22 peer-review + attribution — all pass @ 3342cd31
```

CLI spot-check:

```bash
python3 hooks/block-destructive-commands/pre-tool-use.py <<'EOF'
{"tool_name":"Bash","tool_input":{"command":"rm -rf /"}}
EOF
# → hookSpecificOutput permissionDecision: deny

python3 hooks/block-destructive-commands/pre-tool-use.py <<'EOF'
{"tool_name":"Bash","tool_input":{"command":"git status"}}
EOF
# → empty stdout (allowed)
```

## Peer review notes

- Uses official `hookSpecificOutput` schema, not legacy `decision: block`
- Branch diff must stay hook-only (16 paths); reject stray `internal/exporter/*` or `node_modules/*`
- CI workflow uses `pull_request` (not `pull_request_target`) and `fetch-depth: 0` for merge-base scope checks
- Documented limitations: shell variable indirection (`x=rm; $x -rf /`) not blocked
- Log fields sanitized against newline/Unicode injection
- PR #2556 MERGEABLE/CLEAN; no maintainer review blockers as of 2026-06-21

## Reuse guide

For bounty hook PRs in this repo:

1. Check out `feat/issue-3-destructive-command-hook-v2` (PR #2556), not unrelated bounty branches.
2. Run `npm test` from repo root — delivery tests fail if branch scope includes non-hook paths.
3. If fleet parallel work reintroduces `internal/exporter/mkdocs.go`, remove it from the branch (file may exist gitignored locally but must not appear in `git diff main...HEAD`).
4. Extend patterns in `pre-tool-use.py` with matching regression tests in `test_pre_tool_use.py`; update README blocked-patterns table.
5. Do not rewrite working hook logic during merge-nurture if all tests pass — update PROOF.md verification only.
