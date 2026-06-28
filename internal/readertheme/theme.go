package readertheme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/kiwifs/kiwifs/internal/config"
)

// PageContext carries branding and theme CSS for published reader HTML.
type PageContext struct {
	PageTitle         string
	DocumentTitle     string
	FaviconURL        string
	FaviconType       string
	LogoURL           string
	BrandName         string
	HasCustomBranding bool
	ThemeCSS          string
}

// BrandingFromConfig builds reader branding fields from server UI config.
func BrandingFromConfig(b config.BrandingConfig, pageTitle string) PageContext {
	ctx := PageContext{
		PageTitle:     pageTitle,
		DocumentTitle: pageTitle,
		FaviconURL:    b.ResolvedFaviconURL(),
		FaviconType:   faviconMIME(b.ResolvedFaviconURL()),
		LogoURL:       b.ResolvedLogoURL(),
		BrandName:     b.ResolvedName(),
	}
	if b.Name != "" {
		ctx.DocumentTitle = b.Name + " | " + pageTitle
		ctx.HasCustomBranding = true
	} else if b.HasCustomLogo() || b.FaviconURL != "" {
		ctx.HasCustomBranding = true
	}
	return ctx
}

func faviconMIME(href string) string {
	lower := strings.ToLower(href)
	switch {
	case strings.HasSuffix(lower, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(lower, ".ico"):
		return "image/x-icon"
	default:
		return "image/png"
	}
}

// Cache loads and memoizes workspace theme.json by root path and file mtime.
type Cache struct {
	mu      sync.RWMutex
	root    string
	modTime time.Time
	theme   map[string]any
}

// Get returns parsed theme.json for root, or nil when missing or invalid.
func (c *Cache) Get(root string) map[string]any {
	path := filepath.Join(root, ".kiwi", "theme.json")
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.invalidate()
		}
		return nil
	}

	c.mu.RLock()
	if c.root == root && c.modTime.Equal(info.ModTime()) && c.theme != nil {
		theme := c.theme
		c.mu.RUnlock()
		return theme
	}
	c.mu.RUnlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var theme map[string]any
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil
	}

	c.mu.Lock()
	c.root = root
	c.modTime = info.ModTime()
	c.theme = theme
	c.mu.Unlock()
	return theme
}

func (c *Cache) invalidate() {
	c.mu.Lock()
	c.root = ""
	c.modTime = time.Time{}
	c.theme = nil
	c.mu.Unlock()
}

// BuildCSS generates :root overrides from theme.json. Public pages default to light
// mode when mode is unset. Returns empty string when theme is nil or has no tokens.
func BuildCSS(theme map[string]any) string {
	if len(theme) == 0 {
		return ""
	}

	mode := "light"
	if m, ok := theme["mode"].(string); ok {
		switch strings.ToLower(strings.TrimSpace(m)) {
		case "dark", "light", "system":
			mode = strings.ToLower(strings.TrimSpace(m))
		}
	}

	light := tokensFrom(theme["light"])
	dark := tokensFrom(theme["dark"])

	var parts []string
	switch mode {
	case "dark":
		tokens := dark
		if len(tokens) == 0 {
			tokens = light
		}
		if css := tokensToBlock(":root", tokens); css != "" {
			parts = append(parts, css)
		}
	case "system":
		if css := tokensToBlock(":root", light); css != "" {
			parts = append(parts, css)
		}
		if css := tokensToBlock(":root", dark); css != "" {
			parts = append(parts, "@media (prefers-color-scheme: dark) {\n"+css+"}\n")
		}
	default:
		if css := tokensToBlock(":root", light); css != "" {
			parts = append(parts, css)
		}
	}
	return strings.Join(parts, "")
}

func tokensFrom(v any) map[string]string {
	raw, ok := v.(map[string]any)
	if !ok || len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, val := range raw {
		s, ok := val.(string)
		if !ok || !safeCSSKey(k) || !safeCSSValue(s) {
			continue
		}
		out[k] = strings.TrimSpace(s)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func safeCSSValue(v string) bool {
	if v == "" || len(v) > 512 {
		return false
	}
	if strings.ContainsAny(v, "{}<>\n\r") {
		return false
	}
	return true
}

func safeCSSKey(k string) bool {
	if k == "" || len(k) > 64 {
		return false
	}
	if strings.ContainsAny(k, "{}<>\n\r ;:\"'") {
		return false
	}
	for i, r := range k {
		if i == 0 {
			if !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func tokensToBlock(selector string, tokens map[string]string) string {
	if len(tokens) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s {\n", selector)
	for k, v := range tokens {
		fmt.Fprintf(&b, "  --%s: %s;\n", k, v)
	}
	b.WriteString("}\n")
	return b.String()
}

// ApplyTheme merges workspace theme CSS into page context.
func ApplyTheme(ctx PageContext, theme map[string]any) PageContext {
	ctx.ThemeCSS = BuildCSS(theme)
	return ctx
}
