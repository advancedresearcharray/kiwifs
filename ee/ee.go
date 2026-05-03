// Package ee is the entry point for KiwiFS Enterprise features.
//
// All enterprise functionality lives under ee/ and is gated at runtime
// by a valid license key (KIWI_LICENSE_KEY environment variable).
// Without a license, these packages are compiled into the binary but
// remain dormant — no enterprise code paths execute.
//
// Enterprise features:
//   - ee/auth:        SAML 2.0, LDAP, SCIM provisioning, MFA/TOTP
//   - ee/permissions: Page-level and folder-level ACLs, custom roles
//   - ee/audit:       Audit log UI, data retention policies, compliance export
//   - ee/ai:          AI Chat (RAG), AI writing assistant, AI search reranking
//   - ee/collab:      CRDT/Yjs real-time collaborative editing, live cursors
//   - ee/connectors:  Enterprise data sources (Postgres, Confluence, Notion, etc.)
//   - ee/license:     License key validation
//
// See ee/LICENSE for the enterprise license terms.
package ee

import "github.com/kiwifs/kiwifs/ee/license"

// Loader is the global license validator, initialized at server startup.
var Loader = license.New()
