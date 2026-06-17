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

func TestValidateCiteIdentifiers(t *testing.T) {
	invalidDOIs := []string{
		"",
		"not-a-doi",
		"10.1234",
		"10.1234/",
		"10.1234/../evil",
		"10.1234/foo//bar",
		"https://evil.example/10.1234/foo",
		strings.Repeat("a", 300),
	}
	for _, raw := range invalidDOIs {
		if got := normalizeDOI(raw); got != "" {
			t.Fatalf("normalizeDOI(%q) = %q, want empty", raw, got)
		}
	}

	invalidArxiv := []string{
		"not-arxiv",
		"99.12345",
		"2301.12",
		"2301.12345/../../../etc",
	}
	for _, raw := range invalidArxiv {
		if got := normalizeArxivID(raw); got != "" {
			t.Fatalf("normalizeArxivID(%q) = %q, want empty", raw, got)
		}
	}

	if _, err := sanitizeCiteInput("10.1234/evil\ninjection"); err == nil {
		t.Fatal("expected sanitize error for newline injection")
	}
	if _, err := sanitizeCiteInput("10.1234/evil\\path"); err == nil {
		t.Fatal("expected sanitize error for backslash")
	}
}

func TestHandleCiteInvalidDOI(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusOK, http.StatusOK, sampleCrossrefJSON, sampleArxivXML)

	res, err := callCiteTool(t, b, client, map[string]any{"doi": "not-a-valid-doi"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for invalid DOI")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "invalid DOI") {
		t.Fatalf("unexpected error text: %s", text)
	}
}

func TestHandleCiteInvalidArxiv(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusOK, http.StatusOK, sampleCrossrefJSON, sampleArxivXML)

	res, err := callCiteTool(t, b, client, map[string]any{"arxiv_id": "bad-id"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for invalid arXiv ID")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "invalid arXiv") {
		t.Fatalf("unexpected error text: %s", text)
	}
}

func TestHandleCiteArxivNotFound(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	emptyFeed := `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"></feed>`
	client := setupMockCiteClient(t, http.StatusOK, http.StatusOK, sampleCrossrefJSON, emptyFeed)

	res, err := callCiteTool(t, b, client, map[string]any{"arxiv_id": "2301.12345"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for missing arXiv entry")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "not found") {
		t.Fatalf("unexpected error text: %s", text)
	}
}

func TestHandleCiteNetworkError(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	client := &citeHTTPClient{
		http:        &http.Client{Timeout: 1},
		userAgent:   "kiwifs-test",
		crossrefURL: "http://127.0.0.1:1/works/",
		arxivURL:    defaultArxivQueryURL,
	}

	res, err := callCiteTool(t, b, client, map[string]any{"doi": "10.1234/example"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for network failure")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "crossref request failed") {
		t.Fatalf("unexpected error text: %s", text)
	}
}

func TestHandleCiteCrossrefBadJSON(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusOK, http.StatusOK, `{not json`, sampleArxivXML)

	res, err := callCiteTool(t, b, client, map[string]any{"doi": "10.1234/example"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for malformed Crossref response")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "parse Crossref response") {
		t.Fatalf("unexpected error text: %s", text)
	}
}

func TestHandleCiteMaliciousIdentifier(t *testing.T) {
	b, _ := setupTestBackend(t)
	defer b.Close()
	client := setupMockCiteClient(t, http.StatusOK, http.StatusOK, sampleCrossrefJSON, sampleArxivXML)

	cases := []map[string]any{
		{"identifier": "10.1234/evil/../../../admin"},
		{"identifier": "10.1234/foo?bar=baz"},
		{"identifier": "10.1234/evil\nheader: injection"},
	}
	for _, args := range cases {
		res, err := callCiteTool(t, b, client, args)
		if err != nil {
			t.Fatalf("call %v: %v", args, err)
		}
		if !res.IsError {
			t.Fatalf("expected rejection for malicious input %v", args)
		}
	}
}

func TestValidateBibtexKey(t *testing.T) {
	if err := validateBibtexKey("vaswani2017attention"); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
	for _, key := range []string{"", "../evil", "bad/key", "UPPER"} {
		if err := validateBibtexKey(key); err == nil {
			t.Fatalf("expected rejection for key %q", key)
		}
	}
}

func TestAssertCiteRequestURLRejectsUnexpectedHost(t *testing.T) {
	client := newDefaultCiteHTTPClient()
	reqURL := "https://evil.example/works/10.1234/foo"
	if err := client.assertCrossrefURL(reqURL); err == nil {
		t.Fatal("expected host validation failure")
	}
}
