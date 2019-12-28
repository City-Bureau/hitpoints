package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "hitpoints",
	Short: "A simple server for counting embedded page views",
	Long:  `...`,
}

// Execute runs the Cobra command
func Execute() error {
	return rootCmd.Execute()
}
