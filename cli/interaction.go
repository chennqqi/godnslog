package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var interactionCmd = &cobra.Command{
	Use:   "interaction",
	Short: "Manage interactions",
}

var interactionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List interactions",
	RunE:  runInteractionList,
}

var interactionPollCmd = &cobra.Command{
	Use:   "poll [payload-id]",
	Short: "Poll for interactions from a payload",
	Args:  cobra.ExactArgs(1),
	RunE:  runInteractionPoll,
}

var (
	interactionCaseID string
	interactionType   string
	interactionLimit  int
	pollTimeout       time.Duration
	pollInterval      time.Duration
)

func init() {
	interactionCmd.AddCommand(interactionListCmd)
	interactionCmd.AddCommand(interactionPollCmd)

	interactionListCmd.Flags().StringVar(&interactionCaseID, "case-id", "", "Filter by case ID")
	interactionListCmd.Flags().StringVar(&interactionType, "type", "", "Filter by type (dns, http, smtp, ldap)")
	interactionListCmd.Flags().IntVarP(&interactionLimit, "limit", "l", 50, "Limit number of results")

	interactionPollCmd.Flags().DurationVar(&pollTimeout, "timeout", 5*time.Minute, "Polling timeout")
	interactionPollCmd.Flags().DurationVar(&pollInterval, "interval", 5*time.Second, "Polling interval")
}

type Interaction struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	SourceIP    string            `json:"source_ip"`
	Token       string            `json:"token"`
	Domain      string            `json:"domain"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	UserAgent   string            `json:"user_agent"`
	ContentType string            `json:"content_type"`
	Timestamp   string            `json:"timestamp"`
}

type InteractionListResponse struct {
	Items []Interaction `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
}

func runInteractionList(cmd *cobra.Command, args []string) error {
	query := fmt.Sprintf("/interactions?page=1&page_size=%d", interactionLimit)
	if interactionCaseID != "" {
		query += "&case_id=" + interactionCaseID
	}
	if interactionType != "" {
		query += "&type=" + interactionType
	}

	body, err := apiRequest("GET", query, nil)
	if err != nil {
		return fmt.Errorf("failed to list interactions: %w", err)
	}

	var resp struct {
		Code    int                      `json:"code"`
		Message string                   `json:"message"`
		Data    *InteractionListResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	fmt.Printf("Total interactions: %d\n\n", resp.Data.Total)
	for _, i := range resp.Data.Items {
		fmt.Printf("ID: %s\n", i.ID)
		fmt.Printf("  Type: %s\n", i.Type)
		fmt.Printf("  Source IP: %s\n", i.SourceIP)
		fmt.Printf("  Token: %s\n", i.Token)
		if i.Domain != "" {
			fmt.Printf("  Domain: %s\n", i.Domain)
		}
		if i.Method != "" && i.Path != "" {
			fmt.Printf("  Request: %s %s\n", i.Method, i.Path)
		}
		fmt.Printf("  Timestamp: %s\n", i.Timestamp)
		fmt.Println()
	}

	return nil
}

func runInteractionPoll(cmd *cobra.Command, args []string) error {
	payloadID := args[0]
	seenIDs := make(map[string]bool)

	fmt.Printf("Polling for interactions from payload %s...\n", payloadID)
	fmt.Printf("Timeout: %v, Interval: %v\n\n", pollTimeout, pollInterval)

	timeout := time.After(pollTimeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			fmt.Println("\nPolling timeout reached")
			return nil
		case <-ticker.C:
			query := fmt.Sprintf("/interactions?token=%s&page=1&page_size=50", payloadID)
			body, err := apiRequest("GET", query, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error polling: %v\n", err)
				continue
			}

			var resp struct {
				Code    int                      `json:"code"`
				Message string                   `json:"message"`
				Data    *InteractionListResponse `json:"data"`
			}
			if err := json.Unmarshal(body, &resp); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				continue
			}

			if resp.Code != 0 {
				fmt.Fprintf(os.Stderr, "API error: %s\n", resp.Message)
				continue
			}

			newCount := 0
			for _, i := range resp.Data.Items {
				if !seenIDs[i.ID] {
					seenIDs[i.ID] = true
					newCount++
					fmt.Printf("[%s] New interaction detected:\n", time.Now().Format("15:04:05"))
					fmt.Printf("  ID: %s\n", i.ID)
					fmt.Printf("  Type: %s\n", i.Type)
					fmt.Printf("  Source IP: %s\n", i.SourceIP)
					if i.Domain != "" {
						fmt.Printf("  Domain: %s\n", i.Domain)
					}
					if i.Method != "" && i.Path != "" {
						fmt.Printf("  Request: %s %s\n", i.Method, i.Path)
					}
					fmt.Println()
				}
			}
		}
	}
}
