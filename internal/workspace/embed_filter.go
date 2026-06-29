package workspace

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// runbookLegacyEmbedDirs lists pre-UC-6 scaffold paths that must not ship with
// kiwifs init --template runbook. They contain placeholder wiki links that fail
// schema.Lint / kiwifs check on generated workspaces.
var runbookLegacyEmbedDirs = map[string]bool{
	"templates/runbook/incidents":   true,
	"templates/runbook/postmortems": true,
	"templates/runbook/procedures":  true,
}

func isRunbookLegacyEmbedPath(path string) bool {
	path = filepath.ToSlash(path)
	if runbookLegacyEmbedDirs[path] {
		return true
	}
	for dir := range runbookLegacyEmbedDirs {
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
	if isRunbookLegacyEmbedPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return f.inner.Open(name)
}

func (f filteredTemplatesFS) ReadFile(name string) ([]byte, error) {
	if isRunbookLegacyEmbedPath(name) {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	if rf, ok := f.inner.(fs.ReadFileFS); ok {
		return rf.ReadFile(name)
	}
	return fs.ReadFile(f.inner, name)
}

func (f filteredTemplatesFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if isRunbookLegacyEmbedPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	entries, err := fs.ReadDir(f.inner, name)
	if err != nil {
		return nil, err
	}
	out := make([]fs.DirEntry, 0, len(entries))
	for _, e := range entries {
		path := filepath.Join(name, e.Name())
		if isRunbookLegacyEmbedPath(path) {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}
