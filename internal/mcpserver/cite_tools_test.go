package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

const sampleCrossrefJSON = `{
  "message": {
    "title": ["Attention Is All You Need"],
    "author": [
      {"given": "Ashish", "family": "Vaswani"},
      {"given": "Noam", "family": "Shazeer"}
    ],
    "DOI": "10.1234/example.attention",
    "abstract": "<p>We propose a transformer architecture.</p>",
    "container-title": ["NeurIPS"],
    "published-print": {"date-parts": [[2017, 6, 12]]}
  }
}`

const sampleArxivXML = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:arxiv="http://arxiv.org/schemas/atom">
  <entry>
    <title>Sample arXiv Paper</title>
    <author><name>Jane Doe</name></author>
    <published>2023-01-15T00:00:00Z</published>
    <summary>A sample abstract for testing.</summary>
    <arxiv:doi>10.5555/arxiv.sample</arxiv:doi>
    <id>http://arxiv.org/abs/2301.12345v1</id>
  </entry>
</feed>`

func setupMockCiteClient(t *testing.T, crossrefStatus, arxivStatus int, crossrefBody, arxivBody string) *citeHTTPClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/works/"):
			w.WriteHeader(crossrefStatus)
			w.Write([]byte(crossrefBody))
		case strings.Contains(r.URL.Path, "/api/query"):
			w.WriteHeader(arxivStatus)
			w.Write([]byte(arxivBody))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	return &citeHTTPClient{
		http:        srv.Client(),
		userAgent:   "kiwifs-test",
		crossrefURL: srv.URL + "/works/",
		arxivURL:    srv.URL + "/api/query",
	}
}

func callCiteTool(t *testing.T, b Backend, client *citeHTTPClient, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "kiwi_cite"
	req.Params.Arguments = args
	return handleCite(b, client)(context.Background(), req)
}

func TestBibtexKeyAndBibTeX(t *testing.T) {
	meta := &paperMetadata{
		Title:   "Attention Is All You Need",
		Authors: []string{"Vaswani, Ashish"},
		Year:    2017,
		Venue:   "NeurIPS",
		DOI:     "10.1234/example",
	}
	meta.BibtexKey = bibtexKey(meta)
	meta.BibTeX = buildBibTeX(meta)
	if meta.BibtexKey == "" {
		t.Fatal("expected bibtex key")
	}
	if !strings.Contains(meta.BibTeX, meta.BibtexKey) {
		t.Fatalf("bibtex missing key: %s", meta.BibTeX)
	}
}

func TestHandleCiteDOI(t *testing.T) {
	b, tmp := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusOK, http.StatusOK, sampleCrossrefJSON, sampleArxivXML)

	res, err := callCiteTool(t, b, client, map[string]any{
		"identifier": "10.1234/example.attention",
		"actor":      "test-agent",
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %v", res.Content)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &payload); err != nil {
		t.Fatalf("parse result: %v", err)
	}
	if payload["success"] != true {
		t.Fatalf("success = %v", payload["success"])
	}
	path, _ := payload["path"].(string)
	if !strings.HasPrefix(path, "papers/") || !strings.HasSuffix(path, ".md") {
		t.Fatalf("unexpected path: %s", path)
	}

	data, err := os.ReadFile(filepath.Join(tmp, path))
	if err != nil {
		t.Fatalf("read paper: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		"title: \"Attention Is All You Need\"",
		"doi: \"10.1234/example.attention\"",
		"venue: \"NeurIPS\"",
		"year: 2017",
		"abstract: |",
		"bibtex: |",
		"## Abstract",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("missing %q in:\n%s", want, content)
		}
	}
}

func TestHandleCiteArxiv(t *testing.T) {
	b, tmp := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusOK, http.StatusOK, sampleCrossrefJSON, sampleArxivXML)

	res, err := callCiteTool(t, b, client, map[string]any{
		"arxiv_id": "2301.12345",
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %v", res.Content)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &payload); err != nil {
		t.Fatalf("parse result: %v", err)
	}
	path, _ := payload["path"].(string)
	data, err := os.ReadFile(filepath.Join(tmp, path))
	if err != nil {
		t.Fatalf("read paper: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		"title: \"Sample arXiv Paper\"",
		"arxiv: \"2301.12345\"",
		"Jane Doe",
		"year: 2023",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("missing %q in:\n%s", want, content)
		}
	}
}

func TestHandleCiteDOINotFound(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusNotFound, http.StatusOK, `{}`, sampleArxivXML)

	res, err := callCiteTool(t, b, client, map[string]any{"doi": "10.1234/missing"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DOI not found") {
		t.Fatalf("unexpected error text: %s", text)
	}
}

func TestHandleCiteRateLimit(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusTooManyRequests, http.StatusOK, `{}`, sampleArxivXML)

	res, err := callCiteTool(t, b, client, map[string]any{"doi": "10.1234/rate-limited"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "rate limit") {
		t.Fatalf("unexpected error text: %s", text)
	}
}

func TestHandleCiteMissingIdentifier(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusOK, http.StatusOK, sampleCrossrefJSON, sampleArxivXML)

	res, err := callCiteTool(t, b, client, map[string]any{})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error")
	}
}

func TestNormalizeDOIAndArxiv(t *testing.T) {
	if got := normalizeDOI("doi:10.1234/example"); got != "10.1234/example" {
		t.Fatalf("normalizeDOI = %q", got)
	}
	if got := normalizeDOI("https://doi.org/10.1234/example"); got != "10.1234/example" {
		t.Fatalf("normalizeDOI url = %q", got)
	}
	if got := normalizeArxivID("arxiv:2301.12345v2"); got != "2301.12345" {
		t.Fatalf("normalizeArxivID = %q", got)
	}
	if got := normalizeArxivID("https://arxiv.org/abs/2301.12345"); got != "2301.12345" {
		t.Fatalf("normalizeArxivID url = %q", got)
	}
}
