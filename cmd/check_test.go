package cmd

import (
	"context"
	"testing"

	"github.com/kiwifs/kiwifs/internal/workspace"
)

func TestRunbookInitCheckPasses(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := workspace.Init(root, "runbook"); err != nil {
		t.Fatal(err)
	}

	checkCmd.SetContext(context.Background())
	args := []string{"--root", root}
	checkCmd.SetArgs(args)
	if err := checkCmd.ParseFlags(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	if code := runCheckWithCode(checkCmd, args); code != 0 {
		t.Fatalf("expected exit 0 for runbook init scaffold, got %d", code)
	}

	result, _, _, _, err := runKnowledgeScan(checkCmd)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	for _, is := range result.Issues {
		if is.Severity == "error" {
			t.Fatalf("unexpected error issue on runbook scaffold: %+v", is)
		}
	}
}
