---
type: runbook
title: "Runbook Title"
trigger: "Alert or condition that should invoke this runbook"
severity: P3
owner: team-or-oncall
services:
  - "[[service-name]]"
status: draft
tags: [operations]
---

# Runbook Title

## 1. Trigger / When to Use

_Describe the alert, metric threshold, or user report that should invoke this runbook._

## 2. Diagnosis

Run these commands to confirm the problem and locate the affected component.

```bash
# Example diagnostic command
# command-here

# Expected output:
# ...
```

## 3. Mitigation

Ordered, idempotent steps. Re-running a step should be safe.

1. _First mitigation step_
2. _Second mitigation step_

```bash
# Example mitigation command
# command-here
```

## 4. Verification

Concrete success criteria — not "monitor the dashboard" without specifics.

- [ ] _Metric or health check returns to normal_
- [ ] _Error rate below threshold_

```bash
# Example verification command
# command-here
```

## 5. Rollback

How to undo mitigation if it makes things worse.

1. _Rollback step_

## 6. RTO and Data Loss Expectations

| Metric | Target |
|--------|--------|
| Recovery time objective (RTO) | _e.g. 30 minutes_ |
| Recovery point objective (RPO) | _e.g. none — read-only mitigation_ |
| Expected data loss | _e.g. none_ |

## 7. Escalation Path

| Severity | Contact | When |
|----------|---------|------|
| P1 | _on-call → team lead → VP_ | Immediate |
| P2 | _on-call → team lead_ | If not resolved in 30 min |
