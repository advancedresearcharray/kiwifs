---
title: Runbooks
owner: platform-oncall
status: active
tags: [meta, operations]
next-review: 2026-09-01
---

# Runbooks

Operational procedures for on-call and platform teams. Every alert should link
to a runbook. Runbooks follow the DevHelm 7-section format validated by
`.kiwi/schemas/runbook.json`.

## Severity Guide

| Level | Criteria | Response Time | Escalation |
|-------|----------|---------------|------------|
| **P1** | Full outage, all users affected | Immediate | VP/Director within 15 min |
| **P2** | Major feature broken, many users | < 30 min | Team lead within 30 min |
| **P3** | Minor feature broken, workaround exists | < 4 hours | During business hours |
| **P4** | Cosmetic or low-impact | Next business day | No escalation |

_Update contacts and rotation schedule for your team._

## Runbooks

| Runbook | Trigger | Severity | Owner | Last Outcome |
|---------|---------|----------|-------|--------------|
| [[example-high-cpu]] | CPU > 80% on api-service | P2 | platform-oncall | success |

_Add rows when creating new runbooks. Copy from `.kiwi/templates/runbook.md` or duplicate [[example-high-cpu]] as a starting point._

## Query Active Runbooks

```dql
TABLE title AS Title, trigger AS Trigger, severity AS Severity, last_executed AS "Last Run", last_outcome AS Outcome
WHERE type = "runbook" AND status = "active"
SORT severity ASC
```

## Quick Reference

- **Blank template:** `.kiwi/templates/runbook.md`
- **Schema reference:** [[SCHEMA]]
- **On-call schedule:** _link to your rotation tool_
- **Monitoring dashboard:** _link to Grafana/Datadog/etc._
- **Incident channel:** `#incidents`

## Related Docs

- [[SCHEMA]] — frontmatter fields and section requirements
- [[example-high-cpu]] — fully worked example with commands and escalation
