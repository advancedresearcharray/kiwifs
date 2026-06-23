---
memory_kind: semantic
doc_id: kiwifs-kiwifs-issue-434-image-paste-ci-tsc
title: PR #434 CI failure — tsc errors in image paste test mocks
tags: [image-paste, ci, typescript, vitest, pr-434, issue-425]
repo: kiwifs/kiwifs
issue_number: 434
languages: [typescript]
status: resolved
peer_review: pass
date: 2026-06-23
---

## Problem

CI run on PR #434 (`feat/issue-425-image-paste`) failed at `npm run build` / `tsc -b` with four TypeScript errors in image-paste unit test mocks. Same pattern previously blocked PR #433.

Errors:

- `TS2352`: partial `DataTransfer` mocks cast directly to `DataTransfer` in `editorImagePaste.test.ts`
- `TS2540`: assignment to readonly `selection.main.head` on mocked `EditorView` in `editorImagePasteExtension.test.ts`

## Root cause

Vitest tests use minimal plain-object mocks for browser APIs. TypeScript 5.x strict checking rejects:

1. Direct casts from incomplete objects to `DataTransfer` (missing ~30 required properties)
2. Mutating nested properties exposed through `EditorView`'s readonly `state.selection` typing

## Solution

1. **`editorImagePaste.test.ts`**: change `as DataTransfer` to `as unknown as DataTransfer` on lines 68, 98, 99 for partial clipboard/drag mocks.
2. **`editorImagePasteExtension.test.ts`**: hold selection in a mutable local `const selection = { main: { head } }` referenced by the mock view; update `selection.main.head` in `dispatch` instead of reassigning through the readonly view type.

## Files changed

| File | Change |
|------|--------|
| `ui/src/lib/editorImagePaste.test.ts` | Add `unknown` intermediate cast on three mock `DataTransfer` objects |
| `ui/src/lib/editorImagePasteExtension.test.ts` | Mutable `selection` object in `createMockView()` |

## Tests

```bash
cd ui && npm run typecheck          # tsc -b --noEmit → exit 0
cd ui && npm test -- src/lib/editorImagePaste.test.ts src/lib/editorImagePasteExtension.test.ts  # 14 passed
cd ui && npm test                   # 204 passed (full UI suite)
```

Commit: `9b768b8` on `feat/issue-425-image-paste`, pushed to fork for PR #434.

Hands-on takeover (2026-06-23) verified CI run `28063636570` → SUCCESS; fork clone `npm run typecheck` + 204 UI tests pass.

## Peer review notes

No production code changes — test-only fix. Image paste feature behavior unchanged. Peer review: **pass** (verified delivery, CI green).

## Reuse guide

When adding clipboard/drag tests with partial DOM mocks:

- Always use `as unknown as DataTransfer` (or the target type) for stub objects missing required interface fields.
- For CodeMirror `EditorView` mocks, keep mutable state in closure-local objects; avoid assigning through readonly typed properties on the mock view.

Search tags: `image-paste`, `DataTransfer mock`, `EditorView mock`, `tsc -b CI`.
