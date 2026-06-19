package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

const sequencesStateFile = ".kiwi/state/sequences.json"

type sequenceState struct {
	Counters map[string]int64 `json:"counters"`
}

type sequenceStore struct {
	path string
	mu   sync.Mutex
}

func newSequenceStore(root string) *sequenceStore {
	return &sequenceStore{path: filepath.Join(root, sequencesStateFile)}
}

func (s *sequenceStore) next(dirKey string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.load()
	if err != nil {
		return 0, err
	}
	if state.Counters == nil {
		state.Counters = map[string]int64{}
	}
	state.Counters[dirKey]++
	next := state.Counters[dirKey]
	if err := s.save(state); err != nil {
		return 0, err
	}
	return next, nil
}

func (s *sequenceStore) load() (*sequenceState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &sequenceState{Counters: map[string]int64{}}, nil
		}
		return nil, err
	}
	var state sequenceState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse sequences state: %w", err)
	}
	if state.Counters == nil {
		state.Counters = map[string]int64{}
	}
	return &state, nil
}

func (s *sequenceStore) save(state *sequenceState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func normalizeSequencePath(userPath string) string {
	slash := strings.ReplaceAll(userPath, "\\", "/")
	clean := path.Clean("/" + slash)
	return strings.TrimPrefix(clean, "/")
}

func sequenceDirKey(userPath string, directories []string) string {
	p := normalizeSequencePath(userPath)
	for _, dir := range directories {
		d := strings.TrimSuffix(filepath.ToSlash(strings.TrimSpace(dir)), "/")
		if d == "" {
			continue
		}
		if p == d || strings.HasPrefix(p, d+"/") {
			return d
		}
	}
	return ""
}

func injectSequenceMarker(content string, seq int64) string {
	marker := fmt.Sprintf("<!-- seq:%d -->", seq)
	if strings.TrimSpace(content) == "" {
		return marker
	}
	return marker + "\n" + content
}

// CheckSequenceGaps reports missing sequence numbers for configured directories.
func CheckSequenceGaps(root string, directories []string) ([]string, error) {
	if len(directories) == 0 {
		return nil, nil
	}
	statePath := filepath.Join(root, sequencesStateFile)
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var state sequenceState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse sequences state: %w", err)
	}
	var issues []string
	for _, dir := range directories {
		key := strings.TrimSuffix(filepath.ToSlash(strings.TrimSpace(dir)), "/")
		max := state.Counters[key]
		if max <= 0 {
			continue
		}
		seen := map[int64]bool{}
		dirPath := filepath.Join(root, filepath.FromSlash(key))
		_ = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
				return nil
			}
			body, rerr := os.ReadFile(path)
			if rerr != nil {
				return nil
			}
			for _, n := range extractSequenceMarkers(string(body)) {
				seen[n] = true
			}
			return nil
		})
		for i := int64(1); i <= max; i++ {
			if !seen[i] {
				issues = append(issues, fmt.Sprintf("%s: missing seq:%d", key, i))
			}
		}
	}
	return issues, nil
}

func extractSequenceMarkers(body string) []int64 {
	var out []int64
	for {
		idx := strings.Index(body, "<!-- seq:")
		if idx < 0 {
			break
		}
		body = body[idx+len("<!-- seq:"):]
		end := strings.Index(body, " -->")
		if end < 0 {
			break
		}
		var n int64
		if _, err := fmt.Sscanf(body[:end], "%d", &n); err == nil && n > 0 {
			out = append(out, n)
		}
		body = body[end:]
	}
	return out
}
