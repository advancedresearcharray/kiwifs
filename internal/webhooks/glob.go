package webhooks

import (
	"path/filepath"
	"strings"
)

func matchGlob(pattern, path string) bool {
	if pattern == "" {
		return false
	}
	if pattern == "**" {
		return true
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(path, prefix+"/") || path == prefix
	}
	matched, _ := filepath.Match(pattern, path)
	return matched
}
