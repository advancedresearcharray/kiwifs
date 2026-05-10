package markdown

import (
	"testing"
)

func TestLintMarkdown(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string // rule IDs expected
	}{
		{
			"clean document",
			"---\ntitle: Hello\n---\n# Hello\n\nSome text.\n",
			nil,
		},
		{
			"frontmatter yaml invalid",
			"---\n: invalid\n---\n",
			[]string{"frontmatter-yaml-invalid"},
		},
		{
			"frontmatter unterminated",
			"---\ntitle: hi\n",
			[]string{"frontmatter-yaml-invalid"},
		},
		{
			"frontmatter missing title",
			"---\ntags: [a, b]\n---\n# Hello\n",
			[]string{"frontmatter-missing-required"},
		},
		{
			"frontmatter with title OK",
			"---\ntitle: Test\n---\n# Test\n",
			nil,
		},
		{
			"frontmatter invalid date",
			"---\ntitle: Test\ncreated: not-a-date\n---\n# Test\n",
			[]string{"frontmatter-date-invalid"},
		},
		{
			"frontmatter valid date",
			"---\ntitle: Test\ncreated: 2024-01-15\n---\n# Test\n",
			nil,
		},
		{
			"heading skip level",
			"# H1\n### H3\n",
			[]string{"frontmatter-missing-required", "heading-skip-level"},
		},
		{
			"heading duplicate slug",
			"---\ntitle: Test\n---\n# Setup\n\nText.\n\n# Setup\n",
			[]string{"heading-duplicate-slug"},
		},
		{
			"fence unclosed",
			"---\ntitle: Test\n---\n```go\nfunc main() {}\n",
			[]string{"fence-unclosed"},
		},
		{
			"mermaid invalid no keyword",
			"---\ntitle: Test\n---\n```mermaid\nthis is not mermaid\n```\n",
			[]string{"fence-mermaid-invalid"},
		},
		{
			"mermaid valid flowchart",
			"---\ntitle: Test\n---\n```mermaid\nflowchart LR\n  A --> B\n```\n",
			nil,
		},
		{
			"mermaid empty block",
			"---\ntitle: Test\n---\n```mermaid\n```\n",
			[]string{"fence-mermaid-invalid"},
		},
		{
			"image broken empty url",
			"---\ntitle: Test\n---\n# Page\n\n![alt text]()\n",
			[]string{"link-image-broken"},
		},
		{
			"image with url OK",
			"---\ntitle: Test\n---\n# Page\n\n![alt](https://example.com/img.png)\n",
			nil,
		},
		{
			"no frontmatter at all",
			"# Hello World\n\nSome text.\n",
			[]string{"frontmatter-missing-required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := LintMarkdown([]byte(tt.input))
			got := make(map[string]bool)
			for _, issue := range issues {
				got[issue.Rule] = true
			}

			if tt.expect == nil {
				if len(issues) != 0 {
					t.Errorf("expected no issues, got %d:", len(issues))
					for _, issue := range issues {
						t.Errorf("  %s (line %d): %s", issue.Rule, issue.Line, issue.Message)
					}
				}
				return
			}

			for _, rule := range tt.expect {
				if !got[rule] {
					t.Errorf("expected rule %q not found in issues:", rule)
					for _, issue := range issues {
						t.Errorf("  %s (line %d): %s", issue.Rule, issue.Line, issue.Message)
					}
				}
			}
		})
	}
}

func TestLintSeverity(t *testing.T) {
	// Error-severity rules that should block writes.
	errorRules := map[string]string{
		"frontmatter-yaml-invalid": "---\n: bad\n---\n",
		"fence-unclosed":           "---\ntitle: t\n---\n```go\ncode\n",
		"fence-mermaid-invalid":    "---\ntitle: t\n---\n```mermaid\nnot-mermaid\n```\n",
	}
	for rule, input := range errorRules {
		t.Run(rule, func(t *testing.T) {
			issues := LintMarkdown([]byte(input))
			found := false
			for _, issue := range issues {
				if issue.Rule == rule {
					if issue.Severity != "error" {
						t.Errorf("rule %s should be error severity, got %s", rule, issue.Severity)
					}
					found = true
				}
			}
			if !found {
				t.Errorf("rule %s not found in issues", rule)
			}
		})
	}

	// Warning-severity rules.
	warningRules := map[string]string{
		"frontmatter-missing-required": "---\ntags: [a]\n---\n# Page\n",
		"heading-skip-level":           "---\ntitle: t\n---\n# H1\n### H3\n",
		"frontmatter-date-invalid":     "---\ntitle: t\ncreated: bad\n---\n# Page\n",
		"link-image-broken":            "---\ntitle: t\n---\n# Page\n![]()\n",
	}
	for rule, input := range warningRules {
		t.Run(rule, func(t *testing.T) {
			issues := LintMarkdown([]byte(input))
			found := false
			for _, issue := range issues {
				if issue.Rule == rule {
					if issue.Severity != "warning" {
						t.Errorf("rule %s should be warning severity, got %s", rule, issue.Severity)
					}
					found = true
				}
			}
			if !found {
				t.Errorf("rule %s not found in issues", rule)
			}
		})
	}
}
