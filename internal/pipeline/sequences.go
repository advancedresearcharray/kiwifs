package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var seqMarkerRe = regexp.MustCompile(`<!-- seq:(\d+) -->`)

type sequenceState struct {
	Counter int `json:"counter"`
}

// SequenceStore persists a monotonic counter in .kiwi/state/sequences.json
// and assigns sequence numbers to appends in configured directories.
type SequenceStore struct {
	path        string
	directories []string // normalized with trailing /
	mu          sync.Mutex
	counter     int
}

// NewSequenceStore opens (or creates) the counter file. Returns nil when
// directories is empty — sequence numbering is disabled.
func NewSequenceStore(root string, directories []string) (*SequenceStore, error) {
	dirs := normalizeSequenceDirs(directories)
	if len(dirs) == 0 {
		return nil, nil
	}
	stateDir := filepath.Join(root, ".kiwi", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, fmt.Errorf("sequences: mkdir: %w", err)
	}
	s := &SequenceStore{
		path:        filepath.Join(stateDir, "sequences.json"),
		directories: dirs,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("sequences: load: %w", err)
	}
	return s, nil
}

func normalizeSequenceDirs(dirs []string) []string {
	out := make([]string, 0, len(dirs))
	seen := make(map[string]bool)
	for _, d := range dirs {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		d = filepath.ToSlash(d)
		if !strings.HasSuffix(d, "/") {
			d += "/"
		}
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out
}

// AppliesTo reports whether path is under a configured sequenced directory.
func (s *SequenceStore) AppliesTo(path string) bool {
	if s == nil || len(s.directories) == 0 {
		return false
	}
	path = filepath.ToSlash(path)
	for _, dir := range s.directories {
		if strings.HasPrefix(path, dir) {
			return true
		}
		if path == strings.TrimSuffix(dir, "/") {
			return true
		}
	}
	return false
}

// Counter returns the last assigned sequence number (0 when none assigned).
func (s *SequenceStore) Counter() int {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counter
}

func (s *SequenceStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var st sequenceState
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}
	s.counter = st.Counter
	return nil
}

func (s *SequenceStore) saveLocked() error {
	st := sequenceState{Counter: s.counter}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".sequences-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if tmpName != "" {
			os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return err
	}
	tmpName = ""
	return nil
}

// Next allocates the next sequence number and persists the counter atomically.
func (s *SequenceStore) Next() (int, error) {
	if s == nil {
		return 0, fmt.Errorf("sequences: store not configured")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	seq := s.counter
	if err := s.saveLocked(); err != nil {
		s.counter--
		return 0, err
	}
	return seq, nil
}

// SequenceMarker returns the HTML comment injected before appended content.
func SequenceMarker(seq int) string {
	return fmt.Sprintf("<!-- seq:%d -->", seq)
}

// InjectSequence prepends a sequence marker to content.
func InjectSequence(content string, seq int) string {
	return SequenceMarker(seq) + "\n" + content
}

// SequenceIssue describes a gap or inconsistency found during verification.
type SequenceIssue struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Missing int    `json:"missing,omitempty"`
}

// SequenceCheckResult is returned by CheckSequences.
type SequenceCheckResult struct {
	Issues   []SequenceIssue `json:"issues"`
	Found    []int           `json:"found"`
	Counter  int             `json:"counter"`
	MaxFound int             `json:"max_found"`
}

func (r *SequenceCheckResult) HasIssues() bool {
	return len(r.Issues) > 0
}

func (r *SequenceCheckResult) Summary() string {
	if !r.HasIssues() {
		return fmt.Sprintf("Sequences OK (%d markers, counter=%d)\n", len(r.Found), r.Counter)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Sequence check: %d issue(s)\n", len(r.Issues))
	for _, iss := range r.Issues {
		fmt.Fprintf(&sb, "  [%s] %s\n", iss.Kind, iss.Message)
	}
	return sb.String()
}

// CheckSequences scans configured directories for sequence markers and reports
// gaps, duplicates, and counter mismatches against sequences.json.
func CheckSequences(root string, directories []string) (*SequenceCheckResult, error) {
	dirs := normalizeSequenceDirs(directories)
	result := &SequenceCheckResult{}
	if len(dirs) == 0 {
		return result, nil
	}

	store, err := NewSequenceStore(root, directories)
	if err != nil {
		return nil, err
	}
	if store != nil {
		result.Counter = store.Counter()
	}

	counts := make(map[int]int)
	for _, dirPrefix := range dirs {
		relDir := strings.TrimSuffix(dirPrefix, "/")
		absDir := filepath.Join(root, filepath.FromSlash(relDir))
		info, err := os.Stat(absDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("sequences: stat %s: %w", relDir, err)
		}
		if !info.IsDir() {
			continue
		}
		err = filepath.Walk(absDir, func(path string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil || fi.IsDir() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			for _, m := range seqMarkerRe.FindAllSubmatch(data, -1) {
				n, err := strconv.Atoi(string(m[1]))
				if err != nil {
					continue
				}
				counts[n]++
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	for n, c := range counts {
		result.Found = append(result.Found, n)
		if c > 1 {
			result.Issues = append(result.Issues, SequenceIssue{
				Kind:    "duplicate",
				Message: fmt.Sprintf("sequence %d appears %d times", n, c),
			})
		}
		if n > result.MaxFound {
			result.MaxFound = n
		}
	}
	sort.Ints(result.Found)

	if len(result.Found) > 0 {
		expect := result.Found[0]
		for _, n := range result.Found[1:] {
			for expect+1 < n {
				expect++
				result.Issues = append(result.Issues, SequenceIssue{
					Kind:    "gap",
					Message: fmt.Sprintf("missing sequence %d", expect),
					Missing: expect,
				})
			}
			expect = n
		}
	}

	if result.MaxFound > 0 && result.Counter != result.MaxFound {
		result.Issues = append(result.Issues, SequenceIssue{
			Kind: "counter_mismatch",
			Message: fmt.Sprintf("sequences.json counter=%d but highest marker is %d",
				result.Counter, result.MaxFound),
		})
	}

	return result, nil
}
