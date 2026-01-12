package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Batch operation thresholds
const (
	// BatchConfirmThreshold is the number of items above which we prompt for confirmation
	BatchConfirmThreshold = 10
)

// Global batch flags (set in root.go init)
var (
	dryRun bool
	yesAll bool
)

// isDryRun returns true if --dry-run flag is set
func isDryRun() bool {
	return dryRun
}

// isYesAll returns true if --yes flag is set
func isYesAll() bool {
	return yesAll
}

// confirmBatch prompts for confirmation if the count exceeds threshold
// Returns true if the operation should proceed, false otherwise
// If --dry-run is set, prints preview and returns false
// If --yes is set, skips confirmation and returns true
func confirmBatch(action string, count int, alwaysConfirm bool) bool {
	// Dry run mode - just preview
	if dryRun {
		fmt.Printf("[DRY RUN] Would %s %d item(s)\n", action, count)
		return false
	}

	// Check if confirmation needed
	needsConfirm := alwaysConfirm || count > BatchConfirmThreshold

	if !needsConfirm {
		return true
	}

	// Skip confirmation if --yes flag is set
	if yesAll {
		return true
	}

	// Prompt for confirmation
	fmt.Printf("%s %d item(s)? [y/N] ", strings.Title(action), count)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// printBatchResult prints a summary of a batch operation
func printBatchResult(action string, succeeded, failed []string) {
	if len(succeeded) > 0 {
		fmt.Printf("%s %d item(s)\n", strings.Title(action), len(succeeded))
	}
	if len(failed) > 0 {
		fmt.Printf("Failed to %s %d item(s)\n", action, len(failed))
		if len(failed) <= 5 {
			for _, id := range failed {
				fmt.Printf("  - %s\n", id)
			}
		}
	}
}
