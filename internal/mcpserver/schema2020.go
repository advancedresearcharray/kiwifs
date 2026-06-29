package mcpserver

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

// validateRegisteredToolSchemas rejects tool definitions that reference external JSON Schema URIs (SSRF risk per MCP 2026 spec).
func validateRegisteredToolSchemas(s *server.MCPServer) error {
	for _, st := range s.ListTools() {
		if st == nil {
			continue
		}
		tool := st.Tool
		if tool.RawInputSchema != nil {
			var schema any
			if err := json.Unmarshal(tool.RawInputSchema, &schema); err != nil {
				return fmt.Errorf("tool %q input schema: %w", tool.Name, err)
			}
			if hasExternalSchemaRef(schema) {
				return fmt.Errorf("tool %q input schema: external $ref URIs are not allowed", tool.Name)
			}
		}
		if tool.InputSchema.Type != "" {
			raw, err := json.Marshal(tool.InputSchema)
			if err != nil {
				return fmt.Errorf("tool %q input schema: %w", tool.Name, err)
			}
			var schema any
			if err := json.Unmarshal(raw, &schema); err != nil {
				return fmt.Errorf("tool %q input schema: %w", tool.Name, err)
			}
			if hasExternalSchemaRef(schema) {
				return fmt.Errorf("tool %q input schema: external $ref URIs are not allowed", tool.Name)
			}
		}
	}
	return nil
}
