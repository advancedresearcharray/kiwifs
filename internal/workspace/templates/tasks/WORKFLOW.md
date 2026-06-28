---
tracker:
  kind: kiwi
  endpoint: ${KIWIFS_URL:-http://localhost:3333}
  api_prefix: /api/kiwi

polling:
  interval_seconds: 30

workspace:
  root: ./workspaces

agent:
  executable: codex
  idle_timeout_seconds: 1800

hooks:
  after_create: |
    git clone $REPO_URL .
---

# KiwiFS Task Workflow

You are an autonomous coding agent working on a task from the KiwiFS task board.

## Context
{{task.content}}

## Instructions
1. Read the task description and acceptance criteria
2. Implement the changes described
3. Run tests to verify
4. Create a PR with your changes
5. Update the task status to `review`

## Integration Patterns

### Polling for work
```
GET /api/kiwi/query?q=TABLE _path, title, priority WHERE type = "task" AND status = "todo" AND _blocked = false SORT priority ASC
```

### Claiming a task
```
POST /api/kiwi/claim
X-Actor: agent-name
{"path": "tasks/my-task.md", "lease_duration": "30m"}
```

### Updating status
```
PUT /api/kiwi/file?path=tasks/my-task.md
X-Actor: agent-name
```

### Long-polling for changes
```
GET /api/kiwi/changes?feed=longpoll&since=<seq>&timeout=30s
```

### Webhooks for transitions
```
POST /api/kiwi/webhooks
{"url": "https://orchestrator/hook", "path_glob": "tasks/**", "event_types": ["transition"]}
```

Webhook deliveries include `X-Kiwi-Signature-256: sha256=<hex>` where the hex
value is `HMAC-SHA256(secret, raw_request_body)`. Consumers should compare it
with a constant-time equality check.
