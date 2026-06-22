package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwifs/kiwifs/internal/janitor"
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

func TestRunCheckWithCode_SequenceGapFails(t *testing.T) {
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
		[]byte(`{"counters":{"events":3}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	checkCmd.SetContext(context.Background())
	args := []string{"--root", dir}
	checkCmd.SetArgs(args)
	if err := checkCmd.ParseFlags(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	if code := runCheckWithCode(checkCmd, args); code != 1 {
		t.Fatalf("expected exit 1 for sequence gap, got %d", code)
	}
}

func TestRunKnowledgeScan_ExecutionStalenessFromConfig(t *testing.T) {
	root := t.TempDir()
	cfgDir := filepath.Join(root, ".kiwi")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `
[janitor.execution_staleness]
directory = "runbooks/"
date_field = "last_executed"
max_age_days = 30

[janitor.execution_staleness.flag_values]
last_outcome = "failure"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "runbooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
title: Failed runbook
owner: alice
status: active
last_executed: 2099-01-01
last_outcome: failure
reviewed: 2030-01-01
next-review: 2040-01-01
---

This runbook has enough content to avoid the empty-page threshold in janitor scans.
`
	if err := os.WriteFile(filepath.Join(root, "runbooks", "failed.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	checkCmd.SetContext(context.Background())
	args := []string{"--root", root}
	checkCmd.SetArgs(args)
	if err := checkCmd.ParseFlags(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	result, _, _, _, err := runKnowledgeScan(checkCmd)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	found := false
	for _, is := range result.Issues {
		if is.Kind == janitor.IssueExecutionStale && is.Path == "runbooks/failed.md" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected execution-stale for runbooks/failed.md, got %+v", result.Issues)
	}
}

func TestRunKnowledgeScan_ExecutionStalenessStaleDateFromConfig(t *testing.T) {
	root := t.TempDir()
	cfgDir := filepath.Join(root, ".kiwi")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `
[janitor.execution_staleness]
directory = "runbooks/"
date_field = "last_verified"
max_age_days = 30
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "runbooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
title: Stale verified runbook
owner: alice
status: active
last_verified: 2020-01-01
last_outcome: success
reviewed: 2030-01-01
next-review: 2040-01-01
---

This runbook has enough content to avoid the empty-page threshold in janitor scans.
`
	if err := os.WriteFile(filepath.Join(root, "runbooks", "stale.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	checkCmd.SetContext(context.Background())
	args := []string{"--root", root}
	checkCmd.SetArgs(args)
	if err := checkCmd.ParseFlags(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	result, _, _, _, err := runKnowledgeScan(checkCmd)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	found := false
	for _, is := range result.Issues {
		if is.Kind == janitor.IssueExecutionStale && is.Path == "runbooks/stale.md" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected execution-stale for runbooks/stale.md, got %+v", result.Issues)
	}
}
