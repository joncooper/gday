package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/joncooper/gday/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
)

// Scopes required for Gmail and Calendar access
var Scopes = []string{
	gmail.GmailReadonlyScope,
	gmail.GmailSendScope,
	gmail.GmailModifyScope,
	calendar.CalendarReadonlyScope,
	calendar.CalendarEventsScope,
}

// Google's device authorization endpoint
const deviceAuthURL = "https://oauth2.googleapis.com/device/code"
const tokenURL = "https://oauth2.googleapis.com/token"

// DeviceAuthResponse represents the response from device authorization request
type DeviceAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// GetClient returns an authenticated HTTP client
func GetClient(ctx context.Context) (*http.Client, error) {
	cfg, err := getOAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth config: %w", err)
	}

	token, err := getToken(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return cfg.Client(ctx, token), nil
}

// getOAuthConfig returns the OAuth2 configuration
func getOAuthConfig() (*oauth2.Config, error) {
	credBytes, err := config.ReadCredentials()
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w\n\nRun 'gday auth setup' to configure credentials", err)
	}

	cfg, err := google.ConfigFromJSON(credBytes, Scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	return cfg, nil
}

// getToken retrieves a token from cache or initiates OAuth flow
func getToken(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	tokenBytes, err := config.ReadToken()
	if err == nil {
		var token oauth2.Token
		if err := json.Unmarshal(tokenBytes, &token); err == nil {
			// Check if token is still valid or can be refreshed
			if token.Valid() {
				return &token, nil
			}
			// Try to refresh
			tokenSource := cfg.TokenSource(ctx, &token)
			newToken, err := tokenSource.Token()
			if err == nil {
				config.SaveToken(newToken)
				return newToken, nil
			}
		}
	}

	return nil, fmt.Errorf("not authenticated. Run 'gday auth login' to authenticate")
}

// Login performs the OAuth2 login flow (browser-based)
func Login(ctx context.Context) error {
	cfg, err := getOAuthConfig()
	if err != nil {
		return err
	}

	// Start local server for OAuth callback
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server := &http.Server{Addr: ":8089"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code in callback")
			fmt.Fprintf(w, "<html><body><h1>Error</h1><p>No authorization code received.</p></body></html>")
			return
		}
		codeChan <- code
		fmt.Fprintf(w, "<html><body><h1>Success!</h1><p>You can close this window and return to the terminal.</p></body></html>")
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Generate auth URL
	cfg.RedirectURL = "http://localhost:8089/callback"
	authURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	fmt.Println("\nOpening browser for Google authentication...")
	fmt.Println("\nIf the browser doesn't open, visit this URL:")
	fmt.Printf("\n  %s\n\n", authURL)

	// Try to open browser
	openBrowser(authURL)

	// Wait for callback
	var code string
	select {
	case code = <-codeChan:
	case err := <-errChan:
		server.Shutdown(ctx)
		return fmt.Errorf("OAuth callback error: %w", err)
	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return fmt.Errorf("authentication timeout")
	}

	server.Shutdown(ctx)

	// Exchange code for token
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	// Save token
	if err := config.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("Authentication successful!")
	return nil
}

// LoginDevice performs the OAuth2 device flow (for headless/SSH environments)
func LoginDevice(ctx context.Context) error {
	cfg, err := getOAuthConfig()
	if err != nil {
		return err
	}

	// Request device code
	deviceAuth, err := requestDeviceCode(cfg.ClientID)
	if err != nil {
		return fmt.Errorf("failed to get device code: %w", err)
	}

	// Display instructions to user
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("  Device Authentication")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	fmt.Println("  1. Open this URL in any browser:")
	fmt.Printf("\n     %s\n\n", deviceAuth.VerificationURL)
	fmt.Println("  2. Enter this code:")
	fmt.Printf("\n     %s\n\n", deviceAuth.UserCode)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	fmt.Println("Waiting for authorization...")

	// Poll for token
	token, err := pollForToken(ctx, cfg.ClientID, cfg.ClientSecret, deviceAuth)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	// Save token
	if err := config.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("\nAuthentication successful!")
	return nil
}

// requestDeviceCode requests a device code from Google
func requestDeviceCode(clientID string) (*DeviceAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", strings.Join(Scopes, " "))

	resp, err := http.PostForm(deviceAuthURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device auth request failed: %s", string(body))
	}

	var deviceAuth DeviceAuthResponse
	if err := json.Unmarshal(body, &deviceAuth); err != nil {
		return nil, err
	}

	return &deviceAuth, nil
}

// pollForToken polls the token endpoint until authorization is complete
func pollForToken(ctx context.Context, clientID, clientSecret string, deviceAuth *DeviceAuthResponse) (*oauth2.Token, error) {
	interval := time.Duration(deviceAuth.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}

	deadline := time.Now().Add(time.Duration(deviceAuth.ExpiresIn) * time.Second)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
			token, err := requestToken(clientID, clientSecret, deviceAuth.DeviceCode)
			if err == nil {
				return token, nil
			}

			// Check for specific errors
			if strings.Contains(err.Error(), "authorization_pending") {
				// Still waiting for user, continue polling
				continue
			}
			if strings.Contains(err.Error(), "slow_down") {
				// Increase polling interval
				interval += 5 * time.Second
				continue
			}
			if strings.Contains(err.Error(), "access_denied") {
				return nil, fmt.Errorf("user denied access")
			}
			if strings.Contains(err.Error(), "expired_token") {
				return nil, fmt.Errorf("device code expired, please try again")
			}

			// Other error, return it
			return nil, err
		}
	}

	return nil, fmt.Errorf("authorization timed out")
}

// requestToken exchanges the device code for an access token
func requestToken(clientID, clientSecret, deviceCode string) (*oauth2.Token, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for error response
	var errorResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
		return nil, fmt.Errorf("%s: %s", errorResp.Error, errorResp.ErrorDescription)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed: %s", string(body))
	}

	// Parse successful response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return token, nil
}

// Logout removes the cached token
func Logout() error {
	if err := config.DeleteToken(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	fmt.Println("Logged out successfully")
	return nil
}

// Status prints the current authentication status
func Status() {
	if !config.CredentialsExist() {
		fmt.Println("Status: Not configured")
		fmt.Println("\nRun 'gday auth setup' to configure OAuth credentials")
		return
	}

	if !config.TokenExists() {
		fmt.Println("Status: Credentials configured, not logged in")
		fmt.Println("\nRun 'gday auth login' to authenticate")
		return
	}

	// Try to verify token
	ctx := context.Background()
	client, err := GetClient(ctx)
	if err != nil {
		fmt.Println("Status: Token expired or invalid")
		fmt.Println("\nRun 'gday auth login' to re-authenticate")
		return
	}

	// Quick check with Gmail API
	srv, err := gmail.New(client)
	if err != nil {
		fmt.Println("Status: Error creating Gmail client")
		return
	}

	profile, err := srv.Users.GetProfile("me").Do()
	if err != nil {
		fmt.Println("Status: Token invalid")
		fmt.Println("\nRun 'gday auth login' to re-authenticate")
		return
	}

	fmt.Println("Status: Authenticated")
	fmt.Printf("Email: %s\n", profile.EmailAddress)
}

// openBrowser attempts to open the URL in the default browser
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // Linux and others
		// Try common browser openers in order of preference
		browsers := []string{"xdg-open", "sensible-browser", "x-www-browser", "firefox", "chromium", "google-chrome"}
		for _, browser := range browsers {
			if path, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(path, url)
				break
			}
		}
	}

	if cmd != nil {
		// Run in background, ignore errors (best effort)
		cmd.Start()
	}
}
