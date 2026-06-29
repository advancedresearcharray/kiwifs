<p align="center">
  <a href="../README.md">README</a> · <a href="FAQ.md">FAQ</a> · <a href="POSIX.md">POSIX</a> · <a href="API.md">API</a> · <a href="EXAMPLES.md">Examples</a>
</p>

# Architecture

KiwiFS is a single Go binary that turns a folder of markdown files into a searchable, versioned, multi-protocol knowledge server with an embedded web UI.

---

## System Overview

```
┌──────────────────────────────────────────────────────────┐
│  KiwiFS                                    single Go binary
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Web UI (embedded via go:embed)                    │  │
│  │  shadcn/ui · CodeMirror · react-markdown · Sigma.js │  │
│  └────────────────────┬───────────────────────────────┘  │
│                       │                                  │
│  ┌────────────────────▼───────────────────────────────┐  │
│  │  Access Protocols                                  │  │
│  │  REST :3333 · MCP · NFS :2049 · S3 :3334 · WebDAV │  │
│  └────────────────────┬───────────────────────────────┘  │
│                       │                                  │
│  ┌────────────────────▼───────────────────────────────┐  │
│  │  Write Pipeline                                    │  │
│  │  Storage → Git commit → Index update → SSE event   │  │
│  └────────────────────┬───────────────────────────────┘  │
│                       │                                  │
│  ┌────────────────────▼───────────────────────────────┐  │
│  │  Core                                              │  │
│  │  Storage · Git versioning · FTS5 + Vector search   │  │
│  │  Watcher (fsnotify) · SSE events · Schema/lint     │  │
│  └────────────────────┬───────────────────────────────┘  │
│                       │                                  │
│  ┌────────────────────▼───────────────────────────────┐  │
│  │  State                                             │  │
│  │  .git/ (audit WAL)  ·  .kiwi/state/ (indexes)     │  │
│  └────────────────────┬───────────────────────────────┘  │
│                       │                                  │
│  ┌────────────────────▼───────────────────────────────┐  │
│  │  Filesystem: local · NFS · EFS · JuiceFS · FUSE-S3 │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

---

## Design Principles

1. **Files are the source of truth.** Every artifact is a plain markdown file. No proprietary format. Delete the search index and the files remain. `cat file.md` always works.

2. **Two interfaces, one truth.** The web UI and the agent filesystem read/write the same files. No sync. No eventual consistency. One folder, two ways to access it.

3. **Search is derivative.** The FTS5 index, the vector index, the metadata table are all rebuildable from files. `kiwifs reindex` and you're back. The folder is the truth, never the index.

4. **Storage-agnostic.** KiwiFS depends on `open()`, `read()`, `write()`, `listdir()`. It doesn't care if the folder is on a laptop SSD, NFS, EFS, JuiceFS, or a FUSE-mounted S3 bucket.

5. **Git as the WAL.** Instead of building a custom write-ahead log, every write is a git commit. Crash recovery, audit trail, tamper detection, replication, all for free.

6. **Embeddable.** The Go library (`pkg/kiwi`) lets you embed KiwiFS in any Go application. The web UI components are built for future standalone use as an npm package.

---

## Write Pipeline

Every write, regardless of protocol, flows through the same pipeline:

```
Client (REST / NFS / S3 / WebDAV / FUSE / MCP)
  │
  ▼
Write Pipeline (single mutex)
  │
  ├── 1. Write file to disk (atomic: tmp → fsync → rename → dirsync)
  ├── 2. Git add + commit (atomic, audit trail)
  ├── 3. Update search index (FTS5 + vector + metadata)
  ├── 4. Update wiki link index (backlinks)
  └── 5. Broadcast SSE event to connected clients
```

The mutex serializes all writes. This is intentional: knowledge bases are read-heavy (agents and humans read far more than they write), and serialization eliminates an entire class of concurrency bugs. For write-heavy workloads, the bottleneck is git, not the mutex.

---

## Concurrency Model

KiwiFS uses **optimistic locking** via HTTP ETags:

1. Client reads a file and receives its ETag (git blob hash)
2. Client writes with `If-Match: <etag>`
3. If the file changed since the read, the server returns 409 Conflict with the current version
4. Client merges and retries

This is standard HTTP conditional requests (RFC 7232), not a custom protocol.

---

## Search Architecture

Three tiers, each building on the previous:

```
Tier 1: grep          Tier 2: SQLite FTS5       Tier 3: Vector
(zero deps)           (default, BM25 ranked)    (semantic similarity)
                      ┌──────────────┐          ┌──────────────┐
                      │ search.db    │          │ Embedder     │
                      │ FTS5 index   │          │ (pluggable)  │
                      │ file_meta    │          ├──────────────┤
                      │ links table  │          │ VectorStore  │
                      └──────────────┘          │ (pluggable)  │
                                                └──────────────┘
```

The search index is always a **derivative** of the files. Delete `.kiwi/state/` and run `kiwifs reindex` to rebuild everything.

### Pluggable Vector Search

Two independent interfaces that can be mixed and matched:

| Embedder | Vector Store |
|---|---|
| OpenAI, Ollama, Cohere, Vertex AI, Bedrock, custom HTTP | sqlite-vec (default), Qdrant, pgvector, Pinecone, Weaviate, Milvus |

Default setup (sqlite-vec + OpenAI) needs one env var and zero infrastructure. For fully offline: Ollama + sqlite-vec.

---

## Access Protocols

| Protocol | Port | Use case | Implementation |
|---|---|---|---|
| **REST API** | 3333 | Web frontend, scripts, CI/CD | Echo (Go) |
| **MCP** | stdio / 8080 | AI agents (Claude, Cursor, etc.) | In-process or HTTP |
| **NFS** | 2049 | Docker, Kubernetes (native mount) | `willscott/go-nfs` (userspace, pure Go) |
| **S3** | 3334 | Backup, data pipelines | `gofakes3` (minimal S3 surface) |
| **WebDAV** | 3335 | Windows mapped drives, legacy tools | `golang.org/x/net/webdav` |
| **FUSE** | client | Developer workstations | `hanwen/go-fuse` (Google) |

All protocols funnel through the same write pipeline. Every write gets a git commit, a search index update, and an SSE broadcast.

---

## Directory Structure

### Data Directory

```
knowledge/                    (user content)
├── SCHEMA.md                 Structure and frontmatter conventions
├── index.md                  Table of contents
├── log.md                    Append-only chronological record
├── pages/                    Durable knowledge pages
├── episodes/                 Per-session episodic notes
├── .git/                     Git repository (audit trail, versioning)
│   ├── objects/              Immutable, content-addressed storage
│   ├── refs/                 Branch pointers
│   └── logs/                 Reflog
└── .kiwi/                    KiwiFS system directory
    ├── config.toml           Server and search configuration
    ├── playbook.md           Agent-readable operation guide
    ├── server.lock           Single-instance flock
    ├── state/
    │   └── search.db         SQLite: FTS5 + metadata + vector indexes
    ├── comments/             Inline comment annotations (JSON)
    └── templates/            Page templates for slash commands
```

### Source Code

```
kiwifs/
├── cmd/              CLI commands (serve, init, mcp, query, import, export, ...)
├── internal/
│   ├── api/          REST API handlers
│   ├── bootstrap/    Dependency wiring
│   ├── pipeline/     Write pipeline (git + index + SSE)
│   ├── search/       grep + SQLite FTS5 + metadata index
│   ├── storage/      Filesystem abstraction
│   ├── vectorstore/  Vector search backends
│   ├── versioning/   Git, copy-on-write, noop
│   ├── mcpserver/    MCP server (62 tools)
│   ├── nfs/          NFS server
│   ├── s3/           S3-compatible API
│   ├── webdav/       WebDAV server
│   ├── fuse/         FUSE client
│   ├── spaces/       Multi-space manager
│   ├── dataview/     DQL parser and query engine
│   ├── importer/     Data import from 19 sources
│   ├── exporter/     Export to JSONL/CSV
│   ├── janitor/      Scheduled health scans
│   ├── memory/       Episodic vs semantic memory
│   ├── comments/     Inline comment annotations
│   └── links/        Wiki link extraction and backlinks
├── pkg/kiwi/         Public Go library
├── ui/               React + TypeScript + shadcn/ui
└── main.go
```

---

## Key Dependencies

### Backend (Go)

| Library | Purpose | License |
|---|---|---|
| `echo` | HTTP server, middleware, SSE | MIT |
| `modernc.org/sqlite` | SQLite FTS5 (pure Go, no CGo) | BSD |
| `willscott/go-nfs` | Userspace NFS server | MIT |
| `hanwen/go-fuse` | FUSE client | BSD |
| `golang.org/x/net/webdav` | WebDAV server | BSD |
| `fsnotify` | File watching | BSD |
| `goldmark` | Markdown parsing | MIT |
| `cobra` | CLI framework | Apache 2.0 |

### Frontend (TypeScript)

| Library | Purpose | License |
|---|---|---|
| shadcn/ui + Radix | UI primitives, accessible components | MIT |
| CodeMirror | Markdown source editor | MIT |
| react-markdown | Markdown rendering | MIT |
| Sigma.js + Graphology | Knowledge graph visualization | MIT |
| cmdk | Command palette (Cmd+K) | MIT |

---

## Further Reading

- [API Reference](API.md) for the full REST API surface
- [POSIX Compliance](POSIX.md) for filesystem semantics details
- [Examples](EXAMPLES.md) for agent workflows and query patterns
