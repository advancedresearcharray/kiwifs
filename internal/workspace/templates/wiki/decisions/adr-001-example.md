---
type: decision
title: "ADR-001: Example — Use PostgreSQL for Primary Datastore"
status: accepted
date: 2026-01-01
owner: tech-lead
decision: Use PostgreSQL as the primary relational datastore
decision-drivers:
  - Need ACID transactions for financial data
  - Team has existing PostgreSQL expertise
  - Must support both relational and semi-structured data
alternatives:
  - option: PostgreSQL
    pros: Mature, strong ecosystem, JSONB for semi-structured data
    cons: Requires DBA knowledge for tuning at scale
  - option: MongoDB
    pros: Flexible schema, easy to start
    cons: Weak transactions, harder to enforce data integrity
  - option: SQLite
    pros: Zero-ops, embedded
    cons: Single-writer, no networked access
impact: All services use PostgreSQL; team needs PG expertise
reversal-conditions: If we outgrow single-node PG and need a fundamentally different data model
review-by: 2027-01-01
superseded-by:
supersedes:
linked-docs: []
tags: [adr, database, infrastructure]
---

# ADR-001: Use PostgreSQL for Primary Datastore

> **This is an example ADR.** Replace it with your first real decision,
> or delete it once you've created your own.

## Status

Accepted — 2026-01-01

## Decision Drivers

- Need ACID transactions for financial data integrity
- Team has existing PostgreSQL expertise (3 engineers with production experience)
- Must support both strict relational schemas and semi-structured metadata (JSONB)
- Hosted options must be available in our cloud provider

## Context

We need a primary datastore for the application. The choice affects
every service, developer workflow, and operational procedure.

## Decision

We will use PostgreSQL. It handles our relational data model, supports JSONB
for semi-structured fields, and has a mature ecosystem of tools,
hosting options, and community knowledge.

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| PostgreSQL | Mature, strong ecosystem, JSONB | Requires tuning at scale |
| MongoDB | Flexible schema, easy start | Weak transactions |
| SQLite | Zero-ops, embedded | Single-writer only |

## Consequences

### Positive

- All services share a well-understood data layer
- Strong tooling for migrations, monitoring, and backups
- JSONB covers semi-structured use cases without a second database

### Negative

- Team must invest in PG operational knowledge (connection pooling, vacuuming)
- Schema migrations require coordination across services
- Vertical scaling limits may require read replicas sooner

### Neutral

- We accept vendor-neutral SQL as the query interface

## Reversal Conditions

Revisit this decision if:
- We need a fundamentally different data model (graph, time-series)
- Single-node PG becomes a bottleneck we can't solve with read replicas
- A service has access patterns that are incompatible with relational storage
