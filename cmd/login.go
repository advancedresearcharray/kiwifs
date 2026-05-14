package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultCloudHost = "https://app.kiwifs.com"
	credFileName     = "credentials.json"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with KiwiFS Cloud via your browser",
	Long: `Log in to KiwiFS Cloud so the CLI can manage workspaces on your behalf.

Uses the OAuth 2.0 Device Authorization Flow: the CLI displays a code,
you confirm it in your browser, and the CLI receives a token.

After login, commands like 'kiwifs connect' work without --key.`,
	Example: `  kiwifs login
  kiwifs login --host https://self-hosted.example.com`,
	RunE: runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored KiwiFS Cloud credentials",
	RunE:  runLogout,
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently authenticated user",
	RunE:  runWhoami,
}

func init() {
	loginCmd.Flags().String("host", defaultCloudHost, "KiwiFS Cloud host")
	loginCmd.Flags().String("client-id", "", "WorkOS client ID (auto-detected from cloud host if omitted)")
}

// ---------------------------------------------------------------------------
// Credential storage (~/.kiwifs/credentials.json)
// ---------------------------------------------------------------------------

type storedCredentials struct {
	Host         string `json:"host"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	Email        string `json:"email,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
}

func credDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kiwifs")
}

func credPath() string {
	return filepath.Join(credDir(), credFileName)
}

func loadCredentials() (*storedCredentials, error) {
	data, err := os.ReadFile(credPath())
	if err != nil {
		return nil, err
	}
	var creds storedCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

func saveCredentials(creds *storedCredentials) error {
	if err := os.MkdirAll(credDir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(credPath(), append(data, '\n'), 0600)
}

func clearCredentials() error {
	if err := os.Remove(credPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------
// Device Authorization Flow
// ---------------------------------------------------------------------------

type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type tokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
	Error        string      `json:"error"`
}

func requestDeviceAuth(cloudHost, clientID string) (*deviceAuthResponse, error) {
	endpoint := cloudHost + "/api/auth/device/authorize"

	resp, err := http.PostForm(endpoint, url.Values{
		"client_id": {clientID},
	})
	if err != nil {
		return nil, fmt.Errorf("request device authorization: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("device authorization failed (%d): %s", resp.StatusCode, string(body))
	}

	var dar deviceAuthResponse
	if err := json.Unmarshal(body, &dar); err != nil {
		return nil, fmt.Errorf("parse device authorization response: %w", err)
	}
	if dar.Interval == 0 {
		dar.Interval = 5
	}
	if dar.ExpiresIn == 0 {
		dar.ExpiresIn = 300
	}
	return &dar, nil
}

func pollForTokens(cloudHost, clientID, deviceCode string, interval, expiresIn int) (*tokenResponse, error) {
	endpoint := cloudHost + "/api/auth/device/token"
	deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)

	for time.Now().Before(deadline) {
		resp, err := http.PostForm(endpoint, url.Values{
			"client_id":  {clientID},
			"device_code": {deviceCode},
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		})
		if err != nil {
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var tr tokenResponse
		if err := json.Unmarshal(body, &tr); err != nil {
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		if resp.StatusCode == 200 && tr.AccessToken != "" {
			return &tr, nil
		}

		switch tr.Error {
		case "authorization_pending":
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		case "slow_down":
			interval += 1
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		case "access_denied", "expired_token":
			return nil, fmt.Errorf("authorization %s", tr.Error)
		default:
			if tr.Error != "" {
				return nil, fmt.Errorf("authorization error: %s", tr.Error)
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}

	return nil, fmt.Errorf("authorization timed out after %ds", expiresIn)
}

func fetchClientID(cloudHost string) (string, error) {
	resp, err := http.Get(cloudHost + "/api/auth/client-info")
	if err != nil {
		return "", fmt.Errorf("fetch client info: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("client-info endpoint returned %d", resp.StatusCode)
	}
	var info struct {
		ClientID string `json:"client_id"`
	}
	if err := json.Unmarshal(body, &info); err != nil || info.ClientID == "" {
		return "", fmt.Errorf("could not determine client ID from %s", cloudHost)
	}
	return info.ClientID, nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

func runLogin(cmd *cobra.Command, args []string) error {
	cloudHost, _ := cmd.Flags().GetString("host")
	cloudHost = strings.TrimRight(cloudHost, "/")
	clientID, _ := cmd.Flags().GetString("client-id")

	if clientID == "" {
		var err error
		clientID, err = fetchClientID(cloudHost)
		if err != nil {
			return fmt.Errorf("could not auto-detect client ID: %w\n\nProvide it with --client-id", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Authenticating with %s...\n\n", cloudHost)

	dar, err := requestDeviceAuth(cloudHost, clientID)
	if err != nil {
		return err
	}

	verifyURL := dar.VerificationURIComplete
	if verifyURL == "" {
		verifyURL = dar.VerificationURI
	}

	fmt.Fprintf(os.Stderr, "  Your code: %s\n\n", dar.UserCode)
	fmt.Fprintf(os.Stderr, "  Opening: %s\n", verifyURL)
	fmt.Fprintf(os.Stderr, "  (If the browser doesn't open, visit the URL above and enter the code)\n\n")

	openBrowser(verifyURL)

	fmt.Fprintf(os.Stderr, "Waiting for confirmation...")

	tr, err := pollForTokens(cloudHost, clientID, dar.DeviceCode, dar.Interval, dar.ExpiresIn)
	if err != nil {
		fmt.Fprintln(os.Stderr)
		return err
	}

	// Extract user info
	email := ""
	displayName := ""
	userID := ""
	if userMap, ok := tr.User.(map[string]interface{}); ok {
		if e, ok := userMap["email"].(string); ok {
			email = e
		}
		if fn, ok := userMap["first_name"].(string); ok {
			ln, _ := userMap["last_name"].(string)
			displayName = strings.TrimSpace(fn + " " + ln)
		}
		if id, ok := userMap["id"].(string); ok {
			userID = id
		}
	}

	creds := &storedCredentials{
		Host:         cloudHost,
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		UserID:       userID,
		Email:        email,
		DisplayName:  displayName,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
	}

	if err := saveCredentials(creds); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	fmt.Fprintln(os.Stderr)
	if displayName != "" {
		fmt.Fprintf(os.Stderr, "Logged in as %s (%s)\n", displayName, email)
	} else if email != "" {
		fmt.Fprintf(os.Stderr, "Logged in as %s\n", email)
	} else {
		fmt.Fprintln(os.Stderr, "Logged in successfully")
	}
	fmt.Fprintf(os.Stderr, "Credentials saved to %s\n", credPath())

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	if err := clearCredentials(); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Logged out. Credentials removed.")
	return nil
}

func runWhoami(cmd *cobra.Command, args []string) error {
	creds, err := loadCredentials()
	if err != nil {
		return fmt.Errorf("not logged in. Run: kiwifs login")
	}
	if creds.DisplayName != "" {
		fmt.Printf("%s (%s)\n", creds.DisplayName, creds.Email)
	} else if creds.Email != "" {
		fmt.Println(creds.Email)
	} else {
		fmt.Printf("Authenticated against %s (user ID: %s)\n", creds.Host, creds.UserID)
	}
	return nil
}
