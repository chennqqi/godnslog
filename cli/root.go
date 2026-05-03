package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiURL       string
	apiKey       string
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "godnslog",
	Short: "GODNSLOG 2.0 CLI - OAST interaction verification tool",
	Long: `GODNSLOG 2.0 CLI is a command-line tool for Out-of-Band Application Security Testing (OAST).
It allows you to create cases, generate payloads, poll for interactions, and export reports.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if apiURL == "" {
			apiURL = os.Getenv("GODNSLOG_API_URL")
		}
		if apiKey == "" {
			apiKey = os.Getenv("GODNSLOG_API_KEY")
		}
		if apiURL == "" {
			fmt.Fprintln(os.Stderr, "Error: API URL is required. Set GODNSLOG_API_URL environment variable or use --api-url flag")
			os.Exit(1)
		}
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: API Key is required. Set GODNSLOG_API_KEY environment variable or use --api-key flag")
			os.Exit(1)
		}
	},
}

func Execute() error {
	rootCmd.AddCommand(caseCmd)
	rootCmd.AddCommand(payloadCmd)
	rootCmd.AddCommand(interactionCmd)
	rootCmd.AddCommand(reportCmd)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "API URL (default: GODNSLOG_API_URL)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API Key (default: GODNSLOG_API_KEY)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json, yaml, markdown)")
}
