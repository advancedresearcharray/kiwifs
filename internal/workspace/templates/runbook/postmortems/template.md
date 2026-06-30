---
title: "Postmortem: [Short Description]"
date: 2026-01-01
incident: 
severity: P2
authors: []
duration-minutes: 
tags: [postmortem]
---

> **Blameless Postmortem Notice**
>
> This document follows blameless postmortem principles. We focus on:
> - What happened (not who caused it)
> - Why our systems allowed this to happen
> - How we can prevent similar incidents
>
> Names appear only to establish timeline context, not to assign blame.
> We assume everyone acted with good intentions based on the information
> available to them at the time.

# Postmortem: [Short Description]

**Incident:** [[incidents/template]]
**Date:** YYYY-MM-DD
**Severity:** P1 / P2 / P3 / P4
**Duration:** X hours Y minutes
**Authors:** _who wrote this postmortem_

## Summary

_2-3 sentences for a broad audience: what happened, how long it lasted,
what the impact was. Write this so a VP or customer-facing team can
understand without reading the full document._

## Impact

- **Users affected:** _number or percentage_
- **Error budget consumed:** _e.g., consumed 25% of monthly error budget_
- **Revenue impact:** _if applicable_
- **SLA impact:** _did we breach? Which SLOs were violated?_
- **Data impact:** _any data loss or corruption?_

## Timeline

_All times in UTC. Include human decision points — what was decided and
why it seemed reasonable at the time._

| Time (UTC) | Event |
|------------|-------|
| HH:MM | Deployment triggered the issue |
| HH:MM | Alert fired — `[alert name]` |
| HH:MM | On-call began investigating |
| HH:MM | Root cause identified |
| HH:MM | Mitigation applied — `[action]` |
| HH:MM | Fully resolved — verified via `[check]` |

## Contributing Factors

_What systemic conditions made this incident possible? Use the 5 Whys
technique to dig past surface-level symptoms. There are usually multiple
contributing factors — list them all._

1. **[Factor 1]** — _description_
   - Why? → _because..._
   - Why? → _because..._
   - Why? → _root systemic issue_
2. **[Factor 2]** — _description_
   - Why? → ...

## What Went Well

_Acknowledge what worked. This reinforces good practices._

- _e.g., Alert fired within 2 minutes of threshold breach_
- _e.g., Rollback procedure worked as documented_
- _e.g., Cross-team communication was fast and clear_

## What Went Wrong

_What didn't work? Focus on systems and processes, not people._

- _e.g., Took 30 minutes to identify root cause due to insufficient logging_
- _e.g., No runbook existed for this specific failure mode_
- _e.g., Deploy happened without canary, bypassing standard process_

## Action Items

_Each action item MUST have: a named person (not a team), a firm deadline,
and a tracking ticket. Limit to 3-5 high-impact items — ruthlessly prioritize.
Vague items like "improve monitoring" are not acceptable._

| Action | Owner | Due | Ticket | Status |
|--------|-------|-----|--------|--------|
| Add alerting for [specific failure mode] | @person | YYYY-MM-DD | JIRA-123 | todo |
| Update deploy rollback runbook with [step] | @person | YYYY-MM-DD | JIRA-124 | todo |
| Add integration test for [scenario] | @person | YYYY-MM-DD | JIRA-125 | todo |

## Lessons Learned

_What should the team remember from this incident? What would you tell
a new team member about this failure mode?_

## Related

- Incident: [[incidents/template]]
- Related past incidents: _link any similar events_
- Procedures used: [[procedures/deploy-rollback]]

---

## Postmortem Quality Checklist

- [ ] Blameless — no individual blame language
- [ ] Timeline is precise with UTC timestamps
- [ ] Contributing factors go ≥ 3 levels deep (5 Whys)
- [ ] Action items have named owners (people, not teams)
- [ ] Action items have specific deadlines
- [ ] Action items are filed in project tracker
- [ ] Impact is quantified (users, SLO, revenue)
- [ ] Published within 5 business days of resolution
