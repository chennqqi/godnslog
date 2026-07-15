package scannerhub

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

// BackfillResultsRequest represents a request to import scan results
type BackfillResultsRequest struct {
	Format       string `json:"format"`      // "jsonl" or "sarif"
	RawResults   string `json:"raw_results"` // Raw scan output text
	ScannerRunID string `json:"scanner_run_id"`
}

// BackfillResults imports scan results and associates them with the scanner run.
// It parses Nuclei JSONL or SARIF output and links findings to interactions by token.
func (s *Service) BackfillResults(req *BackfillResultsRequest) (*BackfillResult, error) {
	scannerRun, err := s.GetScannerRunByID(req.ScannerRunID)
	if err != nil {
		return nil, err
	}

	var findings []ScanFinding
	switch strings.ToLower(req.Format) {
	case "jsonl", "nuclei-jsonl":
		findings, err = parseNucleiJSONL(req.RawResults)
	case "sarif":
		findings, err = parseSARIF(req.RawResults)
	case "burp-json":
		findings, err = parseBurpJSON(req.RawResults)
	case "xray-json":
		findings, err = parseXrayJSON(req.RawResults)
	default:
		return nil, fmt.Errorf("unsupported results format: %s", req.Format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	// Count interactions for this scanner run's payload
	interactionCount, err := s.engine.Where("payload_id = ?", scannerRun.PayloadID).Count(&models.Interaction{})
	if err != nil {
		return nil, fmt.Errorf("failed to count interactions: %w", err)
	}

	// Auto-associate findings with interactions by matching token
	associated := 0
	for i := range findings {
		findings[i].ScannerRunID = req.ScannerRunID
		findings[i].PayloadID = scannerRun.PayloadID
		findings[i].CaseID = scannerRun.CaseID
		findings[i].AssociatedAt = time.Now()
		if interactionCount > 0 {
			findings[i].HasInteraction = true
			associated++
		}
	}

	// Update scanner run status to "observed" if interactions exist
	if interactionCount > 0 && scannerRun.Status == models.ScannerRunStatusDistributed {
		_, err := s.engine.ID(req.ScannerRunID).Cols("status", "updated_at").Update(&models.ScannerRun{
			Status:    models.ScannerRunStatusObserved,
			UpdatedAt: time.Now(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update scanner run status: %w", err)
		}
	}

	return &BackfillResult{
		ScannerRunID:     req.ScannerRunID,
		FindingsCount:    len(findings),
		AssociatedCount:  associated,
		InteractionCount: int(interactionCount),
		Findings:         findings,
	}, nil
}

// BackfillResult represents the result of a backfill operation
type BackfillResult struct {
	ScannerRunID     string        `json:"scanner_run_id"`
	FindingsCount    int           `json:"findings_count"`
	AssociatedCount  int           `json:"associated_count"`
	InteractionCount int           `json:"interaction_count"`
	Findings         []ScanFinding `json:"findings"`
}

// ScanFinding represents a single finding from scan results
type ScanFinding struct {
	ID             string    `json:"id"`
	ScannerRunID   string    `json:"scanner_run_id"`
	CaseID         string    `json:"case_id"`
	PayloadID      string    `json:"payload_id"`
	RuleID         string    `json:"rule_id"`
	Name           string    `json:"name"`
	Severity       string    `json:"severity"`
	Description    string    `json:"description,omitempty"`
	URL            string    `json:"url,omitempty"`
	HasInteraction bool      `json:"has_interaction"`
	AssociatedAt   time.Time `json:"associated_at,omitempty"`
}

// parseNucleiJSONL parses Nuclei JSONL output into findings
func parseNucleiJSONL(raw string) ([]ScanFinding, error) {
	var findings []ScanFinding
	lines := strings.Split(strings.TrimSpace(raw), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue // Skip unparseable lines
		}

		finding := ScanFinding{
			ID:          fmt.Sprintf("finding-%d", len(findings)+1),
			RuleID:      getString(record, "template-id"),
			Name:        getString(record, "info", "name"),
			Severity:    getString(record, "info", "severity"),
			Description: getString(record, "info", "description"),
			URL:         getString(record, "matched-at"),
		}

		if finding.RuleID == "" && finding.Name == "" {
			continue // Skip empty findings
		}

		findings = append(findings, finding)
	}

	return findings, nil
}

// parseSARIF parses SARIF 2.1.0 output into findings
func parseSARIF(raw string) ([]ScanFinding, error) {
	var sarif struct {
		Runs []struct {
			Tool struct {
				Driver struct {
					Name  string `json:"name"`
					Rules []struct {
						ID               string `json:"id"`
						ShortDescription struct {
							Text string `json:"text"`
						} `json:"shortDescription"`
						FullDescription struct {
							Text string `json:"text"`
						} `json:"fullDescription"`
					} `json:"rules"`
				} `json:"driver"`
			} `json:"tool"`
			Results []struct {
				RuleID  string `json:"ruleId"`
				Message struct {
					Text string `json:"text"`
				} `json:"message"`
				Level     string `json:"level"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
	}

	if err := json.Unmarshal([]byte(raw), &sarif); err != nil {
		return nil, fmt.Errorf("failed to parse SARIF: %w", err)
	}

	ruleMap := make(map[string]struct {
		ShortDesc string
		FullDesc  string
	})
	for _, run := range sarif.Runs {
		for _, rule := range run.Tool.Driver.Rules {
			ruleMap[rule.ID] = struct {
				ShortDesc string
				FullDesc  string
			}{
				ShortDesc: rule.ShortDescription.Text,
				FullDesc:  rule.FullDescription.Text,
			}
		}
	}

	var findings []ScanFinding
	idx := 0
	for _, run := range sarif.Runs {
		for _, result := range run.Results {
			idx++
			finding := ScanFinding{
				ID:       fmt.Sprintf("finding-%d", idx),
				RuleID:   result.RuleID,
				Name:     result.Message.Text,
				Severity: sarifLevelToSeverity(result.Level),
			}

			if rule, ok := ruleMap[result.RuleID]; ok {
				if finding.Name == "" {
					finding.Name = rule.ShortDesc
				}
				finding.Description = rule.FullDesc
			}

			if len(result.Locations) > 0 {
				finding.URL = result.Locations[0].PhysicalLocation.ArtifactLocation.URI
			}

			findings = append(findings, finding)
		}
	}

	return findings, nil
}

// sarifLevelToSeverity converts SARIF level to severity string
func sarifLevelToSeverity(level string) string {
	switch strings.ToLower(level) {
	case "error":
		return "high"
	case "warning":
		return "medium"
	case "note":
		return "low"
	case "none":
		return "info"
	default:
		return "info"
	}
}

// getString safely extracts a nested string from a map
func getString(m map[string]interface{}, keys ...string) string {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			if v, ok := current[key].(string); ok {
				return v
			}
			return ""
		}
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return ""
		}
		current = next
	}
	return ""
}

// GenerateSARIF generates SARIF 2.1.0 output from scan findings
func GenerateSARIF(scannerRun *models.ScannerRun, findings []ScanFinding) ([]byte, error) {
	type sarifRule struct {
		ID               string `json:"id"`
		ShortDescription struct {
			Text string `json:"text"`
		} `json:"shortDescription"`
	}
	type sarifResult struct {
		RuleID  string `json:"ruleId"`
		Level   string `json:"level"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
	}
	type sarifRun struct {
		Tool struct {
			Driver struct {
				Name  string      `json:"name"`
				Rules []sarifRule `json:"rules"`
			} `json:"driver"`
		} `json:"tool"`
		Results []sarifResult `json:"results"`
	}
	type sarifDoc struct {
		Schema  string     `json:"$schema"`
		Version string     `json:"version"`
		Runs    []sarifRun `json:"runs"`
	}

	rules := make([]sarifRule, 0)
	results := make([]sarifResult, 0)
	seenRules := make(map[string]bool)

	for _, f := range findings {
		if !seenRules[f.RuleID] {
			seenRules[f.RuleID] = true
			rule := sarifRule{ID: f.RuleID}
			rule.ShortDescription.Text = f.Name
			rules = append(rules, rule)
		}

		result := sarifResult{
			RuleID: f.RuleID,
			Level:  severityToSARIFLevel(f.Severity),
		}
		result.Message.Text = f.Name
		results = append(results, result)
	}

	doc := sarifDoc{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{{
			Tool: struct {
				Driver struct {
					Name  string      `json:"name"`
					Rules []sarifRule `json:"rules"`
				} `json:"driver"`
			}{
				Driver: struct {
					Name  string      `json:"name"`
					Rules []sarifRule `json:"rules"`
				}{
					Name:  fmt.Sprintf("godnslog-%s", scannerRun.Scanner),
					Rules: rules,
				},
			},
			Results: results,
		}},
	}

	return json.MarshalIndent(doc, "", "  ")
}

// severityToSARIFLevel converts severity string to SARIF level
func severityToSARIFLevel(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "note"
	default:
		return "none"
	}
}

// parseBurpJSON parses Burp Suite issue JSON export into findings.
func parseBurpJSON(raw string) ([]ScanFinding, error) {
	var doc struct {
		Issues []struct {
			Name       string `json:"name"`
			Severity   string `json:"severity"`
			Confidence string `json:"confidence"`
			Host       string `json:"host"`
			Path       string `json:"path"`
		} `json:"issues"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("failed to parse Burp JSON: %w", err)
	}
	var findings []ScanFinding
	for _, issue := range doc.Issues {
		if issue.Name == "" {
			continue
		}
		url := issue.Host + issue.Path
		findings = append(findings, ScanFinding{
			ID:       fmt.Sprintf("burp-%d", len(findings)+1),
			RuleID:   issue.Name,
			Name:     issue.Name,
			Severity: burpSeverityToStandard(issue.Severity),
			URL:      url,
		})
	}
	return findings, nil
}

// burpSeverityToStandard converts Burp severity to standard severity string.
func burpSeverityToStandard(severity string) string {
	switch strings.ToLower(severity) {
	case "high", "medium", "low", "info":
		return severity
	case "certain", "firm":
		return "medium"
	default:
		return "info"
	}
}

// parseXrayJSON parses xray JSON output into findings.
func parseXrayJSON(raw string) ([]ScanFinding, error) {
	var results []struct {
		VulnID   string `json:"vuln_id"`
		Plugin   string `json:"plugin"`
		Severity string `json:"severity"`
		URL      string `json:"url"`
		Payload  string `json:"payload"`
		Detail   string `json:"detail"`
	}
	if err := json.Unmarshal([]byte(raw), &results); err != nil {
		return nil, fmt.Errorf("failed to parse xray JSON: %w", err)
	}
	var findings []ScanFinding
	for _, r := range results {
		if r.VulnID == "" && r.Plugin == "" {
			continue
		}
		ruleID := r.VulnID
		if ruleID == "" {
			ruleID = r.Plugin
		}
		desc := r.Detail
		if desc == "" && r.Payload != "" {
			desc = "Payload: " + r.Payload
		}
		findings = append(findings, ScanFinding{
			ID:          fmt.Sprintf("xray-%d", len(findings)+1),
			RuleID:      ruleID,
			Name:        r.Plugin,
			Severity:    r.Severity,
			URL:         r.URL,
			Description: desc,
		})
	}
	return findings, nil
}
