# Agent Playbook — Runbook

Operational knowledge for on-call and platform teams. When connected
via MCP, use these operations to document incidents, write procedures,
and run postmortems.

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
   status: active | mitigated | resolved
   on-call: person-name
   tags: [service, area]
   ---
   ```
3. `kiwi_append` timeline entries as the incident progresses.
4. Link to relevant `[[procedures/<name>]]` used during response.
5. Update `index.md` with the new incident.

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
   ---
   ```
3. Include step-by-step instructions with commands.
4. Cross-link to related procedures with `[[wikilinks]]`.
5. Update `index.md`.

## Run Postmortem

After an incident is resolved:

1. `kiwi_search` for the incident file.
2. `kiwi_write` to `postmortems/YYYY-MM-DD-<slug>.md` linking
   back to the incident with `[[incidents/<file>]]`.
3. Include: timeline, root cause, impact, action items.
4. `kiwi_search` for related past incidents to find patterns.
5. Update `index.md`.

## Maintain

Run periodically:

1. `kiwi_analytics` — find orphaned procedures and stale docs.
2. `kiwi_search` for procedures with outdated commands.
3. Review incidents without linked postmortems.
4. Bump `last-reviewed` on procedures that are still accurate.

## Quality Rules

- **One incident per file.** Named `YYYY-MM-DD-<slug>.md`.
- **Procedures are reusable.** Write them generically, link from incidents.
- **Every incident gets a postmortem** (even brief ones).
- **Frontmatter required.** At least `title`, `date`, `severity` for incidents.
- **No orphans.** All procedures reachable from `index.md`.
