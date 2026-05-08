package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func handleDraftCreate(b Backend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		actor, _ := args["actor"].(string)
		if actor == "" {
			actor = "mcp-agent"
		}
		d, err := b.DraftCreate(ctx, actor)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create draft: %v", err)), nil
		}
		data, _ := json.Marshal(d)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleDraftList(b Backend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		drafts, err := b.DraftList(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list drafts: %v", err)), nil
		}
		if len(drafts) == 0 {
			return mcp.NewToolResultText("No active drafts."), nil
		}
		var sb strings.Builder
		fmt.Fprintf(&sb, "%d active draft(s):\n", len(drafts))
		for _, d := range drafts {
			fmt.Fprintf(&sb, "  %s (branch: %s, actor: %s, created: %s)\n", d.ID, d.Branch, d.Actor, d.CreatedAt)
		}
		return mcp.NewToolResultText(sb.String()), nil
	}
}

func handleDraftWrite(b Backend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		draftID, _ := args["draft_id"].(string)
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		actor, _ := args["actor"].(string)
		if draftID == "" {
			return mcp.NewToolResultError("draft_id is required"), nil
		}
		if path == "" {
			return mcp.NewToolResultError("path is required"), nil
		}
		if content == "" {
			return mcp.NewToolResultError("content is required"), nil
		}
		if actor == "" {
			actor = "mcp-agent"
		}
		etag, err := b.DraftWrite(ctx, draftID, path, content, actor)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write to draft: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Written %s to draft %s (ETag: %s)", path, draftID, etag)), nil
	}
}

func handleDraftRead(b Backend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		draftID, _ := args["draft_id"].(string)
		path, _ := args["path"].(string)
		if draftID == "" {
			return mcp.NewToolResultError("draft_id is required"), nil
		}
		if path == "" {
			return mcp.NewToolResultError("path is required"), nil
		}
		content, etag, err := b.DraftRead(ctx, draftID, path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read from draft: %v", err)), nil
		}
		result := content
		if etag != "" {
			result = fmt.Sprintf("[ETag: %s]\n\n%s", etag, content)
		}
		return mcp.NewToolResultText(result), nil
	}
}

func handleDraftDiff(b Backend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		draftID, _ := args["draft_id"].(string)
		if draftID == "" {
			return mcp.NewToolResultError("draft_id is required"), nil
		}
		diff, err := b.DraftDiff(ctx, draftID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get diff: %v", err)), nil
		}
		if diff == "" {
			return mcp.NewToolResultText("No changes in draft."), nil
		}
		return mcp.NewToolResultText(diff), nil
	}
}

func handleDraftMerge(b Backend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		draftID, _ := args["draft_id"].(string)
		if draftID == "" {
			return mcp.NewToolResultError("draft_id is required"), nil
		}
		if err := b.DraftMerge(ctx, draftID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Merge failed: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Draft %s merged into main.", draftID)), nil
	}
}

func handleDraftDiscard(b Backend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		draftID, _ := args["draft_id"].(string)
		if draftID == "" {
			return mcp.NewToolResultError("draft_id is required"), nil
		}
		if err := b.DraftDiscard(ctx, draftID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Discard failed: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Draft %s discarded.", draftID)), nil
	}
}
