package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/kiwifs/kiwifs/internal/pipeline"
)

func buildTestServerWithValidateWrite(t *testing.T, rules []config.ValidateWriteRuleConfig) *Server {
	t.Helper()
	dir, pipe, cstore := buildTestPipeline(t)
	if len(rules) > 0 {
		wv := pipeline.NewWriteRuleValidator(pipe.Store, rules)
		pipe.ValidateWrite = func(ctx context.Context, path string, content []byte, kind pipeline.WriteKind) error {
			return wv.Validate(ctx, path, content, kind)
		}
	}
	cfg := &config.Config{}
	cfg.Storage.Root = dir
	return NewServer(cfg, pipe, nil, cstore, nil, nil, nil)
}

func TestValidateWriteOverwriteReturns409OnPutAllowsAppend(t *testing.T) {
	s := buildTestServerWithValidateWrite(t, []config.ValidateWriteRuleConfig{{
		Name:    "append-only",
		Reject:  "overwrite",
		Message: "This file is append-only. Use POST /api/kiwi/file/append.",
		Match:   config.ValidateWriteMatchConfig{Frontmatter: "append_only", Value: "true"},
	}})
	initial := "---\nappend_only: true\n---\nentry one\n"
	mustPutFile(t, s, "log.md", initial)

	req := httptest.NewRequest(http.MethodPut, "/api/kiwi/file?path=log.md", strings.NewReader("---\nappend_only: true\n---\nreplaced\n"))
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("PUT overwrite: want 409, got %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "append-only") {
		t.Fatalf("expected custom message in body, got %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/kiwi/file/append?path=log.md", strings.NewReader("entry two"))
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST append: want 200, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestValidateWriteBodyChangeReturns409OnBodyEditAllowsFrontmatterPatch(t *testing.T) {
	s := buildTestServerWithValidateWrite(t, []config.ValidateWriteRuleConfig{{
		Name:    "immutable-after-status",
		Reject:  "body_change",
		Message: "Accepted decisions cannot be edited.",
		Match:   config.ValidateWriteMatchConfig{Frontmatter: "status", Values: []string{"accepted", "deprecated", "superseded"}},
	}})
	initial := "---\nstatus: accepted\n---\n# Decision\n\nBody text.\n"
	mustPutFile(t, s, "adr.md", initial)

	req := httptest.NewRequest(http.MethodPut, "/api/kiwi/file?path=adr.md", strings.NewReader("---\nstatus: accepted\n---\n# Decision\n\nEdited body.\n"))
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("PUT body change: want 409, got %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPatch, "/api/kiwi/file/frontmatter?path=adr.md", strings.NewReader(`{"fields":{"reviewed_by":"alice"}}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH frontmatter: want 200, got %d %s", rec.Code, rec.Body.String())
	}
}
