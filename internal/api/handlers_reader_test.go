package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishedPageContentNegotiation(t *testing.T) {
	s := buildTestServer(t)
	pageContent := "---\npublished: true\ntitle: API Guide\ndescription: Headless CMS demo\n---\n# API Guide\n\nSee [diagram](./assets/chart.png).\n"
	mustPutFile(t, s, "docs/api-guide.md", pageContent)

	t.Run("default HTML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/p/docs/api-guide.md", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", ct)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "<!doctype html>") {
			t.Fatalf("expected HTML document, got: %s", body[:min(120, len(body))])
		}
		if !strings.Contains(body, "API Guide") {
			t.Fatalf("expected page title in HTML")
		}
	})

	t.Run("text/markdown", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/p/docs/api-guide.md", nil)
		req.Header.Set("Accept", "text/markdown")
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/markdown; charset=utf-8" {
			t.Fatalf("Content-Type = %q, want text/markdown; charset=utf-8", ct)
		}
		if rec.Body.String() != pageContent {
			t.Fatalf("expected raw markdown source, got:\n%s", rec.Body.String())
		}
	})

	t.Run("application/json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/p/docs/api-guide.md", nil)
		req.Header.Set("Accept", "application/json")
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}

		var out publishedPageJSON
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if out.Frontmatter["title"] != "API Guide" {
			t.Fatalf("frontmatter.title = %v, want API Guide", out.Frontmatter["title"])
		}
		if out.Frontmatter["description"] != "Headless CMS demo" {
			t.Fatalf("frontmatter.description = %v", out.Frontmatter["description"])
		}
		if !strings.Contains(out.HTML, "<h1>API Guide</h1>") {
			t.Fatalf("expected rendered HTML body, got: %s", out.HTML)
		}
		if !strings.Contains(out.HTML, `href="/p/docs/assets/chart.png"`) {
			t.Fatalf("expected rewritten asset links in HTML, got: %s", out.HTML)
		}
		wantMarkdown := "# API Guide\n\nSee [diagram](./assets/chart.png).\n"
		if out.Markdown != wantMarkdown {
			t.Fatalf("markdown = %q, want %q", out.Markdown, wantMarkdown)
		}
	})

	t.Run("prefers higher q value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/p/docs/api-guide.md", nil)
		req.Header.Set("Accept", "text/html;q=0.5, application/json;q=0.9")
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}
	})

	t.Run("unsupported Accept returns 406", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/p/docs/api-guide.md", nil)
		req.Header.Set("Accept", "image/png")
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotAcceptable {
			t.Fatalf("GET: %d %s, want 406", rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Accept"); got != readerSupportedFormats {
			t.Fatalf("Accept response header = %q, want %q", got, readerSupportedFormats)
		}
	})

	t.Run("invalid Accept with CRLF returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/p/docs/api-guide.md", nil)
		req.Header.Set("Accept", "text/html\r\nX-Injected: true")
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("GET: %d %s, want 400", rec.Code, rec.Body.String())
		}
	})

	t.Run("wildcard application type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/p/docs/api-guide.md", nil)
		req.Header.Set("Accept", "application/*;q=0.9, text/html;q=0.5")
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}
	})

	t.Run("large JSON payload", func(t *testing.T) {
		largeBody := strings.Repeat("Paragraph with content.\n\n", 500)
		largePage := "---\npublished: true\ntitle: Large Page\n---\n" + largeBody
		mustPutFile(t, s, "docs/large-page.md", largePage)

		req := httptest.NewRequest(http.MethodGet, "/p/docs/large-page.md", nil)
		req.Header.Set("Accept", "application/json")
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
		}
		var out publishedPageJSON
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if out.Frontmatter["title"] != "Large Page" {
			t.Fatalf("frontmatter.title = %v", out.Frontmatter["title"])
		}
		if len(out.Markdown) < len(largeBody)/2 {
			t.Fatalf("markdown payload too small: %d bytes", len(out.Markdown))
		}
		if !strings.Contains(out.HTML, "<p>Paragraph with content.</p>") {
			t.Fatalf("expected rendered HTML in large payload")
		}
	})
}
