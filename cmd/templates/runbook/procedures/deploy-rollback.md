---
title: Deploy Rollback
tags: [deployment, rollback, ci-cd]
status: draft
last-reviewed: 2026-01-01
---

# Deploy Rollback

Roll back the most recent deployment when it causes issues in production.

## When to Use

- Health checks failing after a deploy
- Error rate spike correlated with a recent deployment
- Customer reports of broken functionality after a release

## Prerequisites

- [ ] Access to CI/CD dashboard
- [ ] Permission to trigger deployments
- [ ] Know the previous good version/commit

## Steps

### 1. Identify the Bad Deploy

```bash
# Check recent deployments (replace with your actual commands)
git log --oneline -5 main

# Or check your CI/CD tool
# e.g., kubectl rollout history deployment/<name> -n <namespace>
```

### 2. Roll Back

```bash
# Option A: Revert the commit
git revert <bad-commit-sha> --no-edit
git push origin main

# Option B: Redeploy previous version
# kubectl rollout undo deployment/<name> -n <namespace>

# Option C: Feature flag
# Disable the flag in your feature flag dashboard
```

### 3. Verify

```bash
# Check health endpoint
curl -s https://api.example.com/health
# Expected: {"status":"ok"}

# Check error rate is declining
# Open monitoring dashboard: <URL>
```

### 4. Communicate

- Post in `#team-urgent`: "Rolled back deploy [SHA]. Error rate recovering."
- If customer-facing: update status page

## Rollback of the Rollback

If the rollback itself causes issues:

1. Check if the revert introduced conflicts
2. Consider deploying a known-good tag instead
3. Escalate if neither works

## Related

- [[procedures/scale-up]] — if rollback isn't enough and you need capacity
- [[incidents/template]] — document the incident
