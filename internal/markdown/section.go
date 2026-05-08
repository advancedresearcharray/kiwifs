package markdown

import (
	"fmt"
	"strings"
)

// Section represents a single heading section in a markdown document.
type Section struct {
	Heading   string `json:"heading"`
	Level     int    `json:"level"`
	Content   string `json:"content"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
}

// ExtractSection finds a section by heading text. It supports:
//   - Exact match (case-insensitive): "Training" matches "Training" before "Training Deep Networks"
//   - Hierarchical addressing with ">": "Infrastructure Design > Training" matches "Training" under "Infrastructure Design"
//   - Partial match fallback: "Training" will still match "Training Deep Networks" if no exact match exists
func ExtractSection(body []byte, heading string) (*Section, error) {
	sections := SplitSections(body)

	// Support hierarchical addressing: "Parent > Child"
	if parts := strings.SplitN(heading, ">", 2); len(parts) == 2 {
		parent := strings.ToLower(strings.TrimSpace(parts[0]))
		child := strings.ToLower(strings.TrimSpace(parts[1]))
		inParent := false
		parentLevel := 0
		for i := range sections {
			hLower := strings.ToLower(sections[i].Heading)
			if !inParent {
				if strings.EqualFold(hLower, parent) || strings.Contains(hLower, parent) {
					inParent = true
					parentLevel = sections[i].Level
				}
				continue
			}
			if sections[i].Level <= parentLevel {
				break // left the parent scope
			}
			if strings.EqualFold(hLower, child) || strings.Contains(hLower, child) {
				return &sections[i], nil
			}
		}
		return nil, fmt.Errorf("section %q > %q not found", parts[0], parts[1])
	}

	headingLower := strings.ToLower(strings.TrimSpace(heading))

	// First pass: exact match
	for i := range sections {
		if strings.EqualFold(sections[i].Heading, headingLower) {
			return &sections[i], nil
		}
	}

	// Second pass: partial match
	for i := range sections {
		if strings.Contains(strings.ToLower(sections[i].Heading), headingLower) {
			return &sections[i], nil
		}
	}
	return nil, fmt.Errorf("section %q not found", heading)
}

// ExtractSectionByIndex returns the Nth section (0-indexed).
func ExtractSectionByIndex(body []byte, index int) (*Section, error) {
	sections := SplitSections(body)
	if index < 0 || index >= len(sections) {
		return nil, fmt.Errorf("section index %d out of range (have %d sections)", index, len(sections))
	}
	return &sections[index], nil
}

// SplitSections splits markdown into sections based on headings.
// Each section spans from its heading line to just before the next heading
// of the same or higher level (or EOF).
func SplitSections(body []byte) []Section {
	lines := strings.Split(string(body), "\n")
	var sections []Section
	var current *Section

	for i, line := range lines {
		level := headingLevel(line)
		if level > 0 {
			if current != nil {
				current.LineEnd = i - 1
				current.Content = strings.TrimSpace(current.Content)
				sections = append(sections, *current)
			}
			title := strings.TrimSpace(strings.TrimLeft(line, "# "))
			current = &Section{
				Heading:   title,
				Level:     level,
				LineStart: i,
			}
		} else if current != nil {
			current.Content += line + "\n"
		}
	}

	if current != nil {
		current.LineEnd = len(lines) - 1
		current.Content = strings.TrimSpace(current.Content)
		sections = append(sections, *current)
	}

	return sections
}

func headingLevel(line string) int {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return 0
	}
	level := 0
	for _, ch := range trimmed {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level > 0 && level < len(trimmed) && trimmed[level] == ' ' {
		return level
	}
	return 0
}
