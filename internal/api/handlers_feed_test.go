package api

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
)

func TestFeedAtomHasSingleXMLDeclaration(t *testing.T) {
	s, root := buildTestServerWithRoot(t)

	runGit(t, root, "init")
	runGit(t, root, "config", "user.name", "Test User")
	runGit(t, root, "config", "user.email", "test@example.com")
	mustPutFile(t, s, "pages/feed.md", "# Feed\n")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "add feed page")

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/feed.xml", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /feed.xml: %d %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if count := strings.Count(body, "<?xml"); count != 1 {
		t.Fatalf("expected exactly one XML declaration, got %d:\n%s", count, body[:min(len(body), 200)])
	}
	if !strings.HasPrefix(body, "<?xml") {
		t.Fatalf("feed should start with XML declaration, got %q", body[:min(len(body), 40)])
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}
