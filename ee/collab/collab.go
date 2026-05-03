// Package collab provides real-time collaborative editing via CRDT (Yjs).
//
// Gated behind [license.FeatureCollab].
//
// The core open-source editor supports single-user editing with SSE-based
// live updates and ETag optimistic locking. This package adds:
//   - Yjs/HocusPocus CRDT document synchronization
//   - Live cursors and presence indicators
//   - Threaded comments with resolve/unresolve workflow
//   - @mentions with notifications
//   - Periodic CRDT state → file + git commit flush
package collab
