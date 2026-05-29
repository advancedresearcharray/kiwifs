---
title: "Postmortem: [Short Description]"
date: 2026-01-01
incident: 
severity: P2
authors: []
tags: [postmortem]
---

# Postmortem: [Short Description]

**Incident:** [[incidents/YYYY-MM-DD-slug]]
**Date:** YYYY-MM-DD
**Severity:** P1 / P2 / P3 / P4
**Duration:** X hours Y minutes
**Authors:** _who wrote this postmortem_

## Summary

_2-3 sentences: what happened, how long it lasted, what the impact was._

## Timeline

| Time (UTC) | Event |
|------------|-------|
| HH:MM | Deployment triggered the issue |
| HH:MM | Alert fired |
| HH:MM | On-call began investigating |
| HH:MM | Root cause identified |
| HH:MM | Mitigation applied |
| HH:MM | Fully resolved |

## Impact

- **Users affected:** _number or percentage_
- **Revenue impact:** _if applicable_
- **SLA impact:** _did we breach?_

## Root Cause

_What actually went wrong? Be specific._

## Contributing Factors

_What conditions made this possible or made detection/resolution slower?_

- _e.g., Missing monitoring for this failure mode_
- _e.g., Runbook was outdated_
- _e.g., Deploy happened on Friday afternoon_

## What Went Well

- _e.g., Alert fired within 2 minutes_
- _e.g., Rollback procedure worked as documented_

## What Went Wrong

- _e.g., Took 30 minutes to identify root cause_
- _e.g., No runbook for this specific failure_

## Action Items

| Action | Owner | Due | Status |
|--------|-------|-----|--------|
| Add monitoring for [failure mode] | @engineer | YYYY-MM-DD | todo |
| Update deploy rollback runbook | @on-call | YYYY-MM-DD | todo |
| Add integration test for [scenario] | @engineer | YYYY-MM-DD | todo |

## Lessons Learned

_What should the team remember from this incident?_

## Related

- Incident: [[incidents/YYYY-MM-DD-slug]]
- Related past incidents: _link any similar events_
