---
memory_kind: episodic
episode_id: cursor-issue-120-2026-06-14
title: Issue #120 MySQL importer integration test
tags: [kiwifs, issue-120, importer, mysql, testcontainers]
date: 2026-06-14
---

## Run log

Implemented kiwifs/kiwifs#120 on branch `issue-120-mysql-importer-test`.

1. Searched Kiwi cluster for prior issue-120 docs (found local draft fix doc, not yet on cluster).
2. Recovered WIP `mysql_test.go` from git stash; added integration test using testcontainers `mysql:8`.
3. First test run failed: `active=1, want true` — MySQL driver returns BOOLEAN as `int64`.
4. Root-caused: `ColumnType.DatabaseTypeName()` is `TINYINT` but `Length()` is unset; must use `information_schema.COLUMNS` with `COLUMN_TYPE = 'tinyint(1)'`.
5. Added `detectBoolColumns()` + `mapMySQLColumnValue()` in `mysql.go`.
6. Tests pass: `TestMySQLImporterIntegration` (~14s with Docker), short mode skips cleanly.

## Outcome

Ready for fleet publish (local commit only; no push/PR from Cursor).
