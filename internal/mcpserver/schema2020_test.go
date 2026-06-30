package mcpserver

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestValidateRegisteredToolSchemasAllowsLocalRef(t *testing.T) {
	s := server.NewMCPServer("kiwifs", "1.0.0")
	s.AddTool(
		mcp.NewTool("good_tool", mcp.WithRawInputSchema(json.RawMessage(`{"$ref":"#/$defs/foo","$defs":{"foo":{"type":"string"}}}`))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		},
	)
	if err := validateRegisteredToolSchemas(s); err != nil {
		t.Fatalf("local $ref should be allowed: %v", err)
	}
}

func TestIsExternalRef(t *testing.T) {
	if !isExternalRef("https://example.com/schema.json") {
		t.Fatal("https ref should be external")
	}
	if isExternalRef("#/$defs/foo") {
		t.Fatal("fragment ref should be local")
	}
}
