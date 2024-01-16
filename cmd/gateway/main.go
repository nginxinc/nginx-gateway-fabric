package main

import (
	"fmt"
	"os"
)

// Set during go build
var (
	version string
	commit  string
	date    string

	// telemetryReportPeriod is the period at which telemetry reports are sent.
	telemetryReportPeriod string
)

func main() {
	rootCmd := createRootCommand()

	rootCmd.AddCommand(
		createStaticModeCommand(),
		createProvisionerModeCommand(),
		createSleepCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
