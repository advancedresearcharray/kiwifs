package api

import (
	"testing"
)

// TestParseGitLog tests the git log parsing logic.
func TestParseGitLog(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		actorFilter  string
		typeFilter   string
		wantCount    int
		wantFirstPath string
		wantFirstType string
	}{
		{
			name: "basic write events",
			output: `COMMIT:abc123|2024-01-15T10:30:00Z|alice|Add new page
A	pages/new.md
M	pages/existing.md`,
			actorFilter:   "",
			typeFilter:    "",
			wantCount:     2,
			wantFirstPath: "pages/new.md",
			wantFirstType: "write",
		},
		{
			name: "delete event",
			output: `COMMIT:def456|2024-01-15T11:00:00Z|bob|Remove old page
D	pages/old.md`,
			actorFilter:   "",
			typeFilter:    "",
			wantCount:     1,
			wantFirstPath: "pages/old.md",
			wantFirstType: "delete",
		},
		{
			name: "filter by type write",
			output: `COMMIT:abc123|2024-01-15T10:30:00Z|alice|Changes
A	pages/new.md
D	pages/old.md`,
			actorFilter:   "",
			typeFilter:    "write",
			wantCount:     1,
			wantFirstPath: "pages/new.md",
			wantFirstType: "write",
		},
		{
			name: "filter by type delete",
			output: `COMMIT:abc123|2024-01-15T10:30:00Z|alice|Changes
A	pages/new.md
D	pages/old.md`,
			actorFilter:   "",
			typeFilter:    "delete",
			wantCount:     1,
			wantFirstPath: "pages/old.md",
			wantFirstType: "delete",
		},
		{
			name: "filter by actor",
			output: `COMMIT:abc123|2024-01-15T10:30:00Z|alice|Alice changes
A	pages/alice.md
COMMIT:def456|2024-01-15T11:00:00Z|bob|Bob changes
A	pages/bob.md`,
			actorFilter:   "alice",
			typeFilter:    "",
			wantCount:     1,
			wantFirstPath: "pages/alice.md",
			wantFirstType: "write",
		},
		{
			name: "skip .kiwi files",
			output: `COMMIT:abc123|2024-01-15T10:30:00Z|alice|Config change
M	.kiwi/config.toml
M	pages/real.md`,
			actorFilter:   "",
			typeFilter:    "",
			wantCount:     1,
			wantFirstPath: "pages/real.md",
			wantFirstType: "write",
		},
		{
			name: "rename event",
			output: `COMMIT:abc123|2024-01-15T10:30:00Z|alice|Rename file
R100	pages/old.md	pages/new.md`,
			actorFilter:   "",
			typeFilter:    "",
			wantCount:     1,
			wantFirstPath: "pages/new.md",
			wantFirstType: "write",
		},
		{
			name:          "empty output",
			output:        "",
			actorFilter:   "",
			typeFilter:    "",
			wantCount:     0,
			wantFirstPath: "",
			wantFirstType: "",
		},
		{
			name: "multiple commits",
			output: `COMMIT:abc123|2024-01-15T10:30:00Z|alice|First commit
A	pages/first.md
COMMIT:def456|2024-01-15T11:00:00Z|bob|Second commit
M	pages/second.md`,
			actorFilter:   "",
			typeFilter:    "",
			wantCount:     2,
			wantFirstPath: "pages/first.md",
			wantFirstType: "write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := parseGitLog(tt.output, tt.actorFilter, tt.typeFilter, "")
			if err != nil {
				t.Fatalf("parseGitLog() error = %v", err)
			}

			if len(events) != tt.wantCount {
				t.Errorf("parseGitLog() got %d events, want %d", len(events), tt.wantCount)
			}

			if tt.wantCount > 0 {
				if events[0].Path != tt.wantFirstPath {
					t.Errorf("first event path = %v, want %v", events[0].Path, tt.wantFirstPath)
				}
				if events[0].Type != tt.wantFirstType {
					t.Errorf("first event type = %v, want %v", events[0].Type, tt.wantFirstType)
				}
			}
		})
	}
}

// TestParseGitLogTimestamp tests that timestamps are preserved correctly.
func TestParseGitLogTimestamp(t *testing.T) {
	output := `COMMIT:abc123|2024-01-15T10:30:00Z|alice|Add page
A	pages/test.md`

	events, err := parseGitLog(output, "", "", "")
	if err != nil {
		t.Fatalf("parseGitLog() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	// Check timestamp format is RFC3339
	want := "2024-01-15T10:30:00Z"
	if events[0].Timestamp != want {
		t.Errorf("timestamp = %v, want %v", events[0].Timestamp, want)
	}
}

// TestParseGitLogActor tests that actor names are preserved.
func TestParseGitLogActor(t *testing.T) {
	output := `COMMIT:abc123|2024-01-15T10:30:00Z|Alice Smith|Add page
A	pages/test.md`

	events, err := parseGitLog(output, "", "", "")
	if err != nil {
		t.Fatalf("parseGitLog() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	want := "Alice Smith"
	if events[0].Actor != want {
		t.Errorf("actor = %v, want %v", events[0].Actor, want)
	}
}

// --- Edge case tests ---

func TestParseGitLogPathPrefix(t *testing.T) {
	output := "COMMIT:abc|2024-01-01T00:00:00Z|alice|msg\nA\tpages/foo.md\nA\tdocs/bar.md"
	events, err := parseGitLog(output, "", "", "pages/")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event filtered by path_prefix, got %d", len(events))
	}
	if events[0].Path != "pages/foo.md" {
		t.Errorf("expected pages/foo.md, got %s", events[0].Path)
	}
}

func TestParseGitLogUnicodeActor(t *testing.T) {
	output := "COMMIT:abc|2024-01-01T00:00:00Z|日本語ユーザー|日本語メッセージ\nA\tpages/日本語.md"
	events, err := parseGitLog(output, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Actor != "日本語ユーザー" {
		t.Errorf("actor = %v", events[0].Actor)
	}
	if events[0].Path != "pages/日本語.md" {
		t.Errorf("path = %v", events[0].Path)
	}
}

func TestParseGitLogPipeInSubject(t *testing.T) {
	output := "COMMIT:abc|2024-01-01T00:00:00Z|alice|fix: handle pipe | in messages\nA\tpages/test.md"
	events, err := parseGitLog(output, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if events[0].Message != "fix: handle pipe | in messages" {
		t.Errorf("message = %v", events[0].Message)
	}
}

func TestParseGitLogCopyStatus(t *testing.T) {
	output := "COMMIT:abc|2024-01-01T00:00:00Z|alice|copy file\nC100\tpages/src.md\tpages/dst.md"
	events, err := parseGitLog(output, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Path != "pages/dst.md" {
		t.Errorf("copy should use destination: got %s", events[0].Path)
	}
}

func TestParseGitLogMalformedLines(t *testing.T) {
	output := "COMMIT:abc|2024-01-01T00:00:00Z|alice|msg\nGARBAGE LINE\n\n\tonly-tab\nA\tvalid.md"
	events, err := parseGitLog(output, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 valid event, got %d", len(events))
	}
}
