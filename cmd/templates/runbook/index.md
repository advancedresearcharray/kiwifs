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

| Level | Criteria | Response Time | Example |
|-------|----------|---------------|---------|
| **P1** | Full outage, all users affected | Immediate | API returns 500 for all requests |
| **P2** | Major feature broken, many users affected | < 30 min | Payments failing, search down |
| **P3** | Minor feature broken, workaround exists | < 4 hours | Export slow, non-critical UI bug |
| **P4** | Cosmetic or low-impact issue | Next business day | Typo in email, minor UI glitch |

## Procedures

Reusable step-by-step guides for common operational tasks.

| Procedure | Tags | Last Reviewed |
|-----------|------|---------------|
| [[procedures/deploy-rollback]] | deployment, rollback | 2026-01-01 |
| [[procedures/rotate-secrets]] | security, secrets | 2026-01-01 |
| [[procedures/scale-up]] | capacity, scaling | 2026-01-01 |

## Incidents

One file per incident, created from [[incidents/template|the template]].
Named `incidents/YYYY-MM-DD-<slug>.md`.

## Postmortems

Root cause analysis written after incidents are resolved. Use
[[postmortems/template|the postmortem template]].

## Quick Reference

- **On-call schedule:** _link to your rotation tool_
- **Monitoring dashboard:** _link to Grafana/Datadog/etc._
- **Status page:** _link to your status page_
- **Incident channel:** `#incidents` or `#team-urgent`
