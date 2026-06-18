package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/janitor"
	"github.com/kiwifs/kiwifs/internal/pipeline"
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

func TestCheckCommandSequenceGap(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".kiwi"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `[sequences]
directories = ["events/"]
`
	if err := os.WriteFile(filepath.Join(dir, ".kiwi", "config.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	eventsDir := filepath.Join(dir, "events")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eventsDir, "log.md"),
		[]byte("<!-- seq:1 -->\na\n<!-- seq:3 -->\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(dir, ".kiwi", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "sequences.json"),
		[]byte("{\n  \"counter\": 3\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := pipeline.CheckSequences(dir, []string{"events/"})
	if err != nil {
		t.Fatalf("CheckSequences: %v", err)
	}
	if !result.HasIssues() {
		t.Fatal("expected sequence gap issue")
	}
}

func TestCheckCommandNoSequencesConfigured(t *testing.T) {
	dir := t.TempDir()
	checkCmd.SetContext(context.Background())
	checkCmd.SetArgs([]string{"--root", dir})
	if err := checkCmd.ParseFlags([]string{"--root", dir}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	result, _, _, _, err := runKnowledgeScan(checkCmd)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if result.HasErrors() {
		t.Fatalf("expected clean scan, got %+v", result.Issues)
	}
}
