---
title: Scale Up
tags: [capacity, scaling, performance]
status: draft
last-reviewed: 2026-01-01
---

# Scale Up

Add capacity during a traffic spike, performance degradation, or
when approaching resource limits.

## When to Use

- CPU/memory utilization exceeding 80% sustained
- Request latency increasing beyond SLA thresholds
- Known upcoming traffic event (launch, promotion, migration)
- Auto-scaling is not responding fast enough

## Prerequisites

- [ ] Access to cloud console or CLI
- [ ] Permission to modify instance counts / resource limits
- [ ] Understanding of current architecture bottlenecks

## Steps

### 1. Identify the Bottleneck

```bash
# Check which component is saturated
# CPU? Memory? Connections? Disk I/O?

# Example: Kubernetes pod resource usage
kubectl top pods -n <namespace>

# Example: check connection pool
# Query your monitoring dashboard: <URL>
```

### 2. Scale the Service

```bash
# Option A: Kubernetes horizontal scaling
kubectl scale deployment/<name> -n <namespace> --replicas=<N>

# Option B: Cloud provider auto-scaling group
# aws autoscaling set-desired-capacity \
#   --auto-scaling-group-name <asg-name> \
#   --desired-capacity <N>

# Option C: Vertical scaling (resize instance)
# Requires downtime — schedule or use rolling update
```

### 3. Verify

```bash
# Confirm new instances are running
kubectl get pods -n <namespace>

# Confirm load is distributed
# Check monitoring dashboard for reduced latency / CPU

# Confirm health checks pass
curl -s https://api.example.com/health
```

### 4. Monitor

- Watch for 15-30 minutes to confirm stability
- Verify the bottleneck metric has improved
- Check for cascading issues (e.g., database connection limits)

## Scale Down

After the traffic event passes:

1. Monitor metrics for 1-2 hours at reduced load
2. Reduce replicas gradually (not all at once)
3. Verify latency and error rates remain stable

## Rollback

If scaling up causes issues (e.g., database overwhelmed by new connections):

1. Scale back to previous replica count
2. Investigate the cascading failure
3. Address the downstream bottleneck before scaling again

## Related

- [[procedures/deploy-rollback]] — if scaling was triggered by a bad deploy
