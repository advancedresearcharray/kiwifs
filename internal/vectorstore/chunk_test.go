package vectorstore

import (
	"strings"
	"testing"
)

func TestChunkRespectsSize(t *testing.T) {
	// 300 chars; size=100 overlap=10 → several chunks, each ≤ size.
	input := strings.Repeat("a", 300)
	chunks := chunk(input, 100, 10)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if len([]rune(c)) > 100 {
			t.Fatalf("chunk %d exceeded size: %d", i, len([]rune(c)))
		}
	}
}

func TestChunkKeepsShortText(t *testing.T) {
	chunks := chunk("short", 100, 10)
	if len(chunks) != 1 || chunks[0] != "short" {
		t.Fatalf("unexpected chunks: %v", chunks)
	}
}

func TestChunkEmpty(t *testing.T) {
	if got := chunk("", 100, 10); got != nil {
		t.Fatalf("want nil, got %v", got)
	}
}

func TestChunkMarkdownHeadings(t *testing.T) {
	md := `# Introduction

This is the intro section with some content.

## Authentication

This section covers authentication flows.

### OAuth Flow

OAuth uses tokens for authorization.

## Database

This section covers database design.`

	chunks := chunkMarkdown(md, 1500, 50)
	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks for 3 heading sections, got %d", len(chunks))
	}

	// Each chunk should contain its heading context
	foundAuth := false
	foundOAuth := false
	foundDB := false
	for _, c := range chunks {
		if strings.Contains(c, "Authentication") && strings.Contains(c, "authentication flows") {
			foundAuth = true
		}
		if strings.Contains(c, "OAuth") && strings.Contains(c, "tokens for authorization") {
			foundOAuth = true
		}
		if strings.Contains(c, "Database") && strings.Contains(c, "database design") {
			foundDB = true
		}
	}
	if !foundAuth {
		t.Error("missing Authentication section in chunks")
	}
	if !foundOAuth {
		t.Error("missing OAuth Flow section in chunks")
	}
	if !foundDB {
		t.Error("missing Database section in chunks")
	}
}

func TestChunkMarkdownLongSectionFallback(t *testing.T) {
	// Single heading with body that exceeds maxSize → falls back to char splitting
	body := strings.Repeat("word ", 400) // ~2000 chars
	md := "## Big Section\n\n" + body

	chunks := chunkMarkdown(md, 500, 50)
	if len(chunks) < 2 {
		t.Fatalf("expected long section to be split, got %d chunks", len(chunks))
	}
}

func TestChunkMarkdownNonMarkdown(t *testing.T) {
	// Content with no headings falls back to paragraph chunking
	plain := "Just some text.\n\nAnother paragraph.\n\nThird paragraph."
	chunks := chunkMarkdown(plain, 1500, 10)
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk for plain content")
	}
}

func TestChunkMarkdownMergesSmallSections(t *testing.T) {
	md := `## A

Hi.

## B

Hello.

## C

This is a longer section with more content to push it over the minimum size threshold for chunks.`

	chunks := chunkMarkdown(md, 1500, 200)
	// Sections A and B are small (<200), so they should be merged
	if len(chunks) > 2 {
		t.Fatalf("expected small sections to be merged, got %d chunks", len(chunks))
	}
}

func TestChunkMarkdownHeadingHierarchy(t *testing.T) {
	md := `# Top

## Sub

Content under sub heading.`

	chunks := chunkMarkdown(md, 1500, 50)
	found := false
	for _, c := range chunks {
		if strings.Contains(c, "Top > Sub") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected heading hierarchy prefix 'Top > Sub', got chunks: %v", chunks)
	}
}
