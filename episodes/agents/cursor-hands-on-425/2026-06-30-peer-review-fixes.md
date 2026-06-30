---
memory_kind: episodic
episode_id: cursor-hands-on-425-peer-review
title: Issue #425 peer review fixes and PR delivery
tags: [issue-425, image-paste, peer-review, ui]
date: 2026-06-30
---

## Context

Hands-on takeover after fleet agent delivery check failed (no PR, peer review not passed). Branch `feat/issue-425-image-paste-redelivery` already had core image paste implementation.

## Peer review fixes

1. **Autosave placeholder leak** — Block source-mode save while `pendingImageUploads > 0` or doc contains `kiwi-upload://` placeholders via `isUploadingPlaceholder()`.
2. **Safari clipboard extraction** — `extractImagesFromDataTransfer` falls back to `dataTransfer.files` when items have empty MIME types.
3. **Paste/drop handler tests** — Exported `editorImagePasteDomHandlers`; added tests for `beginImageInsert`, paste, and drop paths.
4. **Dead code** — Removed unused `imagePasteProsemirrorPlugin.ts` (BlockNote `uploadFile` handles visual mode).
5. **Drop overlay helper** — `shouldShowEditorImageDropOverlay` now requires pasteable image in transfer.

## Tests

```bash
cd ui && npm test -- --run
# 204 passed (35 files)
```

## Outcome

Committed fixes, pushed branch, opened PR closing #425.
