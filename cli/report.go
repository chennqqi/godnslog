package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate and export reports",
}

var reportExportCmd = &cobra.Command{
	Use:   "export [case-id]",
	Short: "Export case report",
	Args:  cobra.ExactArgs(1),
	RunE:  runReportExport,
}

var (
	reportFormat string
	reportOutput string
	includeRaw   bool
)

func init() {
	reportCmd.AddCommand(reportExportCmd)

	reportExportCmd.Flags().StringVarP(&reportFormat, "format", "f", "json", "Output format (json, markdown, csv)")
	reportExportCmd.Flags().StringVarP(&reportOutput, "output", "o", "", "Output file (default: stdout)")
	reportExportCmd.Flags().BoolVar(&includeRaw, "include-raw", false, "Include raw data in export")
}

type EvidenceExportRequest struct {
	CaseID      string `json:"case_id"`
	Format      string `json:"format"`
	IncludeRaw  bool   `json:"include_raw"`
}

func runReportExport(cmd *cobra.Command, args []string) error {
	caseID := args[0]

	req := EvidenceExportRequest{
		CaseID:     caseID,
		Format:     reportFormat,
		IncludeRaw: includeRaw,
	}

	body, err := apiRequest("POST", "/evidence/export", req)
	if err != nil {
		return fmt.Errorf("failed to export report: %w", err)
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	output := resp.Data
	if reportOutput != "" {
		if err := os.WriteFile(reportOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Report exported to: %s\n", reportOutput)
	} else {
		fmt.Println(output)
	}

	return nil
}
