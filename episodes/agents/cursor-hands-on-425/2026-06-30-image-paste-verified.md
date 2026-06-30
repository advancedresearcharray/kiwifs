---
memory_kind: episodic
episode_id: cursor-hands-on-425-verified-2026-06-30
title: Issue #425 image paste verified delivery
tags: [kiwifs, ui, editor, image-paste, issue-425, hands-on, verified]
date: 2026-06-30
---

# Issue #425 — verified delivery (hands-on takeover)

## Task

Re-verify and ship clipboard image paste + drag-and-drop for KiwiEditor after fleet delivery check failed (`no_committed_diff`, `peer_review_not_passed`).

## Actions

1. Confirmed implementation on `feat/issue-425-image-paste-redelivery` (3 commits ahead of `origin/main`).
2. Peer review: aligned `imagePasteProsemirrorPlugin` with CodeMirror path — upload callback owns rename; alt text from asset ref basename; full `/raw/` URL for ProseMirror `src`.
3. Removed "Made with Cursor" from fork PR #61 body.
4. Ran full UI suite — 200 passed (11 image-paste tests).

## Tests

```bash
cd ui && npm test -- --run editorImagePaste   # 11 passed
cd ui && npm test -- --run                    # 200 passed (35 files)
```

## PR

- https://github.com/advancedresearcharray/kiwifs/pull/61 (fork; upstream kiwifs/kiwifs restricts PR creation to collaborators)

## Kiwi MCP

Gateway at `192.168.167.240:3333` unreachable; fix doc at `pages/fixes/kiwifs-kiwifs/issue-425-image-paste-clipboard.md`.
