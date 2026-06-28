# ADR Playbook

Architecture Decision Records with MADR format and enforced status lifecycle.
When connected via MCP, use these operations to propose, review, and query decisions.

## Quick Start

1. Call `kiwi_context` to get this playbook + schema + decision log
2. Call `kiwi_workflow_board` with `workflow: adr` to see decisions by status
3. Use the operations below to propose and advance ADRs

## Propose a Decision

When a significant technical choice needs recording:

1. `kiwi_search` to check if a related ADR already exists.
2. Copy `.kiwi/templates/adr.md` or `kiwi_write` to `decisions/ADR-NNN-slug.md`:
   ```yaml
   ---
   type: adr
   title: "ADR-NNN: Short Title"
   status: proposed
   date: 2026-06-19
   deciders: [team-or-person]
   workflow: adr
   state: proposed
   domain: auth
   decision: One-line summary
   tags: [adr, topic]
   ---
   ```
3. Fill MADR sections: Context, Decision Drivers, Considered Options,
   Decision Outcome, Consequences.
4. Add a row to `index.md` or rely on embedded DQL views.

The pipeline auto-assigns `adr_number` when writing to `decisions/` without one.

## Advance Status

Move ADRs through the lifecycle after review:

```
kiwi_workflow_advance(path: "decisions/my-adr.md", workflow: "adr", target_state: "accepted")
```

Valid progression:

- `proposed → accepted` (approved)
- `proposed → deprecated` (rejected proposal)
- `accepted → deprecated` (no longer recommended)
- `accepted → superseded` (replaced by newer ADR)
- `deprecated → superseded` (cleanup)

Terminal state `superseded` cannot transition further. Skipping states
(e.g. `proposed → superseded`) is rejected.

After advancing, update `status` in frontmatter to match `state`.

## Supersede an Accepted ADR

When a decision changes:

1. Create a new ADR with `supersedes: decisions/ADR-NNN-old.md`.
2. Advance the old ADR: `kiwi_workflow_advance(..., target_state: "superseded")`.
3. Set `superseded_by` on the old ADR pointing to the new file.
4. Never edit the body of the accepted ADR — the git history is the audit trail.

## Query Decisions

Before proposing architecture, check existing constraints:

```
kiwi_query("TABLE adr_number, title, status, domain FROM 'decisions/' WHERE type = 'adr' AND status = 'accepted' SORT adr_number ASC")
```

Find ADRs affecting a domain:

```
kiwi_search("authentication ADR")
```

Navigate supersession chains via `kiwi_backlinks` on `supersedes` and
`superseded_by` typed links.

## Validate

Run `kiwifs check --root .` in CI to enforce:

- Required frontmatter (`status`, `date`, `deciders`)
- Valid workflow transitions
- No broken wikilinks in the decision log

## Quarterly Review

Query ADRs past their `review-by` date:

```dql
TABLE adr_number, title, review-by, deciders
FROM "decisions/"
WHERE type = "adr" AND status = "accepted" AND review-by < today()
SORT review-by ASC
```

Revisit or extend the review date after assessment.
