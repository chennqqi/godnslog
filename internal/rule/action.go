package rule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Executor executes rule actions
type Executor struct {
	httpClient *http.Client
}

// NewExecutor creates a new action executor
func NewExecutor() *Executor {
	return &Executor{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute executes all actions for a matched rule
func (e *Executor) Execute(ctx context.Context, rule *Rule, inter map[string]interface{}) error {
	actions := rule.Actions

	// Discard noise action
	if actions.DiscardNoise {
		// Mark interaction as noise (implementation depends on storage)
		// This is a placeholder for noise filtering
	}

	// Execute notifications
	for _, notif := range actions.Notifications {
		if err := e.executeNotification(ctx, notif, inter); err != nil {
			return fmt.Errorf("notification failed: %w", err)
		}
	}

	// Execute tag actions
	for _, tag := range actions.Tags {
		if err := e.executeTagAction(ctx, tag, inter); err != nil {
			return fmt.Errorf("tag action failed: %w", err)
		}
	}

	// Execute webhooks
	for _, webhook := range actions.Webhooks {
		if err := e.executeWebhook(ctx, webhook, inter); err != nil {
			return fmt.Errorf("webhook failed: %w", err)
		}
	}

	// Execute report generation
	for _, report := range actions.Reports {
		if err := e.executeReport(ctx, report, inter); err != nil {
			return fmt.Errorf("report generation failed: %w", err)
		}
	}

	return nil
}

// executeNotification sends a notification
func (e *Executor) executeNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	switch notif.Type {
	case "feishu":
		return e.sendFeishuNotification(ctx, notif, inter)
	case "wecom":
		return e.sendWecomNotification(ctx, notif, inter)
	case "dingtalk":
		return e.sendDingtalkNotification(ctx, notif, inter)
	case "slack":
		return e.sendSlackNotification(ctx, notif, inter)
	case "discord":
		return e.sendDiscordNotification(ctx, notif, inter)
	case "telegram":
		return e.sendTelegramNotification(ctx, notif, inter)
	case "email":
		return e.sendEmailNotification(ctx, notif, inter)
	case "webhook":
		return e.sendWebhookNotification(ctx, notif, inter)
	default:
		return fmt.Errorf("unsupported notification type: %s", notif.Type)
	}
}

// sendFeishuNotification sends a Feishu notification
func (e *Executor) sendFeishuNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	webhookURL, ok := notif.Config["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("missing webhook_url in config")
	}

	message := e.renderTemplate(notif.Template, inter)
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": message,
		},
	}

	return e.sendHTTP(ctx, webhookURL, "POST", nil, payload)
}

// sendWecomNotification sends a WeCom notification
func (e *Executor) sendWecomNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	webhookURL, ok := notif.Config["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("missing webhook_url in config")
	}

	message := e.renderTemplate(notif.Template, inter)
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	return e.sendHTTP(ctx, webhookURL, "POST", nil, payload)
}

// sendDingtalkNotification sends a DingTalk notification
func (e *Executor) sendDingtalkNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	webhookURL, ok := notif.Config["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("missing webhook_url in config")
	}

	message := e.renderTemplate(notif.Template, inter)
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	return e.sendHTTP(ctx, webhookURL, "POST", nil, payload)
}

// sendSlackNotification sends a Slack notification
func (e *Executor) sendSlackNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	webhookURL, ok := notif.Config["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("missing webhook_url in config")
	}

	message := e.renderTemplate(notif.Template, inter)
	payload := map[string]string{
		"text": message,
	}

	return e.sendHTTP(ctx, webhookURL, "POST", nil, payload)
}

// sendDiscordNotification sends a Discord notification
func (e *Executor) sendDiscordNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	webhookURL, ok := notif.Config["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("missing webhook_url in config")
	}

	message := e.renderTemplate(notif.Template, inter)
	payload := map[string]string{
		"content": message,
	}

	return e.sendHTTP(ctx, webhookURL, "POST", nil, payload)
}

// sendTelegramNotification sends a Telegram notification
func (e *Executor) sendTelegramNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	botToken, ok := notif.Config["bot_token"].(string)
	if !ok {
		return fmt.Errorf("missing bot_token in config")
	}
	chatID, ok := notif.Config["chat_id"].(string)
	if !ok {
		return fmt.Errorf("missing chat_id in config")
	}

	message := e.renderTemplate(notif.Template, inter)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]string{
		"chat_id": chatID,
		"text":    message,
	}

	return e.sendHTTP(ctx, url, "POST", nil, payload)
}

// sendEmailNotification sends an email notification
func (e *Executor) sendEmailNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	// Email implementation requires SMTP configuration
	// This is a placeholder for email notification
	return fmt.Errorf("email notification not implemented")
}

// sendWebhookNotification sends a generic webhook notification
func (e *Executor) sendWebhookNotification(ctx context.Context, notif Notification, inter map[string]interface{}) error {
	webhookURL, ok := notif.Config["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("missing webhook_url in config")
	}

	message := e.renderTemplate(notif.Template, inter)
	payload := map[string]string{
		"message": message,
	}

	return e.sendHTTP(ctx, webhookURL, "POST", nil, payload)
}

// executeTagAction executes tag actions
func (e *Executor) executeTagAction(ctx context.Context, tag TagAction, inter map[string]interface{}) error {
	// Tag actions require interaction with the interaction storage
	// This is a placeholder for tag action implementation
	// In production, this would update the interaction's tags in the database
	return nil
}

// executeWebhook executes a webhook forwarding action
func (e *Executor) executeWebhook(ctx context.Context, webhook Webhook, inter map[string]interface{}) error {
	body := e.renderTemplate(webhook.Body, inter)
	
	var payload interface{}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		// If not JSON, send as plain text
		payload = body
	}

	return e.sendHTTP(ctx, webhook.URL, webhook.Method, webhook.Headers, payload)
}

// executeReport executes a report generation action
func (e *Executor) executeReport(ctx context.Context, report Report, inter map[string]interface{}) error {
	// Report generation requires interaction with the evidence export system
	// This is a placeholder for report generation
	return nil
}

// sendHTTP sends an HTTP request
func (e *Executor) sendHTTP(ctx context.Context, url, method string, headers map[string]string, body interface{}) error {
	var bodyReader *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	return nil
}

// renderTemplate renders a template with interaction data
func (e *Executor) renderTemplate(template string, data map[string]interface{}) string {
	result := template
	for k, v := range data {
		placeholder := fmt.Sprintf("{{.%s}}", k)
		value := fmt.Sprintf("%v", v)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
