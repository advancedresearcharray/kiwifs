package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPut_RejectsAppendOnlyOverwrite(t *testing.T) {
	s := buildTestServer(t)
	initial := "---\nappend_only: true\n---\nentry one\n"
	mustPutFile(t, s, "events/log.md", initial)

	req := httptest.NewRequest(http.MethodPut, "/api/kiwi/file?path=events/log.md", strings.NewReader("replaced\n"))
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("PUT overwrite: want 409, got %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "append-only") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestPut_AllowsCreateWithAppendOnly(t *testing.T) {
	s := buildTestServer(t)
	content := "---\nappend_only: true\n---\nfirst\n"
	mustPutFile(t, s, "events/new-log.md", content)
}

func TestAppend_AllowsAppendOnlyFile(t *testing.T) {
	s := buildTestServer(t)
	mustPutFile(t, s, "events/log.md", "---\nappend_only: true\n---\nline1\n")

	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/file/append?path=events/log.md", strings.NewReader("line2"))
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("append: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=events/log.md", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if got := rec.Body.String(); !strings.Contains(got, "line1") || !strings.Contains(got, "line2") {
		t.Fatalf("content = %q", got)
	}
}

func TestPut_UnaffectedWithoutAppendOnly(t *testing.T) {
	s := buildTestServer(t)
	mustPutFile(t, s, "note.md", "v1\n")
	mustPutFile(t, s, "note.md", "v2\n")
}
