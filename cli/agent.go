package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agent runs",
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agent runs",
	RunE:  runAgentList,
}

var agentGetCmd = &cobra.Command{
	Use:   "get [run-id]",
	Short: "Get agent run details",
	Args:  cobra.ExactArgs(1),
	RunE:  runAgentGet,
}

var agentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new agent run",
	RunE:  runAgentCreate,
}

var agentCompleteCmd = &cobra.Command{
	Use:   "complete [run-id]",
	Short: "Complete an agent run (generate review, export evidence, update status)",
	Args:  cobra.ExactArgs(1),
	RunE:  runAgentComplete,
}

var (
	agentID      string
	agentOperator string
	agentTarget  string
	agentTitle   string
	agentCaseID  string
	agentPayloadID string
	agentFormat  string
	agentDecision string
	agentDecisionReason string
	agentIncludeAudit bool
	agentStatusFilter string
)

func init() {
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentGetCmd)
	agentCmd.AddCommand(agentCreateCmd)
	agentCmd.AddCommand(agentCompleteCmd)

	agentCreateCmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID (required)")
	agentCreateCmd.Flags().StringVar(&agentOperator, "operator-id", "cli", "Operator ID")
	agentCreateCmd.Flags().StringVar(&agentTarget, "target", "", "Target URL")
	agentCreateCmd.Flags().StringVar(&agentTitle, "title", "", "Agent run title")
	agentCreateCmd.Flags().StringVar(&agentCaseID, "case-id", "", "Case ID to bind")
	agentCreateCmd.Flags().StringVar(&agentPayloadID, "payload-id", "", "Payload ID to bind")
	agentCreateCmd.MarkFlagRequired("agent-id")

	agentListCmd.Flags().StringVar(&agentID, "agent-id", "", "Filter by Agent ID")
	agentListCmd.Flags().StringVar(&agentStatusFilter, "status", "", "Filter by status")

	agentCompleteCmd.Flags().StringVar(&agentFormat, "format", "json", "Output format (json or markdown)")
	agentCompleteCmd.Flags().StringVar(&agentDecision, "decision", "", "Review decision (accepted, false_positive, needs_manual_followup, insufficient_evidence)")
	agentCompleteCmd.Flags().StringVar(&agentDecisionReason, "decision-reason", "", "Reason for the decision")
	agentCompleteCmd.Flags().BoolVar(&agentIncludeAudit, "include-audit", false, "Include audit references in export")
}

type AgentRun struct {
	ID         string `json:"id"`
	AgentID    string `json:"agent_id"`
	OperatorID string `json:"operator_id"`
	CaseID     string `json:"case_id"`
	PayloadID  string `json:"payload_id"`
	Target     string `json:"target"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	StartedAt  string `json:"started_at"`
	EndedAt    string `json:"ended_at"`
}

type AgentRunListResponse struct {
	Items []AgentRun `json:"items"`
	Total int        `json:"total"`
}

func runAgentList(cmd *cobra.Command, args []string) error {
	query := "/agent-runs?page=1&page_size=50"
	if agentID != "" {
		query += "&agent_id=" + agentID
	}
	if agentStatusFilter != "" {
		query += "&status=" + agentStatusFilter
	}

	body, err := apiRequest("GET", query, nil)
	if err != nil {
		return fmt.Errorf("failed to list agent runs: %w", err)
	}

	var resp struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    *AgentRunListResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Total agent runs: %d\n\n", resp.Data.Total)
	for _, r := range resp.Data.Items {
		fmt.Printf("ID: %s\n", r.ID)
		fmt.Printf("  Agent: %s\n", r.AgentID)
		fmt.Printf("  Title: %s\n", r.Title)
		fmt.Printf("  Status: %s\n", r.Status)
		fmt.Printf("  Target: %s\n", r.Target)
		fmt.Printf("  Created: %s\n", r.CreatedAt)
		fmt.Println()
	}

	return nil
}

func runAgentGet(cmd *cobra.Command, args []string) error {
	runID := args[0]

	body, err := apiRequest("GET", "/agent-runs/"+runID, nil)
	if err != nil {
		return fmt.Errorf("failed to get agent run: %w", err)
	}

	var resp struct {
		Code    int        `json:"code"`
		Message string     `json:"message"`
		Data    *AgentRun  `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("ID: %s\n", resp.Data.ID)
	fmt.Printf("Agent ID: %s\n", resp.Data.AgentID)
	fmt.Printf("Operator ID: %s\n", resp.Data.OperatorID)
	fmt.Printf("Title: %s\n", resp.Data.Title)
	fmt.Printf("Status: %s\n", resp.Data.Status)
	fmt.Printf("Target: %s\n", resp.Data.Target)
	if resp.Data.CaseID != "" {
		fmt.Printf("Case ID: %s\n", resp.Data.CaseID)
	}
	if resp.Data.PayloadID != "" {
		fmt.Printf("Payload ID: %s\n", resp.Data.PayloadID)
	}
	fmt.Printf("Created: %s\n", resp.Data.CreatedAt)
	if resp.Data.StartedAt != "" {
		fmt.Printf("Started: %s\n", resp.Data.StartedAt)
	}
	if resp.Data.EndedAt != "" {
		fmt.Printf("Ended: %s\n", resp.Data.EndedAt)
	}

	return nil
}

func runAgentCreate(cmd *cobra.Command, args []string) error {
	req := map[string]interface{}{
		"agent_id":    agentID,
		"operator_id": agentOperator,
		"target":      agentTarget,
		"title":       agentTitle,
	}
	if agentCaseID != "" {
		req["case_id"] = agentCaseID
	}
	if agentPayloadID != "" {
		req["payload_id"] = agentPayloadID
	}

	body, err := apiRequest("POST", "/agent-runs", req)
	if err != nil {
		return fmt.Errorf("failed to create agent run: %w", err)
	}

	var resp struct {
		Code    int        `json:"code"`
		Message string     `json:"message"`
		Data    *AgentRun  `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Agent run created successfully\n")
	fmt.Printf("ID: %s\n", resp.Data.ID)
	fmt.Printf("Status: %s\n", resp.Data.Status)

	return nil
}

func runAgentComplete(cmd *cobra.Command, args []string) error {
	runID := args[0]

	req := map[string]interface{}{
		"format":        agentFormat,
		"include_audit": agentIncludeAudit,
	}
	if agentDecision != "" {
		req["decision"] = agentDecision
	}
	if agentDecisionReason != "" {
		req["decision_reason"] = agentDecisionReason
	}

	body, err := apiRequest("POST", "/agent-runs/"+runID+"/complete", req)
	if err != nil {
		return fmt.Errorf("failed to complete agent run: %w", err)
	}

	var resp struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	if outputFormat == "json" {
		var prettyJSON map[string]interface{}
		if err := json.Unmarshal(resp.Data, &prettyJSON); err == nil {
			formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
			fmt.Println(string(formatted))
			return nil
		}
	}

	// Print summary
	var result struct {
		AgentRunID       string `json:"agent_run_id"`
		Status           string `json:"status"`
		PackageHash      string `json:"package_hash"`
		Decision         string `json:"decision"`
		InteractionCount int    `json:"interaction_count"`
		EvidenceStrength string `json:"evidence_strength"`
		OperationID      string `json:"operation_id"`
		AuditRefID       string `json:"audit_ref_id"`
	}
	json.Unmarshal(resp.Data, &result)

	fmt.Printf("Agent run completed successfully\n")
	fmt.Printf("ID: %s\n", result.AgentRunID)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Interactions: %d\n", result.InteractionCount)
	if result.EvidenceStrength != "" {
		fmt.Printf("Evidence Strength: %s\n", result.EvidenceStrength)
	}
	if result.Decision != "" {
		fmt.Printf("Decision: %s\n", result.Decision)
	}
	if result.PackageHash != "" {
		fmt.Printf("Package Hash: %s\n", result.PackageHash)
	}
	if result.OperationID != "" {
		fmt.Printf("Operation ID: %s\n", result.OperationID)
	}
	if result.AuditRefID != "" {
		fmt.Printf("Audit Ref ID: %s\n", result.AuditRefID)
	}

	return nil
}
