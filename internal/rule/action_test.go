package rule

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteTagAction_AddTags(t *testing.T) {
	exec := NewExecutor()
	tag := TagAction{Add: []string{"malicious", "confirmed"}}
	inter := map[string]interface{}{"id": "interaction-1"}

	err := exec.executeTagAction(context.Background(), tag, inter)
	assert.NoError(t, err)

	tags, ok := inter["_tags"].([]string)
	assert.True(t, ok)
	assert.Contains(t, tags, "malicious")
	assert.Contains(t, tags, "confirmed")
}

func TestExecuteTagAction_RemoveTags(t *testing.T) {
	exec := NewExecutor()
	tag := TagAction{Remove: []string{"noise"}}
	inter := map[string]interface{}{
		"id":    "interaction-2",
		"_tags": []string{"noise", "suspicious"},
	}

	err := exec.executeTagAction(context.Background(), tag, inter)
	assert.NoError(t, err)

	tags, ok := inter["_tags"].([]string)
	assert.True(t, ok)
	assert.NotContains(t, tags, "noise")
	assert.Contains(t, tags, "suspicious")
}

func TestExecuteTagAction_AddDuplicate(t *testing.T) {
	exec := NewExecutor()
	tag := TagAction{Add: []string{"tag1"}}
	inter := map[string]interface{}{
		"id":    "interaction-3",
		"_tags": []string{"tag1", "tag2"},
	}

	err := exec.executeTagAction(context.Background(), tag, inter)
	assert.NoError(t, err)

	tags, ok := inter["_tags"].([]string)
	assert.True(t, ok)
	assert.Len(t, tags, 2)
}

func TestExecuteReport_GeneratesReport(t *testing.T) {
	exec := NewExecutor()
	report := Report{Format: "json", Title: "Test Report"}
	inter := map[string]interface{}{
		"id":        "interaction-3",
		"type":      "dns",
		"source_ip": "10.0.0.1",
	}

	err := exec.executeReport(context.Background(), report, inter)
	assert.NoError(t, err)

	reportData, ok := inter["_report"]
	assert.True(t, ok)
	assert.NotNil(t, reportData)
}

func TestExecute_DiscardNoise(t *testing.T) {
	exec := NewExecutor()
	rule := &Rule{
		Actions: Actions{
			DiscardNoise: true,
		},
	}
	inter := map[string]interface{}{"id": "interaction-4"}

	err := exec.Execute(context.Background(), rule, inter)
	assert.NoError(t, err)

	noiseFlag, ok := inter["_noise"]
	assert.True(t, ok)
	assert.True(t, noiseFlag.(bool))
}
