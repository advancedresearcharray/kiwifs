package workspace

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// excludedEmbedDirs lists template paths that must not ship via kiwifs init.
// go:embed all:templates includes every file on disk under templates/, including
// local dev scaffolds and pre-UC-6 runbook dirs with placeholder wiki links
// that fail schema.Lint / kiwifs check on generated workspaces.
var excludedEmbedDirs = map[string]bool{
	"templates/runbook/incidents":     true,
	"templates/runbook/postmortems": true,
	"templates/runbook/procedures":    true,
	"templates/knowledge":             true, // superseded by memory/kb templates
	"templates/research/experiments":  true, // dev-only; not part of research init
	"templates/research/literature":   true, // dev-only; research uses papers/
}

func isExcludedEmbedPath(path string) bool {
	path = filepath.ToSlash(path)
	if excludedEmbedDirs[path] {
		return true
	}
	for dir := range excludedEmbedDirs {
		if strings.HasPrefix(path, dir+"/") {
			return true
		}
	}
	return false
}

type filteredTemplatesFS struct {
	inner fs.FS
}

func (f filteredTemplatesFS) Open(name string) (fs.File, error) {
	if isExcludedEmbedPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return f.inner.Open(name)
}

func (f filteredTemplatesFS) ReadFile(name string) ([]byte, error) {
	if isExcludedEmbedPath(name) {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	if rf, ok := f.inner.(fs.ReadFileFS); ok {
		return rf.ReadFile(name)
	}
	return fs.ReadFile(f.inner, name)
}

func (f filteredTemplatesFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if isExcludedEmbedPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	entries, err := fs.ReadDir(f.inner, name)
	if err != nil {
		return nil, err
	}
	out := make([]fs.DirEntry, 0, len(entries))
	for _, e := range entries {
		path := filepath.Join(name, e.Name())
		if isExcludedEmbedPath(path) {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}
