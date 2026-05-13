package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_Valid(t *testing.T) {
	w := Workflow{
		Name: "review",
		States: []State{
			{Name: "draft"},
			{Name: "review"},
			{Name: "published", Terminal: true},
		},
		Transitions: []Transition{
			{From: "draft", To: "review"},
			{From: "review", To: "published"},
			{From: "review", To: "draft"},
		},
	}
	if err := Validate(w); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidate_EmptyName(t *testing.T) {
	w := Workflow{
		States: []State{{Name: "a"}},
	}
	if err := Validate(w); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestValidate_NoStates(t *testing.T) {
	w := Workflow{Name: "x"}
	if err := Validate(w); err == nil {
		t.Fatal("expected error for no states")
	}
}

func TestValidate_DuplicateState(t *testing.T) {
	w := Workflow{
		Name:   "dup",
		States: []State{{Name: "a"}, {Name: "a"}},
	}
	if err := Validate(w); err == nil {
		t.Fatal("expected error for duplicate state")
	}
}

func TestValidate_TerminalWithOutbound(t *testing.T) {
	w := Workflow{
		Name: "bad",
		States: []State{
			{Name: "open"},
			{Name: "closed", Terminal: true},
		},
		Transitions: []Transition{
			{From: "open", To: "closed"},
			{From: "closed", To: "open"}, // not allowed
		},
	}
	if err := Validate(w); err == nil {
		t.Fatal("expected error for terminal with outbound")
	}
}

func TestValidate_UnknownTransitionFrom(t *testing.T) {
	w := Workflow{
		Name:   "bad",
		States: []State{{Name: "a"}},
		Transitions: []Transition{
			{From: "nonexistent", To: "a"},
		},
	}
	if err := Validate(w); err == nil {
		t.Fatal("expected error for unknown from state")
	}
}

func TestValidate_UnknownTransitionTo(t *testing.T) {
	w := Workflow{
		Name:   "bad",
		States: []State{{Name: "a"}},
		Transitions: []Transition{
			{From: "a", To: "nonexistent"},
		},
	}
	if err := Validate(w); err == nil {
		t.Fatal("expected error for unknown to state")
	}
}

func TestValidateTransition(t *testing.T) {
	w := Workflow{
		Name: "test",
		States: []State{
			{Name: "open"},
			{Name: "closed"},
		},
		Transitions: []Transition{
			{From: "open", To: "closed"},
		},
	}

	if err := ValidateTransition(w, "open", "closed"); err != nil {
		t.Fatalf("expected valid transition, got: %v", err)
	}

	if err := ValidateTransition(w, "closed", "open"); err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

func TestSaveLoadGet(t *testing.T) {
	dir := t.TempDir()

	w := Workflow{
		Name: "review",
		States: []State{
			{Name: "draft"},
			{Name: "published", Terminal: true},
		},
		Transitions: []Transition{
			{From: "draft", To: "published"},
		},
	}

	if err := Save(dir, w); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify file exists
	path := filepath.Join(dir, ".kiwi", "workflows", "review.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	// Load all
	workflows, err := Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}
	if workflows[0].Name != "review" {
		t.Fatalf("expected name=review, got %s", workflows[0].Name)
	}

	// Get specific
	got, err := Get(dir, "review")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "review" {
		t.Fatalf("expected name=review, got %s", got.Name)
	}
	if len(got.States) != 2 {
		t.Fatalf("expected 2 states, got %d", len(got.States))
	}
}

func TestLoad_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	workflows, err := Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if workflows != nil {
		t.Fatalf("expected nil, got %v", workflows)
	}
}

func TestGet_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Get(dir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent workflow")
	}
}

func TestSave_InvalidWorkflow(t *testing.T) {
	dir := t.TempDir()
	w := Workflow{Name: ""} // no name
	if err := Save(dir, w); err == nil {
		t.Fatal("expected error for invalid workflow")
	}
}
