---
memory_kind: episodic
episode_id: cursor-issue-425-2026-06-30
title: Issue #425 image paste from clipboard
tags: [kiwifs, ui, editor, image-paste, issue-425]
date: 2026-06-30
---

# Issue #425 — image paste from clipboard

## Task

Implement clipboard image paste and drag-and-drop in KiwiEditor (visual BlockNote + source CodeMirror) per [kiwifs/kiwifs#425](https://github.com/kiwifs/kiwifs/issues/425).

## Approach

- Shared helpers in `editorImagePaste.ts` (MIME detection, `paste-YYYYMMDD-HHMMSS.ext` naming, relative markdown refs).
- Source mode: CodeMirror `editorImagePasteExtension` inserts `![Uploading...]()` placeholder, uploads via `api.uploadAsset`, replaces with `![name](relative-path)`.
- Visual mode: BlockNote native `uploadFile` path with `renameFileForPaste`, block removal + toast on failure.
- Drop-zone overlay on editor area for both modes.
- Error toast bottom-left (mirrors slash-command alert pattern).

## Tests

```bash
cd ui && npm test -- src/lib/editorImagePaste.test.ts src/lib/editorImagePasteExtension.test.ts
# 11 passed

cd ui && npm test
# 217 passed
```

## Branch

`feat/issue-425-image-paste` (local commit only; fleet publishes PR).
