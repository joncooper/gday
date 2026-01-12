package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// Default timeout for operations
const defaultTimeout = 2 * time.Minute

// Global flags
var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "gday",
	Short: "Gmail and Google Calendar CLI",
	Long: `G'day! A command-line interface for Gmail and Google Calendar.

Provides fast access to your email and calendar from the terminal,
designed to work seamlessly with Claude Code and other CLI tools.

First time setup:
  gday auth login    # Authenticate with Google

Gmail commands:
  gday mail list     # List recent emails
  gday mail read ID  # Read a specific email
  gday mail send     # Send an email
  gday mail search   # Search emails

Calendar commands:
  gday cal list      # List upcoming events
  gday cal create    # Create an event
  gday cal delete    # Delete an event

Use --json flag with any command for machine-readable output.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview changes without executing")
	rootCmd.PersistentFlags().BoolVar(&yesAll, "yes", false, "Skip confirmation prompts")
}

// Helper to print errors and exit
func exitError(msg string, args ...interface{}) {
	if jsonOutput {
		outputJSON(map[string]interface{}{
			"error": fmt.Sprintf(msg, args...),
		})
	} else {
		fmt.Fprintf(os.Stderr, "Error: "+msg+"\n", args...)
	}
	os.Exit(1)
}

// outputJSON prints data as JSON
func outputJSON(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

// isJSONOutput returns true if JSON output is enabled
func isJSONOutput() bool {
	return jsonOutput
}

// newContext creates a context that cancels on SIGINT/SIGTERM (Ctrl-C)
// and has a default timeout
func newContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(sigChan)
	}()

	return ctx, cancel
}
