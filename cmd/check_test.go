package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/janitor"
	"github.com/kiwifs/kiwifs/internal/workspace"
)

func TestRunKnowledgeScan_DetectsBrokenLinks(t *testing.T) {
	root := t.TempDir()
	content := `---
title: Broken
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

This page links to [[missing-page]] and has enough text to avoid empty-page.
`
	if err := os.WriteFile(filepath.Join(root, "broken.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	args := []string{"--root", root}
	checkCmd.SetContext(context.Background())
	checkCmd.SetArgs(args)
	if err := checkCmd.ParseFlags(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	result, _, _, _, err := runKnowledgeScan(checkCmd)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if !result.HasErrors() {
		t.Fatalf("expected broken link error, got %+v", result.Issues)
	}
}

func TestScanResult_HasWarnings(t *testing.T) {
	r := &janitor.ScanResult{Issues: []janitor.Issue{{Severity: "warning"}}}
	if !r.HasWarnings() {
		t.Fatal("expected warnings")
	}
}

func TestRunbookInitCheckPasses(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "runbook-ws")
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
