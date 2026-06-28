package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssignWorkflowAddsWorkflowFrontmatterToExistingPage(t *testing.T) {
	s, root := buildTestServerWithRoot(t)
	workflowDir := filepath.Join(root, ".kiwi", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workflowDir, "tasks.json"), []byte(`{
		"name":"tasks",
		"states":[{"name":"todo","color":"#111111"},{"name":"done","color":"#222222"}],
		"transitions":[{"from":"todo","to":"done"}]
	}`), 0644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	mustPutFile(t, s, "pages/existing.md", "---\ntitle: Existing\ntags: [demo]\n---\n# Existing\n")

	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/workflow/assign", strings.NewReader(`{"path":"pages/existing.md","workflow":"tasks","state":"todo","actor":"tester"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST assign: %d %s", rec.Code, rec.Body.String())
	}
	read := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=pages/existing.md", nil)
	out := httptest.NewRecorder()
	s.echo.ServeHTTP(out, read)
	content := out.Body.String()
	if !strings.Contains(content, "workflow: tasks") || !strings.Contains(content, "state: todo") {
		t.Fatalf("assigned frontmatter missing workflow/state:\n%s", content)
	}
	if !strings.Contains(content, "title: Existing") || !strings.Contains(content, "# Existing") {
		t.Fatalf("assign should preserve existing metadata and body:\n%s", content)
	}
}

func TestAssignWorkflowPrependsFrontmatterWhenMissing(t *testing.T) {
	s, root := buildTestServerWithRoot(t)
	workflowDir := filepath.Join(root, ".kiwi", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workflowDir, "tasks.json"), []byte(`{
		"name":"tasks",
		"states":[{"name":"todo","color":"#111111"}],
		"transitions":[]
	}`), 0644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	mustPutFile(t, s, "pages/plain.md", "# Plain\n")

	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/workflow/assign", strings.NewReader(`{"path":"pages/plain.md","workflow":"tasks","state":"todo"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST assign: %d %s", rec.Code, rec.Body.String())
	}
	read := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=pages/plain.md", nil)
	out := httptest.NewRecorder()
	s.echo.ServeHTTP(out, read)
	content := out.Body.String()
	wantPrefix := "---\nstate: todo\nworkflow: tasks\n---\n# Plain"
	if !strings.HasPrefix(content, wantPrefix) {
		t.Fatalf("expected frontmatter prefix %q, got:\n%s", wantPrefix, content)
	}
}

func TestAssignWorkflowRejectsUnknownState(t *testing.T) {
	s, root := buildTestServerWithRoot(t)
	workflowDir := filepath.Join(root, ".kiwi", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workflowDir, "tasks.json"), []byte(`{
		"name":"tasks",
		"states":[{"name":"todo","color":"#111111"}],
		"transitions":[]
	}`), 0644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	mustPutFile(t, s, "pages/plain.md", "# Plain\n")

	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/workflow/assign", strings.NewReader(`{"path":"pages/plain.md","workflow":"tasks","state":"done"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown state, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestAssignWorkflowRejectsInvalidWorkflowName(t *testing.T) {
	s, _ := buildTestServerWithRoot(t)
	mustPutFile(t, s, "pages/plain.md", "# Plain\n")

	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/workflow/assign", strings.NewReader(`{"path":"pages/plain.md","workflow":"../tasks","state":"todo"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid workflow name, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestAssignWorkflowRejectsNonMarkdownPath(t *testing.T) {
	s, root := buildTestServerWithRoot(t)
	workflowDir := filepath.Join(root, ".kiwi", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workflowDir, "tasks.json"), []byte(`{
		"name":"tasks",
		"states":[{"name":"todo","color":"#111111"}],
		"transitions":[]
	}`), 0644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	mustPutFile(t, s, "data.json", `{"name":"tasks"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/workflow/assign", strings.NewReader(`{"path":"data.json","workflow":"tasks","state":"todo"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-markdown path, got %d %s", rec.Code, rec.Body.String())
	}
}
