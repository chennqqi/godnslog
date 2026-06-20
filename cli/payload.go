package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var payloadCmd = &cobra.Command{
	Use:   "payload",
	Short: "Manage payloads",
}

var payloadCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new payload",
	RunE:  runPayloadCreate,
}

var payloadListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all payloads",
	RunE:  runPayloadList,
}

var payloadRevokeCmd = &cobra.Command{
	Use:   "revoke [payload-id]",
	Short: "Revoke a payload",
	Args:  cobra.ExactArgs(1),
	RunE:  runPayloadRevoke,
}

var payloadPreviewCmd = &cobra.Command{
	Use:   "preview [payload-id]",
	Short: "Preview payload details and rendered output",
	Args:  cobra.ExactArgs(1),
	RunE:  runPayloadPreview,
}

var (
	payloadTemplate string
	payloadCaseID   string
	payloadVars     map[string]string
	payloadExpires  string
)

func init() {
	payloadCmd.AddCommand(payloadCreateCmd)
	payloadCmd.AddCommand(payloadListCmd)
	payloadCmd.AddCommand(payloadRevokeCmd)
	payloadCmd.AddCommand(payloadPreviewCmd)

	payloadCreateCmd.Flags().StringVar(&payloadTemplate, "template", "", "Payload template (required)")
	payloadCreateCmd.Flags().StringVar(&payloadCaseID, "case-id", "", "Case ID to bind payload to")
	payloadCreateCmd.Flags().StringToStringVar(&payloadVars, "var", map[string]string{}, "Template variables")
	payloadCreateCmd.Flags().StringVar(&payloadExpires, "expires", "", "Expiration time (e.g., 1h, 24h)")
	payloadCreateCmd.MarkFlagRequired("template")
}

type Payload struct {
	ID              string            `json:"id"`
	Token           string            `json:"token"`
	Template        string            `json:"template"`
	CaseID          string            `json:"case_id"`
	Variables       map[string]string `json:"variables"`
	RenderedPayload string            `json:"rendered_payload"`
	Status          string            `json:"status"`
	ExpiresAt       string            `json:"expires_at"`
	CreatedAt       string            `json:"created_at"`
}

type PayloadCreateRequest struct {
	Template  string            `json:"template"`
	CaseID    string            `json:"case_id,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
	ExpiresIn string            `json:"expires_in,omitempty"`
}

type PayloadListResponse struct {
	Items []Payload `json:"items"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
}

func runPayloadCreate(cmd *cobra.Command, args []string) error {
	req := PayloadCreateRequest{
		Template:  payloadTemplate,
		CaseID:    payloadCaseID,
		Variables: payloadVars,
		ExpiresIn: payloadExpires,
	}

	body, err := apiRequest("POST", "/payloads", req)
	if err != nil {
		return fmt.Errorf("failed to create payload: %w", err)
	}

	var resp struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    *Payload `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Payload created successfully\n")
	fmt.Printf("ID: %s\n", resp.Data.ID)
	fmt.Printf("Token: %s\n", resp.Data.Token)
	fmt.Printf("Rendered Payload: %s\n", resp.Data.RenderedPayload)
	fmt.Printf("Status: %s\n", resp.Data.Status)

	return nil
}

func runPayloadList(cmd *cobra.Command, args []string) error {
	body, err := apiRequest("GET", "/payloads?page=1&page_size=100", nil)
	if err != nil {
		return fmt.Errorf("failed to list payloads: %w", err)
	}

	var resp struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    *PayloadListResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Total payloads: %d\n\n", resp.Data.Total)
	for _, p := range resp.Data.Items {
		fmt.Printf("ID: %s\n", p.ID)
		fmt.Printf("  Token: %s\n", p.Token)
		fmt.Printf("  Template: %s\n", p.Template)
		fmt.Printf("  Status: %s\n", p.Status)
		fmt.Printf("  Created: %s\n", p.CreatedAt)
		fmt.Println()
	}

	return nil
}

func runPayloadRevoke(cmd *cobra.Command, args []string) error {
	payloadID := args[0]

	req := map[string]string{"status": "revoked"}
	body, err := apiRequest("PUT", "/payloads/"+payloadID+"/status", req)
	if err != nil {
		return fmt.Errorf("failed to revoke payload: %w", err)
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Payload %s revoked successfully\n", payloadID)

	return nil
}

func runPayloadPreview(cmd *cobra.Command, args []string) error {
	payloadID := args[0]

	body, err := apiRequest("GET", "/payloads/"+payloadID, nil)
	if err != nil {
		return fmt.Errorf("failed to get payload: %w", err)
	}

	var resp struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    *Payload `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("ID: %s\n", resp.Data.ID)
	fmt.Printf("Token: %s\n", resp.Data.Token)
	fmt.Printf("Template: %s\n", resp.Data.Template)
	fmt.Printf("Case ID: %s\n", resp.Data.CaseID)
	fmt.Printf("Status: %s\n", resp.Data.Status)
	fmt.Printf("Rendered Payload: %s\n", resp.Data.RenderedPayload)
	fmt.Printf("Variables: %v\n", resp.Data.Variables)
	fmt.Printf("Expires At: %s\n", resp.Data.ExpiresAt)
	fmt.Printf("Created At: %s\n", resp.Data.CreatedAt)

	return nil
}
