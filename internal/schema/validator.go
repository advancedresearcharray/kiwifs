package schema

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Validator struct {
	mu      sync.RWMutex
	schemas map[string]*jsonschema.Schema
	root    string
}

type ValidationError struct {
	Type   string   `json:"type"`
	Errors []string `json:"errors"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for type %q: %s", e.Type, strings.Join(e.Errors, "; "))
}

func NewValidator(root string) *Validator {
	v := &Validator{
		schemas: make(map[string]*jsonschema.Schema),
		root:    root,
	}
	v.loadSchemas()
	return v
}

func (v *Validator) loadSchemas() {
	dir := filepath.Join(v.root, ".kiwi", "schemas")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	compiler := jsonschema.NewCompiler()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		typeName := strings.TrimSuffix(entry.Name(), ".json")
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("schema: read %s: %v", path, err)
			continue
		}
		var raw any
		if err := json.Unmarshal(data, &raw); err != nil {
			log.Printf("schema: parse %s: %v", path, err)
			continue
		}
		url := "file:///" + typeName + ".json"
		if err := compiler.AddResource(url, raw); err != nil {
			log.Printf("schema: add resource %s: %v", typeName, err)
			continue
		}
		sch, err := compiler.Compile(url)
		if err != nil {
			log.Printf("schema: compile %s: %v", typeName, err)
			continue
		}
		v.mu.Lock()
		v.schemas[typeName] = sch
		v.mu.Unlock()
	}
}

func (v *Validator) Reload() {
	v.mu.Lock()
	v.schemas = make(map[string]*jsonschema.Schema)
	v.mu.Unlock()
	v.loadSchemas()
}

func (v *Validator) Validate(frontmatter map[string]any) *ValidationError {
	if v == nil {
		return nil
	}
	typeName, ok := frontmatter["type"].(string)
	if !ok || typeName == "" {
		return nil
	}

	v.mu.RLock()
	sch, exists := v.schemas[typeName]
	v.mu.RUnlock()

	if !exists {
		return nil
	}

	err := sch.Validate(frontmatter)
	if err == nil {
		return nil
	}

	var errs []string
	if ve, ok := err.(*jsonschema.ValidationError); ok {
		for _, cause := range ve.Causes {
			errs = append(errs, cause.Error())
		}
		if len(errs) == 0 {
			errs = append(errs, ve.Error())
		}
	} else {
		errs = append(errs, err.Error())
	}

	return &ValidationError{Type: typeName, Errors: errs}
}

func (v *Validator) ListTypes() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	types := make([]string, 0, len(v.schemas))
	for t := range v.schemas {
		types = append(types, t)
	}
	return types
}
