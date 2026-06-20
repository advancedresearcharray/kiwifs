package workspace

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed all:templates
var templates embed.FS

// InitTemplate describes a workspace scaffold available at space creation.
type InitTemplate struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// nonSpaceDirs are embedded paths that are not selectable workspace templates.
var nonSpaceDirs = map[string]bool{
	"workflow": true,
}

var templateLabels = map[string]string{
	"knowledge":      "Knowledge Base",
	"wiki":           "Wiki",
	"runbook":        "Runbook",
	"research":       "Research",
	"tasks":          "Tasks",
	"prompt-library": "Prompt Library",
	"adr":            "Architecture Decision Records",
	"blank":          "Blank",
}

var templateDescriptions = map[string]string{
	"knowledge":      "LLM-maintained knowledge base with schema, episodes, and agent playbook",
	"wiki":           "Wiki with onboarding, ADRs, processes, and reference docs",
	"runbook":        "Operational runbooks and incident response procedures",
	"research":       "Research library with paper tracking, reading workflow, and literature reviews",
	"tasks":          "Task tracking with priorities and status workflows",
	"prompt-library": "Versioned prompt registry with schemas, eval rubrics, and DQL metrics",
	"adr":            "Architecture Decision Records with MADR format, status workflow, and JSON schema",
	"blank":          "Empty workspace with Kiwi config only",
}

// EmbeddedTemplates returns the embedded template filesystem (for tests).
func EmbeddedTemplates() fs.FS {
	return templates
}

// ListInitTemplates returns workspace scaffolds derived from embedded templates.
func ListInitTemplates() ([]InitTemplate, error) {
	entries, err := fs.ReadDir(templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("read templates: %w", err)
	}

	seen := map[string]bool{"blank": true}
	out := []InitTemplate{metaFor("blank")}

	for _, e := range entries {
		if !e.IsDir() || nonSpaceDirs[e.Name()] {
			continue
		}
		if _, err := fs.Stat(templates, filepath.Join("templates", e.Name(), "index.md")); err != nil {
			if _, err := fs.Stat(templates, filepath.Join("templates", e.Name(), "playbook.md")); err != nil {
				continue
			}
		}
		seen[e.Name()] = true
		out = append(out, metaFor(e.Name()))
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ID == "blank" {
			return false
		}
		if out[j].ID == "blank" {
			return true
		}
		return out[i].Label < out[j].Label
	})
	return out, nil
}

func metaFor(id string) InitTemplate {
	label := templateLabels[id]
	if label == "" {
		label = strings.ReplaceAll(id, "-", " ")
		if label != "" {
			label = strings.ToUpper(label[:1]) + label[1:]
		}
	}
	desc := templateDescriptions[id]
	if desc == "" {
		desc = fmt.Sprintf("%s workspace template", label)
	}
	return InitTemplate{ID: id, Label: label, Description: desc}
}

// Init scaffolds a workspace root with the given template id.
// Blank creates only Kiwi config; other ids copy the embedded scaffold.
func Init(root, template string) error {
	if template == "" {
		template = "blank"
	}

	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("create root: %w", err)
	}

	switch template {
	case "knowledge", "wiki", "runbook", "research", "tasks", "prompt-library", "adr":
		if err := copyEmbedDir("templates/"+template, root); err != nil {
			return err
		}
	case "memory":
		return fmt.Errorf("the 'memory' template has been merged into 'knowledge' — use --template knowledge instead")
	case "blank":
		// directory only
	default:
		return fmt.Errorf("unknown template %q", template)
	}

	kiwiDir := filepath.Join(root, ".kiwi")
	if err := os.MkdirAll(kiwiDir, 0755); err != nil {
		return fmt.Errorf("create .kiwi: %w", err)
	}

	gitignorePath := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		data, _ := fs.ReadFile(templates, "templates/gitignore.txt")
		if err := os.WriteFile(gitignorePath, data, 0644); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
	}

	templatesDir := filepath.Join(kiwiDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("create .kiwi/templates: %w", err)
	}
	if err := copyEmbedDir("templates/workflow", templatesDir); err != nil {
		return err
	}

	configPath := filepath.Join(kiwiDir, "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		data, _ := fs.ReadFile(templates, "templates/config.toml")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}

	if template != "blank" {
		playbookSrc := "templates/" + template + "/playbook.md"
		if data, err := fs.ReadFile(templates, playbookSrc); err == nil {
			playbookDest := filepath.Join(kiwiDir, "playbook.md")
			if _, err := os.Stat(playbookDest); os.IsNotExist(err) {
				_ = os.WriteFile(playbookDest, data, 0644)
			}
		}
	}

	rulesPath := filepath.Join(kiwiDir, "rules.md")
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		data, _ := fs.ReadFile(templates, "templates/rules.md")
		if len(data) > 0 {
			_ = os.WriteFile(rulesPath, data, 0644)
		}
	}

	readmePath := filepath.Join(root, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		data, _ := fs.ReadFile(templates, "templates/README.md")
		if len(data) > 0 {
			_ = os.WriteFile(readmePath, data, 0644)
		}
	}

	return nil
}

func copyEmbedDir(srcDir, destRoot string) error {
	return fs.WalkDir(templates, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(path, srcDir+"/")
		if rel == srcDir {
			return nil
		}
		dest := filepath.Join(destRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		if _, err := os.Stat(dest); err == nil {
			return nil
		}
		data, err := fs.ReadFile(templates, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
}
