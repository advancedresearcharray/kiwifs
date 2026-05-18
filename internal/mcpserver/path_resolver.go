package mcpserver

import (
	"fmt"
	"net/url"
	pathpkg "path"
	"strings"
)

type mcpPathMode int

const (
	mcpPathReadOnly mcpPathMode = iota
	mcpPathMutation
)

type resolvedMCPPath struct {
	Path    string
	Reasons []string
}

func resolveMCPPath(input string, mode mcpPathMode) (resolvedMCPPath, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return resolvedMCPPath{}, fmt.Errorf("path is required")
	}

	reasons := []string{}
	if u, err := url.Parse(input); err == nil && u.Scheme != "" && u.Host != "" {
		if mode != mcpPathReadOnly {
			return resolvedMCPPath{}, fmt.Errorf("page URLs are only accepted for read-only MCP tools; pass the decoded logical KiwiFS path explicitly")
		}
		if !strings.HasPrefix(u.Path, "/page/") {
			return resolvedMCPPath{}, fmt.Errorf("unsupported URL path %q; expected /page/<path>", u.Path)
		}
		input = strings.TrimPrefix(u.Path, "/page/")
		reasons = append(reasons, "page URL converted to KiwiFS path")
	}

	if strings.Contains(input, "%") {
		if mode != mcpPathReadOnly {
			return resolvedMCPPath{}, fmt.Errorf("percent-encoded paths are only accepted for read-only MCP tools; pass the decoded logical KiwiFS path explicitly")
		}
		decoded, err := url.PathUnescape(input)
		if err != nil {
			return resolvedMCPPath{}, fmt.Errorf("invalid percent-encoded path: %w", err)
		}
		input = decoded
		reasons = append(reasons, "encoded path decoded")
	}

	if input == "/" {
		return resolvedMCPPath{Path: "", Reasons: append(reasons, "root path normalized")}, nil
	}
	if strings.HasPrefix(input, "/") {
		return resolvedMCPPath{}, fmt.Errorf("absolute paths are not allowed; pass a relative KiwiFS path")
	}
	if strings.Contains(input, "\\") {
		return resolvedMCPPath{}, fmt.Errorf("backslash path separators are not allowed; use '/' in KiwiFS paths")
	}

	if err := validateLogicalMCPPath(input); err != nil {
		return resolvedMCPPath{}, err
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "path allowed unchanged")
	}
	return resolvedMCPPath{Path: input, Reasons: reasons}, nil
}

func validateLogicalMCPPath(p string) error {
	if p == "" {
		return fmt.Errorf("path is required")
	}
	if pathpkg.IsAbs(p) {
		return fmt.Errorf("absolute paths are not allowed; pass a relative KiwiFS path")
	}
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if seg == "" && i == len(parts)-1 {
			continue
		}
		if seg == "" || seg == "." || seg == ".." {
			return fmt.Errorf("unsafe path segment %q", seg)
		}
	}
	return nil
}

func readOnlyPathArg(args map[string]any, key string) (string, error) {
	raw, _ := args[key].(string)
	resolved, err := resolveMCPPath(raw, mcpPathReadOnly)
	if err != nil {
		return "", err
	}
	return resolved.Path, nil
}

func optionalReadOnlyPathArg(args map[string]any, key string) (string, error) {
	raw, _ := args[key].(string)
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	return readOnlyPathArg(args, key)
}

func mutationPathArg(args map[string]any, key string) (string, error) {
	raw, _ := args[key].(string)
	resolved, err := resolveMCPPath(raw, mcpPathMutation)
	if err != nil {
		return "", err
	}
	return resolved.Path, nil
}
