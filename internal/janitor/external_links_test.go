package janitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckExternalLinks_Flags404(t *testing.T) {
	var headCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			headCount++
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	store, root := buildStore(t, map[string]string{
		"docs/setup.md": `---
title: Setup
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Follow the guide at ` + srv.URL + `/guide for installation steps beyond this intro text.
`,
	})

	client := srv.Client()
	sc := New(root, store, nil, 90, WithExternalLinkCheck(ExternalLinkCheckConfig{
		Enabled:    true,
		Timeout:    2 * time.Second,
		Ignore:     []string{},
		CacheDir:   root,
		HTTPClient: client,
		Now:        func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	}))

	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(res.ExternalLinks) != 1 {
		t.Fatalf("expected 1 external link finding, got %+v", res.ExternalLinks)
	}
	if res.ExternalLinks[0].Status != http.StatusNotFound {
		t.Fatalf("status = %d", res.ExternalLinks[0].Status)
	}
	if res.ExternalLinks[0].Rule != externalLinkRuleName {
		t.Fatalf("rule = %q", res.ExternalLinks[0].Rule)
	}
	by := issuesByKind(res.Issues)
	if len(by[IssueExternalLinkRot]) != 1 {
		t.Fatalf("expected external-link-rot issue, got %+v", by[IssueExternalLinkRot])
	}
	if headCount == 0 {
		t.Fatal("expected at least one HEAD request")
	}
}

func TestCheckExternalLinks_UsesCache(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Broken link: ` + srv.URL + `/missing and enough filler text to exceed the empty page threshold easily.
`,
	})

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	cfg := ExternalLinkCheckConfig{
		Enabled:    true,
		Timeout:    2 * time.Second,
		Ignore:     []string{},
		CacheDir:   root,
		HTTPClient: srv.Client(),
		Now:        func() time.Time { return now },
	}
	sc := New(root, store, nil, 90, WithExternalLinkCheck(cfg))

	if _, err := sc.Scan(context.Background()); err != nil {
		t.Fatalf("first scan: %v", err)
	}
	first := requests

	cfg.Now = func() time.Time { return now.Add(time.Hour) }
	sc = New(root, store, nil, 90, WithExternalLinkCheck(cfg))
	if _, err := sc.Scan(context.Background()); err != nil {
		t.Fatalf("second scan: %v", err)
	}
	if requests != first {
		t.Fatalf("expected cached URL to skip re-check, requests before=%d after=%d", first, requests)
	}
}

func TestCheckExternalLinks_IgnoresConfiguredHosts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	host := strings.TrimPrefix(srv.URL, "http://")
	if i := strings.Index(host, "/"); i >= 0 {
		host = host[:i]
	}
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

See ` + srv.URL + `/ignored for details and enough filler text to exceed the empty page threshold easily.
`,
	})

	sc := New(root, store, nil, 90, WithExternalLinkCheck(ExternalLinkCheckConfig{
		Enabled:    true,
		Timeout:    2 * time.Second,
		Ignore:     []string{host},
		CacheDir:   root,
		HTTPClient: srv.Client(),
	}))
	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(res.ExternalLinks) != 0 {
		t.Fatalf("expected ignored host to produce no findings, got %+v", res.ExternalLinks)
	}
}

func TestCheckExternalLinks_HEADFallbackToGET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Range") != "bytes=0-0" {
			t.Fatalf("expected Range header on GET fallback, got %q", r.Header.Get("Range"))
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Link: ` + srv.URL + `/doc with enough filler text to exceed the empty page threshold easily.
`,
	})

	sc := New(root, store, nil, 90, WithExternalLinkCheck(ExternalLinkCheckConfig{
		Enabled:    true,
		Timeout:    2 * time.Second,
		Ignore:     []string{},
		CacheDir:   root,
		HTTPClient: srv.Client(),
	}))
	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(res.ExternalLinks) != 1 {
		t.Fatalf("expected finding after GET fallback, got %+v", res.ExternalLinks)
	}
}

func TestCheckExternalLinks_SkipsURLsInCodeBlocks(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

` + "```" + `
` + srv.URL + `
` + "```" + `

Enough filler text here to exceed the empty page threshold without any live external URLs.
`,
	})

	sc := New(root, store, nil, 90, WithExternalLinkCheck(ExternalLinkCheckConfig{
		Enabled:    true,
		Timeout:    2 * time.Second,
		Ignore:     []string{},
		CacheDir:   root,
		HTTPClient: srv.Client(),
	}))
	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(res.ExternalLinks) != 0 {
		t.Fatalf("expected no external link checks from fenced code, got %+v", res.ExternalLinks)
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestShouldIgnoreExternalURL(t *testing.T) {
	ignore := []string{"localhost", "example.com"}
	cases := map[string]bool{
		"https://localhost/docs":      true,
		"https://api.example.com/x":   true,
		"https://kiwifs.com/docs":     false,
		"not-a-url":                   true,
	}
	for raw, want := range cases {
		if got := shouldIgnoreExternalURL(raw, ignore); got != want {
			t.Fatalf("%q: got %v want %v", raw, got, want)
		}
	}
}
