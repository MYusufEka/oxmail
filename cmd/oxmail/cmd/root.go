package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	apiURL     string
)

var rootCmd = &cobra.Command{
	Use:   "oxmail",
	Short: "Oxmail CLI — manage your mail server",
	Long:  "Oxmail is a CLI tool for managing domains, users, aliases, viewing health, tailing logs, and sending test emails.",
}

// Execute runs the root command.
func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	defaultAPI := os.Getenv("OXMAIL_API_URL")
	if defaultAPI == "" {
		defaultAPI = "http://localhost:8080"
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", defaultAPI, "API server URL (env: OXMAIL_API_URL)")
}
