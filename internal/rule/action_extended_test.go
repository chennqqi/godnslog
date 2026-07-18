package rule

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecute_MultipleActions(t *testing.T) {
	exec := NewExecutor()
	rule := &Rule{
		Actions: Actions{
			DiscardNoise: true,
			Tags: []TagAction{
				{Add: []string{"malicious", "high-priority"}},
			},
			Reports: []Report{
				{Format: "json", Title: "Test"},
			},
		},
	}
	inter := map[string]interface{}{
		"id":   "interaction-1",
		"type": "dns",
	}

	err := exec.Execute(context.Background(), rule, inter)
	assert.NoError(t, err)

	// Verify noise flag
	assert.True(t, inter["_noise"].(bool))

	// Verify tags
	tags, ok := inter["_tags"].([]string)
	assert.True(t, ok)
	assert.Contains(t, tags, "malicious")
	assert.Contains(t, tags, "high-priority")

	// Verify report
	assert.Contains(t, inter, "_report")
}

func TestExecute_TagAndReportOnly(t *testing.T) {
	exec := NewExecutor()
	rule := &Rule{
		Actions: Actions{
			Tags: []TagAction{
				{Add: []string{"tag-only"}},
			},
			Reports: []Report{
				{Format: "markdown", Title: "My Report"},
			},
		},
	}
	inter := map[string]interface{}{"id": "inter-2", "source_ip": "10.0.0.1"}

	err := exec.Execute(context.Background(), rule, inter)
	assert.NoError(t, err)

	tags, ok := inter["_tags"].([]string)
	assert.True(t, ok)
	assert.Equal(t, "tag-only", tags[0])

	_, reportOk := inter["_report"]
	assert.True(t, reportOk)

	// Should not have noise flag
	_, noiseOk := inter["_noise"]
	assert.False(t, noiseOk)
}

func TestExecute_AddDuplicateTags(t *testing.T) {
	exec := NewExecutor()
	rule := &Rule{
		Actions: Actions{
			Tags: []TagAction{
				{Add: []string{"dup-tag"}},
			},
		},
	}
	inter := map[string]interface{}{
		"id":    "inter-3",
		"_tags": []string{"dup-tag"},
	}

	err := exec.Execute(context.Background(), rule, inter)
	assert.NoError(t, err)

	tags := inter["_tags"].([]string)
	assert.Len(t, tags, 1)
}

func TestExecute_AddAndRemoveTags(t *testing.T) {
	exec := NewExecutor()
	rule := &Rule{
		Actions: Actions{
			Tags: []TagAction{
				{Remove: []string{"old-tag"}},
				{Add: []string{"new-tag"}},
			},
		},
	}
	inter := map[string]interface{}{
		"id":    "inter-4",
		"_tags": []string{"old-tag", "keep-tag"},
	}

	err := exec.Execute(context.Background(), rule, inter)
	assert.NoError(t, err)

	tags := inter["_tags"].([]string)
	assert.Contains(t, tags, "new-tag")
	assert.Contains(t, tags, "keep-tag")
	assert.NotContains(t, tags, "old-tag")
}

func TestExecute_EmptyActions(t *testing.T) {
	exec := NewExecutor()
	rule := &Rule{Actions: Actions{}}
	inter := map[string]interface{}{"id": "inter-5"}

	err := exec.Execute(context.Background(), rule, inter)
	assert.NoError(t, err)
	assert.Len(t, inter, 1) // only the original "id"
}

func TestRenderTemplate(t *testing.T) {
	exec := NewExecutor()
	data := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	result := exec.renderTemplate("hello {{.name}}", data)
	assert.Equal(t, "hello test", result)

	result2 := exec.renderTemplate("value={{.value}}", data)
	assert.Equal(t, "value=42", result2)
}

func TestRenderTemplate_WithMap(t *testing.T) {
	exec := NewExecutor()
	data := map[string]interface{}{
		"name": "test",
		"tags": []string{"a", "b"},
	}

	result := exec.renderTemplate("{{.name}}:{{(index .tags 0)}}", data)
	assert.Contains(t, result, "test")
}

func TestRenderTemplate_MissingKey(t *testing.T) {
	exec := NewExecutor()
	data := map[string]interface{}{"name": "test"}

	// Missing key should not panic
	result := exec.renderTemplate("hello {{.name}} {{.missing}}", data)
	assert.NotContains(t, result, "<no value>") // Should handle gracefully or include empty
}

func TestReport_JSONFormat(t *testing.T) {
	exec := NewExecutor()
	report := Report{Format: "json", Title: "JSON Report"}
	inter := map[string]interface{}{
		"id":        "inter-1",
		"type":      "http",
		"source_ip": "10.0.0.1",
		"path":      "/admin",
	}

	err := exec.executeReport(context.Background(), report, inter)
	assert.NoError(t, err)

	reportData, ok := inter["_report"]
	assert.True(t, ok)
	assert.NotNil(t, reportData)

	reportMap, ok := reportData.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "JSON Report", reportMap["title"])
	assert.Equal(t, "json", reportMap["format"])
}

func TestReport_MarkdownFormat(t *testing.T) {
	exec := NewExecutor()
	report := Report{Format: "markdown", Title: "MD Report"}
	inter := map[string]interface{}{
		"id":        "inter-1",
		"type":      "dns",
		"source_ip": "10.0.0.55",
	}

	err := exec.executeReport(context.Background(), report, inter)
	assert.NoError(t, err)

	reportData, ok := inter["_report"]
	assert.True(t, ok)
	assert.NotNil(t, reportData)

	reportMap, ok := reportData.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "markdown", reportMap["format"])
}

func TestExecute_WebhookTimeout(t *testing.T) {
	exec := NewExecutor()
	rule := &Rule{
		Actions: Actions{
			Webhooks: []Webhook{
				{
					URL:    "http://localhost:19999/nonexistent",
					Method: "POST",
				},
			},
		},
	}
	inter := map[string]interface{}{"id": "inter-1"}

	// Use a context with timeout to ensure it doesn't hang
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := exec.Execute(ctx, rule, inter)
	// Should either time out or fail with connection refused
	assert.Error(t, err)
}

func TestNewExecutor(t *testing.T) {
	exec := NewExecutor()
	assert.NotNil(t, exec)
}

func TestExecuteNotification_UnknownType(t *testing.T) {
	exec := NewExecutor()
	notif := Notification{
		Type: "unknown-notification-type",
	}
	inter := map[string]interface{}{"id": "inter-1"}

	// Unknown type should return an error or be silently skipped
	err := exec.executeNotification(context.Background(), notif, inter)
	assert.Error(t, err)
}

func TestExecute_WithNotificationsUnknown(t *testing.T) {
	exec := NewExecutor()
	rule := &Rule{
		Actions: Actions{
			Notifications: []Notification{
				{Type: "nonexistent-channel"},
			},
		},
	}
	inter := map[string]interface{}{"id": "inter-1"}

	err := exec.Execute(context.Background(), rule, inter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification failed")
}

func TestExecuteTagAction_RemoveFromEmpty(t *testing.T) {
	exec := NewExecutor()
	tag := TagAction{Remove: []string{"something"}}
	inter := map[string]interface{}{"id": "inter-1"}

	err := exec.executeTagAction(context.Background(), tag, inter)
	assert.NoError(t, err)
}
