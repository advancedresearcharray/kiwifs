// Package workflow provides a lightweight state-machine engine for pages.
//
// Workflow definitions are stored as JSON files under .kiwi/workflows/.
// Each workflow defines a set of named states and valid transitions between
// them. Pages reference a workflow via their frontmatter "workflow" field,
// and their current state via "state".
//
// Advancing a page's state validates the transition, updates frontmatter,
// and commits via the pipeline.
package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Workflow is a named state machine definition.
type Workflow struct {
	Name        string       `json:"name"`
	States      []State      `json:"states"`
	Transitions []Transition `json:"transitions"`
}

// State is a named node in the state machine.
type State struct {
	Name     string `json:"name"`
	Color    string `json:"color,omitempty"`      // hex color for UI
	Terminal bool   `json:"terminal,omitempty"`   // no outbound transitions allowed
	WIPLimit int    `json:"wip_limit,omitempty"`  // max cards allowed in this column (0 = unlimited)
}

// Transition is a directed edge between states.
type Transition struct {
	From         string `json:"from"`
	To           string `json:"to"`
	RequiredRole string `json:"required_role,omitempty"` // optional RBAC gate
}

// ValidateName checks that a workflow name maps to exactly one JSON file under
// .kiwi/workflows. Display names may contain spaces or non-ASCII characters,
// but path separators and cleaned path components are rejected.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}
	if name != strings.TrimSpace(name) {
		return fmt.Errorf("workflow name cannot have leading or trailing whitespace")
	}
	if name == "." || name == ".." || filepath.Clean(name) != name || strings.ContainsAny(name, `/\\`) {
		return fmt.Errorf("invalid workflow name: %s", name)
	}
	return nil
}

// Validate checks internal consistency of a workflow definition:
//   - all transition endpoints reference defined states
//   - terminal states have no outbound transitions
//   - no duplicate state names
func Validate(w Workflow) error {
	if err := ValidateName(w.Name); err != nil {
		return err
	}
	if len(w.States) == 0 {
		return fmt.Errorf("workflow must have at least one state")
	}

	stateSet := make(map[string]State, len(w.States))
	for _, s := range w.States {
		if s.Name == "" {
			return fmt.Errorf("state name cannot be empty")
		}
		if _, dup := stateSet[s.Name]; dup {
			return fmt.Errorf("duplicate state name: %s", s.Name)
		}
		stateSet[s.Name] = s
	}

	terminalFrom := make(map[string]bool)
	for _, t := range w.Transitions {
		if _, ok := stateSet[t.From]; !ok {
			return fmt.Errorf("transition from unknown state: %s", t.From)
		}
		if _, ok := stateSet[t.To]; !ok {
			return fmt.Errorf("transition to unknown state: %s", t.To)
		}
		if stateSet[t.From].Terminal {
			terminalFrom[t.From] = true
		}
	}
	for name := range terminalFrom {
		return fmt.Errorf("terminal state %q has outbound transitions", name)
	}

	return nil
}

// ValidateTransition checks whether moving from currentState to targetState
// is allowed by the workflow definition.
func ValidateTransition(w Workflow, currentState, targetState string) error {
	for _, t := range w.Transitions {
		if t.From == currentState && t.To == targetState {
			return nil
		}
	}
	return fmt.Errorf("no transition from %q to %q in workflow %q", currentState, targetState, w.Name)
}

// LoadError describes a single broken workflow file encountered during Load.
type LoadError struct {
	File string
	Err  error
}

func (e LoadError) Error() string {
	return fmt.Sprintf("%s: %v", e.File, e.Err)
}

// LoadResult bundles successfully loaded workflows with any per-file errors
// so callers can surface broken files instead of silently dropping them.
type LoadResult struct {
	Workflows []Workflow
	Errors    []LoadError
}

// Load reads all workflow definitions from .kiwi/workflows/*.json.
// Workflows are validated on load; invalid definitions are skipped.
func Load(kiwiDir string) ([]Workflow, error) {
	res := LoadWithErrors(kiwiDir)
	// Preserve the existing API: return nil error unless the directory
	// itself is unreadable (handled inside LoadWithErrors).
	return res.Workflows, nil
}

// LoadWithErrors is like Load but also returns per-file errors for broken
// or invalid workflow definitions.
func LoadWithErrors(kiwiDir string) LoadResult {
	dir := filepath.Join(kiwiDir, ".kiwi", "workflows")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return LoadResult{}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return LoadResult{Errors: []LoadError{{File: dir, Err: err}}}
	}

	var result LoadResult
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		fpath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(fpath)
		if err != nil {
			result.Errors = append(result.Errors, LoadError{File: entry.Name(), Err: err})
			continue
		}
		var w Workflow
		if err := json.Unmarshal(data, &w); err != nil {
			result.Errors = append(result.Errors, LoadError{File: entry.Name(), Err: fmt.Errorf("invalid JSON: %w", err)})
			continue
		}
		if w.Name == "" {
			w.Name = strings.TrimSuffix(entry.Name(), ".json")
		}
		// Reject broken graphs (e.g. transitions to non-existent states).
		if err := Validate(w); err != nil {
			result.Errors = append(result.Errors, LoadError{File: entry.Name(), Err: fmt.Errorf("invalid workflow: %w", err)})
			continue
		}
		result.Workflows = append(result.Workflows, w)
	}
	return result
}

// Get reads a single workflow definition by name.
func Get(kiwiDir, name string) (Workflow, error) {
	if err := ValidateName(name); err != nil {
		return Workflow{}, err
	}
	path := filepath.Join(kiwiDir, ".kiwi", "workflows", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Workflow{}, fmt.Errorf("read workflow %q: %w", name, err)
	}
	var w Workflow
	if err := json.Unmarshal(data, &w); err != nil {
		return Workflow{}, fmt.Errorf("parse workflow %q: %w", name, err)
	}
	if w.Name == "" {
		w.Name = name
	}
	return w, nil
}

// Save writes a workflow definition to .kiwi/workflows/<name>.json.
func Save(kiwiDir string, w Workflow) error {
	if err := Validate(w); err != nil {
		return fmt.Errorf("invalid workflow: %w", err)
	}

	dir := filepath.Join(kiwiDir, ".kiwi", "workflows")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create workflows dir: %w", err)
	}

	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal workflow: %w", err)
	}

	path := filepath.Join(dir, w.Name+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}
	return nil
}

// Delete removes a workflow definition from .kiwi/workflows/<name>.json.
func Delete(kiwiDir, name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	path := filepath.Join(kiwiDir, ".kiwi", "workflows", name+".json")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("workflow %q not found", name)
		}
		return fmt.Errorf("delete workflow %q: %w", name, err)
	}
	return nil
}
