package evidencehub

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// SummaryRequest identifies the evidence scope an Agent or automation client wants to consume.
type SummaryRequest struct {
	CaseID       string `json:"case_id,omitempty"`
	PayloadID    string `json:"payload_id,omitempty"`
	ScannerRunID string `json:"scanner_run_id,omitempty"`
}

// SummaryScope is the normalized scope used to build the summary.
type SummaryScope struct {
	CaseID       string `json:"case_id,omitempty"`
	PayloadID    string `json:"payload_id,omitempty"`
	ScannerRunID string `json:"scanner_run_id,omitempty"`
	Source       string `json:"source"`
}

// SummaryResponse is a single-call, Agent-friendly evidence bundle.
type SummaryResponse struct {
	Scope         SummaryScope           `json:"scope"`
	Evidence      *interaction.Evidence  `json:"evidence"`
	ScannerRuns   []models.ScannerRun    `json:"scanner_runs"`
	PackageHashes []string               `json:"package_hashes"`
	SummaryHash   string                 `json:"summary_hash"`
	NextActions   []string               `json:"next_actions"`
	GeneratedAt   time.Time              `json:"generated_at"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Service builds evidence summaries across interactions and scanner packages.
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new evidence summary service.
func NewService(engine *xorm.Engine) *Service {
	return &Service{engine: engine}
}

// BuildSummary returns a structured evidence summary for a case, payload, or scanner run.
func (s *Service) BuildSummary(req *SummaryRequest, baseURL string) (*SummaryResponse, error) {
	scope, scannerRuns, err := s.resolveScope(req)
	if err != nil {
		return nil, err
	}

	interactionService := interaction.NewService(s.engine, nil, nil)
	evidenceService := interaction.NewEvidenceService(interactionService)
	evidenceResp, err := evidenceService.GenerateEvidence(scope.CaseID, scope.PayloadID, "json")
	if err != nil {
		return nil, err
	}

	if len(scannerRuns) == 0 {
		scannerRuns, err = s.listScannerRuns(scope.CaseID, scope.PayloadID)
		if err != nil {
			return nil, err
		}
	}

	packageHashes := collectPackageHashes(scannerRuns)
	summary := &SummaryResponse{
		Scope:         scope,
		Evidence:      evidenceResp.Evidence,
		ScannerRuns:   scannerRuns,
		PackageHashes: packageHashes,
		NextActions:   buildNextActions(evidenceResp.Evidence, scannerRuns),
		GeneratedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"base_url":          baseURL,
			"scanner_run_count": len(scannerRuns),
			"package_count":     len(packageHashes),
		},
	}

	summaryHash, err := computeSummaryHash(summary)
	if err != nil {
		return nil, err
	}
	summary.SummaryHash = summaryHash

	return summary, nil
}

func (s *Service) resolveScope(req *SummaryRequest) (SummaryScope, []models.ScannerRun, error) {
	if req == nil || (req.CaseID == "" && req.PayloadID == "" && req.ScannerRunID == "") {
		return SummaryScope{}, nil, errors.New("case_id, payload_id, or scanner_run_id is required")
	}

	if req.ScannerRunID != "" {
		var scannerRun models.ScannerRun
		has, err := s.engine.ID(req.ScannerRunID).Get(&scannerRun)
		if err != nil {
			return SummaryScope{}, nil, err
		}
		if !has {
			return SummaryScope{}, nil, fmt.Errorf("scanner run not found: %s", req.ScannerRunID)
		}
		return SummaryScope{
			CaseID:       scannerRun.CaseID,
			PayloadID:    scannerRun.PayloadID,
			ScannerRunID: scannerRun.ID,
			Source:       "scanner_run",
		}, []models.ScannerRun{scannerRun}, nil
	}

	caseID := req.CaseID

	source := "case"
	if req.PayloadID != "" {
		source = "payload"
	}

	return SummaryScope{
		CaseID:    caseID,
		PayloadID: req.PayloadID,
		Source:    source,
	}, nil, nil
}

func (s *Service) listScannerRuns(caseID, payloadID string) ([]models.ScannerRun, error) {
	var scannerRuns []models.ScannerRun
	session := s.engine.NewSession()
	defer session.Close()

	if caseID != "" {
		session = session.Where("case_i_d = ?", caseID)
	}
	if payloadID != "" {
		session = session.Where("payload_i_d = ?", payloadID)
	}
	if err := session.Desc("created_at").Limit(100).Find(&scannerRuns); err != nil {
		return nil, err
	}
	return scannerRuns, nil
}

func collectPackageHashes(scannerRuns []models.ScannerRun) []string {
	hashes := make([]string, 0, len(scannerRuns))
	seen := map[string]bool{}
	for _, run := range scannerRuns {
		if run.PackageHash == "" || seen[run.PackageHash] {
			continue
		}
		seen[run.PackageHash] = true
		hashes = append(hashes, run.PackageHash)
	}
	sort.Strings(hashes)
	return hashes
}

func buildNextActions(evidence *interaction.Evidence, scannerRuns []models.ScannerRun) []string {
	actions := []string{
		"Review the evidence strength, confidence, and timeline before closing the validation.",
		"Use package_hashes to correlate scanner distribution packages with observed callbacks.",
	}
	if evidence != nil && evidence.EvidenceStrength == interaction.EvidenceStrengthLow {
		actions = append(actions, "Wait for more interactions or run a follow-up probe before marking the finding as confirmed.")
	}
	if len(scannerRuns) == 0 {
		actions = append(actions, "Create a Scanner Run if this evidence needs scanner-side distribution provenance.")
	}
	return actions
}

func computeSummaryHash(summary *SummaryResponse) (string, error) {
	interactionIDs := make([]string, 0)
	if summary.Evidence != nil {
		for _, item := range summary.Evidence.Interactions {
			interactionIDs = append(interactionIDs, item.ID)
		}
	}
	sort.Strings(interactionIDs)

	return models.ComputeDeterministicHash(map[string]interface{}{
		"scope":           summary.Scope,
		"interaction_ids": interactionIDs,
		"evidence": map[string]interface{}{
			"case_id":           summary.Evidence.CaseID,
			"payload_id":        summary.Evidence.PayloadID,
			"evidence_strength": summary.Evidence.EvidenceStrength,
			"confidence":        summary.Evidence.Confidence,
			"interaction_count": summary.Evidence.InteractionCount,
			"unique_sources":    summary.Evidence.UniqueSources,
			"explainability":    summary.Evidence.Explainability,
		},
		"package_hashes": summary.PackageHashes,
		"next_actions":   summary.NextActions,
	})
}
