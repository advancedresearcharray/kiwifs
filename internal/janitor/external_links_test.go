package janitor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestExtractExternalURLs_SkipsCodeBlocks(t *testing.T) {
	body := `
See https://good.example.com/page for docs.

` + "```go\n" + `url := "https://ignored.example.com/in-code"
` + "```" + `

Inline ` + "`https://ignored.example.com/inline`" + ` stays hidden.

<https://autolink.example.com/ok>
`
	got := extractExternalURLs(body)
	if len(got) != 2 {
		t.Fatalf("expected 2 URLs, got %v", got)
	}
	if got[0] != "https://good.example.com/page" || got[1] != "https://autolink.example.com/ok" {
		t.Fatalf("unexpected URLs: %v", got)
	}
}

func TestHostIgnored(t *testing.T) {
	ignore := []string{"localhost", "127.0.0.1", "example.com"}
	cases := []struct {
		url  string
		want bool
	}{
		{"https://localhost/docs", true},
		{"https://127.0.0.1:8080/x", true},
		{"https://sub.example.com/x", true},
		{"https://real.example.org/x", false},
	}
	for _, tc := range cases {
		if got := hostIgnored(tc.url, ignore); got != tc.want {
			t.Fatalf("hostIgnored(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestURLAllowedForProbe_SSRFBlocksPrivateIPs(t *testing.T) {
	ignore := []string{"example.com"}
	cases := []struct {
		url  string
		want bool
	}{
		{"https://127.0.0.1/admin", false},
		{"https://10.0.0.1/internal", false},
		{"https://192.168.1.1/router", false},
		{"https://169.254.169.254/latest/meta-data", false},
		{"http://[::1]/local", false},
		{"ftp://public.example.com/file", false},
		{"not-a-url", false},
		{"https://public.example.org/docs", true},
	}
	for _, tc := range cases {
		if got := urlAllowedForProbe(tc.url, ignore, nil, false); got != tc.want {
			t.Fatalf("urlAllowedForProbe(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestURLAllowedForProbe_Whitelist(t *testing.T) {
	allow := []string{"github.com", "docs.example.com"}
	if urlAllowedForProbe("https://evil.example.org/x", nil, allow, false) {
		t.Fatal("expected non-allowlisted host to be blocked")
	}
	if !urlAllowedForProbe("https://api.github.com/repos", nil, allow, false) {
		t.Fatal("expected allowlisted host to pass")
	}
	if !urlAllowedForProbe("https://docs.example.com/guide", nil, allow, false) {
		t.Fatal("expected subdomain of allowlisted host to pass")
	}
}

func TestValidateWorkspaceRoot(t *testing.T) {
	root := t.TempDir()
	got, err := validateWorkspaceRoot(root)
	if err != nil {
		t.Fatalf("validateWorkspaceRoot: %v", err)
	}
	if got != root {
		t.Fatalf("got %q want %q", got, root)
	}
	if _, err := validateWorkspaceRoot(""); err == nil {
		t.Fatal("expected error for empty root")
	}
	if _, err := validateWorkspaceRoot(filepath.Join(root, "missing-dir")); err == nil {
		t.Fatal("expected error for missing root")
	}
}

func TestLinkCheckCachePath_StaysInsideRoot(t *testing.T) {
	root := t.TempDir()
	path, err := linkCheckCachePath(root)
	if err != nil {
		t.Fatalf("linkCheckCachePath: %v", err)
	}
	want := filepath.Join(root, ".kiwi", "cache", "link-check.json")
	if path != want {
		t.Fatalf("got %q want %q", path, want)
	}
}

func TestExternalLinkConfigFrom_ValidatesRoot(t *testing.T) {
	if _, err := ExternalLinkConfigFrom(true, time.Second, time.Hour, 0, nil, nil, 10, 5, ""); err == nil {
		t.Fatal("expected error for empty root")
	}
}

func TestScan_FlagsBrokenExternalLink(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	okURL := srv.URL + "/ok"
	badURL := srv.URL + "/missing"

	store, root := buildStore(t, map[string]string{
		"docs/setup.md": `---
title: Setup
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Install guide at ` + okURL + ` and legacy docs at ` + badURL + ` for reference.
`,
	})

	cachePath := filepath.Join(root, ".kiwi", "cache", "link-check.json")
	client := srv.Client()
	sc := New(root, store, nil, 90, WithExternalLinks(ExternalLinkConfig{
		Enabled:       true,
		Timeout:       2 * time.Second,
		CacheTTL:      time.Minute,
		CachePath:     cachePath,
		Client:        client,
		Ignore:        []string{"example.com"},
		MaxConcurrent: 10,
		RequestDelay:  0,
	}))

	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	byKind := issuesByKind(res.Issues)
	if len(byKind[IssueExternalLinkRot]) != 1 {
		t.Fatalf("expected 1 external-link-rot issue, got %+v", res.Issues)
	}
	if byKind[IssueExternalLinkRot][0].Path != "docs/setup.md" {
		t.Fatalf("unexpected path: %s", byKind[IssueExternalLinkRot][0].Path)
	}
	if !strings.Contains(byKind[IssueExternalLinkRot][0].Message, badURL) {
		t.Fatalf("message should mention bad URL: %q", byKind[IssueExternalLinkRot][0].Message)
	}
	if len(res.ExternalLinks) != 1 {
		t.Fatalf("expected 1 external_links entry, got %+v", res.ExternalLinks)
	}
	if res.ExternalLinks[0].URL != badURL || res.ExternalLinks[0].Status != 404 {
		t.Fatalf("unexpected external link finding: %+v", res.ExternalLinks[0])
	}
	if res.ExternalLinks[0].Rule != externalLinkRuleName {
		t.Fatalf("rule = %q", res.ExternalLinks[0].Rule)
	}
	if hits.Load() < 2 {
		t.Fatalf("expected HTTP probes for ok and bad URLs, hits=%d", hits.Load())
	}
}

func TestScan_FlagsServerErrorExternalLink(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Server error: ` + srv.URL + `/fail
`,
	})

	sc := New(root, store, nil, 90, WithExternalLinks(ExternalLinkConfig{
		Enabled:       true,
		Timeout:       2 * time.Second,
		CachePath:     filepath.Join(root, ".kiwi", "cache", "link-check.json"),
		Client:        srv.Client(),
		Ignore:        []string{"example.com"},
		RequestDelay:  0,
	}))
	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(issuesByKind(res.Issues)[IssueExternalLinkRot]) != 1 {
		t.Fatalf("expected 500 to be flagged, got %+v", res.Issues)
	}
	if res.ExternalLinks[0].Status != 500 {
		t.Fatalf("status = %d", res.ExternalLinks[0].Status)
	}
}

func TestScan_SkipsBlockedSSRFURLs(t *testing.T) {
	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Internal: https://127.0.0.1/secret
Private: https://10.0.0.5/internal
Metadata: https://169.254.169.254/latest/meta-data
`,
	})

	// No custom Client → production SSRF guard applies; no HTTP probes run.
	sc := New(root, store, nil, 90, WithExternalLinks(ExternalLinkConfig{
		Enabled:      true,
		Timeout:      2 * time.Second,
		CachePath:    filepath.Join(root, ".kiwi", "cache", "link-check.json"),
		Ignore:       []string{"example.com"},
		RequestDelay: 0,
	}))
	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(res.ExternalLinks) != 0 {
		t.Fatalf("SSRF-blocked URLs should not be probed or flagged, got %+v", res.ExternalLinks)
	}
}

func TestScan_ExternalLinkMaxChecksCap(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	body := `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

`
	for i := 0; i < 5; i++ {
		body += "Link: " + srv.URL + "/page" + string(rune('a'+i)) + "\n"
	}

	store, root := buildStore(t, map[string]string{"page.md": body})
	sc := New(root, store, nil, 90, WithExternalLinks(ExternalLinkConfig{
		Enabled:       true,
		Timeout:       2 * time.Second,
		CachePath:     filepath.Join(root, ".kiwi", "cache", "link-check.json"),
		Client:        srv.Client(),
		Ignore:        []string{"example.com"},
		MaxChecks:     2,
		RequestDelay:  0,
	}))
	if _, err := sc.Scan(context.Background()); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if hits.Load() != 2 {
		t.Fatalf("expected max 2 probes, got %d", hits.Load())
	}
}

func TestScan_ExternalLinkCacheSkipsReprobe(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Broken link: ` + srv.URL + `/gone
`,
	})

	cachePath := filepath.Join(root, ".kiwi", "cache", "link-check.json")
	cfg := ExternalLinkConfig{
		Enabled:       true,
		Timeout:       2 * time.Second,
		CacheTTL:      24 * time.Hour,
		CachePath:     cachePath,
		Client:        srv.Client(),
		Ignore:        []string{"example.com"},
		RequestDelay:  0,
	}
	sc := New(root, store, nil, 90, WithExternalLinks(cfg))

	if _, err := sc.Scan(context.Background()); err != nil {
		t.Fatalf("first scan: %v", err)
	}
	firstHits := hits.Load()
	if firstHits != 1 {
		t.Fatalf("expected 1 probe on first scan, got %d", firstHits)
	}

	if _, err := sc.Scan(context.Background()); err != nil {
		t.Fatalf("second scan: %v", err)
	}
	if hits.Load() != firstHits {
		t.Fatalf("cache should skip second probe, hits=%d", hits.Load())
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read cache: %v", err)
	}
	var cache linkCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		t.Fatalf("unmarshal cache: %v", err)
	}
	if cache.Entries[srv.URL+"/gone"].Status != 404 {
		t.Fatalf("cache entry = %+v", cache.Entries)
	}
}

func TestScan_ExternalLinkHEADFallbackToGET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Works via GET fallback: ` + srv.URL + `/doc
`,
	})

	sc := New(root, store, nil, 90, WithExternalLinks(ExternalLinkConfig{
		Enabled:       true,
		Timeout:       2 * time.Second,
		CachePath:     filepath.Join(root, ".kiwi", "cache", "link-check.json"),
		Client:        srv.Client(),
		Ignore:        []string{"example.com"},
		RequestDelay:  0,
	}))
	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(issuesByKind(res.Issues)[IssueExternalLinkRot]) != 0 {
		t.Fatalf("expected no external-link-rot, got %+v", res.Issues)
	}
}

func TestScan_ExternalLinksDisabledByDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Link: ` + srv.URL + `/missing
`,
	})
	sc := New(root, store, nil, 90)
	res, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(res.ExternalLinks) != 0 || len(issuesByKind(res.Issues)[IssueExternalLinkRot]) != 0 {
		t.Fatalf("expected no external link checks without opt-in, got %+v", res)
	}
}

func TestScan_ExternalLinkWhitelistOnlyProbesAllowed(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store, root := buildStore(t, map[string]string{
		"page.md": `---
title: Page
owner: alice
status: verified
reviewed: 2030-01-01
next-review: 2040-01-01
---

Allowed: ` + srv.URL + `/ok
Blocked: https://other.example.org/nope
`,
	})

	allowHost := strings.TrimPrefix(strings.TrimPrefix(srv.URL, "http://"), "https://")
	allowHost = strings.Split(allowHost, ":")[0]

	sc := New(root, store, nil, 90, WithExternalLinks(ExternalLinkConfig{
		Enabled:       true,
		Timeout:       2 * time.Second,
		CachePath:     filepath.Join(root, ".kiwi", "cache", "link-check.json"),
		Client:        srv.Client(),
		Ignore:        []string{"example.com"},
		Allow:         []string{allowHost},
		RequestDelay:  0,
	}))
	if _, err := sc.Scan(context.Background()); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("expected 1 probe for allowlisted host, hits=%d", hits.Load())
	}
}

func TestOptionsFromExternalLinks_DisabledWhenFalse(t *testing.T) {
	root := t.TempDir()
	if opts := OptionsFromExternalLinks(false, time.Second, time.Hour, 0, nil, nil, 10, 5, root); opts != nil {
		t.Fatalf("expected nil opts when disabled, got %v", opts)
	}
}

func TestOptionsFromExternalLinks_InvalidRootReturnsNil(t *testing.T) {
	if opts := OptionsFromExternalLinks(true, time.Second, time.Hour, 0, nil, nil, 10, 5, ""); opts != nil {
		t.Fatalf("expected nil opts for invalid root, got %v", opts)
	}
}
