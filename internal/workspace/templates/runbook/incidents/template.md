---
title: "Incident: [Short Description]"
date: 2026-01-01
severity: P2
status: active
on-call: 
related-alert: 
detection-minutes: 
mitigation-minutes: 
resolution-minutes: 
users-affected: 
error-budget-impact: 
tags: [service-name]
postmortem: 
---

# Incident: [Short Description]

## Trigger

_What alert or report initiated this incident?_

- Alert: `[alert name or link]`
- Detected by: monitoring / customer report / engineer
- Detection lag: _how long between incident start and detection_

## Impact

- **Affected services:** _list services_
- **Blast radius:** _e.g., all users, 10% of requests, internal only_
- **User-facing symptoms:** _what users experienced_
- **SLA impact:** _e.g., breaches 99.9% availability after 15 min_
- **Error budget consumed:** _percentage of monthly/quarterly budget_

## Timeline

_All times in UTC._

| Time (UTC) | Event |
|------------|-------|
| HH:MM | Incident started (estimated) |
| HH:MM | Detected — alert fired / report received |
| HH:MM | Investigating — [who] started looking |
| HH:MM | Mitigated — [action taken] |
| HH:MM | Resolved — root cause fixed |
| HH:MM | Verified — confirmed via [check] |

## Communication Log

| Time (UTC) | Channel | Message |
|------------|---------|---------|
| HH:MM | #incident-channel | Incident declared, investigating |
| HH:MM | Status page | Updated to degraded performance |
| HH:MM | #incident-channel | Mitigated, monitoring |
| HH:MM | Status page | Resolved |

## Diagnostics

_What did you check? Include exact commands and what they showed._

```bash
# Example: check pod health
kubectl get pods -n <namespace> | grep -v Running

# Example: check error rate
curl -s https://monitoring.example.com/api/v1/query?query=rate(http_errors[5m])
```

## Resolution

_What fixed it?_

1. _Action taken (e.g., rolled back deploy, restarted service)_
2. _Verification: how you confirmed it was fixed_

## Escalation

_Who was involved? Was it escalated?_

- First responder: @on-call
- Escalated to: @team-lead (if applicable)
- Stakeholders notified: #incident-channel

## Metrics

| Metric | Value |
|--------|-------|
| Time to detect | _X minutes_ |
| Time to mitigate | _X minutes_ |
| Time to resolve | _X minutes_ |
| Total duration | _X minutes_ |

## Follow-ups

- [ ] Write postmortem → `postmortems/YYYY-MM-DD-slug.md`
- [ ] File ticket for permanent fix
- [ ] Update runbook if procedures were wrong or missing
- [ ] Update monitoring if detection was too slow
- [ ] Update escalation docs if escalation was unclear

## Related

- Procedure used: [[procedures/deploy-rollback]]
- Postmortem: _link when written_
- Similar past incidents: _link if any_
