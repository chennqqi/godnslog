package rule

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExecuteNotification_Feishu verifies Feishu notification sends a POST with the expected payload.
func TestExecuteNotification_Feishu(t *testing.T) {
	var receivedMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exec := NewExecutor()
	notif := Notification{
		Type:     "feishu",
		Template: "Alert: {{.type}} from {{.source_ip}}",
		Config:   map[string]interface{}{"webhook_url": srv.URL},
	}
	inter := map[string]interface{}{"type": "dns", "source_ip": "10.0.0.1"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.NoError(t, err)
	assert.Equal(t, "POST", receivedMethod)
}

// TestExecuteNotification_Wecom verifies WeCom notification sends successfully.
func TestExecuteNotification_Wecom(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exec := NewExecutor()
	notif := Notification{
		Type:     "wecom",
		Template: "Alert: {{.type}}",
		Config:   map[string]interface{}{"webhook_url": srv.URL},
	}
	inter := map[string]interface{}{"type": "http"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.NoError(t, err)
}

// TestExecuteNotification_Dingtalk verifies DingTalk notification sends successfully.
func TestExecuteNotification_Dingtalk(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exec := NewExecutor()
	notif := Notification{
		Type:     "dingtalk",
		Template: "Alert: {{.type}}",
		Config:   map[string]interface{}{"webhook_url": srv.URL},
	}
	inter := map[string]interface{}{"type": "ldap"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.NoError(t, err)
}

// TestExecuteNotification_Slack verifies Slack notification sends successfully.
func TestExecuteNotification_Slack(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exec := NewExecutor()
	notif := Notification{
		Type:     "slack",
		Template: "Alert: {{.type}}",
		Config:   map[string]interface{}{"webhook_url": srv.URL},
	}
	inter := map[string]interface{}{"type": "rmi"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.NoError(t, err)
}

// TestExecuteNotification_Discord verifies Discord notification sends successfully.
func TestExecuteNotification_Discord(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exec := NewExecutor()
	notif := Notification{
		Type:     "discord",
		Template: "Alert: {{.type}}",
		Config:   map[string]interface{}{"webhook_url": srv.URL},
	}
	inter := map[string]interface{}{"type": "dns"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.NoError(t, err)
}

// TestExecuteNotification_Telegram_MissingBotToken verifies error when bot_token is missing.
func TestExecuteNotification_Telegram_MissingBotToken(t *testing.T) {
	exec := NewExecutor()
	notif := Notification{
		Type:     "telegram",
		Template: "Alert",
		Config:   map[string]interface{}{"chat_id": "123"},
	}
	inter := map[string]interface{}{"type": "http"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing bot_token")
}

// TestExecuteNotification_Telegram_MissingChatID verifies error when chat_id is missing.
func TestExecuteNotification_Telegram_MissingChatID(t *testing.T) {
	exec := NewExecutor()
	notif := Notification{
		Type:     "telegram",
		Template: "Alert",
		Config:   map[string]interface{}{"bot_token": "token123"},
	}
	inter := map[string]interface{}{"type": "http"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing chat_id")
}

// TestExecuteNotification_Webhook verifies generic webhook notification sends successfully.
func TestExecuteNotification_Webhook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exec := NewExecutor()
	notif := Notification{
		Type:     "webhook",
		Template: "Alert: {{.type}} from {{.source_ip}}",
		Config:   map[string]interface{}{"webhook_url": srv.URL},
	}
	inter := map[string]interface{}{"type": "dns", "source_ip": "10.0.0.1"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.NoError(t, err)
}

// TestExecuteNotification_Webhook_HTTPError verifies error on non-200 response.
func TestExecuteNotification_Webhook_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	exec := NewExecutor()
	notif := Notification{
		Type:     "webhook",
		Template: "Alert",
		Config:   map[string]interface{}{"webhook_url": srv.URL},
	}
	inter := map[string]interface{}{"type": "dns"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP request failed")
}

// TestExecuteNotification_MissingWebhookURL verifies error when webhook_url is missing.
func TestExecuteNotification_MissingWebhookURL(t *testing.T) {
	exec := NewExecutor()
	notif := Notification{
		Type:     "feishu",
		Template: "Alert",
		Config:   map[string]interface{}{},
	}
	inter := map[string]interface{}{"type": "dns"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing webhook_url")
}

// TestExecuteNotification_UnsupportedType verifies error for unsupported notification type.
func TestExecuteNotification_UnsupportedType(t *testing.T) {
	exec := NewExecutor()
	notif := Notification{
		Type:     "unsupported",
		Template: "Alert",
		Config:   map[string]interface{}{},
	}
	inter := map[string]interface{}{"type": "dns"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported notification type")
}

// TestExecuteNotification_Email_MissingSMTPHost verifies error when smtp_host is missing.
func TestExecuteNotification_Email_MissingSMTPHost(t *testing.T) {
	exec := NewExecutor()
	notif := Notification{
		Type:     "email",
		Template: "Alert",
		Config:   map[string]interface{}{},
	}
	inter := map[string]interface{}{"type": "dns"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing smtp_host")
}

// TestExecuteNotification_Email_MissingFrom verifies error when from is missing.
func TestExecuteNotification_Email_MissingFrom(t *testing.T) {
	exec := NewExecutor()
	notif := Notification{
		Type:     "email",
		Template: "Alert",
		Config:   map[string]interface{}{"smtp_host": "localhost"},
	}
	inter := map[string]interface{}{"type": "dns"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing from")
}

// TestExecuteNotification_Email_MissingTo verifies error when to is missing.
func TestExecuteNotification_Email_MissingTo(t *testing.T) {
	exec := NewExecutor()
	notif := Notification{
		Type:     "email",
		Template: "Alert",
		Config:   map[string]interface{}{"smtp_host": "localhost", "from": "a@b.com"},
	}
	inter := map[string]interface{}{"type": "dns"}

	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing to")
}

// TestExecuteWebhook verifies webhook forwarding action.
func TestExecuteWebhook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exec := NewExecutor()
	wh := Webhook{
		URL:    srv.URL,
		Method: "POST",
		Body:   `{"event":"{{.type}}"}`,
	}
	inter := map[string]interface{}{"type": "dns"}

	err := exec.executeWebhook(context.Background(), wh, inter)
	assert.NoError(t, err)
}

// TestExecuteWebhook_PlainText verifies webhook with non-JSON body sends as plain text.
func TestExecuteWebhook_PlainText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exec := NewExecutor()
	wh := Webhook{
		URL:    srv.URL,
		Method: "POST",
		Body:   `plain text event={{.type}}`,
	}
	inter := map[string]interface{}{"type": "http"}

	err := exec.executeWebhook(context.Background(), wh, inter)
	assert.NoError(t, err)
}

// TestSendHTTP_InvalidURL verifies error handling for unreachable URL.
func TestSendHTTP_InvalidURL(t *testing.T) {
	exec := NewExecutor()
	err := exec.sendHTTP(context.Background(), "http://127.0.0.1:1", "GET", nil, "test")
	assert.Error(t, err)
}
