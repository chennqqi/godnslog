package workflow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestQueue_NewQueue(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	ctx := context.Background()

	q := NewQueue(ctx, service, 2, 3)
	assert.NotNil(t, q)
	assert.Equal(t, 2, q.workers)
	assert.Equal(t, 3, q.maxRetries)
}

func TestQueue_StartStop(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	ctx := context.Background()
	q := NewQueue(ctx, service, 2, 3)
	q.Start()

	// Give workers time to start
	time.Sleep(10 * time.Millisecond)

	q.Stop()
	// Should not block
}

func TestQueue_EnqueueAndProcess(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)
	service.SetOutboundSecurity(NewOutboundSecurity([]string{"127.0.0.1"}))

	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := NewQueue(ctx, service, 2, 1)
	q.Start()
	defer q.Stop()

	token := "test-token"
	interaction := &models.Interaction{
		ID:       models.GenerateID(),
		Type:     "dns",
		SourceIP: "192.168.1.1",
		Token:    &token,
	}

	actions := models.Actions{
		{
			Type:    models.ActionTypeHTTP,
			Enabled: true,
			Config: map[string]interface{}{
				"url":    srv.URL,
				"method": "GET",
			},
		},
	}

	err = q.EnqueueWorkflow("wf-test", actions, interaction)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	logs := q.GetLogs()
	assert.Len(t, logs, 1)
	assert.True(t, logs[0].Success)
	assert.Equal(t, "http", logs[0].ActionType)
}

func TestQueue_RetryOnFailure(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)
	service.SetOutboundSecurity(NewOutboundSecurity([]string{"127.0.0.1"}))

	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := NewQueue(ctx, service, 1, 3)
	q.Start()
	defer q.Stop()

	token := "test-token"
	interaction := &models.Interaction{
		ID:       models.GenerateID(),
		Type:     "dns",
		SourceIP: "192.168.1.1",
		Token:    &token,
	}

	actions := models.Actions{
		{
			Type:    models.ActionTypeHTTP,
			Enabled: true,
			Config: map[string]interface{}{
				"url":    srv.URL,
				"method": "GET",
			},
		},
	}

	err = q.EnqueueWorkflow("wf-retry", actions, interaction)
	assert.NoError(t, err)

	// Wait for retries (1s + 4s backoff = ~5s max, but we check after 7s)
	time.Sleep(8 * time.Second)

	assert.GreaterOrEqual(t, atomic.LoadInt32(&callCount), int32(3))

	logs := q.GetLogs()
	assert.GreaterOrEqual(t, len(logs), 1)
	// At least one should be successful
	foundSuccess := false
	for _, l := range logs {
		if l.Success {
			foundSuccess = true
			break
		}
	}
	assert.True(t, foundSuccess, "expected at least one successful execution")
}

func TestQueue_EnqueueWorkflow_SkipsDisabled(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := NewQueue(ctx, service, 1, 1)
	q.Start()
	defer q.Stop()

	token := "test-token"
	interaction := &models.Interaction{
		ID:    models.GenerateID(),
		Type:  "dns",
		Token: &token,
	}

	actions := models.Actions{
		{Type: models.ActionTypeHTTP, Enabled: false, Config: map[string]interface{}{}},
		{Type: models.ActionTypeDNS, Enabled: false, Config: map[string]interface{}{}},
	}

	err = q.EnqueueWorkflow("wf-disabled", actions, interaction)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	logs := q.GetLogs()
	assert.Len(t, logs, 0)
}

func TestQueue_EnqueueFull(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create queue with small buffer, don't start workers
	q := NewQueue(ctx, service, 1, 1)

	token := "test-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	// Fill the queue (capacity 100)
	for i := 0; i < 100; i++ {
		job := &QueueJob{
			WorkflowID:  "wf-full",
			Action:      models.Action{Type: models.ActionTypeDNS, Enabled: true, Config: map[string]interface{}{"domain": "localhost"}},
			Interaction: interaction,
		}
		err := q.Enqueue(job)
		assert.NoError(t, err)
	}

	// Next enqueue should fail
	job := &QueueJob{
		WorkflowID:  "wf-full",
		Action:      models.Action{Type: models.ActionTypeDNS, Enabled: true, Config: map[string]interface{}{"domain": "localhost"}},
		Interaction: interaction,
	}
	err = q.Enqueue(job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue is full")
}
