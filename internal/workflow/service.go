package workflow

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/internal/models"
)

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
)

// Service provides workflow management services
type Service struct {
	engine   *xorm.Engine
	security *OutboundSecurity
}

// NewService creates a new workflow service
func NewService(engine *xorm.Engine) *Service {
	return &Service{engine: engine}
}

// SetOutboundSecurity sets the outbound security policy for action executors.
func (s *Service) SetOutboundSecurity(sec *OutboundSecurity) {
	s.security = sec
}

// CreateWorkflow creates a new workflow
func (s *Service) CreateWorkflow(workflow *models.Workflow) error {
	if workflow.ID == "" {
		workflow.ID = models.GenerateID()
	}
	if workflow.Actions == nil {
		workflow.Actions = models.Actions{}
	}
	if workflow.CreatedAt.IsZero() {
		workflow.CreatedAt = time.Now()
	}
	if workflow.UpdatedAt.IsZero() {
		workflow.UpdatedAt = time.Now()
	}

	_, err := s.engine.Insert(workflow)
	return err
}

// GetWorkflowByID retrieves a workflow by its ID
func (s *Service) GetWorkflowByID(id string) (*models.Workflow, error) {
	var workflow models.Workflow
	has, err := s.engine.ID(id).Get(&workflow)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrWorkflowNotFound
	}
	return &workflow, nil
}

// ListWorkflows retrieves workflows with filtering
func (s *Service) ListWorkflows(caseID string, enabled *bool, page, pageSize int) (*models.WorkflowListResponse, error) {
	var workflows []models.Workflow
	session := s.engine.NewSession()
	defer session.Close()

	if caseID != "" {
		session = session.Where("case_id = ?", caseID)
	}
	if enabled != nil {
		session = session.Where("enabled = ?", *enabled)
	}

	total, err := session.Count(&models.Workflow{})
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := session.Desc("created_at").Limit(pageSize, offset).Find(&workflows); err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.WorkflowListResponse{
		Items:      workflows,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateWorkflow updates an existing workflow
func (s *Service) UpdateWorkflow(workflow *models.Workflow) error {
	workflow.UpdatedAt = time.Now()

	_, err := s.engine.ID(workflow.ID).Update(workflow)
	return err
}

// DeleteWorkflow deletes a workflow by its ID
func (s *Service) DeleteWorkflow(id string) error {
	_, err := s.engine.ID(id).Delete(&models.Workflow{})
	return err
}

// ExecuteWorkflow executes a workflow's actions
func (s *Service) ExecuteWorkflow(workflowID string, interaction *models.Interaction) error {
	workflow, err := s.GetWorkflowByID(workflowID)
	if err != nil {
		return err
	}

	if !workflow.Enabled {
		return nil // Skip disabled workflows
	}

	// Execute each action in the workflow
	for _, action := range workflow.Actions {
		if !action.Enabled {
			continue
		}

		if err := s.executeAction(action, interaction); err != nil {
			// Log error but continue with other actions
			continue
		}
	}

	return nil
}

// executeAction executes a single action
func (s *Service) executeAction(action models.Action, interaction *models.Interaction) error {
	switch action.Type {
	case models.ActionTypeHTTP:
		return s.executeHTTPAction(action, interaction)
	case models.ActionTypeDNS:
		return s.executeDNSAction(action, interaction)
	case models.ActionTypeSMTP:
		return s.executeSMTPAction(action, interaction)
	case models.ActionTypeWebhook:
		return s.executeWebhookAction(action, interaction)
	case models.ActionTypeNotify:
		return s.executeNotifyAction(action, interaction)
	default:
		return errors.New("unsupported action type")
	}
}

// executeHTTPAction executes an HTTP action with outbound security constraints.
// Config fields: url (required), method (default GET), body, headers (map[string]string).
func (s *Service) executeHTTPAction(action models.Action, interaction *models.Interaction) error {
	urlStr, _ := action.Config["url"].(string)
	if urlStr == "" {
		return errors.New("missing url in action config")
	}

	method, _ := action.Config["method"].(string)
	if method == "" {
		method = "GET"
	}

	if s.security == nil {
		s.security = NewOutboundSecurity(nil)
	}
	actionID := action.ID
	if actionID == "" {
		actionID = urlStr
	}
	if err := s.security.ValidateURL(actionID, urlStr); err != nil {
		return err
	}

	bodyStr, _ := action.Config["body"].(string)
	var bodyReader *bytes.Buffer
	if bodyStr != "" {
		bodyReader = bytes.NewBufferString(bodyStr)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if headers, ok := action.Config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	if bodyStr != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP action request failed: %w", err)
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, 1<<20)
	respBody, _ := io.ReadAll(limitedReader)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP action returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// executeDNSAction executes a DNS lookup action on hit.
// Config fields: domain (required), type (default A; supports A, TXT, MX, NS).
func (s *Service) executeDNSAction(action models.Action, interaction *models.Interaction) error {
	domain, _ := action.Config["domain"].(string)
	if domain == "" {
		return errors.New("missing domain in DNS action config")
	}

	recordType, _ := action.Config["type"].(string)
	if recordType == "" {
		recordType = "A"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resolver := net.Resolver{}
	switch recordType {
	case "A":
		_, err := resolver.LookupHost(ctx, domain)
		if err != nil {
			return fmt.Errorf("DNS A lookup failed for %s: %w", domain, err)
		}
	case "TXT":
		_, err := resolver.LookupTXT(ctx, domain)
		if err != nil {
			return fmt.Errorf("DNS TXT lookup failed for %s: %w", domain, err)
		}
	case "MX":
		_, err := resolver.LookupMX(ctx, domain)
		if err != nil {
			return fmt.Errorf("DNS MX lookup failed for %s: %w", domain, err)
		}
	case "NS":
		_, err := resolver.LookupNS(ctx, domain)
		if err != nil {
			return fmt.Errorf("DNS NS lookup failed for %s: %w", domain, err)
		}
	default:
		return fmt.Errorf("unsupported DNS record type: %s", recordType)
	}

	return nil
}

// executeSMTPAction sends an email via SMTP on interaction hit.
// Config fields: smtp_host (required), smtp_port (default 587), username, password,
// from (required), to (required, comma-separated), subject (default "OAST Alert"),
// message (template with {{.field}} placeholders).
func (s *Service) executeSMTPAction(action models.Action, interaction *models.Interaction) error {
	smtpHost, _ := action.Config["smtp_host"].(string)
	if smtpHost == "" {
		return errors.New("missing smtp_host in SMTP action config")
	}

	smtpPort, _ := action.Config["smtp_port"].(string)
	if smtpPort == "" {
		smtpPort = "587"
	}

	from, _ := action.Config["from"].(string)
	if from == "" {
		return errors.New("missing from in SMTP action config")
	}

	toStr, _ := action.Config["to"].(string)
	if toStr == "" {
		return errors.New("missing to in SMTP action config")
	}
	toList := strings.Split(toStr, ",")
	for i := range toList {
		toList[i] = strings.TrimSpace(toList[i])
	}

	subject, _ := action.Config["subject"].(string)
	if subject == "" {
		subject = "OAST Alert"
	}

	messageTemplate, _ := action.Config["message"].(string)
	if messageTemplate == "" {
		messageTemplate = "OAST interaction: {{.id}} from {{.source_ip}}"
	}
	message := renderActionTemplate(messageTemplate, interaction)

	username, _ := action.Config["username"].(string)
	password, _ := action.Config["password"].(string)

	body := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		from, toStr, subject, message)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, smtpHost)
	}

	err := smtp.SendMail(addr, auth, from, toList, []byte(body))
	if err != nil {
		return fmt.Errorf("failed to send SMTP action email: %w", err)
	}
	return nil
}

// executeWebhookAction executes a webhook action with template rendering and security.
// Config fields: url (required), method (default POST), headers, body (supports {{.field}} placeholders).
func (s *Service) executeWebhookAction(action models.Action, interaction *models.Interaction) error {
	urlStr, _ := action.Config["url"].(string)
	if urlStr == "" {
		return errors.New("missing url in webhook config")
	}

	method, _ := action.Config["method"].(string)
	if method == "" {
		method = "POST"
	}

	if s.security == nil {
		s.security = NewOutboundSecurity(nil)
	}
	actionID := action.ID
	if actionID == "" {
		actionID = urlStr
	}
	if err := s.security.ValidateURL(actionID, urlStr); err != nil {
		return err
	}

	bodyTemplate, _ := action.Config["body"].(string)
	renderedBody := renderActionTemplate(bodyTemplate, interaction)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bytes.NewBufferString(renderedBody))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	if headers, ok := action.Config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// renderActionTemplate renders {{.field}} placeholders with interaction data.
// Supported fields: id, type, source_ip, token, domain, path.
func renderActionTemplate(template string, inter *models.Interaction) string {
	data := map[string]string{
		"id":        inter.ID,
		"type":      inter.Type,
		"source_ip": inter.SourceIP,
	}
	if inter.Token != nil {
		data["token"] = *inter.Token
	}
	if inter.Domain != nil {
		data["domain"] = *inter.Domain
	}
	if inter.Path != nil {
		data["path"] = *inter.Path
	}
	result := template
	for k, v := range data {
		result = strings.ReplaceAll(result, fmt.Sprintf("{{.%s}}", k), v)
	}
	return result
}

// executeNotifyAction triggers a notification via various channels.
// Config fields: channel (required), webhook_url (for webhook/feishu/wecom/dingtalk/slack/discord),
// bot_token + chat_id (for telegram), message (template with {{.field}} placeholders).
func (s *Service) executeNotifyAction(action models.Action, interaction *models.Interaction) error {
	channel, _ := action.Config["channel"].(string)
	if len(channel) == 0 {
		return errors.New("missing channel in notify action config")
	}

	messageTemplate, _ := action.Config["message"].(string)
	if messageTemplate == "" {
		messageTemplate = "OAST interaction: {{.id}} from {{.source_ip}}"
	}
	message := renderActionTemplate(messageTemplate, interaction)

	switch channel {
	case "webhook":
		return s.sendNotifyWebhook(action, message)
	case "feishu":
		return s.sendNotifyFeishu(action, message)
	case "wecom":
		return s.sendNotifyWecom(action, message)
	case "dingtalk":
		return s.sendNotifyDingtalk(action, message)
	case "slack":
		return s.sendNotifySlack(action, message)
	case "discord":
		return s.sendNotifyDiscord(action, message)
	case "telegram":
		return s.sendNotifyTelegram(action, message)
	case "email":
		return s.sendNotifyEmail(action, message)
	default:
		return fmt.Errorf("unsupported notify channel: %s", channel)
	}
}

// sendNotifyWebhook sends a generic webhook notification
func (s *Service) sendNotifyWebhook(action models.Action, message string) error {
	webhookURL, _ := action.Config["webhook_url"].(string)
	if len(webhookURL) == 0 {
		return errors.New("missing webhook_url for webhook notify channel")
	}
	if err := s.validateNotifyURL(action.ID, webhookURL); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]string{"message": message})
	return s.sendNotifyRequest(webhookURL, payload)
}

// sendNotifyFeishu sends a Feishu (Lark) notification
func (s *Service) sendNotifyFeishu(action models.Action, message string) error {
	webhookURL, _ := action.Config["webhook_url"].(string)
	if len(webhookURL) == 0 {
		return errors.New("missing webhook_url for feishu notify channel")
	}
	if err := s.validateNotifyURL(action.ID, webhookURL); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"msg_type": "text",
		"content":  map[string]string{"text": message},
	})
	return s.sendNotifyRequest(webhookURL, payload)
}

// sendNotifyWecom sends a WeChat Work notification
func (s *Service) sendNotifyWecom(action models.Action, message string) error {
	webhookURL, _ := action.Config["webhook_url"].(string)
	if len(webhookURL) == 0 {
		return errors.New("missing webhook_url for wecom notify channel")
	}
	if err := s.validateNotifyURL(action.ID, webhookURL); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": message},
	})
	return s.sendNotifyRequest(webhookURL, payload)
}

// sendNotifyDingtalk sends a DingTalk notification
func (s *Service) sendNotifyDingtalk(action models.Action, message string) error {
	webhookURL, _ := action.Config["webhook_url"].(string)
	if len(webhookURL) == 0 {
		return errors.New("missing webhook_url for dingtalk notify channel")
	}
	if err := s.validateNotifyURL(action.ID, webhookURL); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": message},
	})
	return s.sendNotifyRequest(webhookURL, payload)
}

// sendNotifySlack sends a Slack notification
func (s *Service) sendNotifySlack(action models.Action, message string) error {
	webhookURL, _ := action.Config["webhook_url"].(string)
	if len(webhookURL) == 0 {
		return errors.New("missing webhook_url for slack notify channel")
	}
	if err := s.validateNotifyURL(action.ID, webhookURL); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]string{"text": message})
	return s.sendNotifyRequest(webhookURL, payload)
}

// sendNotifyDiscord sends a Discord notification
func (s *Service) sendNotifyDiscord(action models.Action, message string) error {
	webhookURL, _ := action.Config["webhook_url"].(string)
	if len(webhookURL) == 0 {
		return errors.New("missing webhook_url for discord notify channel")
	}
	if err := s.validateNotifyURL(action.ID, webhookURL); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]string{"content": message})
	return s.sendNotifyRequest(webhookURL, payload)
}

// sendNotifyTelegram sends a Telegram notification
func (s *Service) sendNotifyTelegram(action models.Action, message string) error {
	botToken, _ := action.Config["bot_token"].(string)
	if len(botToken) == 0 {
		return errors.New("missing bot_token for telegram notify channel")
	}
	chatID, _ := action.Config["chat_id"].(string)
	if len(chatID) == 0 {
		return errors.New("missing chat_id for telegram notify channel")
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	if err := s.validateNotifyURL(action.ID, url); err != nil {
		return fmt.Errorf("telegram URL validation failed: %w", err)
	}
	payload, _ := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    message,
	})
	return s.sendNotifyRequest(url, payload)
}

// sendNotifyEmail sends an email notification via SMTP.
// Config fields: smtp_host (required), smtp_port (default 587), username, password,
// from (required), to (required, comma-separated), subject (default "OAST Alert"),
// use_tls (default true, set false to disable STARTTLS requirement).
func (s *Service) sendNotifyEmail(action models.Action, message string) error {
	smtpHost, _ := action.Config["smtp_host"].(string)
	if len(smtpHost) == 0 {
		return errors.New("missing smtp_host for email notify channel")
	}

	smtpPort, _ := action.Config["smtp_port"].(string)
	if smtpPort == "" {
		smtpPort = "587"
	}

	from, _ := action.Config["from"].(string)
	if len(from) == 0 {
		return errors.New("missing from for email notify channel")
	}

	toStr, _ := action.Config["to"].(string)
	if len(toStr) == 0 {
		return errors.New("missing to for email notify channel")
	}
	toList := strings.Split(toStr, ",")
	for i := range toList {
		toList[i] = strings.TrimSpace(toList[i])
	}

	subject, _ := action.Config["subject"].(string)
	if subject == "" {
		subject = "OAST Alert"
	}

	username, _ := action.Config["username"].(string)
	password, _ := action.Config["password"].(string)

	useTLS, _ := action.Config["use_tls"].(bool)
	if !useTLS {
		// Default to requiring TLS unless explicitly disabled
		useTLS = true
	}

	body := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		from, toStr, subject, message)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, smtpHost)
	}

	if useTLS {
		if err := s.sendEmailWithTLS(addr, auth, from, toList, []byte(body)); err != nil {
			return fmt.Errorf("failed to send email via TLS: %w", err)
		}
	} else {
		if err := smtp.SendMail(addr, auth, from, toList, []byte(body)); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}
	return nil
}

// sendEmailWithTLS sends an email with mandatory STARTTLS upgrade.
// If the server does not support STARTTLS, the connection is aborted.
func (s *Service) sendEmailWithTLS(addr string, auth smtp.Auth, from string, to []string, body []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, strings.Split(addr, ":")[0])
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer c.Close()

	if err = c.Hello("localhost"); err != nil {
		return fmt.Errorf("HELO failed: %w", err)
	}

	// Check if STARTTLS is supported
	if ok, _ := c.Extension("STARTTLS"); !ok {
		return errors.New("SMTP server does not support STARTTLS, refusing to send without TLS")
	}

	if err = c.StartTLS(&tls.Config{
		ServerName:         strings.Split(addr, ":")[0],
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	if auth != nil {
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	if err = c.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return fmt.Errorf("RCPT TO failed for %s: %w", addr, err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}
	if _, err = w.Write(body); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close DATA: %w", err)
	}

	return c.Quit()
}

// validateNotifyURL validates a URL against the outbound security policy
func (s *Service) validateNotifyURL(actionID, urlStr string) error {
	if s.security == nil {
		s.security = NewOutboundSecurity(nil)
	}
	if len(actionID) == 0 {
		actionID = urlStr
	}
	return s.security.ValidateURL(actionID, urlStr)
}

// sendNotifyRequest sends an HTTP POST with JSON payload for notifications
func (s *Service) sendNotifyRequest(url string, payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create notify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("notify request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("notify returned status %d", resp.StatusCode)
	}
	return nil
}
