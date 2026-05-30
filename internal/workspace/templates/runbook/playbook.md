# Agent Playbook — Runbook

Operational knowledge for on-call and platform teams. When connected
via MCP, use these operations to document incidents, write procedures,
and run blameless postmortems.

## Quick Start

1. Call `kiwi_context` to get this playbook + schema + index in one call
2. Call `kiwi_tree` to see the current structure
3. Use the operations below to document and maintain

## Document Incident

When an incident occurs:

1. `kiwi_read` `incidents/template.md` for the incident template.
2. `kiwi_write` to `incidents/YYYY-MM-DD-<slug>.md` with:
   ```yaml
   ---
   title: "Incident title"
   date: YYYY-MM-DD
   severity: P1 | P2 | P3 | P4
   status: active
   on-call: person-name
   related-alert: alert-name
   detection-minutes:
   mitigation-minutes:
   users-affected:
   error-budget-impact:
   tags: [service, area]
   ---
   ```
3. Fill in Trigger, Impact, and initial Timeline entries.
4. `kiwi_append` timeline entries as the incident progresses.
5. Link to relevant `[[procedures/<name>]]` used during response.
6. Once resolved: set `status: resolved`, fill in resolution metrics.
7. Update `index.md` with the new incident.

## Write Procedure

When documenting a new operational procedure:

1. `kiwi_search` to check if a similar procedure exists.
2. `kiwi_write` to `procedures/<slug>.md` with:
   ```yaml
   ---
   title: "Procedure Name"
   tags: [service, area]
   status: active
   last-reviewed: YYYY-MM-DD
   last-tested: YYYY-MM-DD
   test-cadence: quarterly
   estimated-time: "15 minutes"
   ---
   ```
3. Structure the body with these sections:
   - **When to Use** — trigger conditions
   - **Prerequisites** — access, tools, permissions needed
   - **Steps** — numbered, with exact commands and expected output
   - **Verification** — how to confirm success
   - **Rollback** — how to undo if something goes wrong
4. Cross-link to related procedures with `[[wikilinks]]`.
5. Update `index.md`.

## Run Blameless Postmortem

After an incident is resolved (within 3-5 business days):

1. `kiwi_read` `postmortems/template.md` for the template.
2. `kiwi_write` to `postmortems/YYYY-MM-DD-<slug>.md` linking
   back to the incident with `[[incidents/<file>]]`.
3. Include the blameless notice at the top (from template).
4. Fill in all sections:
   - **Summary** — 2-3 sentences for broad audience
   - **Impact** — quantified (users, SLO, revenue)
   - **Timeline** — precise UTC timestamps with decision points
   - **Contributing Factors** — use 5 Whys, go ≥ 3 levels deep
   - **What Went Well** — reinforce good practices
   - **What Went Wrong** — systems focus, not people
   - **Action Items** — named owner + deadline + ticket for each
   - **Lessons Learned** — what to remember
5. Verify against the Quality Checklist at the bottom.
6. `kiwi_search` for related past incidents to find patterns.
7. Update `index.md` and link from the incident file.

### Postmortem Anti-Patterns

Avoid these common mistakes:
- **Stopping at 2 Whys** — surface causes get fixed, root causes don't
- **Blame language** — "X broke it" → "The system allowed X to happen"
- **Team-level owners** — "Engineering will fix" → "@jane will add the test by June 15"
- **Too many action items** — 3-5 high-impact items > 15 that never get done
- **Writing weeks later** — details fade; schedule within 3-5 business days

## Maintain

Run periodically:

1. `kiwi_lint` with `path` — check individual files for structural issues.
2. `kiwi_analytics` — find orphaned procedures and stale docs.
3. Find procedures past their test cadence:
   ```
   kiwi_query("TABLE _path, title, last-tested, test-cadence WHERE status = 'active' AND last-tested < date_sub(now(), 90)")
   ```
4. Review incidents without linked postmortems:
   ```
   kiwi_query("TABLE _path, title, date WHERE postmortem IS NULL AND status = 'resolved'")
   ```
5. Check open postmortem action items.
6. Bump `last-reviewed` on procedures that are still accurate.

**Best practice:** After every `kiwi_write`, call `kiwi_lint` on the same path.
The server auto-formats cosmetic issues; `kiwi_lint` only reports semantic fixes.

### Game Day Maintenance

Track game day testing of procedures:

1. Find untested procedures:
   ```
   kiwi_query("TABLE _path, title, last-tested WHERE last-tested IS NULL OR last-tested < date_sub(now(), 180)")
   ```
2. After a game day test: update `last-tested` in the procedure's frontmatter.
3. If the procedure failed during the game day: fix it and add a note to `log.md`.

## Quality Rules

- **One incident per file.** Named `YYYY-MM-DD-<slug>.md`.
- **Procedures are reusable.** Write them generically, link from incidents.
- **Every incident gets a postmortem** (even brief ones).
- **Postmortems are blameless.** Systems focus, not individual blame.
- **Action items are actionable.** Named person, deadline, ticket.
- **Frontmatter required.** At least `title`, `date`, `severity` for incidents.
- **No orphans.** All procedures reachable from `index.md`.
- **Procedures are tested.** `last-tested` should never be empty on active procedures.
- **Metrics are tracked.** Detection, mitigation, and resolution times on every incident.
