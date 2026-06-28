package rbac

import (
	"testing"
)

func TestPagePublished(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "published true",
			content: "---\ntitle: Hello\npublished: true\n---\nBody\n",
			want:    true,
		},
		{
			name:    "published false",
			content: "---\ntitle: Hello\npublished: false\n---\nBody\n",
			want:    false,
		},
		{
			name:    "no published field",
			content: "---\ntitle: Hello\n---\nBody\n",
			want:    false,
		},
		{
			name:    "no frontmatter",
			content: "Just content\n",
			want:    false,
		},
		{
			name:    "empty file",
			content: "",
			want:    false,
		},
		{
			name:    "malformed yaml",
			content: "---\n: bad: yaml: {{{\n---\nBody\n",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PagePublished([]byte(tt.content))
			if got != tt.want {
				t.Errorf("PagePublished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPagePublishedAt(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantNil bool
		wantStr string
	}{
		{
			name:    "valid timestamp",
			content: "---\npublished_at: 2026-05-16T12:00:00Z\n---\nBody\n",
			wantNil: false,
			wantStr: "2026-05-16 12:00:00 +0000 UTC",
		},
		{
			name:    "no published_at",
			content: "---\ntitle: Hello\n---\nBody\n",
			wantNil: true,
		},
		{
			name:    "no frontmatter",
			content: "Just body\n",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PagePublishedAt([]byte(tt.content))
			if tt.wantNil {
				if got != nil {
					t.Errorf("PagePublishedAt() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Fatal("PagePublishedAt() = nil, want non-nil")
				}
				if got.String() != tt.wantStr {
					t.Errorf("PagePublishedAt() = %v, want %v", got.String(), tt.wantStr)
				}
			}
		})
	}
}
