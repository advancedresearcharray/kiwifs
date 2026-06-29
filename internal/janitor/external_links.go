package janitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kiwifs/kiwifs/internal/markdown"
)

const (
	IssueExternalLinkRot = "external-link-rot"
	externalLinkRuleName = "external-link-rot"

	defaultExternalLinkTimeout  = 5 * time.Second
	defaultExternalLinkCacheTTL = 24 * time.Hour
	defaultExternalLinkDelay    = 100 * time.Millisecond
	defaultExternalLinkWorkers  = 10
	linkCheckerUserAgent        = "KiwiFS-LinkChecker/1.0"
)

var defaultExternalLinkIgnore = []string{"localhost", "127.0.0.1", "example.com"}

// ExternalLinkFinding is one broken or unreachable external URL in a page.
type ExternalLinkFinding struct {
	Path   string `json:"path"`
	URL    string `json:"url"`
	Status int    `json:"status"`
	Rule   string `json:"rule"`
}

// ExternalLinkCheckConfig controls outbound link verification.
type ExternalLinkCheckConfig struct {
	Enabled    bool
	Timeout    time.Duration
	Ignore     []string
	CacheTTL   time.Duration
	CacheDir   string
	HTTPClient *http.Client
	Now        func() time.Time
}

func (c ExternalLinkCheckConfig) enabled() bool {
	return c.Enabled
}

func (c ExternalLinkCheckConfig) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return defaultExternalLinkTimeout
}

func (c ExternalLinkCheckConfig) cacheTTL() time.Duration {
	if c.CacheTTL > 0 {
		return c.CacheTTL
	}
	return defaultExternalLinkCacheTTL
}

func (c ExternalLinkCheckConfig) ignoreHosts() []string {
	if c.Ignore != nil {
		return c.Ignore
	}
	return defaultExternalLinkIgnore
}

func (c ExternalLinkCheckConfig) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now().UTC()
}

type linkCheckCache struct {
	UpdatedAt string                       `json:"updated_at"`
	Entries   map[string]linkCheckCacheEntry `json:"entries"`
}

type linkCheckCacheEntry struct {
	Status    int    `json:"status"`
	CheckedAt string `json:"checked_at"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

func linkCheckCachePath(root string) string {
	return filepath.Join(root, ".kiwi", "cache", "link-check.json")
}

func loadLinkCheckCache(path string) (*linkCheckCache, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &linkCheckCache{Entries: map[string]linkCheckCacheEntry{}}, nil
		}
		return nil, err
	}
	var cache linkCheckCache
	if err := json.Unmarshal(raw, &cache); err != nil {
		return nil, err
	}
	if cache.Entries == nil {
		cache.Entries = map[string]linkCheckCacheEntry{}
	}
	return &cache, nil
}

func saveLinkCheckCache(path string, cache *linkCheckCache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func shouldIgnoreExternalURL(rawURL string, ignore []string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return true
	}
	host := strings.ToLower(u.Hostname())
	for _, ig := range ignore {
		ig = strings.ToLower(strings.TrimSpace(ig))
		if ig == "" {
			continue
		}
		if host == ig || strings.HasSuffix(host, "."+ig) {
			return true
		}
	}
	return false
}

func externalURLBroken(status int) bool {
	return status >= 400
}

type pageExternalRef struct {
	path string
	url  string
}

type externalURLTarget struct {
	url string
}

func collectExternalURLRefs(pages []pageInfo, ignore []string) []pageExternalRef {
	var refs []pageExternalRef
	for _, p := range pages {
		for _, u := range markdown.ExtractExternalURLs(p.bodyText) {
			if shouldIgnoreExternalURL(u, ignore) {
				continue
			}
			refs = append(refs, pageExternalRef{path: p.path, url: u})
		}
	}
	return refs
}

func (s *Scanner) checkExternalLinks(ctx context.Context, pages []pageInfo) ([]ExternalLinkFinding, error) {
	if s.externalLinkCheck == nil || !s.externalLinkCheck.enabled() {
		return nil, nil
	}
	cfg := *s.externalLinkCheck
	cachePath := linkCheckCachePath(cfg.CacheDir)
	cache, err := loadLinkCheckCache(cachePath)
	if err != nil {
		return nil, fmt.Errorf("load link-check cache: %w", err)
	}

	refs := collectExternalURLRefs(pages, cfg.ignoreHosts())
	if len(refs) == 0 {
		return nil, nil
	}

	checker := newLinkChecker(cfg)
	now := cfg.now()
	ttl := cfg.cacheTTL()

	uniqueURLs := make(map[string]struct{})
	for _, ref := range refs {
		uniqueURLs[ref.url] = struct{}{}
	}

	var toCheck []externalURLTarget
	for u := range uniqueURLs {
		if ent, ok := cache.Entries[u]; ok {
			checkedAt, parseErr := time.Parse(time.RFC3339, ent.CheckedAt)
			if parseErr == nil && now.Sub(checkedAt) < ttl {
				continue
			}
		}
		toCheck = append(toCheck, externalURLTarget{url: u})
	}

	if len(toCheck) > 0 {
		results := checker.checkMany(ctx, toCheck)
		for u, res := range results {
			ent := linkCheckCacheEntry{
				Status:    res.status,
				CheckedAt: now.Format(time.RFC3339),
				OK:        !externalURLBroken(res.status) && res.err == nil,
			}
			if res.err != nil {
				ent.Error = res.err.Error()
				ent.OK = false
			}
			cache.Entries[u] = ent
		}
		cache.UpdatedAt = now.Format(time.RFC3339)
		if err := saveLinkCheckCache(cachePath, cache); err != nil {
			return nil, fmt.Errorf("save link-check cache: %w", err)
		}
	}

	var findings []ExternalLinkFinding
	for _, ref := range refs {
		ent, ok := cache.Entries[ref.url]
		if !ok {
			continue
		}
		if ent.OK {
			continue
		}
		status := ent.Status
		if status == 0 && ent.Error != "" {
			status = 0
		}
		findings = append(findings, ExternalLinkFinding{
			Path:   ref.path,
			URL:    ref.url,
			Status: status,
			Rule:   externalLinkRuleName,
		})
	}
	return findings, nil
}

type linkCheckResult struct {
	status int
	err    error
}

type linkChecker struct {
	client *http.Client
	delay  time.Duration
	workers int
}

func newLinkChecker(cfg ExternalLinkCheckConfig) *linkChecker {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: cfg.timeout(),
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("stopped after 3 redirects")
				}
				return nil
			},
		}
	}
	return &linkChecker{
		client:  client,
		delay:   defaultExternalLinkDelay,
		workers: defaultExternalLinkWorkers,
	}
}

func (lc *linkChecker) checkMany(ctx context.Context, targets []externalURLTarget) map[string]linkCheckResult {
	out := make(map[string]linkCheckResult, len(targets))
	if len(targets) == 0 {
		return out
	}

	sem := make(chan struct{}, lc.workers)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, target := range targets {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			select {
			case <-time.After(lc.delay):
			case <-ctx.Done():
				return
			}

			status, err := lc.checkURL(ctx, u)
			mu.Lock()
			out[u] = linkCheckResult{status: status, err: err}
			mu.Unlock()
		}(target.url)
	}
	wg.Wait()
	return out
}

func (lc *linkChecker) checkURL(ctx context.Context, rawURL string) (int, error) {
	status, err := lc.doRequest(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return 0, err
	}
	if status == http.StatusMethodNotAllowed || status == http.StatusNotImplemented {
		return lc.doRequest(ctx, http.MethodGet, rawURL, map[string]string{"Range": "bytes=0-0"})
	}
	return status, nil
}

func (lc *linkChecker) doRequest(ctx context.Context, method, rawURL string, extraHeaders map[string]string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", linkCheckerUserAgent)
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := lc.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func externalLinkFindingsToIssues(findings []ExternalLinkFinding) []Issue {
	issues := make([]Issue, 0, len(findings))
	for _, f := range findings {
		msg := fmt.Sprintf("external URL returned HTTP %d", f.Status)
		if f.Status == 0 {
			msg = "external URL check failed"
		}
		issues = append(issues, Issue{
			Kind:     IssueExternalLinkRot,
			Path:     f.Path,
			Message:  msg,
			Related:  []string{f.URL},
			Severity: "warning",
		})
	}
	return issues
}

// WithExternalLinkCheck enables outbound URL verification during scans.
func WithExternalLinkCheck(cfg ExternalLinkCheckConfig) Option {
	return func(s *Scanner) {
		if cfg.Enabled {
			cp := cfg
			if cp.CacheDir == "" {
				cp.CacheDir = s.root
			}
			s.externalLinkCheck = &cp
		}
	}
}

// ExternalLinkCheckFromConfig builds scanner options from config.toml fields.
func ExternalLinkCheckFromConfig(enabled bool, timeoutRaw string, ignore []string, root string) Option {
	if !enabled {
		return func(*Scanner) {}
	}
	timeout := defaultExternalLinkTimeout
	if strings.TrimSpace(timeoutRaw) != "" {
		if d, err := time.ParseDuration(timeoutRaw); err == nil && d > 0 {
			timeout = d
		}
	}
	return WithExternalLinkCheck(ExternalLinkCheckConfig{
		Enabled:  true,
		Timeout:  timeout,
		Ignore:   ignore,
		CacheDir: root,
	})
}

// OptionsFromExternalLinkCheck returns nil when disabled.
func OptionsFromExternalLinkCheck(enabled bool, timeoutRaw string, ignore []string, root string) []Option {
	if !enabled {
		return nil
	}
	return []Option{ExternalLinkCheckFromConfig(enabled, timeoutRaw, ignore, root)}
}
