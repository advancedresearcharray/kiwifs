package markdown

import (
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Step 1: line endings
		{
			name:  "crlf normalization",
			input: "# Hello\r\n\r\nWorld\r\n",
			want:  "# Hello\n\nWorld\n",
		},
		{
			name:  "lone cr normalization",
			input: "# Hello\rWorld\r",
			want:  "# Hello\nWorld\n",
		},

		// Step 2: trailing newline
		{
			name:  "trailing newline added",
			input: "# Hello",
			want:  "# Hello\n",
		},
		{
			name:  "trailing newline preserved",
			input: "# Hello\n",
			want:  "# Hello\n",
		},
		{
			name:  "empty input",
			input: "",
			want:  "\n",
		},

		// Step 3: unclosed fences
		{
			name:  "unclosed backtick fence",
			input: "```go\nfunc main() {}\n",
			want:  "```go\nfunc main() {}\n```\n",
		},
		{
			name:  "unclosed tilde fence",
			input: "~~~python\nprint('hello')\n",
			want:  "~~~python\nprint('hello')\n~~~\n",
		},
		{
			name:  "closed fence left alone",
			input: "```go\nfunc main() {}\n```\n",
			want:  "```go\nfunc main() {}\n```\n",
		},
		{
			name:  "multiple fences one unclosed",
			input: "```go\nfoo()\n```\n\n```js\nbar()\n",
			want:  "```go\nfoo()\n```\n\n```js\nbar()\n```\n",
		},

		// Step 4: tables
		{
			name:  "table alignment simple",
			input: "| a | b |\n|---|---|\n| 1 | 2 |\n",
			want:  "| a   | b   |\n| --- | --- |\n| 1   | 2   |\n",
		},
		{
			name:  "table missing separator",
			input: "| a | b |\n| 1 | 2 |\n",
			want:  "| a   | b   |\n| --- | --- |\n| 1   | 2   |\n",
		},
		{
			name:  "table uneven columns padded",
			input: "| name | value |\n|---|---|\n| x | 1 |\n",
			want:  "| name | value |\n| ---- | ----- |\n| x    | 1     |\n",
		},
		{
			name:  "table separator too few columns",
			input: "| a | b |\n|---|\n| 1 | 2 |\n",
			want:  "| a   | b   |\n| --- | --- |\n| 1   | 2   |\n",
		},

		// Step 5: list markers — unordered
		{
			name:  "unordered list star to dash",
			input: "* item 1\n* item 2\n* item 3\n",
			want:  "- item 1\n- item 2\n- item 3\n",
		},
		{
			name:  "unordered list plus to dash",
			input: "+ item 1\n+ item 2\n",
			want:  "- item 1\n- item 2\n",
		},
		{
			name:  "unordered list dash preserved",
			input: "- item 1\n- item 2\n",
			want:  "- item 1\n- item 2\n",
		},

		// Step 6: trailing whitespace
		{
			name:  "trailing spaces stripped and normalized",
			input: "hello   \nworld  \n",
			want:  "hello  \nworld  \n",
		},
		{
			name:  "trailing tabs stripped",
			input: "hello\t\t\n",
			want:  "hello\n",
		},
		{
			name:  "hard break preserved",
			input: "line one  \nline two\n",
			want:  "line one  \nline two\n",
		},

		// Step 7: blank lines
		{
			name:  "collapse 3 blank lines to 2",
			input: "a\n\n\n\nb\n",
			want:  "a\n\n\nb\n",
		},
		{
			name:  "collapse 5 blank lines to 2",
			input: "a\n\n\n\n\n\nb\n",
			want:  "a\n\n\nb\n",
		},
		{
			name:  "2 blank lines preserved",
			input: "a\n\n\nb\n",
			want:  "a\n\n\nb\n",
		},

		// Step 8: frontmatter
		{
			name:  "preserves frontmatter verbatim",
			input: "---\ntitle: hi\n---\n# Hello",
			want:  "---\ntitle: hi\n---\n# Hello\n",
		},
		{
			name:  "frontmatter with trailing whitespace in body",
			input: "---\ntitle: test\n---\nhello   \n",
			want:  "---\ntitle: test\n---\nhello  \n",
		},

		// Combined
		{
			name:  "combined: crlf + fence + trailing newline",
			input: "```go\r\nfunc main() {}\r\n",
			want:  "```go\nfunc main() {}\n```\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(Format([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("Format() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello\r\nworld\r\n", "hello\nworld\n"},
		{"hello\rworld\r", "hello\nworld\n"},
		{"hello\nworld\n", "hello\nworld\n"},
		{"no newline", "no newline"},
	}
	for _, tt := range tests {
		got := normalizeLineEndings(tt.input)
		if got != tt.want {
			t.Errorf("normalizeLineEndings(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCloseUnclosedFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"no fences",
			"hello\nworld",
			"hello\nworld",
		},
		{
			"closed fence",
			"```\ncode\n```",
			"```\ncode\n```",
		},
		{
			"unclosed backtick",
			"```go\ncode",
			"```go\ncode\n```",
		},
		{
			"unclosed tilde",
			"~~~\ncode",
			"~~~\ncode\n~~~",
		},
		{
			"4-backtick fence",
			"````\ncode with ``` inside",
			"````\ncode with ``` inside\n````",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := closeUnclosedFences(tt.input)
			if got != tt.want {
				t.Errorf("closeUnclosedFences() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTables(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"simple table",
			"| a | b |\n|---|---|\n| 1 | 2 |",
			"| a   | b   |\n| --- | --- |\n| 1   | 2   |",
		},
		{
			"table with long content",
			"| name | description |\n|---|---|\n| x | short |",
			"| name | description |\n| ---- | ----------- |\n| x    | short       |",
		},
		{
			"non-table text untouched",
			"hello\nworld",
			"hello\nworld",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTables(tt.input)
			if got != tt.want {
				t.Errorf("formatTables() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestCollapseBlankLines(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a\n\n\n\nb", "a\n\n\nb"},
		{"a\n\nb", "a\n\nb"},
		{"a\nb", "a\nb"},
		{"a\n\n\n\n\n\nb", "a\n\n\nb"},
	}
	for _, tt := range tests {
		got := collapseBlankLines(tt.input)
		if got != tt.want {
			t.Errorf("collapseBlankLines(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStripTrailingWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello   \nworld", "hello  \nworld"},  // 3 spaces → normalized to 2 (hard break)
		{"line  \nend", "line  \nend"},         // hard break (exactly 2 spaces) preserved
		{"hello\t\t\nworld", "hello\nworld"},
		{"clean\nlines", "clean\nlines"},
	}
	for _, tt := range tests {
		got := stripTrailingWhitespace(tt.input)
		if got != tt.want {
			t.Errorf("stripTrailingWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
