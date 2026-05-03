# Agent Playbook — Team Wiki

This wiki is organized by functional area, with each folder containing
an `index.md` landing page. When connected via MCP, use these
operations to maintain it.

## Quick Start

1. Call `kiwi_context` to get this playbook + schema + index in one call
2. Call `kiwi_tree` to see the current folder structure
3. Use the operations below to create, organize, and maintain

## Create Page

When adding new content:

1. **Find the right folder.** `kiwi_tree` to see the area structure.
   If no folder fits, create one with an `index.md` landing page.
2. **Check for duplicates.** `kiwi_search` for key terms first.
3. **Write the page.** `kiwi_write` to `<area>/<slug>.md` with:
   ```yaml
   ---
   title: "Page Title"
   owner: team-or-person
   status: active
   tags: [area, topic]
   last-reviewed: YYYY-MM-DD
   ---
   ```
4. **Cross-link.** Add `[[wikilinks]]` to related pages.
   Use `kiwi_search` to find them.
5. **Update the area index.** `kiwi_read` the area's `index.md`,
   add a link to the new page, `kiwi_write` it back.

## Organize

When restructuring or cleaning up:

1. `kiwi_tree` to understand current layout.
2. `kiwi_rename` to move pages — links auto-update.
3. Ensure every folder has an `index.md` landing page.
4. Keep folder depth ≤ 3 levels for discoverability.

## Maintain

Run periodically or when asked:

1. `kiwi_analytics` — find stale pages, orphans, broken links.
2. For stale pages: update content, bump `last-reviewed`.
3. For orphans: add links from related pages or the area index.
4. For broken links: `kiwi_search` for the intended target, fix.
5. `kiwi_search` for "TODO" or "TBD" to find incomplete pages.

## Quality Rules

- **Folder-per-area.** Each top-level folder is a functional area.
- **Every folder has `index.md`.** The landing page for that area.
- **Keep pages short.** Split when a page exceeds ~300 lines.
- **Frontmatter required.** At least `title`, `owner`, and `tags`.
- **No orphans.** Every page reachable from its area's `index.md`.
