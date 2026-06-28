---
title: Runbooks
owner: on-call
status: active
tags: [meta, operations]
last-reviewed: 2026-01-01
---

# Runbooks

Operational knowledge for on-call and platform teams. Every alert
should link to a runbook. Every incident should produce a postmortem.

## Severity Guide

| Level | Criteria | Response Time | Escalation |
|-------|----------|---------------|------------|
| **P1** | Full outage, all users affected | Immediate | VP/Director within 15 min |
| **P2** | Major feature broken, many users affected | < 30 min | Team lead within 30 min |
| **P3** | Minor feature broken, workaround exists | < 4 hours | During business hours |
| **P4** | Cosmetic or low-impact issue | Next business day | No escalation |

## Escalation Matrix

| Severity | Primary | Secondary | Executive |
|----------|---------|-----------|-----------|
| P1 | On-call engineer | Team lead → Eng. Manager | VP Engineering |
| P2 | On-call engineer | Team lead | Eng. Manager (if > 1hr) |
| P3 | Assigned engineer | Team lead (if > 4hr) | — |
| P4 | Backlog | — | — |

_Update this table with your team's actual contacts and rotation schedule._

## Procedures

Reusable step-by-step guides for common operational tasks.

| Procedure | Tags | Last Reviewed | Last Tested |
|-----------|------|---------------|-------------|
| [[procedures/deploy-rollback]] | deployment, rollback | 2026-01-01 | 2026-01-01 |
| [[procedures/rotate-secrets]] | security, secrets | 2026-01-01 | — |
| [[procedures/scale-up]] | capacity, scaling | 2026-01-01 | — |

## Incidents

One file per incident, created from [[incidents/template|the template]].
Named `incidents/YYYY-MM-DD-<slug>.md`.

## Postmortems

Blameless post-incident reviews. Written within 3-5 business days
of incident resolution. Use [[postmortems/template|the postmortem template]].

## Quick Reference

- **On-call schedule:** _link to your rotation tool_
- **Monitoring dashboard:** _link to Grafana/Datadog/etc._
- **Status page:** _link to your status page_
- **Incident channel:** `#incidents` or `#team-urgent`
- **Escalation contacts:** _see matrix above_

## Game Day Schedule

Track periodic testing of procedures:

| Quarter | Procedures Tested | Results |
|---------|-------------------|---------|
| _Q1 2026_ | _deploy-rollback, scale-up_ | _passed / issues found_ |
