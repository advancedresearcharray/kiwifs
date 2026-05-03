// Package permissions provides page-level and folder-level ACLs with custom roles.
//
// Gated behind [license.FeaturePagePerms].
//
// The core open-source RBAC (internal/rbac) handles space-level admin/editor/viewer.
// This package extends it with:
//   - Per-page read/write ACLs
//   - Per-folder inherited permissions
//   - Custom role definitions beyond the built-in three
//   - User and group management UI API endpoints
//   - Directory sync (LDAP/SCIM groups → KiwiFS groups)
package permissions
