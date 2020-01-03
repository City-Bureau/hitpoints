package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "hitpoints",
	Short: "Minimal tool for counting page hits on embedded content",
}

// Execute runs the Cobra command
func Execute() error {
	return rootCmd.Execute()
}
