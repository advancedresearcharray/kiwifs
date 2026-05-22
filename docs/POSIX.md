<p align="center">
  <a href="../README.md">README</a> · <a href="FAQ.md">FAQ</a> · <a href="ARCHITECTURE.md">Architecture</a> · <a href="API.md">API</a> · <a href="EXAMPLES.md">Examples</a>
</p>

# POSIX Compliance

KiwiFS stores real files on a real filesystem, not blobs in a database. The degree of POSIX compliance depends on how you access the data.

---

## Compliance by Access Path

| Access path | POSIX level | Notes |
|---|---|---|
| **Direct filesystem** | Full | Real files, crash-safe atomic writes, mmap works |
| **NFS mount** | Near-full | Userspace NFSv3, symlinks, open-unlink, advisory locking, stable handles across restarts |
| **FUSE mount** | Near-full | Remote FUSE client, symlinks, directory rename, sub-second mtime, O_APPEND |
| **WebDAV** | Partial | MOVE/COPY/MKCOL/DELETE, buffered writes with spill-to-disk |
| **REST API** | N/A | HTTP semantics (ETag concurrency, not POSIX) |
| **S3 API** | N/A | S3-compatible, not POSIX |
| **MCP** | N/A | Tool calls, not file ops |

---

## What Works

| POSIX semantic | NFS | FUSE | How |
|---|---|---|---|
| **Atomic writes** | Yes | Yes | `write -> fsync -> rename(tmp, target) -> fsync(dir)` |
| **rename(2)** | Yes | Yes | Files via pipeline (atomic); directories via bulk endpoint |
| **O_APPEND** | Yes | Yes | FUSE fetches existing content on open, writes at correct EOF offset |
| **O_TRUNC / ftruncate** | Yes | Yes | NFS `Truncate()` with 64MB bounds; FUSE `Setattr` with `FATTR_SIZE` |
| **Symlinks** | Yes | Yes | Real `os.Symlink` on NFS; `Content-Type: application/x-symlink` on FUSE + REST |
| **readlink** | Yes | Yes | NFS via `os.Readlink`; FUSE via `/api/kiwi/readlink`; REST API endpoint |
| **Open-then-delete** | Yes | -- | NFS defers deletion until last file handle closes (POSIX unlink semantics) |
| **fsync** | Yes | Yes | NFS `Sync()` pushes through pipeline; FUSE `Fsync()` PUTs to server |
| **Sub-second mtime** | -- | Yes | FUSE `Getattr` reports `Mtimensec` from `Last-Modified` header |
| **Advisory locking** | Yes | -- | NFS has process-local `Lock()`/`Unlock()` per file handle |
| **Directory rename** | Yes | Yes | FUSE calls `/api/kiwi/rename-dir`; NFS uses `os.Rename` + bulk re-index |
| **readdir** | Yes | Yes | Both hide internal dirs (`.git`, `.kiwi`) |
| **stat** | Yes | Yes | Size, mode, mtime are real values, not synthetic |
| **EFBIG on oversize** | Yes | Yes | 64MB `maxFileSize` limit returns proper errno / HTTP error |
| **mmap** | -- | Passthrough | Works on NFS mount (kernel handles it); FUSE is over HTTP so no kernel mmap |
| **Path safety** | Yes | Yes | `GuardPath` blocks traversal, null bytes, control chars, 255-byte segment limit |

---

## Concurrency and Durability

- **Optimistic locking** via ETags (content hash). Writes with `If-Match` headers get HTTP 409 on conflict. `If-Match: *` is handled per RFC 7232 section 3.1.
- **Serialized writes** through a single mutex. Concurrent writers are safely queued regardless of protocol.
- **Single-instance guard** via `flock(2)` on `.kiwi/server.lock`. The kernel releases it automatically on any form of process exit (including SIGKILL), so there's no stale lock problem.
- **Crash recovery** for stale `index.lock` files via a background watcher (10s interval, 60s threshold). Git subprocesses receive SIGTERM before SIGKILL, giving them a chance to release locks.
- **Line-ending integrity** via `core.autocrlf=false` + `* -text` in `.gitattributes`. ETags always match raw content. Writes to `.gitattributes` are blocked by the API.
- **Frontmatter bomb protection**: YAML frontmatter blocks exceeding 64KB are silently treated as empty (headings still extracted).
- **Stable NFS handles** derived from `SHA-256(namespaceUUID + path)`, surviving server restarts. No more ESTALE.

---

## Intentionally Not Supported

| POSIX semantic | Why |
|---|---|
| **Hard links** | Would break git versioning (one blob, multiple paths) |
| **chmod / chown** | No user/group model. Auth is API-key / OIDC, not POSIX uid/gid |
| **POSIX ACLs** | Access control is at the HTTP/space level |
| **Extended attributes (xattr)** | Frontmatter serves the same purpose |
| **Distributed locking** | Locks are process-local; use `If-Match` for cross-client concurrency |
