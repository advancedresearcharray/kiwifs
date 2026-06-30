package janitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	IssueExternalLinkRot     = "external-link-rot"
	externalLinkRuleName     = "external-link-rot"
	defaultLinkCheckTimeout  = 5 * time.Second
	defaultLinkCacheTTL      = 24 * time.Hour
	defaultLinkUserAgent     = "KiwiFS-LinkChecker/1.0"
	maxLinkRedirects         = 3
	maxConcurrentLinkChecks  = 10
	linkCheckRequestDelay    = 100 * time.Millisecond
	linkCheckCacheVersion    = 1
)

var (
	externalURLRe  = regexp.MustCompile(`https?://[^\s\)\]\"'<>]+`)
	fencedCodeRe   = regexp.MustCompile("(?s)```.*?```")
	inlineCodeRe   = regexp.MustCompile("`[^`]+`")
	defaultIgnore  = []string{"localhost", "127.0.0.1", "example.com"}
)

// ExternalLinkFinding is one broken or errored external URL in a markdown page.
type ExternalLinkFinding struct {
	Path   string `json:"path"`
	URL    string `json:"url"`
	Status int    `json:"status"`
	Rule   string `json:"rule"`
}

// ExternalLinkConfig controls HTTP link rot checks during janitor scans.
type ExternalLinkConfig struct {
	Enabled   bool
	Timeout   time.Duration
	Ignore    []string
	CacheTTL  time.Duration
	CachePath string
	Client    *http.Client // nil → default client (tests inject httptest transport)
}

func (c ExternalLinkConfig) enabled() bool {
	return c.Enabled
}

func (c ExternalLinkConfig) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return defaultLinkCheckTimeout
}

func (c ExternalLinkConfig) cacheTTL() time.Duration {
	if c.CacheTTL > 0 {
		return c.CacheTTL
	}
	return defaultLinkCacheTTL
}

func (c ExternalLinkConfig) ignoreHosts() []string {
	if len(c.Ignore) > 0 {
		return c.Ignore
	}
	return defaultIgnore
}

// ExternalLinkConfigFrom builds checker settings from janitor scan options.
func ExternalLinkConfigFrom(enabled bool, timeout, cacheTTL time.Duration, ignore []string, root string) ExternalLinkConfig {
	return ExternalLinkConfig{
		Enabled:   enabled,
		Timeout:   timeout,
		Ignore:    ignore,
		CacheTTL:  cacheTTL,
		CachePath: filepath.Join(root, ".kiwi", "cache", "link-check.json"),
	}
}

// WithExternalLinks enables external URL rot detection on the scanner.
func WithExternalLinks(cfg ExternalLinkConfig) Option {
	return func(s *Scanner) {
		if cfg.enabled() {
			cp := cfg
			s.externalLinks = &cp
		}
	}
}

// OptionsFromExternalLinks returns nil when external link checks are disabled.
func OptionsFromExternalLinks(enabled bool, timeout, cacheTTL time.Duration, ignore []string, root string) []Option {
	if !enabled {
		return nil
	}
	return []Option{WithExternalLinks(ExternalLinkConfigFrom(enabled, timeout, cacheTTL, ignore, root))}
}

type linkCacheEntry struct {
	Status    int       `json:"status"`
	CheckedAt time.Time `json:"checked_at"`
}

type linkCacheFile struct {
	Version int                       `json:"version"`
	Entries map[string]linkCacheEntry `json:"entries"`
}

func extractExternalURLs(body string) []string {
	body = fencedCodeRe.ReplaceAllString(body, "")
	body = inlineCodeRe.ReplaceAllString(body, "")

	seen := make(map[string]bool)
	var urls []string
	for _, m := range externalURLRe.FindAllString(body, -1) {
		u := strings.TrimRight(m, ".,;:!?)'\"]")
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		urls = append(urls, u)
	}
	return urls
}

func hostIgnored(rawURL string, ignore []string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return true
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return true
	}
	for _, pattern := range ignore {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		if host == pattern || strings.HasSuffix(host, "."+pattern) {
			return true
		}
	}
	return false
}

func loadLinkCache(path string) linkCacheFile {
	data, err := os.ReadFile(path)
	if err != nil {
		return linkCacheFile{Version: linkCheckCacheVersion, Entries: map[string]linkCacheEntry{}}
	}
	var cache linkCacheFile
	if err := json.Unmarshal(data, &cache); err != nil || cache.Entries == nil {
		return linkCacheFile{Version: linkCheckCacheVersion, Entries: map[string]linkCacheEntry{}}
	}
	if cache.Version != linkCheckCacheVersion {
		return linkCacheFile{Version: linkCheckCacheVersion, Entries: map[string]linkCacheEntry{}}
	}
	return cache
}

func saveLinkCache(path string, cache linkCacheFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	cache.Version = linkCheckCacheVersion
	if cache.Entries == nil {
		cache.Entries = map[string]linkCacheEntry{}
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *Scanner) checkExternalLinks(ctx context.Context, pages []pageInfo) ([]ExternalLinkFinding, []Issue) {
	if s.externalLinks == nil || !s.externalLinks.enabled() {
		return nil, nil
	}
	cfg := *s.externalLinks
	ignore := cfg.ignoreHosts()
	cache := loadLinkCache(cfg.CachePath)
	ttl := cfg.cacheTTL()
	now := time.Now().UTC()

	type pageURL struct {
		path string
		url  string
	}
	var toCheck []pageURL
	for _, p := range pages {
		for _, u := range extractExternalURLs(p.bodyText) {
			if hostIgnored(u, ignore) {
				continue
			}
			if ent, ok := cache.Entries[u]; ok && now.Sub(ent.CheckedAt) < ttl {
				continue
			}
			toCheck = append(toCheck, pageURL{path: p.path, url: u})
		}
	}

	client := cfg.Client
	if client == nil {
		client = newLinkHTTPClient(cfg.timeout())
	}

	var (
		mu       sync.Mutex
		sem      = make(chan struct{}, maxConcurrentLinkChecks)
		wg       sync.WaitGroup
		updated  bool
		findings []ExternalLinkFinding
		issues   []Issue
	)

	checkOne := func(pu pageURL) {
		defer wg.Done()
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return
		}
		defer func() { <-sem }()

		select {
		case <-time.After(linkCheckRequestDelay):
		case <-ctx.Done():
			return
		}

		status, probeErr := probeURL(ctx, client, pu.url)
		if probeErr != nil {
			status = 0
		}

		mu.Lock()
		defer mu.Unlock()
		cache.Entries[pu.url] = linkCacheEntry{Status: status, CheckedAt: now}
		updated = true
		if status >= 400 || status == 0 {
			findings = append(findings, ExternalLinkFinding{
				Path:   pu.path,
				URL:    pu.url,
				Status: status,
				Rule:   externalLinkRuleName,
			})
			msg := fmt.Sprintf("%s returned HTTP %d", pu.url, status)
			if status == 0 {
				msg = fmt.Sprintf("%s is unreachable", pu.url)
				if probeErr != nil {
					msg = fmt.Sprintf("%s is unreachable (%v)", pu.url, probeErr)
				}
			}
			issues = append(issues, Issue{
				Kind:     IssueExternalLinkRot,
				Path:     pu.path,
				Message:  msg,
				Related:  []string{pu.url},
				Severity: externalLinkSeverity(status),
			})
		}
	}

	for _, pu := range toCheck {
		wg.Add(1)
		go checkOne(pu)
	}
	wg.Wait()

	// Include cached broken links not re-checked this run.
	for _, p := range pages {
		for _, u := range extractExternalURLs(p.bodyText) {
			if hostIgnored(u, ignore) {
				continue
			}
			ent, ok := cache.Entries[u]
			if !ok || now.Sub(ent.CheckedAt) >= ttl {
				continue
			}
			if ent.Status < 400 && ent.Status != 0 {
				continue
			}
			if findingExists(findings, p.path, u) {
				continue
			}
			findings = append(findings, ExternalLinkFinding{
				Path:   p.path,
				URL:    u,
				Status: ent.Status,
				Rule:   externalLinkRuleName,
			})
			msg := fmt.Sprintf("%s returned HTTP %d", u, ent.Status)
			if ent.Status == 0 {
				msg = fmt.Sprintf("%s is unreachable", u)
			}
			issues = append(issues, Issue{
				Kind:     IssueExternalLinkRot,
				Path:     p.path,
				Message:  msg,
				Related:  []string{u},
				Severity: externalLinkSeverity(ent.Status),
			})
		}
	}

	if updated {
		_ = saveLinkCache(cfg.CachePath, cache)
	}
	return findings, issues
}

func findingExists(findings []ExternalLinkFinding, path, rawURL string) bool {
	for _, f := range findings {
		if f.Path == path && f.URL == rawURL {
			return true
		}
	}
	return false
}

func externalLinkSeverity(status int) string {
	if status >= 500 || status == 0 {
		return "error"
	}
	if status >= 400 {
		return "error"
	}
	return "info"
}

func newLinkHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxLinkRedirects {
				return fmt.Errorf("stopped after %d redirects", maxLinkRedirects)
			}
			return nil
		},
	}
}

func probeURL(ctx context.Context, client *http.Client, rawURL string) (int, error) {
	status, err := doLinkRequest(ctx, client, rawURL, http.MethodHead, nil)
	if err == nil && !headNeedsFallback(status) {
		return status, nil
	}
	rangeHeader := http.Header{"Range": []string{"bytes=0-0"}}
	status, getErr := doLinkRequest(ctx, client, rawURL, http.MethodGet, rangeHeader)
	if getErr != nil {
		if err != nil {
			return 0, err
		}
		return 0, getErr
	}
	return status, nil
}

func headNeedsFallback(status int) bool {
	switch status {
	case http.StatusMethodNotAllowed, http.StatusNotImplemented, http.StatusForbidden:
		return true
	default:
		return false
	}
}

func doLinkRequest(ctx context.Context, client *http.Client, rawURL, method string, extra http.Header) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", defaultLinkUserAgent)
	for k, v := range extra {
		req.Header[k] = v
	}
	resp, err := client.Do(req)
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return 0, err
		}
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}
