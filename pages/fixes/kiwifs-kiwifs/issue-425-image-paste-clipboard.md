---
memory_kind: semantic
doc_id: kiwifs-issue-425-image-paste-clipboard
title: Image paste from clipboard in KiwiEditor
tags: [ui, editor, image-paste, clipboard, codemirror, blocknote]
repo: kiwifs/kiwifs
issue_number: 425
languages: [TypeScript]
status: resolved
date: 2026-06-23
---

## Problem

KiwiEditor had no way to paste or drag-and-drop images from the OS clipboard into either visual (BlockNote) or source (CodeMirror) editing modes. Users expected Confluence/Obsidian-style image paste.

PR #433 initially failed CI (`npm run build` / `tsc -b`) after the feature landed.

## Root cause

1. **Feature gap** — No clipboard/drag handlers wired for image MIME types in either editor mode.
2. **KiwiEditor prop regression (CI)** — The image-paste branch removed `editorModePref` / `onEditorModeChange` from `KiwiEditor`, but `App.tsx` still passes them → TS2322.
3. **Test mock typing (CI)** — `editorImagePaste.test.ts` used direct `as DataTransfer` casts; `tsc -b` requires `as unknown as DataTransfer`.
4. **Read-only mock (CI)** — `editorImagePasteExtension.test.ts` assigned to `view.state.selection.main.head` (readonly on `EditorView`).

## Solution

1. Added `editorImagePaste.ts` helpers: MIME detection, paste filename standardization (`paste-YYYYMMDD-HHMMSS.ext`), clipboard extraction, drag detection.
2. CodeMirror path: `editorImagePasteExtension.ts` with paste/drop handlers, uploading placeholder, error callback.
3. BlockNote path: `imagePasteProsemirrorPlugin.ts` ProseMirror plugin with same upload flow.
4. `KiwiEditor.tsx` wires both modes, drop-zone overlay (`EditorImageDropOverlay`), and error toasts.
5. CI fix (commit `e5b62e6`): restored `editorModePref`/`onEditorModeChange` props and fixed test mock types.

## Files changed

| File | Purpose |
|------|---------|
| `ui/src/lib/editorImagePaste.ts` | Shared paste/drag helpers |
| `ui/src/lib/editorImagePasteExtension.ts` | CodeMirror paste/drop extension |
| `ui/src/lib/imagePasteProsemirrorPlugin.ts` | BlockNote ProseMirror plugin |
| `ui/src/components/EditorImageDropOverlay.tsx` | Visual drop-zone overlay |
| `ui/src/components/KiwiMarkdownSourceEditor.tsx` | Source editor with paste extension |
| `ui/src/components/KiwiEditor.tsx` | Wires upload, overlay, mode props |
| `ui/src/lib/editorImagePaste.test.ts` | Unit tests for helpers |
| `ui/src/lib/editorImagePasteExtension.test.ts` | Unit tests for CodeMirror upload |

## Tests

```bash
cd ui && npm test -- src/lib/editorImagePaste.test.ts src/lib/editorImagePasteExtension.test.ts
# 14 passed

cd ui && npx tsc -b
# clean

cd ui && npm test
# 204 passed (full UI suite)
```

CI on PR #433: all checks green after `e5b62e6`.

## Peer review notes

- Paste filenames are standardized to avoid collisions and match wiki asset conventions.
- Upload placeholder is replaced atomically; failures remove placeholder and surface toast errors.
- `editorModePref` sync preserves user preference API contract with `App.tsx`.
- No unrelated demo/widget files included in PR diff.

## Reuse guide

- To add paste support in another CodeMirror surface: import `editorImagePasteExtension({ uploadImage, onError })`.
- To add paste in ProseMirror/BlockNote: import `imagePasteProsemirrorPlugin({ uploadImage, onError })`.
- Always use `renameFileForPaste()` before upload for consistent naming.
- Test mocks for `DataTransfer` must use `as unknown as DataTransfer` under strict TS.
- CodeMirror test mocks need a mutable `selection` object, not assignment to readonly `selection.main.head`.
