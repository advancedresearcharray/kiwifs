---
title: "Experiment 001: Baseline Measurement"
date: 2026-01-01
hypothesis: "Establishing baseline metrics for comparison with future experiments"
research-question: "Q1"
status: completed
result: positive
protocol: "Standard load test protocol"
environment: "Linux 6.1, 8-core, 32GB RAM, Go 1.22"
duration: "24 hours"
raw-data: "data/exp-001/"
sample-size: "86,400 data points (1/sec)"
tags: [baseline, setup, performance]
references: [literature/example-paper.md]
---

# Experiment 001 — Baseline Measurement

> **This is an example experiment.** Replace it with your first real
> experiment, or delete it once you've created your own.

## Hypothesis

Establish baseline performance metrics so future experiments have
a reference point for comparison.

## Variables

- **Independent:** none (baseline — default configuration)
- **Dependent:** throughput, latency (p50, p99), error rate
- **Controlled:** hardware, OS, network conditions, data set

## Environment

- **OS:** Linux 6.1 (Ubuntu 22.04)
- **Hardware:** 8-core CPU, 32GB RAM, NVMe SSD
- **Software:** Go 1.22, PostgreSQL 16.1
- **Configuration:** default / unmodified
- **Network:** isolated test network, 1Gbps

## Protocol

1. Configure the standard test environment
2. Run the default configuration with no modifications
3. Collect metrics at 1-second intervals over 24 hours
4. Aggregate results into p50, p95, p99 percentiles
5. Record results below

## Observations

_Notes taken during the experiment. Include timestamps if relevant._

- _e.g., System stable throughout the measurement period_
- _e.g., Noted periodic GC pauses every ~30 minutes_

## Results

| Metric | Value | Notes |
|--------|-------|-------|
| _Throughput_ | _X req/s_ | _baseline_ |
| _P50 latency_ | _X ms_ | _baseline_ |
| _P99 latency_ | _X ms_ | _baseline_ |
| _Error rate_ | _X%_ | _baseline_ |

## Conclusions

_What did you learn? How does this inform the next experiment?_

## Reproduction Steps

To re-run this experiment:

1. Provision a machine matching the environment above
2. Deploy the application at commit `abc123`
3. Run: `./benchmark --duration=24h --rate=100 --output=data/exp-001/`
4. Compare output against the results table above

## Next Steps

- [ ] Design [[experiments/exp-002-placeholder|Experiment 002]] to test first variation
- [ ] Document any anomalies for investigation

## Related

- Literature: [[literature/example-paper]]
- Research question: Q1 (see `questions.md`)
