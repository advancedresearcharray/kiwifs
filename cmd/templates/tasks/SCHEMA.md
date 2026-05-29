# Task Schema

Each task is a markdown file with YAML frontmatter. Tasks are displayed
on the Kanban board and queryable via DQL.

## Required Fields

| Field | Type | Values / Notes |
|-------|------|----------------|
| `type` | string | Always `task` |
| `title` | string | Short summary of what needs to be done |
| `status` | string | See lifecycle below |
| `priority` | integer | See priority scale below |

## Optional Fields

| Field | Type | Values / Notes |
|-------|------|----------------|
| `assignee` | string | Agent or human identifier |
| `blocked-by` | string[] | List of task paths that must be `done` first |
| `due` | date | ISO 8601 deadline |
| `tags` | string[] | Topic tags for filtering |
| `claimed-by` | string | Set automatically by claim endpoint |
| `claimed-at` | datetime | Set automatically by claim endpoint |
| `lease-expires` | datetime | Set automatically by claim endpoint |

## Status Lifecycle

```
backlog → todo → in_progress → review → done
                      ↓                   ↑
                   blocked ──────────────→┘
                                    cancelled
```

| Status | Meaning |
|--------|---------|
| `backlog` | Acknowledged but not yet ready to work on |
| `todo` | Ready to be picked up |
| `in_progress` | Someone (human or agent) is actively working on it |
| `review` | Work complete, awaiting review/approval |
| `blocked` | Cannot proceed — see `blocked-by` for dependencies |
| `done` | Complete and verified |
| `cancelled` | No longer needed |

## Priority Scale

| Priority | Label | Meaning |
|----------|-------|---------|
| 0 | **Urgent** | Drop everything. Production is broken or critical deadline. |
| 1 | **High** | Do this next. Important and time-sensitive. |
| 2 | **Medium** | Normal priority. Do in order. |
| 3 | **Low** | Nice to have. Do when higher-priority work is clear. |
| 4 | **Backlog** | Someday/maybe. Not urgent, not blocking anything. |

## Task Decomposition

- **Use sub-tasks** (checkboxes in the body) when steps are sequential
  and one agent/person handles them all in a single session.
- **Use separate task files** with `blocked-by` when steps can be
  parallelized, span multiple sessions, or need different assignees.
- Rule of thumb: can one agent finish everything in a single session?
  → Sub-tasks. Otherwise → separate files.

## Acceptance Criteria

Every task should include acceptance criteria in the body — a clear
definition of "done" that the assignee (human or agent) can verify.

```markdown
## Acceptance Criteria

- [ ] First criterion
- [ ] Second criterion
- [ ] Tests pass
```
