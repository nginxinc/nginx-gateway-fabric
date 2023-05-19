package main

import (
	"fmt"
	"os"
)

var (
	// Set during go build
	version string
	commit  string
	date    string
)

func main() {
	rootCmd := createRootCommand()

	rootCmd.AddCommand(
		createControlPlaneCommand(),
		createProvisionerCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
