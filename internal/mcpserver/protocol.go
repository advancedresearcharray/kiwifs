package mcpserver

import "time"

// MCP 2026-07-28 protocol constants (SEP-2575, SEP-2243, SEP-2549).
const (
	ProtocolVersion20260728 = "2026-07-28"

	HeaderMCPMethod          = "Mcp-Method"
	HeaderMCPName            = "Mcp-Name"
	HeaderMCPProtocolVersion = "Mcp-Protocol-Version"

	CacheScopePublic  = "public"
	CacheScopePrivate = "private"

	MethodServerDiscover = "server/discover"

	defaultListCacheTTL   = 5 * time.Minute
	defaultDiscoverCacheTTL = time.Hour
)

var supportedProtocolVersions = []string{
	ProtocolVersion20260728,
	"2025-11-25",
	"2025-03-26",
	"2024-11-05",
}

const jsonSchema2020URI = "https://json-schema.org/draft/2020-12/schema"
