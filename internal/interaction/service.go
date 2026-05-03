package interaction

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"time"

	"xorm.io/xorm"
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
func (s *Service) CreateInteraction(interaction *Interaction) error {
	if interaction.ID == "" {
		interaction.ID = generateID()
	}
	if interaction.Timestamp.IsZero() {
		interaction.Timestamp = time.Now()
	}
	
	_, err := s.engine.Insert(interaction)
	return err
}

// GetInteractionByID retrieves an interaction by its ID
func (s *Service) GetInteractionByID(id string) (*Interaction, error) {
	var interaction Interaction
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
func (s *Service) ListInteractions(caseID, payloadID, interactionType string, startTime, endTime *time.Time, page, pageSize int) (*InteractionListResponse, error) {
	var interactions []Interaction
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
	
	total, err := session.Count(&Interaction{})
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
	
	return &InteractionListResponse{
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
	
	_, err := s.engine.In("id", ids).Delete(&Interaction{})
	return err
}

// ExportInteractions exports interactions to specified format
func (s *Service) ExportInteractions(req *ExportRequest) (string, error) {
	var interactions []Interaction
	session := s.engine.NewSession()
	defer session.Close()
	
	if req.CaseID != nil {
		session = session.Where("case_id = ?", *req.CaseID)
	}
	if req.PayloadID != nil {
		session = session.Where("payload_id = ?", *req.PayloadID)
	}
	if req.StartTime != nil {
		session = session.Where("timestamp >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		session = session.Where("timestamp <= ?", req.EndTime)
	}
	
	if err := session.Find(&interactions); err != nil {
		return "", err
	}
	
	switch req.Format {
	case "json":
		data, err := json.MarshalIndent(interactions, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "csv":
		return s.exportToCSV(interactions, req.IncludeRaw)
	case "markdown":
		return s.exportToMarkdown(interactions, req.IncludeRaw)
	default:
		return "", errors.New("unsupported format")
	}
}

// exportToCSV exports interactions to CSV format
func (s *Service) exportToCSV(interactions []Interaction, includeRaw bool) (string, error) {
	// Simple CSV implementation
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
func (s *Service) exportToMarkdown(interactions []Interaction, includeRaw bool) (string, error) {
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

// generateID generates a unique ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}
