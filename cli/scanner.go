package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var scannerCmd = &cobra.Command{
	Use:   "scanner",
	Short: "Manage scanner runs",
}

var scannerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Create a scanner run",
	RunE:  runScannerRun,
}

var scannerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scanner runs",
	RunE:  runScannerList,
}

var scannerGetCmd = &cobra.Command{
	Use:   "get [run-id]",
	Short: "Get scanner run details",
	Args:  cobra.ExactArgs(1),
	RunE:  runScannerGet,
}

var (
	scannerName      string
	scannerTarget    string
	scannerTemplate  string
	scannerDelivery  string
	scannerCaseID    string
	scannerPayloadID string
	scannerStatus    string
)

func init() {
	scannerCmd.AddCommand(scannerRunCmd)
	scannerCmd.AddCommand(scannerListCmd)
	scannerCmd.AddCommand(scannerGetCmd)

	scannerRunCmd.Flags().StringVar(&scannerCaseID, "case-id", "", "Case ID (required)")
	scannerRunCmd.Flags().StringVar(&scannerPayloadID, "payload-id", "", "Payload ID (required)")
	scannerRunCmd.Flags().StringVar(&scannerName, "scanner", "nuclei", "Scanner type (nuclei, burp, yakit, zap, xray, rad, postman, apifox)")
	scannerRunCmd.Flags().StringVar(&scannerTarget, "target", "", "Target URL (required)")
	scannerRunCmd.Flags().StringVar(&scannerTemplate, "template", "ssrf-basic", "Template name")
	scannerRunCmd.Flags().StringVar(&scannerDelivery, "delivery", "", "Delivery method (auto-selected based on scanner)")
	scannerRunCmd.MarkFlagRequired("case-id")
	scannerRunCmd.MarkFlagRequired("payload-id")
	scannerRunCmd.MarkFlagRequired("target")

	scannerListCmd.Flags().StringVar(&scannerCaseID, "case-id", "", "Filter by case ID")
	scannerListCmd.Flags().StringVar(&scannerStatus, "status", "", "Filter by status (created, distributed, observed, evidenced)")
}

type ScannerRunCreateRequest struct {
	CaseID         string `json:"case_id"`
	PayloadID      string `json:"payload_id"`
	Scanner        string `json:"scanner"`
	Target         string `json:"target"`
	Template       string `json:"template"`
	DeliveryMethod string `json:"delivery_method"`
}

type ScannerRun struct {
	ID             string `json:"id"`
	CaseID         string `json:"case_id"`
	PayloadID      string `json:"payload_id"`
	Scanner        string `json:"scanner"`
	Target         string `json:"target"`
	Template       string `json:"template"`
	DeliveryMethod string `json:"delivery_method"`
	Command        string `json:"command"`
	Jsonl          string `json:"jsonl"`
	PackageHash    string `json:"package_hash"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

type ScannerRunListResponse struct {
	Items []ScannerRun `json:"items"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
}

// defaultDeliveryMethod returns the default delivery method for a scanner
func defaultDeliveryMethod(scanner string) string {
	switch scanner {
	case "nuclei":
		return "nuclei-jsonl"
	case "burp":
		return "burp-extension"
	case "yakit":
		return "yakit-script"
	case "zap":
		return "zap-script"
	case "xray":
		return "xray-webhook"
	case "rad":
		return "rad-webhook"
	case "postman":
		return "postman-env"
	case "apifox":
		return "apifox-env"
	default:
		return "nuclei-jsonl"
	}
}

func runScannerRun(cmd *cobra.Command, args []string) error {
	delivery := scannerDelivery
	if delivery == "" {
		delivery = defaultDeliveryMethod(scannerName)
	}

	req := ScannerRunCreateRequest{
		CaseID:         scannerCaseID,
		PayloadID:      scannerPayloadID,
		Scanner:        scannerName,
		Target:         scannerTarget,
		Template:       scannerTemplate,
		DeliveryMethod: delivery,
	}

	body, err := apiRequest("POST", "/scanner-runs", req)
	if err != nil {
		return fmt.Errorf("failed to create scanner run: %w", err)
	}

	var resp struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    *ScannerRun `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Scanner run created successfully\n")
	fmt.Printf("ID: %s\n", resp.Data.ID)
	fmt.Printf("Scanner: %s\n", resp.Data.Scanner)
	fmt.Printf("Target: %s\n", resp.Data.Target)
	fmt.Printf("Status: %s\n", resp.Data.Status)
	fmt.Printf("Delivery Method: %s\n", resp.Data.DeliveryMethod)
	fmt.Printf("\nCommand:\n%s\n", resp.Data.Command)
	fmt.Printf("\nJSONL:\n%s\n", resp.Data.Jsonl)

	return nil
}

func runScannerList(cmd *cobra.Command, args []string) error {
	query := "/scanner-runs?page=1&page_size=50"
	if scannerCaseID != "" {
		query += "&case_id=" + scannerCaseID
	}
	if scannerStatus != "" {
		query += "&status=" + scannerStatus
	}

	body, err := apiRequest("GET", query, nil)
	if err != nil {
		return fmt.Errorf("failed to list scanner runs: %w", err)
	}

	var resp struct {
		Code    int                     `json:"code"`
		Message string                  `json:"message"`
		Data    *ScannerRunListResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Total scanner runs: %d\n\n", resp.Data.Total)
	for _, r := range resp.Data.Items {
		fmt.Printf("ID: %s\n", r.ID)
		fmt.Printf("  Scanner: %s\n", r.Scanner)
		fmt.Printf("  Target: %s\n", r.Target)
		fmt.Printf("  Template: %s\n", r.Template)
		fmt.Printf("  Status: %s\n", r.Status)
		fmt.Printf("  Created: %s\n", r.CreatedAt)
		fmt.Println()
	}

	return nil
}

func runScannerGet(cmd *cobra.Command, args []string) error {
	runID := args[0]

	body, err := apiRequest("GET", "/scanner-runs/"+runID, nil)
	if err != nil {
		return fmt.Errorf("failed to get scanner run: %w", err)
	}

	var resp struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    *ScannerRun `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("ID: %s\n", resp.Data.ID)
	fmt.Printf("Scanner: %s\n", resp.Data.Scanner)
	fmt.Printf("Target: %s\n", resp.Data.Target)
	fmt.Printf("Template: %s\n", resp.Data.Template)
	fmt.Printf("Delivery Method: %s\n", resp.Data.DeliveryMethod)
	fmt.Printf("Status: %s\n", resp.Data.Status)
	fmt.Printf("Package Hash: %s\n", resp.Data.PackageHash)
	fmt.Printf("Created: %s\n", resp.Data.CreatedAt)
	fmt.Printf("\nCommand:\n%s\n", resp.Data.Command)
	fmt.Printf("\nJSONL:\n%s\n", resp.Data.Jsonl)

	return nil
}
