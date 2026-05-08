package search

import "sync"

// TagIndex maintains a bidirectional mapping between tags and file paths.
// Built from frontmatter at startup/reindex, kept in memory.
// Files are the source of truth — the index rebuilds from them.
type TagIndex struct {
	mu    sync.RWMutex
	index map[string][]string // tag → [paths]
	tags  map[string][]string // path → [tags]
}

func NewTagIndex() *TagIndex {
	return &TagIndex{
		index: make(map[string][]string),
		tags:  make(map[string][]string),
	}
}

// Update replaces tags for a given path.
func (t *TagIndex) Update(path string, tags []string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, oldTag := range t.tags[path] {
		filtered := t.index[oldTag][:0]
		for _, p := range t.index[oldTag] {
			if p != path {
				filtered = append(filtered, p)
			}
		}
		t.index[oldTag] = filtered
	}

	t.tags[path] = tags
	for _, tag := range tags {
		t.index[tag] = append(t.index[tag], path)
	}
}

// Remove drops all entries for a path.
func (t *TagIndex) Remove(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, tag := range t.tags[path] {
		filtered := t.index[tag][:0]
		for _, p := range t.index[tag] {
			if p != path {
				filtered = append(filtered, p)
			}
		}
		t.index[tag] = filtered
	}
	delete(t.tags, path)
}

// ByTag returns all paths with the given tag.
func (t *TagIndex) ByTag(tag string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]string, len(t.index[tag]))
	copy(result, t.index[tag])
	return result
}

// TagsFor returns all tags for a given path.
func (t *TagIndex) TagsFor(path string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]string, len(t.tags[path]))
	copy(result, t.tags[path])
	return result
}

// All returns the full tag → paths map.
func (t *TagIndex) All() map[string][]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make(map[string][]string, len(t.index))
	for k, v := range t.index {
		cp := make([]string, len(v))
		copy(cp, v)
		result[k] = cp
	}
	return result
}
