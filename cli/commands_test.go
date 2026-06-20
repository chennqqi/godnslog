package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// findSubcommand searches for a subcommand by name in a cobra command's commands slice.
func findSubcommand(t *testing.T, cmd *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, sub := range cmd.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	t.Fatalf("subcommand %q not found on %q", name, cmd.Name())
	return nil
}

// TestCaseSubcommands verifies that all required case subcommands are registered.
func TestCaseSubcommands(t *testing.T) {
	subs := []string{"create", "list", "get", "delete", "close"}
	for _, name := range subs {
		findSubcommand(t, caseCmd, name)
	}
}

// TestPayloadSubcommands verifies that all required payload subcommands are registered.
func TestPayloadSubcommands(t *testing.T) {
	subs := []string{"create", "list", "revoke", "preview"}
	for _, name := range subs {
		findSubcommand(t, payloadCmd, name)
	}
}

// TestInteractionSubcommands verifies that all required interaction subcommands are registered.
func TestInteractionSubcommands(t *testing.T) {
	subs := []string{"list", "poll"}
	for _, name := range subs {
		findSubcommand(t, interactionCmd, name)
	}
}

// TestReportSubcommands verifies that the report export subcommand is registered.
func TestReportSubcommands(t *testing.T) {
	findSubcommand(t, reportCmd, "export")
}

// TestScannerSubcommands verifies that all required scanner subcommands are registered.
func TestScannerSubcommands(t *testing.T) {
	subs := []string{"run", "list", "get"}
	for _, name := range subs {
		findSubcommand(t, scannerCmd, name)
	}
}

// TestAgentSubcommands verifies that all required agent subcommands are registered.
func TestAgentSubcommands(t *testing.T) {
	subs := []string{"list", "get", "create", "complete"}
	for _, name := range subs {
		findSubcommand(t, agentCmd, name)
	}
}

// TestRootCommand verifies that the root command has all expected subcommands registered.
func TestRootCommand(t *testing.T) {
	// Execute() adds subcommands to rootCmd, but we can't call it without
	// triggering PersistentPreRun. Instead, manually add them as Execute() does.
	rootCmd.AddCommand(caseCmd)
	rootCmd.AddCommand(payloadCmd)
	rootCmd.AddCommand(interactionCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(scannerCmd)
	rootCmd.AddCommand(agentCmd)

	expected := []string{"case", "payload", "interaction", "report", "scanner", "agent"}
	for _, name := range expected {
		findSubcommand(t, rootCmd, name)
	}
}

// TestRootFlags verifies that persistent flags are configured.
func TestRootFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	if flag := flags.Lookup("api-url"); flag == nil {
		t.Error("api-url flag not found")
	}
	if flag := flags.Lookup("api-key"); flag == nil {
		t.Error("api-key flag not found")
	}
	if flag := flags.Lookup("format"); flag == nil {
		t.Error("format flag not found")
	}
}
