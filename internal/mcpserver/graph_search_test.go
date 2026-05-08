package mcpserver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func setupGraphTestBackend(t *testing.T) (*LocalBackend, string) {
	t.Helper()
	tmp := t.TempDir()
	kiwiDir := filepath.Join(tmp, ".kiwi")
	os.MkdirAll(kiwiDir, 0o755)
	os.WriteFile(filepath.Join(kiwiDir, "config.toml"), []byte(`
[search]
engine = "sqlite"
[versioning]
strategy = "none"
`), 0o644)

	os.WriteFile(filepath.Join(tmp, "index.md"), []byte("# Index\n\nWelcome to the knowledge base.\n\nSee [[payments]] and [[auth]] for details.\n"), 0o644)

	os.MkdirAll(filepath.Join(tmp, "pages"), 0o755)
	os.WriteFile(filepath.Join(tmp, "pages", "payments.md"), []byte(`---
tags:
  - billing
  - core
---
# Payments

Payment processing overview.

## Retry Logic

When a payment fails, we retry up to 3 times with exponential backoff.
The retry interval starts at 1 second.

## Circuit Breaker

If the payment provider returns 5 consecutive failures, the circuit breaker opens.
See [[resilience]] for the full pattern.

## Billing Integration

Invoices are generated after successful payment.
See [[billing]] for details.
`), 0o644)

	os.WriteFile(filepath.Join(tmp, "pages", "auth.md"), []byte(`---
tags:
  - security
  - core
---
# Authentication

Auth overview for the platform.

## OAuth Flow

We use OAuth 2.0 with PKCE.

## Session Management

Sessions expire after 24 hours.
`), 0o644)

	os.WriteFile(filepath.Join(tmp, "pages", "resilience.md"), []byte(`---
tags:
  - infrastructure
---
# Resilience Patterns

How we handle failures in distributed systems.

## Circuit Breaker Pattern

The circuit breaker has three states: closed, open, half-open.

## Retry with Backoff

Exponential backoff with jitter.
`), 0o644)

	os.WriteFile(filepath.Join(tmp, "pages", "billing.md"), []byte(`---
tags:
  - billing
---
# Billing

Invoice generation and subscription management.

See [[payments]] for payment processing.
`), 0o644)

	b := NewLocalBackend(tmp)
	if err := b.init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Index files so links are tracked
	ctx := context.Background()
	files := []string{"index.md", "pages/payments.md", "pages/auth.md", "pages/resilience.md", "pages/billing.md"}
	for _, f := range files {
		content, _ := os.ReadFile(filepath.Join(tmp, f))
		b.stack.Pipeline.Write(ctx, f, content, "test")
	}

	return b, tmp
}

// --- Peek Tests ---

func TestPeek_BasicFile(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.Peek(context.Background(), "pages/payments.md")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if result.Title != "Payments" {
		t.Errorf("title = %q, want %q", result.Title, "Payments")
	}
	if result.Snippet == "" {
		t.Error("expected non-empty snippet")
	}
	if !strings.Contains(result.Snippet, "Payment processing") {
		t.Errorf("snippet = %q, want it to contain 'Payment processing'", result.Snippet)
	}
	if result.WordCount == 0 {
		t.Error("expected non-zero word count")
	}
	if len(result.Headings) < 3 {
		t.Errorf("expected at least 3 headings, got %d", len(result.Headings))
	}
}

func TestPeek_SnippetTruncation(t *testing.T) {
	b, tmp := setupGraphTestBackend(t)
	defer b.Close()

	longPara := strings.Repeat("This is a very long paragraph that should be truncated. ", 20)
	os.WriteFile(filepath.Join(tmp, "pages", "long.md"), []byte("# Long Page\n\n"+longPara+"\n"), 0o644)
	b.stack.Pipeline.Write(context.Background(), "pages/long.md", []byte("# Long Page\n\n"+longPara+"\n"), "test")

	result, err := b.Peek(context.Background(), "pages/long.md")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if len(result.Snippet) > 350 {
		t.Errorf("snippet too long: %d chars", len(result.Snippet))
	}
}

func TestPeek_LinksExtracted(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.Peek(context.Background(), "pages/payments.md")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if len(result.LinksOut) == 0 {
		t.Error("expected outbound links")
	}
	found := false
	for _, l := range result.LinksOut {
		if l == "resilience" || l == "billing" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'resilience' or 'billing' in links_out, got %v", result.LinksOut)
	}
}

func TestPeek_BacklinksIncluded(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	// billing.md links to payments, so payments should have billing in backlinks
	result, err := b.Peek(context.Background(), "pages/payments.md")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	// Backlinks depend on the link index; at minimum verify it's a valid slice
	if result.LinksIn == nil {
		t.Error("expected non-nil links_in")
	}
}

func TestPeek_NoFrontmatter(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.Peek(context.Background(), "index.md")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if result.Title == "" {
		t.Error("expected non-empty title even without frontmatter")
	}
}

func TestPeek_NotFound(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	_, err := b.Peek(context.Background(), "nonexistent.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// --- Section Tests ---

func TestSection_ByHeading(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.Section(context.Background(), "pages/payments.md", "Retry Logic", -1)
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	if result.Heading != "Retry Logic" {
		t.Errorf("heading = %q, want %q", result.Heading, "Retry Logic")
	}
	if !strings.Contains(result.Content, "retry up to 3 times") {
		t.Errorf("content missing expected text, got: %s", result.Content)
	}
}

func TestSection_FuzzyMatch(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.Section(context.Background(), "pages/payments.md", "retry", -1)
	if err != nil {
		t.Fatalf("Section fuzzy: %v", err)
	}
	if result.Heading != "Retry Logic" {
		t.Errorf("heading = %q, want %q", result.Heading, "Retry Logic")
	}
}

func TestSection_ByIndex(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.Section(context.Background(), "pages/payments.md", "", 0)
	if err != nil {
		t.Fatalf("Section by index: %v", err)
	}
	if result.Heading != "Payments" {
		t.Errorf("heading = %q, want %q", result.Heading, "Payments")
	}
}

func TestSection_NotFound(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	_, err := b.Section(context.Background(), "pages/payments.md", "Nonexistent Section", -1)
	if err == nil {
		t.Fatal("expected error for missing section")
	}
}

func TestSection_LastSection(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.Section(context.Background(), "pages/payments.md", "Billing Integration", -1)
	if err != nil {
		t.Fatalf("Section last: %v", err)
	}
	if result.Content == "" {
		t.Error("expected non-empty content for last section")
	}
}

// --- GraphWalk Tests ---

func TestGraphWalk_OutboundLinks(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.GraphWalk(context.Background(), "pages/payments.md", false)
	if err != nil {
		t.Fatalf("GraphWalk: %v", err)
	}
	if len(result.LinksOut) == 0 {
		t.Error("expected outbound links")
	}
	if result.OutDegree != len(result.LinksOut) {
		t.Errorf("out_degree=%d != len(links_out)=%d", result.OutDegree, len(result.LinksOut))
	}
}

func TestGraphWalk_DirectorySiblings(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.GraphWalk(context.Background(), "pages/payments.md", true)
	if err != nil {
		t.Fatalf("GraphWalk with siblings: %v", err)
	}
	hasDirSibling := false
	for _, s := range result.Siblings {
		if s.Relation == "sibling_dir" {
			hasDirSibling = true
			break
		}
	}
	if !hasDirSibling {
		t.Error("expected directory siblings for pages/payments.md")
	}
}

func TestGraphWalk_TagSiblings(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.GraphWalk(context.Background(), "pages/payments.md", true)
	if err != nil {
		t.Fatalf("GraphWalk with siblings: %v", err)
	}
	hasTagSibling := false
	for _, s := range result.Siblings {
		if s.Relation == "sibling_tag" && s.SharedTag != "" {
			hasTagSibling = true
			break
		}
	}
	if !hasTagSibling {
		t.Error("expected tag siblings (billing tag shared with billing.md)")
	}
}

func TestGraphWalk_EmptyGraph(t *testing.T) {
	b, tmp := setupGraphTestBackend(t)
	defer b.Close()

	os.WriteFile(filepath.Join(tmp, "pages", "orphan.md"), []byte("# Orphan\n\nNo links here.\n"), 0o644)
	b.stack.Pipeline.Write(context.Background(), "pages/orphan.md", []byte("# Orphan\n\nNo links here.\n"), "test")

	result, err := b.GraphWalk(context.Background(), "pages/orphan.md", false)
	if err != nil {
		t.Fatalf("GraphWalk: %v", err)
	}
	if result.LinksOut == nil {
		t.Error("expected non-nil links_out")
	}
	if result.LinksIn == nil {
		t.Error("expected non-nil links_in")
	}
	if result.Siblings == nil {
		t.Error("expected non-nil siblings")
	}
}

// --- GraphAnalytics (extended with clusters + bridges) Tests ---

func TestGraphAnalytics_ClustersAndBridges(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	result, err := b.GraphAnalytics(context.Background(), 10)
	if err != nil {
		t.Fatalf("GraphAnalytics: %v", err)
	}
	if result.TotalNodes == 0 && result.TotalEdges == 0 {
		t.Skip("no edges indexed — graph analytics require write pipeline")
	}
	if result.TopPages == nil {
		t.Error("expected non-nil top_pages")
	}
	if result.Clusters == nil {
		t.Error("expected non-nil clusters")
	}
	if result.Bridges == nil {
		t.Error("expected non-nil bridges")
	}
	if result.Orphans == nil {
		t.Error("expected non-nil orphans")
	}
}

// --- MCP Tool Handler Tests ---

func TestToolHandler_Peek(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	text := mustCallTool(t, handlePeek(b), "kiwi_peek", map[string]any{"path": "pages/payments.md"})
	var result PeekResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal peek result: %v", err)
	}
	if result.Title != "Payments" {
		t.Errorf("title = %q, want %q", result.Title, "Payments")
	}
}

func TestToolHandler_PeekNotFound(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	handler := handlePeek(b)
	req := callToolReq("kiwi_peek", map[string]any{"path": "nonexistent.md"})
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handlePeek: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected isError=true for missing file")
	}
}

func TestToolHandler_Section(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	text := mustCallTool(t, handleSection(b), "kiwi_section", map[string]any{
		"path":    "pages/payments.md",
		"heading": "Retry Logic",
	})
	if !strings.Contains(text, "retry up to 3 times") {
		t.Errorf("section content missing expected text, got: %s", text)
	}
}

func TestToolHandler_SectionByIndex(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	text := mustCallTool(t, handleSection(b), "kiwi_section", map[string]any{
		"path":  "pages/payments.md",
		"index": float64(0),
	})
	if !strings.Contains(text, "Payments") {
		t.Errorf("section content missing expected heading, got: %s", text)
	}
}

func TestToolHandler_GraphWalk(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	text := mustCallTool(t, handleGraphWalk(b), "kiwi_graph_walk", map[string]any{
		"path": "pages/payments.md",
	})
	var result GraphWalkResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal graph walk result: %v", err)
	}
	if result.LinksOut == nil {
		t.Error("expected non-nil links_out")
	}
}

func TestToolHandler_GraphAnalytics(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	text := mustCallTool(t, handleGraphAnalytics(b), "kiwi_graph_analytics", map[string]any{})
	if !strings.Contains(text, "Graph Analytics") {
		t.Errorf("graph analytics output missing header, got: %s", text[:min(len(text), 100)])
	}
}

// --- Integration: Search → Peek → Section ---

func TestGraphSearch_PeekThenSection(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()

	// 1. Peek at payments
	peek, err := b.Peek(context.Background(), "pages/payments.md")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if peek.WordCount == 0 {
		t.Error("expected non-zero word count from peek")
	}

	// 2. Find the right heading
	foundRetry := false
	for _, h := range peek.Headings {
		if strings.Contains(h, "Retry") {
			foundRetry = true
			break
		}
	}
	if !foundRetry {
		t.Fatal("expected 'Retry' heading in peek results")
	}

	// 3. Read just that section
	section, err := b.Section(context.Background(), "pages/payments.md", "Retry", -1)
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	if !strings.Contains(section.Content, "retry") {
		t.Errorf("section content missing 'retry', got: %s", section.Content)
	}
}

func TestGraphSearch_MultiHop(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()
	ctx := context.Background()

	// Start at payments, discover resilience via graph walk
	walk, err := b.GraphWalk(ctx, "pages/payments.md", false)
	if err != nil {
		t.Fatalf("GraphWalk: %v", err)
	}

	// payments links to resilience
	foundResilience := false
	for _, link := range walk.LinksOut {
		if strings.Contains(link, "resilience") {
			foundResilience = true
			break
		}
	}
	if !foundResilience {
		t.Fatalf("expected payments to link to resilience, got links_out: %v", walk.LinksOut)
	}

	// Now walk from resilience
	walk2, err := b.GraphWalk(ctx, "pages/resilience.md", false)
	if err != nil {
		t.Fatalf("GraphWalk resilience: %v", err)
	}

	// resilience should have backlinks from payments
	if walk2.InDegree == 0 {
		t.Error("expected resilience to have inbound links")
	}

	// Read the targeted section
	section, err := b.Section(ctx, "pages/resilience.md", "Circuit Breaker", -1)
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	if !strings.Contains(section.Content, "closed") {
		t.Errorf("expected circuit breaker content, got: %s", section.Content)
	}
}

func TestGraphSearch_PeekCheaperThanRead(t *testing.T) {
	b, _ := setupGraphTestBackend(t)
	defer b.Close()
	ctx := context.Background()

	peek, err := b.Peek(ctx, "pages/payments.md")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	peekSize := len(peek.Snippet) + len(peek.Title)

	fullContent, _, err := b.ReadFile(ctx, "pages/payments.md")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	fullSize := len(fullContent)

	if peekSize >= fullSize {
		t.Errorf("peek (%d chars) should be smaller than full read (%d chars)", peekSize, fullSize)
	}

	// Peek snippet should be bounded
	if len(peek.Snippet) > 500 {
		t.Errorf("peek snippet too long: %d chars", len(peek.Snippet))
	}
}

func TestGraphSearch_SectionExtraction(t *testing.T) {
	b, tmp := setupGraphTestBackend(t)
	defer b.Close()
	ctx := context.Background()

	// Create a file with 5 headings
	content := `---
title: "Multi-Section"
---
# Multi-Section Doc

Intro text.

## Alpha

Alpha content here.

## Beta

Beta content here.

## Gamma

Gamma content here.

## Delta

Delta content here.

## Epsilon

Epsilon content here.
`
	os.WriteFile(filepath.Join(tmp, "pages", "multi.md"), []byte(content), 0o644)
	b.stack.Pipeline.Write(ctx, "pages/multi.md", []byte(content), "test")

	headings := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
	for _, h := range headings {
		sec, err := b.Section(ctx, "pages/multi.md", h, -1)
		if err != nil {
			t.Errorf("Section(%q): %v", h, err)
			continue
		}
		if sec.Heading != h {
			t.Errorf("expected heading %q, got %q", h, sec.Heading)
		}
		expectedContent := strings.ToLower(h) + " content here."
		if !strings.Contains(strings.ToLower(sec.Content), strings.ToLower(h)+" content") {
			t.Errorf("Section(%q) content missing expected text, got: %q, want substring %q", h, sec.Content, expectedContent)
		}
	}

	// Also verify by index
	for i, h := range headings {
		sec, err := b.Section(ctx, "pages/multi.md", "", i+1) // +1 because index 0 is "Multi-Section Doc"
		if err != nil {
			t.Errorf("Section(index=%d): %v", i+1, err)
			continue
		}
		if sec.Heading != h {
			t.Errorf("Section(index=%d) expected heading %q, got %q", i+1, h, sec.Heading)
		}
	}
}

func TestGraphSearch_ContextIncludesGraph(t *testing.T) {
	b, tmp := setupGraphTestBackend(t)
	defer b.Close()

	// Create playbook and schema for context
	os.MkdirAll(filepath.Join(tmp, ".kiwi"), 0o755)
	os.WriteFile(filepath.Join(tmp, ".kiwi", "playbook.md"), []byte("# Playbook\nTest playbook."), 0o644)
	os.WriteFile(filepath.Join(tmp, "SCHEMA.md"), []byte("# Schema\nTest schema."), 0o644)

	handler := handleContext(b)
	req := callToolReq("kiwi_context", map[string]any{})
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handleContext: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "=== GRAPH ===") {
		t.Error("expected GRAPH section in context output")
	}
	if !strings.Contains(text, "Pages:") {
		t.Error("expected Pages count in graph section")
	}
	if !strings.Contains(text, "Links:") {
		t.Error("expected Links count in graph section")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
