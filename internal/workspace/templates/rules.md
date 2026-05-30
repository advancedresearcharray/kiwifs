# Agent Rules

Rules in this file are included in `kiwi_context` for every connected agent.

## Always

- Search existing pages before creating new ones — avoid duplicates
- After implementing a feature, update relevant documentation
- Link related pages with [[wiki links]]
- Set `X-Actor` header on every write operation
- Include frontmatter with at least `title` and `tags` on every page
- Run `kiwi_lint` after every write to catch structural issues
- Use lowercase-hyphenated slugs for filenames (e.g., `my-topic.md`)
- Cite sources with [[wikilinks]] or URLs in `source-uri` / `derived-from`
- Read a page before overwriting — preserve existing content
- Resolve contradictions explicitly — never silently overwrite

## Never

- Delete pages without confirmation from the user
- Overwrite content without reading it first
- Store secrets, API keys, passwords, or credentials in pages
- Create duplicate pages without checking via `kiwi_search` first
- Leave orphan pages unreachable from `index.md`
- Use vague filenames like `notes.md` or `temp.md` — be descriptive
- Skip frontmatter on any page

## Quality Standards

- One concept per page — split pages over 300 lines
- Every page reachable from `index.md` within 2 hops
- Every `[[wikilink]]` must resolve to an existing page
- Frontmatter fields use the schema defined in `SCHEMA.md`
- Provenance is recorded on agent-created pages (`derived-from` or `provenance` parameter)
