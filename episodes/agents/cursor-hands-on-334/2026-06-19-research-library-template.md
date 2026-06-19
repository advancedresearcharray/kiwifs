---
memory_kind: episodic
episode_id: cursor-hands-on-334-2026-06-19
title: Issue #334 research library init template delivery
tags: [kiwifs, workspace, research, issue-334, uc-research]
date: 2026-06-19
---

## Work item

kiwifs/kiwifs#334 — feat(workspace): ship research library init template with reading workflow

## Actions

1. Took over from fleet agent after delivery check failed (not committed, peer review gaps).
2. Checked out `feat/issue-334-research-library-template`; fixed overlay git index via alternate `GIT_INDEX_FILE`.
3. Peer review (bugbot) flagged schema/doc mismatch and missing lint tests vs prompt-library bar.
4. Applied fixes:
   - Require `workflow` + `state` in `paper.json`
   - Added research-specific `.kiwi/config.toml` with `cites` typed_fields
   - Normalized example `cites` wikilinks to path form
   - Added `TestResearchTemplateLintClean`, `TestInitResearchTemplateMetadata`, extended init/metadata tests
5. Ran `go test ./internal/workspace/... -count=1` — all research tests green.
6. Committed and pushed to PR #405.

## Test output

```
go test ./internal/workspace/... -count=1 -run 'Research|InitResearch|ListInit'
ok  github.com/kiwifs/kiwifs/internal/workspace  0.009s
```

## PR

https://github.com/kiwifs/kiwifs/pull/405
