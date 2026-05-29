---
title: "Incident: [Short Description]"
date: 2026-01-01
severity: P2
status: active
on-call: 
related-alert: 
tags: [service-name]
postmortem: 
---

# Incident: [Short Description]

## Trigger

_What alert or report initiated this incident?_

- Alert: `[alert name or link]`
- Detected by: monitoring / customer report / engineer

## Impact

- **Affected services:** _list services_
- **Blast radius:** _e.g., all users, 10% of requests, internal only_
- **SLA impact:** _e.g., breaches 99.9% availability after 15 min_

## Timeline

| Time (UTC) | Event |
|------------|-------|
| HH:MM | Detected — alert fired / report received |
| HH:MM | Investigating — [who] started looking |
| HH:MM | Mitigated — [action taken] |
| HH:MM | Resolved — root cause fixed |

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

## Follow-ups

- [ ] Write postmortem → `postmortems/YYYY-MM-DD-slug.md`
- [ ] File ticket for permanent fix
- [ ] Update runbook if procedures were wrong or missing
- [ ] Update monitoring if detection was too slow

## Related

- Procedure used: [[procedures/deploy-rollback]]
- Postmortem: _link when written_
