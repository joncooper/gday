package main

import (
	"os"

	"github.com/joncooper/gday/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
