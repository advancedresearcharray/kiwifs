// Package audit provides the enterprise audit log UI, data retention policies,
// and compliance export.
//
// Gated behind [license.FeatureAuditLog].
//
// The underlying data is always in git (every write is a commit). This package
// adds:
//   - Structured audit event stream (parsed from git log + API request context)
//   - Searchable, filterable audit log API for the web UI
//   - Configurable data retention policies (auto-archive, auto-delete)
//   - Compliance export (SOC 2, ISO 27001, HIPAA audit trail formats)
//   - Page verification / review workflow (request review, approve, mark verified)
package audit
