package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/search"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/kiwifs/kiwifs/internal/versioning"
)

func TestWriteRuleValidatorOverwriteRejectsPutAllowsAppend(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	ctx := context.Background()
	initial := []byte("---\nappend_only: true\n---\nline one\n")
	if err := store.Write(ctx, "log.md", initial); err != nil {
		t.Fatalf("seed: %v", err)
	}

	v := NewWriteRuleValidator(store, []config.ValidateWriteRuleConfig{{
		Name:    "append-only",
		Reject:  "overwrite",
		Message: "This file is append-only. Use POST /api/kiwi/file/append.",
		Match:   config.ValidateWriteMatchConfig{Frontmatter: "append_only", Value: "true"},
	}})

	err = v.Validate(ctx, "log.md", []byte("---\nappend_only: true\n---\nreplaced\n"), WriteKindPut)
	if !errors.Is(err, ErrWriteRejected) {
		t.Fatalf("expected ErrWriteRejected on overwrite, got %v", err)
	}
	if !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("expected custom message, got %q", err.Error())
	}

	err = v.Validate(ctx, "log.md", append(initial, []byte("\nline two\n")...), WriteKindAppend)
	if err != nil {
		t.Fatalf("append should be allowed: %v", err)
	}
}

func TestWriteRuleValidatorBodyChangeAllowsFrontmatterOnly(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	ctx := context.Background()
	initial := []byte("---\nstatus: accepted\n---\n# Decision\n\nBody text.\n")
	if err := store.Write(ctx, "adr.md", initial); err != nil {
		t.Fatalf("seed: %v", err)
	}

	v := NewWriteRuleValidator(store, []config.ValidateWriteRuleConfig{{
		Name:    "immutable-after-status",
		Reject:  "body_change",
		Message: "Accepted decisions cannot be edited.",
		Match:   config.ValidateWriteMatchConfig{Frontmatter: "status", Values: []string{"accepted", "deprecated", "superseded"}},
	}})

	updatedBody, err := markdown.SetFrontmatterField(initial, "reviewed_by", "alice")
	if err != nil {
		t.Fatalf("set frontmatter: %v", err)
	}
	if err := v.Validate(ctx, "adr.md", updatedBody, WriteKindPut); err != nil {
		t.Fatalf("frontmatter-only update should pass: %v", err)
	}

	bodyChanged := []byte("---\nstatus: accepted\n---\n# Decision\n\nEdited body.\n")
	err = v.Validate(ctx, "adr.md", bodyChanged, WriteKindPut)
	if !errors.Is(err, ErrWriteRejected) {
		t.Fatalf("expected body change rejection, got %v", err)
	}
}

func TestWriteRuleValidatorSkipsNewFileAndNonMatchingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	ctx := context.Background()
	v := NewWriteRuleValidator(store, []config.ValidateWriteRuleConfig{{
		Name:   "append-only",
		Reject: "overwrite",
		Match:  config.ValidateWriteMatchConfig{Frontmatter: "append_only", Value: "true"},
	}})

	if err := v.Validate(ctx, "new.md", []byte("---\nappend_only: true\n---\nbody\n"), WriteKindPut); err != nil {
		t.Fatalf("new file should skip rules: %v", err)
	}

	if err := store.Write(ctx, "open.md", []byte("---\nappend_only: false\n---\nbody\n")); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := v.Validate(ctx, "open.md", []byte("---\nappend_only: false\n---\nnew body\n"), WriteKindPut); err != nil {
		t.Fatalf("non-matching frontmatter should skip rules: %v", err)
	}
}

func TestWriteRuleValidatorBodyChangeRejectsAppend(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	ctx := context.Background()
	initial := []byte("---\nstatus: accepted\n---\nBody.\n")
	if err := store.Write(ctx, "adr.md", initial); err != nil {
		t.Fatalf("seed: %v", err)
	}

	v := NewWriteRuleValidator(store, []config.ValidateWriteRuleConfig{{
		Name:   "immutable-after-status",
		Reject: "body_change",
		Match:  config.ValidateWriteMatchConfig{Frontmatter: "status", Value: "accepted"},
	}})

	err = v.Validate(ctx, "adr.md", append(initial, []byte("more\n")...), WriteKindAppend)
	if !errors.Is(err, ErrWriteRejected) {
		t.Fatalf("append that changes body should be rejected, got %v", err)
	}
}

func TestPipelineValidateWriteRulesIntegration(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	ctx := context.Background()
	if err := store.Write(ctx, "log.md", []byte("---\nappend_only: true\n---\nentry\n")); err != nil {
		t.Fatalf("seed: %v", err)
	}

	p := New(store, versioning.NewNoop(), search.NewGrep(dir), nil, nil, nil, dir)
	wv := NewWriteRuleValidator(store, []config.ValidateWriteRuleConfig{{
		Name:    "append-only",
		Reject:  "overwrite",
		Message: "append-only conflict",
		Match:   config.ValidateWriteMatchConfig{Frontmatter: "append_only", Value: "true"},
	}})
	p.ValidateWrite = func(c context.Context, path string, content []byte, kind WriteKind) error {
		return wv.Validate(c, path, content, kind)
	}

	_, err = p.Write(ctx, "log.md", []byte("---\nappend_only: true\n---\nreplaced\n"), "tester")
	if !errors.Is(err, ErrAppendOnly) {
		t.Fatalf("Write should return ErrAppendOnly, got %v", err)
	}

	_, err = p.Append(ctx, "log.md", "entry two", "\n", "tester")
	if err != nil {
		t.Fatalf("Append should succeed: %v", err)
	}
}

func TestFrontmatterValueEquals(t *testing.T) {
	cases := []struct {
		actual   any
		expected string
		want     bool
	}{
		{true, "true", true},
		{false, "true", false},
		{"accepted", "accepted", true},
		{"draft", "accepted", false},
		{float64(3), "3", true},
	}
	for _, tc := range cases {
		if got := frontmatterValueEquals(tc.actual, tc.expected); got != tc.want {
			t.Fatalf("frontmatterValueEquals(%v, %q) = %v, want %v", tc.actual, tc.expected, got, tc.want)
		}
	}
}
