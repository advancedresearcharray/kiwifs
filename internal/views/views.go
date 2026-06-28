package views

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _-]*$`)

// validateName ensures the view name is safe and cannot escape the views
// directory via path traversal.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("view name cannot be empty")
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("view name contains illegal characters")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("view name must start with alphanumeric and contain only alphanumeric, space, hyphen, or underscore")
	}
	return nil
}

type View struct {
	Name    string      `json:"name"`
	Query   string      `json:"query"`    // DQL query string
	Layout  string      `json:"layout"`   // table | list | calendar | kanban
	Columns []Column    `json:"columns,omitempty"`
	Filters []Filter    `json:"filters,omitempty"`
	Sort    []SortField `json:"sort,omitempty"`
	GroupBy string      `json:"group_by,omitempty"`
}

type Column struct {
	Property string `json:"property"`
	Label    string `json:"label,omitempty"`
	Formula  string `json:"formula,omitempty"` // computed column expression
	Summary  string `json:"summary,omitempty"` // sum|avg|count|min|max
}

type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // =, !=, >, <, contains, in
	Value    any    `json:"value"`
}

type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc | desc
}

// List returns all view definitions from .kiwi/views/*.json
func List(kiwiDir string) ([]View, error) {
	viewsDir := filepath.Join(kiwiDir, ".kiwi", "views")
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		return []View{}, nil
	}

	entries, err := os.ReadDir(viewsDir)
	if err != nil {
		return nil, fmt.Errorf("read views dir: %w", err)
	}

	var views []View
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		view, err := Get(kiwiDir, name)
		if err != nil {
			continue // skip invalid views
		}
		views = append(views, view)
	}
	return views, nil
}

// Get reads a single view definition
func Get(kiwiDir, name string) (View, error) {
	if err := validateName(name); err != nil {
		return View{}, err
	}
	path := filepath.Join(kiwiDir, ".kiwi", "views", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return View{}, fmt.Errorf("read view: %w", err)
	}

	var view View
	if err := json.Unmarshal(data, &view); err != nil {
		return View{}, fmt.Errorf("parse view: %w", err)
	}
	view.Name = name
	return view, nil
}

// Save writes a view definition to .kiwi/views/<name>.json
func Save(kiwiDir string, v View) error {
	viewsDir := filepath.Join(kiwiDir, ".kiwi", "views")
	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		return fmt.Errorf("create views dir: %w", err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal view: %w", err)
	}

	path := filepath.Join(viewsDir, v.Name+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write view: %w", err)
	}
	return nil
}

// Delete removes a view definition
func Delete(kiwiDir, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	path := filepath.Join(kiwiDir, ".kiwi", "views", name+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete view: %w", err)
	}
	return nil
}
