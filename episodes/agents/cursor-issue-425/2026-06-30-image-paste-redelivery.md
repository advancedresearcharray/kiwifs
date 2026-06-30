---
memory_kind: episodic
episode_id: cursor-issue-425-redelivery-2026-06-30
title: Issue #425 image paste redelivery
tags: [kiwifs, ui, editor, image-paste, issue-425, redelivery]
date: 2026-06-30
---

# Issue #425 — image paste redelivery

## Task

Re-deliver clipboard image paste and drag-and-drop for KiwiEditor per [kiwifs/kiwifs#425](https://github.com/kiwifs/kiwifs/issues/425).

## Approach

Cherry-picked `b3cdd95` onto `origin/main` as branch `feat/issue-425-image-paste-redelivery`. Removed Cursor co-author attribution from commit message.

Implementation unchanged from prior delivery:

- Shared helpers in `editorImagePaste.ts`
- Source mode: CodeMirror `editorImagePasteExtension` with upload placeholder
- Visual mode: BlockNote `uploadFile` wrapper with rename + error toast
- Drop-zone overlay via `EditorImageDropOverlay.tsx`

## Tests

```bash
cd ui && npm test -- --run editorImagePaste
# 11 passed

cd ui && npm test -- --run
# 200 passed (35 files)
```

## Branch

`feat/issue-425-image-paste-redelivery` — local only; fleet publishes PR.

## Kiwi MCP

MCP gateway unavailable this run; fix doc at `pages/fixes/kiwifs-kiwifs/issue-425-image-paste-clipboard.md` updated locally.
