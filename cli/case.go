package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var caseCmd = &cobra.Command{
	Use:   "case",
	Short: "Manage cases",
}

var caseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new case",
	RunE:  runCaseCreate,
}

var caseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cases",
	RunE:  runCaseList,
}

var caseGetCmd = &cobra.Command{
	Use:   "get [case-id]",
	Short: "Get case details",
	Args:  cobra.ExactArgs(1),
	RunE:  runCaseGet,
}

var caseDeleteCmd = &cobra.Command{
	Use:   "delete [case-id]",
	Short: "Delete a case",
	Args:  cobra.ExactArgs(1),
	RunE:  runCaseDelete,
}

var (
	caseTitle       string
	caseDescription string
	caseTarget      string
	caseTags        []string
)

func init() {
	caseCmd.AddCommand(caseCreateCmd)
	caseCmd.AddCommand(caseListCmd)
	caseCmd.AddCommand(caseGetCmd)
	caseCmd.AddCommand(caseDeleteCmd)

	caseCreateCmd.Flags().StringVar(&caseTitle, "title", "", "Case title (required)")
	caseCreateCmd.Flags().StringVar(&caseDescription, "description", "", "Case description")
	caseCreateCmd.Flags().StringVar(&caseTarget, "target", "", "Target system")
	caseCreateCmd.Flags().StringSliceVar(&caseTags, "tags", []string{}, "Case tags")
	caseCreateCmd.MarkFlagRequired("title")
}

type Case struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Target      string   `json:"target"`
	Tags        []string `json:"tags"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type CaseCreateRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Target      string   `json:"target"`
	Tags        []string `json:"tags"`
}

type CaseListResponse struct {
	Items []Case `json:"items"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
}

func runCaseCreate(cmd *cobra.Command, args []string) error {
	req := CaseCreateRequest{
		Title:       caseTitle,
		Description: caseDescription,
		Target:      caseTarget,
		Tags:        caseTags,
	}

	body, err := apiRequest("POST", "/cases", req)
	if err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *Case  `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Case created successfully\n")
	fmt.Printf("ID: %s\n", resp.Data.ID)
	fmt.Printf("Title: %s\n", resp.Data.Title)

	return nil
}

func runCaseList(cmd *cobra.Command, args []string) error {
	body, err := apiRequest("GET", "/cases?page=1&page_size=100", nil)
	if err != nil {
		return fmt.Errorf("failed to list cases: %w", err)
	}

	var resp struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Data    *CaseListResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Total cases: %d\n\n", resp.Data.Total)
	for _, c := range resp.Data.Items {
		fmt.Printf("ID: %s\n", c.ID)
		fmt.Printf("  Title: %s\n", c.Title)
		fmt.Printf("  Status: %s\n", c.Status)
		fmt.Printf("  Created: %s\n", c.CreatedAt)
		fmt.Println()
	}

	return nil
}

func runCaseGet(cmd *cobra.Command, args []string) error {
	caseID := args[0]

	body, err := apiRequest("GET", "/cases/"+caseID, nil)
	if err != nil {
		return fmt.Errorf("failed to get case: %w", err)
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *Case  `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("ID: %s\n", resp.Data.ID)
	fmt.Printf("Title: %s\n", resp.Data.Title)
	fmt.Printf("Description: %s\n", resp.Data.Description)
	fmt.Printf("Target: %s\n", resp.Data.Target)
	fmt.Printf("Status: %s\n", resp.Data.Status)
	fmt.Printf("Tags: %v\n", resp.Data.Tags)
	fmt.Printf("Created: %s\n", resp.Data.CreatedAt)
	fmt.Printf("Updated: %s\n", resp.Data.UpdatedAt)

	return nil
}

func runCaseDelete(cmd *cobra.Command, args []string) error {
	caseID := args[0]

	body, err := apiRequest("DELETE", "/cases/"+caseID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete case: %w", err)
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

	fmt.Printf("Case %s deleted successfully\n", caseID)

	return nil
}
