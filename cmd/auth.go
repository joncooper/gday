package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joncooper/gday/internal/auth"
	"github.com/joncooper/gday/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Google authentication",
	Long:  `Commands for managing Google OAuth2 authentication.`,
}

var authSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure OAuth credentials",
	Long: `Configure OAuth credentials for gday.

You need to create OAuth2 credentials in Google Cloud Console:

1. Go to https://console.cloud.google.com/apis/credentials
2. Create a new project (or select existing)
3. Enable the Gmail API and Google Calendar API
4. Create OAuth 2.0 credentials (Desktop application type)
5. Download the credentials JSON file

Then run this command and paste the contents of the credentials file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("OAuth2 Credentials Setup")
		fmt.Println("========================")
		fmt.Println()
		fmt.Println("You need OAuth2 credentials from Google Cloud Console.")
		fmt.Println()
		fmt.Println("Quick setup steps:")
		fmt.Println("  1. Go to: https://console.cloud.google.com/apis/credentials")
		fmt.Println("  2. Create a project (or select existing)")
		fmt.Println("  3. Enable Gmail API and Google Calendar API")
		fmt.Println("  4. Configure OAuth consent screen (External, add your email as test user)")
		fmt.Println("  5. Create OAuth 2.0 Client ID (Desktop application)")
		fmt.Println("  6. Download the JSON file")
		fmt.Println()
		fmt.Println("Paste the contents of the credentials JSON file below,")
		fmt.Println("then press Enter twice (empty line) to finish:")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		var lines []string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\n\r")
			if line == "" && len(lines) > 0 {
				break
			}
			lines = append(lines, line)
		}

		jsonData := strings.Join(lines, "\n")
		if jsonData == "" {
			fmt.Println("Error: No credentials provided")
			os.Exit(1)
		}

		// Validate it looks like JSON
		if !strings.Contains(jsonData, "client_id") || !strings.Contains(jsonData, "client_secret") {
			fmt.Println("Error: Invalid credentials format. Must contain client_id and client_secret.")
			os.Exit(1)
		}

		if err := config.SaveCredentials([]byte(jsonData)); err != nil {
			fmt.Printf("Error saving credentials: %v\n", err)
			os.Exit(1)
		}

		configDir, _ := config.GetConfigDir()
		fmt.Printf("\nCredentials saved to %s/credentials.json\n", configDir)
		fmt.Println("\nNext, run 'gday auth login' to authenticate with Google.")
	},
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with Google account",
	Long: `Authenticate with Google using OAuth2 flow.

By default, opens a browser for authentication. Use --device for
headless environments (SSH, containers) where no browser is available.

Examples:
  gday auth login           # Browser-based authentication
  gday auth login --device  # Device flow for headless environments`,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.CredentialsExist() {
			fmt.Println("Error: OAuth credentials not configured")
			fmt.Println("\nRun 'gday auth setup' first to configure credentials")
			os.Exit(1)
		}

		ctx := context.Background()
		device, _ := cmd.Flags().GetBool("device")

		var err error
		if device {
			err = auth.LoginDevice(ctx)
		} else {
			err = auth.Login(ctx)
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear cached tokens",
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.Logout(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Run: func(cmd *cobra.Command, args []string) {
		auth.Status()
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authSetupCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)

	// Login flags
	authLoginCmd.Flags().Bool("device", false, "Use device flow for headless environments (SSH, containers)")
}
