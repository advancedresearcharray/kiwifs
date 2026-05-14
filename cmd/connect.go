package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <workspace-slug>",
	Short: "Generate MCP config for connecting an AI client to a KiwiFS workspace",
	Long: `Generate a ready-to-paste MCP server configuration for your AI tool.

By default, prints the JSON config to stdout with your real API key embedded.
Use --write to automatically write it to a supported client's config file.

Supported --write targets: cursor, claude-code, windsurf, claude-desktop.
All other clients: copy the printed JSON into your tool's MCP config.`,
	Example: `  # Print config (works for any MCP client)
  kiwifs connect my-workspace --key kiwi_sk_abc123

  # Auto-write to a specific client's config file
  kiwifs connect my-workspace --key kiwi_sk_abc123 --write cursor

  # Auto-detect and write to all found clients
  kiwifs connect my-workspace --key kiwi_sk_abc123 --write auto

  # Project-level config (Cursor, Claude Code)
  kiwifs connect my-workspace --key kiwi_sk_abc123 --write cursor --project`,
	Args: cobra.ExactArgs(1),
	RunE: runConnect,
}

func init() {
	connectCmd.Flags().String("key", "", "API key (reads KIWI_API_KEY env if omitted)")
	connectCmd.Flags().String("write", "", "Write config to client file: cursor, claude-code, windsurf, claude-desktop, auto")
	connectCmd.Flags().String("host", "https://api.kiwifs.com", "KiwiFS Cloud API host")
	connectCmd.Flags().Bool("project", false, "Write to project-level config instead of global")
}

// ---------------------------------------------------------------------------
// Client registry
// ---------------------------------------------------------------------------

type mcpClient struct {
	name       string
	globalPath func() string
	projFile   string
	supportsURL bool
}

var clientRegistry = []mcpClient{
	{name: "cursor", globalPath: cursorPath, projFile: ".cursor/mcp.json", supportsURL: true},
	{name: "claude-code", globalPath: claudeCodePath, projFile: ".mcp.json", supportsURL: true},
	{name: "windsurf", globalPath: windsurfPath, projFile: "", supportsURL: true},
	{name: "claude-desktop", globalPath: claudeDesktopPath, projFile: "", supportsURL: false},
}

func cursorPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor", "mcp.json")
}

func claudeCodePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func windsurfPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
}

func claudeDesktopPath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json")
	default:
		return filepath.Join(home, ".config", "claude", "claude_desktop_config.json")
	}
}

func lookupClient(name string) *mcpClient {
	for i := range clientRegistry {
		if clientRegistry[i].name == name {
			return &clientRegistry[i]
		}
	}
	return nil
}

func detectInstalledClients() []*mcpClient {
	var found []*mcpClient
	for i := range clientRegistry {
		dir := filepath.Dir(clientRegistry[i].globalPath())
		if _, err := os.Stat(dir); err == nil {
			found = append(found, &clientRegistry[i])
		}
	}
	return found
}

// ---------------------------------------------------------------------------
// Config generation
// ---------------------------------------------------------------------------

func mcpEntryURL(slug, apiKey, host string, useEnvRef bool) map[string]interface{} {
	authValue := fmt.Sprintf("Bearer %s", apiKey)
	if useEnvRef {
		authValue = "Bearer ${env:KIWI_API_KEY}"
	}
	return map[string]interface{}{
		"url": fmt.Sprintf("%s/api/workspaces/%s/mcp", host, slug),
		"headers": map[string]string{
			"Authorization": authValue,
		},
	}
}

func mcpEntryStdio(slug, apiKey, host string, useEnvRef bool) map[string]interface{} {
	remoteURL := fmt.Sprintf("%s/api/workspaces/%s", host, slug)
	key := apiKey
	if useEnvRef {
		key = "${env:KIWI_API_KEY}"
	}
	return map[string]interface{}{
		"command": "npx",
		"args":    []string{"kiwifs", "mcp", "--remote", remoteURL, "--api-key", key},
	}
}

func fullConfig(slug, apiKey, host string, supportsURL, useEnvRef bool) map[string]interface{} {
	var entry map[string]interface{}
	if supportsURL {
		entry = mcpEntryURL(slug, apiKey, host, useEnvRef)
	} else {
		entry = mcpEntryStdio(slug, apiKey, host, useEnvRef)
	}
	return map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"kiwi": entry,
		},
	}
}

// ---------------------------------------------------------------------------
// File I/O
// ---------------------------------------------------------------------------

func readJSON(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return make(map[string]interface{}), nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}

func writeJSON(path string, data map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0644)
}

func mergeWriteConfig(client *mcpClient, slug, apiKey, host string, projectLevel bool) error {
	var path string
	if projectLevel && client.projFile != "" {
		path = client.projFile
	} else {
		path = client.globalPath()
	}

	existing, err := readJSON(path)
	if err != nil {
		return err
	}

	servers, _ := existing["mcpServers"].(map[string]interface{})
	if servers == nil {
		servers = make(map[string]interface{})
	}

	if client.supportsURL {
		servers["kiwi"] = mcpEntryURL(slug, apiKey, host, false)
	} else {
		servers["kiwi"] = mcpEntryStdio(slug, apiKey, host, false)
	}
	existing["mcpServers"] = servers

	if err := writeJSON(path, existing); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "  ✓ %s → %s\n", client.name, path)
	return nil
}

// ---------------------------------------------------------------------------
// Fetch workspace key using stored login session
// ---------------------------------------------------------------------------

func fetchWorkspaceKey(creds *storedCredentials, slug string) (string, error) {
	url := creds.Host + "/api/workspaces/" + slug + "/key"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("fetch key: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		APIKey string `json:"api_key"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result.APIKey, nil
}

// ---------------------------------------------------------------------------
// Main command
// ---------------------------------------------------------------------------

func runConnect(cmd *cobra.Command, args []string) error {
	slug := args[0]

	apiKey, _ := cmd.Flags().GetString("key")
	if apiKey == "" {
		apiKey = os.Getenv("KIWI_API_KEY")
	}

	host, _ := cmd.Flags().GetString("host")

	// If no explicit key, try fetching workspace key via stored login session
	if apiKey == "" {
		creds, err := loadCredentials()
		if err == nil && creds.AccessToken != "" {
			if host == "https://api.kiwifs.com" && creds.Host != "" {
				host = creds.Host
			}
			fetched, fetchErr := fetchWorkspaceKey(creds, slug)
			if fetchErr == nil && fetched != "" {
				apiKey = fetched
				fmt.Fprintf(os.Stderr, "Using API key from login session (%s)\n\n", creds.Email)
			}
		}
	}

	if apiKey == "" {
		return fmt.Errorf("API key required: use --key, set KIWI_API_KEY, or run 'kiwifs login' first\n\nFind your key at https://app.kiwifs.com (workspace settings)")
	}
	writeTarget, _ := cmd.Flags().GetString("write")
	projectLevel, _ := cmd.Flags().GetBool("project")

	// No --write flag: print to stdout (universal, works for any client)
	if writeTarget == "" {
		// Print env-var version (safe to commit/share)
		config := fullConfig(slug, apiKey, host, true, true)
		out, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(out))

		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Set the environment variable in your shell profile:")
		fmt.Fprintf(os.Stderr, "  export KIWI_API_KEY=%s\n", apiKey)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Paste the config above into your MCP client's config file.")
		fmt.Fprintln(os.Stderr, "Common locations:")
		fmt.Fprintln(os.Stderr, "  Cursor:        ~/.cursor/mcp.json")
		fmt.Fprintln(os.Stderr, "  Claude Code:   ~/.claude/settings.json (or .mcp.json in project root)")
		fmt.Fprintln(os.Stderr, "  VS Code:       .vscode/mcp.json")
		fmt.Fprintln(os.Stderr, "  Windsurf:      ~/.codeium/windsurf/mcp_config.json")
		fmt.Fprintln(os.Stderr, "  Zed:           ~/.config/zed/settings.json")
		fmt.Fprintln(os.Stderr, "  Gemini CLI:    ~/.gemini/settings.json")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "For Claude Desktop (stdio only), use:")
		stdio := fullConfig(slug, apiKey, host, false, true)
		stdioOut, _ := json.MarshalIndent(stdio, "", "  ")
		fmt.Fprintln(os.Stderr, string(stdioOut))
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Or use --write <client> to auto-write (embeds key directly):")
		fmt.Fprintln(os.Stderr, "  kiwifs connect "+slug+" --key <key> --write auto")
		return nil
	}

	// --write auto: detect and write to all found clients
	if writeTarget == "auto" {
		clients := detectInstalledClients()
		if len(clients) == 0 {
			return fmt.Errorf("no supported clients detected. Use --write <name> or paste the config manually")
		}
		fmt.Fprintf(os.Stderr, "Connecting workspace %q...\n\n", slug)
		var errs []string
		for _, c := range clients {
			if err := mergeWriteConfig(c, slug, apiKey, host, projectLevel); err != nil {
				errs = append(errs, fmt.Sprintf("  ✗ %s: %v", c.name, err))
			}
		}
		if len(errs) > 0 {
			fmt.Fprintln(os.Stderr)
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, e)
			}
		}
		fmt.Fprintf(os.Stderr, "\nDone. Restart your client(s) to activate.\n")
		return nil
	}

	// --write <specific-client>
	client := lookupClient(writeTarget)
	if client == nil {
		return fmt.Errorf("unknown client %q\nSupported: cursor, claude-code, windsurf, claude-desktop, auto", writeTarget)
	}
	fmt.Fprintf(os.Stderr, "Connecting workspace %q to %s...\n\n", slug, client.name)
	if err := mergeWriteConfig(client, slug, apiKey, host, projectLevel); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "\nDone. Restart %s to activate.\n", client.name)
	return nil
}
