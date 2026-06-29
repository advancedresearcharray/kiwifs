---
type: runbook
title: High CPU Utilization
trigger: Sustained CPU > 80% for 5+ minutes on api-service pods
severity: P2
owner: platform-oncall
services:
  - "[[api-service]]"
  - "[[monitoring]]"
status: active
tags: [performance, capacity]
last_executed: 2026-06-01
last_outcome: success
execution_count: 3
avg_resolution_time: 20m
reviewed: 2026-06-01
next-review: 2026-09-01
---

# High CPU Utilization

Respond when CPU saturation threatens request latency or pod eviction on
`api-service` (listed in frontmatter `services`).

## 1. Trigger / When to Use

Use this runbook when **any** of the following are true:

- Alert `HighCPU` fires for `api-service` pods (CPU > 80% for 5+ minutes)
- P95 latency exceeds 500ms while CPU graphs show saturation
- Horizontal Pod Autoscaler is at max replicas but CPU remains elevated
- On-call receives user reports of slow API responses during a traffic spike

## 2. Diagnosis

Confirm which pods are hot and whether the load is legitimate traffic or a runaway process.

```bash
# List pods sorted by CPU
kubectl top pods -n production -l app=api-service --sort-by=cpu
```

Expected output (example):

```
NAME                          CPU(cores)   MEMORY(bytes)
api-service-7d4f8b9c6-xk2m9   890m         512Mi
api-service-7d4f8b9c6-pq8n1   120m         480Mi
```

```bash
# Check recent deploy or config change
kubectl rollout history deployment/api-service -n production

# Inspect request rate vs. baseline on the monitoring dashboard
# Grafana: Production > API > Request Rate (compare to 7-day baseline)
```

If a single pod dominates CPU while others are idle, suspect a stuck worker or bad shard assignment before scaling the whole fleet.

## 3. Mitigation

Execute in order. Each step is idempotent — safe to re-run.

1. **Confirm autoscaling headroom** — ensure HPA is not blocked by max replicas or quota.
2. **Scale horizontally** — add replicas before vertical changes.

```bash
# Scale deployment (adjust replica count to your baseline + headroom)
kubectl scale deployment/api-service -n production --replicas=8

# Watch rollout
kubectl rollout status deployment/api-service -n production --timeout=120s
```

3. **If traffic is abusive** — enable rate limiting at the edge (only if product approves).

```bash
# Example: temporarily tighten rate limit (values are illustrative)
kubectl patch configmap edge-ratelimit -n ingress --type merge \
  -p '{"data":{"api_rpm":"600"}}'
kubectl rollout restart deployment/ingress-gateway -n ingress
```

4. **If a single pod is an outlier** — drain and restart it.

```bash
kubectl delete pod -n production api-service-7d4f8b9c6-xk2m9 --grace-period=30
```

## 4. Verification

Success means CPU and latency return to SLO without new errors.

- [ ] All `api-service` pods report CPU < 70% sustained for 10 minutes
- [ ] P95 latency < 300ms on the API dashboard
- [ ] Error rate unchanged from pre-incident baseline (< 0.1%)
- [ ] No CrashLoopBackOff or OOMKilled pods in `production`

```bash
# Health check
curl -sf https://api.example.com/health | jq .

# Expected:
# { "status": "ok", "version": "..." }

kubectl get pods -n production -l app=api-service
# All pods Running, READY 1/1
```

## 5. Rollback

If scaling or rate-limit changes cause **new** failures (503s, DB connection exhaustion):

1. Restore previous replica count:

```bash
kubectl scale deployment/api-service -n production --replicas=4
```

2. Revert rate-limit patch if applied:

```bash
kubectl patch configmap edge-ratelimit -n ingress --type merge \
  -p '{"data":{"api_rpm":"1200"}}'
kubectl rollout restart deployment/ingress-gateway -n ingress
```

3. Page the database on-call if connection pool saturation persists after scale-down.

## 6. RTO and Data Loss Expectations

| Metric | Target |
|--------|--------|
| Recovery time objective (RTO) | 30 minutes to restore latency SLO |
| Recovery point objective (RPO) | None — mitigation is read-only scaling |
| Expected data loss | None |

Scaling and rate limits do not mutate user data. Worst case is temporary throttling of write-heavy clients.

## 7. Escalation Path

| When | Contact | Channel |
|------|---------|---------|
| Not mitigated in 30 min | Platform team lead | `#platform-oncall` |
| Customer-visible outage > 1 hr | Engineering manager | PagerDuty `eng-manager` |
| Suspected bad deploy | Release owner | `#releases` |
| Database saturation after scale-up | DBA on-call | PagerDuty `dba-primary` |

For P1 (full API outage), escalate immediately per the severity guide in [[index]].
