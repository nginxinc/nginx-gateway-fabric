package main

import (
	"fmt"
	"os"
)

// Set during go build
var version string

func main() {
	rootCmd := createRootCommand()

	rootCmd.AddCommand(
		createStaticModeCommand(),
		createProvisionerModeCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
