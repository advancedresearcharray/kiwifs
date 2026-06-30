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
	IssueExternalLinkRot        = "external-link-rot"
	externalLinkRuleName        = "external-link-rot"
	defaultLinkCheckTimeout     = 5 * time.Second
	defaultLinkCacheTTL         = 24 * time.Hour
	defaultLinkUserAgent        = "KiwiFS-LinkChecker/1.0"
	maxLinkRedirects            = 3
	defaultMaxConcurrentChecks  = 10
	defaultLinkCheckRequestDelay = 100 * time.Millisecond
	defaultMaxChecksPerScan     = 200
	linkCheckCacheVersion       = 1
	dnsLookupTimeout            = 2 * time.Second
)

var (
	externalURLRe = regexp.MustCompile(`https?://[^\s\)\]\"'<>]+`)
	fencedCodeRe  = regexp.MustCompile("(?s)```.*?```")
	inlineCodeRe  = regexp.MustCompile("`[^`]+`")
	defaultIgnore = []string{"localhost", "127.0.0.1", "example.com"}
	blockedHosts  = []string{
		"metadata.google.internal",
		"metadata.goog",
	}
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
	Enabled       bool
	Timeout       time.Duration
	Ignore        []string
	Allow         []string // optional whitelist; when set, only matching hosts are probed
	CacheTTL      time.Duration
	CachePath     string
	MaxChecks     int
	MaxConcurrent int
	RequestDelay  time.Duration
	Client        *http.Client // nil → default client (tests inject httptest transport)
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

func (c ExternalLinkConfig) maxChecks() int {
	if c.MaxChecks > 0 {
		return c.MaxChecks
	}
	return defaultMaxChecksPerScan
}

func (c ExternalLinkConfig) maxConcurrent() int {
	if c.MaxConcurrent > 0 {
		return c.MaxConcurrent
	}
	return defaultMaxConcurrentChecks
}

func (c ExternalLinkConfig) requestDelay() time.Duration {
	if c.RequestDelay >= 0 {
		return c.RequestDelay
	}
	return defaultLinkCheckRequestDelay
}

// ExternalLinkConfigFrom builds checker settings from janitor scan options.
// root is validated and used only for the on-disk cache path under .kiwi/cache/.
func ExternalLinkConfigFrom(
	enabled bool,
	timeout, cacheTTL, requestDelay time.Duration,
	ignore, allow []string,
	maxChecks, maxConcurrent int,
	root string,
) (ExternalLinkConfig, error) {
	cachePath, err := linkCheckCachePath(root)
	if err != nil {
		return ExternalLinkConfig{}, err
	}
	return ExternalLinkConfig{
		Enabled:       enabled,
		Timeout:       timeout,
		Ignore:        ignore,
		Allow:         allow,
		CacheTTL:      cacheTTL,
		CachePath:     cachePath,
		MaxChecks:     maxChecks,
		MaxConcurrent: maxConcurrent,
		RequestDelay:  requestDelay,
	}, nil
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
func OptionsFromExternalLinks(
	enabled bool,
	timeout, cacheTTL, requestDelay time.Duration,
	ignore, allow []string,
	maxChecks, maxConcurrent int,
	root string,
) []Option {
	if !enabled {
		return nil
	}
	cfg, err := ExternalLinkConfigFrom(enabled, timeout, cacheTTL, requestDelay, ignore, allow, maxChecks, maxConcurrent, root)
	if err != nil {
		return nil
	}
	return []Option{WithExternalLinks(cfg)}
}

type linkCacheEntry struct {
	Status    int       `json:"status"`
	CheckedAt time.Time `json:"checked_at"`
}

type linkCacheFile struct {
	Version int                       `json:"version"`
	Entries map[string]linkCacheEntry `json:"entries"`
}

// validateWorkspaceRoot returns a clean absolute path to an existing directory.
func validateWorkspaceRoot(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", fmt.Errorf("empty workspace root")
	}
	abs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("workspace root: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("workspace root: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace root is not a directory: %s", abs)
	}
	return abs, nil
}

func linkCheckCachePath(root string) (string, error) {
	validated, err := validateWorkspaceRoot(root)
	if err != nil {
		return "", err
	}
	cachePath := filepath.Join(validated, ".kiwi", "cache", "link-check.json")
	rel, err := filepath.Rel(validated, cachePath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("link check cache path outside workspace root")
	}
	return cachePath, nil
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

func hostMatchesPattern(host, pattern string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" {
		return false
	}
	host = strings.ToLower(host)
	return host == pattern || strings.HasSuffix(host, "."+pattern)
}

func hostMatchesList(host string, patterns []string) bool {
	for _, pattern := range patterns {
		if hostMatchesPattern(host, pattern) {
			return true
		}
	}
	return false
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
	return hostMatchesList(host, ignore)
}

func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

func isBlockedHostname(host string) bool {
	host = strings.ToLower(host)
	for _, blocked := range blockedHosts {
		if hostMatchesPattern(host, blocked) {
			return true
		}
	}
	return false
}

func hostnameResolvesToBlocked(host string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), dnsLookupTimeout)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return false
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return true
		}
	}
	return false
}

// urlAllowedForProbe applies scheme, ignore/allow lists, and SSRF guards.
// skipSSRF disables private-IP blocking (used when tests inject a custom Client).
func urlAllowedForProbe(rawURL string, ignore, allow []string, skipSSRF bool) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return false
	}
	if len(allow) > 0 && !hostMatchesList(host, allow) {
		return false
	}
	if hostIgnored(rawURL, ignore) {
		return false
	}
	if skipSSRF {
		return true
	}
	if isBlockedHostname(host) {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return !isBlockedIP(ip)
	}
	return !hostnameResolvesToBlocked(host)
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
	allow := cfg.Allow
	cache := loadLinkCache(cfg.CachePath)
	ttl := cfg.cacheTTL()
	now := time.Now().UTC()
	maxChecks := cfg.maxChecks()

	type pageURL struct {
		path string
		url  string
	}
	seenURLs := make(map[string]bool)
	var toCheck []pageURL
	for _, p := range pages {
		for _, u := range extractExternalURLs(p.bodyText) {
			if !urlAllowedForProbe(u, ignore, allow, cfg.Client != nil) {
				continue
			}
			if ent, ok := cache.Entries[u]; ok && now.Sub(ent.CheckedAt) < ttl {
				continue
			}
			if seenURLs[u] {
				continue
			}
			seenURLs[u] = true
			toCheck = append(toCheck, pageURL{path: p.path, url: u})
			if len(toCheck) >= maxChecks {
				break
			}
		}
		if len(toCheck) >= maxChecks {
			break
		}
	}

	client := cfg.Client
	if client == nil {
		client = newLinkHTTPClient(cfg.timeout())
	}

	maxConcurrent := cfg.maxConcurrent()
	requestDelay := cfg.requestDelay()

	var (
		mu       sync.Mutex
		sem      = make(chan struct{}, maxConcurrent)
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

		if requestDelay > 0 {
			select {
			case <-time.After(requestDelay):
			case <-ctx.Done():
				return
			}
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
			if !urlAllowedForProbe(u, ignore, allow, cfg.Client != nil) {
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
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return fmt.Errorf("redirect to disallowed scheme: %s", req.URL.Scheme)
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
