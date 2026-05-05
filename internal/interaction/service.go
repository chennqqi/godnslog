package interaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/internal/models"
)

var (
	ErrInteractionNotFound = errors.New("interaction not found")
)

// Service provides interaction management services
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new interaction service
func NewService(engine *xorm.Engine) *Service {
	return &Service{engine: engine}
}

// CreateInteraction creates a new interaction record
func (s *Service) CreateInteraction(interaction *models.Interaction) error {
	if interaction.ID == "" {
		interaction.ID = models.GenerateID()
	}
	if interaction.Timestamp.IsZero() {
		interaction.Timestamp = time.Now()
	}

	// Auto-attribution: associate interaction with payload and case based on token
	if interaction.Token != nil && *interaction.Token != "" {
		// Find payload by token
		var payloadID string
		var caseID string

		// Query payload table for token match
		type PayloadInfo struct {
			ID     int64 `xorm:"id"`
			CaseId int64 `xorm:"case_id"`
		}
		var payloadInfo PayloadInfo
		has, err := s.engine.Table("payloads").Where("token = ?", *interaction.Token).Get(&payloadInfo)
		if err == nil && has {
			payloadID = strconv.FormatInt(payloadInfo.ID, 10)
			caseID = strconv.FormatInt(payloadInfo.CaseId, 10)
		}

		// Set payload_id and case_id if found
		if payloadID != "" {
			interaction.PayloadID = &payloadID
		}
		if caseID != "" {
			interaction.CaseID = &caseID
		}
	}

	_, err := s.engine.InsertOne(interaction)
	return err
}

// GetInteractionByID retrieves an interaction by its ID
func (s *Service) GetInteractionByID(id string) (*models.Interaction, error) {
	var interaction models.Interaction
	has, err := s.engine.ID(id).Get(&interaction)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrInteractionNotFound
	}
	return &interaction, nil
}

// ListInteractions retrieves interactions with filtering
func (s *Service) ListInteractions(caseID, payloadID, interactionType string, startTime, endTime *time.Time, page, pageSize int) (*models.InteractionListResponse, error) {
	var interactions []models.Interaction
	session := s.engine.NewSession()
	defer session.Close()

	if caseID != "" {
		session = session.Where("case_id = ?", caseID)
	}
	if payloadID != "" {
		session = session.Where("payload_id = ?", payloadID)
	}
	if interactionType != "" {
		session = session.Where("type = ?", interactionType)
	}
	if startTime != nil {
		session = session.Where("timestamp >= ?", startTime)
	}
	if endTime != nil {
		session = session.Where("timestamp <= ?", endTime)
	}

	total, err := session.Count(&models.Interaction{})
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := session.Desc("timestamp").Limit(pageSize, offset).Find(&interactions); err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.InteractionListResponse{
		Items:      interactions,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// DeleteInteractions deletes interactions by IDs
func (s *Service) DeleteInteractions(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := s.engine.In("id", ids).Delete(&models.Interaction{})
	return err
}

// ExportInteractions exports interaction data in specified format
func (s *Service) ExportInteractions(caseID, format string, includeRaw bool) (string, error) {
	var interactions []models.Interaction
	session := s.engine.NewSession()
	defer session.Close()

	if caseID != "" {
		session = session.Where("case_id = ?", caseID)
	}

	err := session.OrderBy("timestamp ASC").Find(&interactions)
	if err != nil {
		return "", err
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(interactions, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "csv":
		return s.exportToCSV(interactions, includeRaw)
	case "markdown":
		return s.exportToMarkdown(interactions, includeRaw)
	default:
		return "", errors.New("unsupported format")
	}
}

// GetTimeline retrieves interactions as a timeline grouped by time intervals
func (s *Service) GetTimeline(caseID, payloadID string, startTime, endTime *time.Time, interval string) (*TimelineResponse, error) {
	var interactions []models.Interaction
	session := s.engine.NewSession()
	defer session.Close()

	if caseID != "" {
		session = session.Where("case_id = ?", caseID)
	}
	if payloadID != "" {
		session = session.Where("payload_id = ?", payloadID)
	}
	if startTime != nil {
		session = session.Where("timestamp >= ?", startTime)
	}
	if endTime != nil {
		session = session.Where("timestamp <= ?", endTime)
	}

	if err := session.OrderBy("timestamp ASC").Find(&interactions); err != nil {
		return nil, err
	}

	// Group interactions by time interval
	timeline := &TimelineResponse{
		Total:         int64(len(interactions)),
		Interactions:  interactions,
		GroupedEvents: s.groupByInterval(interactions, interval),
	}

	return timeline, nil
}

// groupByInterval groups interactions by time interval
func (s *Service) groupByInterval(interactions []models.Interaction, interval string) []TimelineGroup {
	if len(interactions) == 0 {
		return []TimelineGroup{}
	}

	groups := make(map[string][]models.Interaction)

	for _, interaction := range interactions {
		key := s.getIntervalKey(interaction.Timestamp, interval)
		groups[key] = append(groups[key], interaction)
	}

	var result []TimelineGroup
	for key, items := range groups {
		result = append(result, TimelineGroup{
			Time:         key,
			Count:        len(items),
			Interactions: items,
		})
	}

	return result
}

// getIntervalKey returns the time interval key for grouping
func (s *Service) getIntervalKey(t time.Time, interval string) string {
	switch interval {
	case "hour":
		return t.Format("2006-01-02 15:00")
	case "day":
		return t.Format("2006-01-02")
	case "week":
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case "month":
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02 15:04")
	}
}

// exportToCSV exports interactions to CSV format
func (s *Service) exportToCSV(interactions []models.Interaction, includeRaw bool) (string, error) {
	csv := "ID,Type,Token,Timestamp,SourceIP,Domain,Method,Path\n"
	for _, i := range interactions {
		domain := ""
		if i.Domain != nil {
			domain = *i.Domain
		}
		method := ""
		if i.Method != nil {
			method = *i.Method
		}
		path := ""
		if i.Path != nil {
			path = *i.Path
		}
		token := ""
		if i.Token != nil {
			token = *i.Token
		}

		csv += i.ID + "," + i.Type + "," + token + "," + i.Timestamp.Format(time.RFC3339) + "," + i.SourceIP + "," + domain + "," + method + "," + path + "\n"
	}
	return csv, nil
}

// exportToMarkdown exports interactions to Markdown format
func (s *Service) exportToMarkdown(interactions []models.Interaction, includeRaw bool) (string, error) {
	md := "# Interactions Report\n\n"
	md += "Generated at: " + time.Now().Format(time.RFC3339) + "\n\n"
	md += "Total: " + string(rune(len(interactions))) + " interactions\n\n"

	for _, i := range interactions {
		md += "## " + i.Type + " Interaction\n"
		md += "- **ID**: " + i.ID + "\n"
		md += "- **Timestamp**: " + i.Timestamp.Format(time.RFC3339) + "\n"
		md += "- **Source IP**: " + i.SourceIP + "\n"

		if i.Token != nil {
			md += "- **Token**: " + *i.Token + "\n"
		}
		if i.Domain != nil {
			md += "- **Domain**: " + *i.Domain + "\n"
		}
		if i.Method != nil {
			md += "- **Method**: " + *i.Method + "\n"
		}
		if i.Path != nil {
			md += "- **Path**: " + *i.Path + "\n"
		}
		if includeRaw && len(i.RawData) > 0 {
			md += "- **Raw Data**:\n```\n" + i.RawData + "\n```\n"
		}
		md += "\n"
	}

	return md, nil
}
