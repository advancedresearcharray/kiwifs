package clipper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClip(t *testing.T) {
	// Create test server serving sample HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Test Article Title</title>
	<meta name="author" content="Test Author">
</head>
<body>
	<article>
		<h1>Test Article Title</h1>
		<p class="byline">By Test Author</p>
		<p>This is the first paragraph of the article. It contains some interesting content about testing.</p>
		<p>This is the second paragraph with more details and information.</p>
		<h2>Section Heading</h2>
		<p>More content under a section.</p>
	</article>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	t.Run("basic clip", func(t *testing.T) {
		req := ClipRequest{
			URL:    server.URL,
			Tags:   []string{"test", "article"},
			Folder: "clips/",
		}

		result, content, err := Clip(context.Background(), req, server.Client())
		if err != nil {
			t.Fatalf("Clip failed: %v", err)
		}

		// Check result fields
		if result.Title == "" {
			t.Error("Expected title to be set")
		}
		if result.Path == "" {
			t.Error("Expected path to be set")
		}
		if !strings.HasPrefix(result.Path, "clips/") {
			t.Errorf("Expected path to start with 'clips/', got %s", result.Path)
		}
		if !strings.HasSuffix(result.Path, ".md") {
			t.Errorf("Expected path to end with '.md', got %s", result.Path)
		}

		// Check content structure
		if !strings.HasPrefix(content, "---\n") {
			t.Error("Expected content to start with YAML frontmatter")
		}
		if !strings.Contains(content, "title:") {
			t.Error("Expected frontmatter to contain title")
		}
		if !strings.Contains(content, "source_url:") {
			t.Error("Expected frontmatter to contain source_url")
		}
		if !strings.Contains(content, "clipped_at:") {
			t.Error("Expected frontmatter to contain clipped_at")
		}
		if !strings.Contains(content, "tags:") {
			t.Error("Expected frontmatter to contain tags")
		}
		if !strings.Contains(content, "- test") {
			t.Error("Expected tags to contain 'test'")
		}

		// Check markdown body exists after frontmatter
		parts := strings.SplitN(content, "---\n", 3)
		if len(parts) < 3 {
			t.Error("Expected frontmatter and body to be separated")
		}
	})

	t.Run("custom title override", func(t *testing.T) {
		req := ClipRequest{
			URL:    server.URL,
			Title:  "Custom Title Override",
			Folder: "test/",
		}

		result, content, err := Clip(context.Background(), req, server.Client())
		if err != nil {
			t.Fatalf("Clip failed: %v", err)
		}

		if result.Title != "Custom Title Override" {
			t.Errorf("Expected title 'Custom Title Override', got %s", result.Title)
		}
		if !strings.Contains(content, "title: Custom Title Override") {
			t.Error("Expected custom title in frontmatter")
		}
		if !strings.HasPrefix(result.Path, "test/") {
			t.Errorf("Expected path to start with 'test/', got %s", result.Path)
		}
	})

	t.Run("default folder", func(t *testing.T) {
		req := ClipRequest{
			URL: server.URL,
		}

		result, _, err := Clip(context.Background(), req, server.Client())
		if err != nil {
			t.Fatalf("Clip failed: %v", err)
		}

		if !strings.HasPrefix(result.Path, "clips/") {
			t.Errorf("Expected default folder 'clips/', got %s", result.Path)
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		req := ClipRequest{
			URL: "://invalid-url",
		}

		_, _, err := Clip(context.Background(), req, nil)
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("404 error", func(t *testing.T) {
		server404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server404.Close()

		req := ClipRequest{
			URL: server404.URL,
		}

		_, _, err := Clip(context.Background(), req, server404.Client())
		if err == nil {
			t.Error("Expected error for 404 response")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("Expected 404 in error message, got: %v", err)
		}
	})
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"Test Article Title", "test-article-title"},
		{"Special!@#$%Characters", "special-characters"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"CamelCaseTitle", "camelcasetitle"},
		{"Title-With-Dashes", "title-with-dashes"},
		{"   Leading and Trailing   ", "leading-and-trailing"},
		{"", "untitled"},
		{"123 Numbers", "123-numbers"},
		{"Übung mit Umlauts", "bung-mit-umlauts"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSlugifyLongTitle(t *testing.T) {
	longTitle := strings.Repeat("a", 150)
	slug := slugify(longTitle)
	if len(slug) > 100 {
		t.Errorf("Expected slug length <= 100, got %d", len(slug))
	}
}
