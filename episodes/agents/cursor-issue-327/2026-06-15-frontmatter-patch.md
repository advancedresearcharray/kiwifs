---
memory_kind: episodic
episode_id: cursor-issue-327-2026-06-15
title: "Issue #327 — implement PATCH merge=frontmatter"
tags: [kiwifs, api, frontmatter, issue-327, runbooks]
date: 2026-06-15
---

## Task

Implement kiwifs/kiwifs#327: `PATCH /api/kiwi/file?merge=frontmatter` for frontmatter-only updates during incident response.

## Investigation

- Legacy handler existed at `PATCH /api/kiwi/file/frontmatter` with `{"fields":{...}}` body and no If-Match.
- Issue spec requires `merge=frontmatter` on `/file`, flat JSON body, ETag/If-Match, body byte preservation, git commit, 404 for missing files.

## Implementation

- Added `PatchFile` route + shared `patchFrontmatterFields` helper.
- Switched frontmatter writes to `WriteWithOpts` with If-Match.
- Added 8 regression tests including git commit verification.

## Tests

```
go test ./internal/api/... -run 'PatchFile|PatchFrontmatter' -count=1
# ok github.com/kiwifs/kiwifs/internal/api 0.239s
```

## Notes

- UI `api.ts` patchFrontmatter still points at legacy endpoint (file not writable in overlay); legacy route remains compatible.
- Branch: `issue-327-frontmatter-patch` (local commit, fleet publishes PR).
