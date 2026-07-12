package interaction

import (
	"sort"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

// AttackChain represents a summary of aggregated interactions for a token.
type AttackChain struct {
	Token            string   `json:"token"`
	InteractionCount int      `json:"interaction_count"`
	Protocols        []string `json:"protocols"`
	ExploitTypes     []string `json:"exploit_types"`
	FirstSeen        string   `json:"first_seen"`
	LastSeen         string   `json:"last_seen"`
	Confidence       string   `json:"confidence"`
}

// AttackChainDetail represents a detailed view of an attack chain for a token,
// including all associated interactions.
type AttackChainDetail struct {
	Token            string                `json:"token"`
	InteractionCount int                   `json:"interaction_count"`
	Protocols        []string              `json:"protocols"`
	ExploitTypes     []string              `json:"exploit_types"`
	FirstSeen        string                `json:"first_seen"`
	LastSeen         string                `json:"last_seen"`
	Confidence       string                `json:"confidence"`
	Interactions     []models.Interaction  `json:"interactions"`
}

// AttackChainListResponse represents a paginated list of attack chains.
type AttackChainListResponse struct {
	Items      []AttackChain `json:"items"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// tokenRow is a helper struct for the GROUP BY aggregation query.
type tokenRow struct {
	Token   string    `xorm:"token"`
	Cnt     int       `xorm:"cnt"`
	FirstTS time.Time `xorm:"first_ts"`
	LastTS  time.Time `xorm:"last_ts"`
}

// GetAttackChains returns a paginated list of attack chains, each being an
// aggregation of all interactions sharing the same token.
func (s *Service) GetAttackChains(page, pageSize int) (*AttackChainListResponse, error) {
	// Count distinct tokens
	type cntResult struct {
		Cnt int64 `xorm:"cnt"`
	}
	var cr cntResult
	_, err := s.engine.SQL(
		"SELECT COUNT(DISTINCT token) as cnt FROM interactions WHERE token IS NOT NULL AND token != ''",
	).Get(&cr)
	if err != nil {
		return nil, err
	}

	total := cr.Cnt

	if total == 0 {
		totalPages := 0
		if pageSize > 0 {
			totalPages = int(total) / pageSize
		}
		return &AttackChainListResponse{
			Items:      []AttackChain{},
			Total:      0,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		}, nil
	}

	// Get tokens with aggregation
	offset := (page - 1) * pageSize
	var rows []tokenRow
	err = s.engine.SQL(
		"SELECT token, COUNT(*) as cnt, MIN(timestamp) as first_ts, MAX(timestamp) as last_ts FROM interactions WHERE token IS NOT NULL AND token != '' GROUP BY token ORDER BY last_ts DESC LIMIT ? OFFSET ?",
		pageSize, offset,
	).Find(&rows)
	if err != nil {
		return nil, err
	}

	items := make([]AttackChain, 0, len(rows))
	for _, row := range rows {
		chain := s.buildAttackChain(row.Token, row.Cnt, row.FirstTS, row.LastTS)
		items = append(items, chain)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &AttackChainListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAttackChainDetail returns detailed information for a single attack chain,
// including all interactions associated with the given token.
func (s *Service) GetAttackChainDetail(token string) (*AttackChainDetail, error) {
	var interactions []models.Interaction
	err := s.engine.Where("token = ?", token).Asc("timestamp").Find(&interactions)
	if err != nil {
		return nil, err
	}

	if len(interactions) == 0 {
		return &AttackChainDetail{
			Token:        token,
			Interactions: []models.Interaction{},
		}, nil
	}

	firstSeen := interactions[0].Timestamp
	lastSeen := interactions[len(interactions)-1].Timestamp

	chain := s.buildAttackChain(token, len(interactions), firstSeen, lastSeen)

	return &AttackChainDetail{
		Token:            chain.Token,
		InteractionCount: chain.InteractionCount,
		Protocols:        chain.Protocols,
		ExploitTypes:     chain.ExploitTypes,
		FirstSeen:        chain.FirstSeen,
		LastSeen:         chain.LastSeen,
		Confidence:       chain.Confidence,
		Interactions:     interactions,
	}, nil
}

// buildAttackChain fetches all interactions for the token and enriches the
// base aggregation (count, first/last seen) with protocols, exploit types,
// and confidence level derived from the interactions.
func (s *Service) buildAttackChain(token string, count int, firstTS, lastTS time.Time) AttackChain {
	var interactions []models.Interaction
	err := s.engine.Where("token = ?", token).Find(&interactions)
	if err != nil {
		return AttackChain{
			Token:            token,
			InteractionCount: count,
			FirstSeen:        firstTS.Format(time.RFC3339),
			LastSeen:         lastTS.Format(time.RFC3339),
		}
	}

	protocolsSet := make(map[string]bool)
	exploitTypesSet := make(map[string]bool)
	order := []string{"high", "medium", "low"}
	highestConfidence := ""

	for _, interaction := range interactions {
		if interaction.Type != "" {
			protocolsSet[interaction.Type] = true
		}
		if interaction.ExploitType != nil && *interaction.ExploitType != "" {
			exploitTypesSet[*interaction.ExploitType] = true
		}
		if interaction.Confidence != nil && *interaction.Confidence != "" {
			if highestConfidence == "" {
				highestConfidence = *interaction.Confidence
			} else {
				highestConfidence = higherConfidence(highestConfidence, *interaction.Confidence, order)
			}
		}
	}

	return AttackChain{
		Token:            token,
		InteractionCount: count,
		Protocols:        sortedKeys(protocolsSet),
		ExploitTypes:     sortedKeys(exploitTypesSet),
		FirstSeen:        firstTS.Format(time.RFC3339),
		LastSeen:         lastTS.Format(time.RFC3339),
		Confidence:       highestConfidence,
	}
}

// higherConfidence returns the higher confidence level between a and b,
// based on their positions in the order slice (earlier = higher).
func higherConfidence(a, b string, order []string) string {
	rankA := len(order)
	rankB := len(order)
	for i, level := range order {
		if a == level {
			rankA = i
		}
		if b == level {
			rankB = i
		}
	}
	if rankA < rankB {
		return a
	}
	return b
}

// sortedKeys returns the sorted keys of a map[string]bool.
func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
