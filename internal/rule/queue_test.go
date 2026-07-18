package rule

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewQueue(t *testing.T) {
	ctx := context.Background()
	q := NewQueue(ctx, 2, 3)
	assert.NotNil(t, q)
	assert.Equal(t, 2, q.workers)
	assert.Equal(t, 3, q.maxRetries)
	assert.NotNil(t, q.jobs)
	assert.NotNil(t, q.ctx)

	// Cleanup
	q.Stop()
}

func TestQueue_StartStop(t *testing.T) {
	ctx := context.Background()
	q := NewQueue(ctx, 2, 0)
	exec := NewExecutor()

	// Start workers
	q.Start(exec)
	// No easy way to verify workers are running without race conditions,
	// but we can stop and ensure no deadlocks
	q.Stop()
}

func TestQueue_Enqueue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		q := NewQueue(ctx, 1, 0)
		exec := NewExecutor()
		q.Start(exec)

		job := &Job{
			RuleID: "rule-1",
			Rule:   &Rule{Actions: Actions{}},
			Interaction: map[string]interface{}{
				"id": "inter-1",
			},
		}

		err := q.Enqueue(job)
		assert.NoError(t, err)

		// Give worker time to process
		time.Sleep(50 * time.Millisecond)
		q.Stop()
	})

	t.Run("queue stopped", func(t *testing.T) {
		// Fill the buffer first so Enqueue will block and pick ctx.Done()
		q := &Queue{
			jobs:       make(chan *Job, 1),
			workers:    0,
			ctx:        context.Background(),
			cancel:     func() {},
			maxRetries: 0,
		}
		q.jobs <- &Job{RuleID: "fill"}

		// Now create a properly cancelable queue
		ctx2, cancel2 := context.WithCancel(context.Background())
		q2 := &Queue{
			jobs:       make(chan *Job, 1),
			workers:    0,
			ctx:        ctx2,
			cancel:     cancel2,
			maxRetries: 0,
		}
		q2.jobs <- &Job{RuleID: "fill"}

		cancel2() // Cancel context

		// Buffer is full and ctx is done, so Enqueue must pick ctx.Done()
		job := &Job{RuleID: "rule-1"}
		err := q2.Enqueue(job)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "queue is stopped")
	})
}

func TestQueue_EnqueueFull(t *testing.T) {
	// Use a queue with buffer 1 and no workers so enqueue blocks into default case
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := &Queue{
		jobs:       make(chan *Job, 1),
		workers:    0,
		ctx:        ctx,
		cancel:     cancel,
		maxRetries: 0,
	}

	// Fill the buffer
	q.jobs <- &Job{RuleID: "first"}

	// Second enqueue should fail because buffer is full
	err := q.Enqueue(&Job{RuleID: "second"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue is full")
}

func TestQueue_ProcessJob_Success(t *testing.T) {
	q := NewQueue(context.Background(), 1, 0)
	exec := NewExecutor()

	job := &Job{
		RuleID:      "rule-1",
		Rule:        &Rule{Actions: Actions{}},
		Interaction: map[string]interface{}{"id": "inter-1"},
	}

	err := q.processJob(exec, job)
	assert.NoError(t, err)

	q.Stop()
}

func TestQueue_ProcessJob_WithActions(t *testing.T) {
	q := NewQueue(context.Background(), 1, 0)
	exec := NewExecutor()

	job := &Job{
		RuleID: "rule-1",
		Rule: &Rule{
			Actions: Actions{
				Tags: []TagAction{
					{Add: []string{"auto-tag"}},
				},
				Reports: []Report{
					{Format: "json", Title: "Auto Report"},
				},
			},
		},
		Interaction: map[string]interface{}{
			"id":   "inter-1",
			"type": "dns",
		},
	}

	err := q.processJob(exec, job)
	assert.NoError(t, err)

	tags := job.Interaction["_tags"].([]string)
	assert.Contains(t, tags, "auto-tag")

	q.Stop()
}

func TestQueue_ProcessJob_Error(t *testing.T) {
	q := NewQueue(context.Background(), 1, 0)
	exec := NewExecutor()

	// A notification with unknown type will fail
	job := &Job{
		RuleID: "rule-1",
		Rule: &Rule{
			Actions: Actions{
				Notifications: []Notification{
					{Type: "unknown-type"},
				},
			},
		},
		Interaction: map[string]interface{}{"id": "inter-1"},
	}

	err := q.processJob(exec, job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification failed")

	q.Stop()
}

func TestQueue_WorkerRetry(t *testing.T) {
	// Create a queue with 1 retry to test retry logic
	// The worker will retry the job once after failure
	ctx := context.Background()
	q := NewQueue(ctx, 1, 1)
	exec := NewExecutor()
	q.Start(exec)

	// This job should fail (unknown notification type) and be retried once
	job := &Job{
		RuleID: "rule-retry",
		Rule: &Rule{
			Actions: Actions{
				Notifications: []Notification{
					{Type: "unknown-retry-type"},
				},
			},
		},
		Interaction: map[string]interface{}{"id": "inter-retry"},
		Attempt:     0,
	}

	// Enqueue - the worker will process, fail, and retry (since maxRetries=1)
	err := q.Enqueue(job)
	assert.NoError(t, err)

	// Wait for retry (backoff = 1*1 = 1 second)
	time.Sleep(1200 * time.Millisecond)
	q.Stop()

	// After retry, attempt should be 1 (incremented from 0)
	assert.Equal(t, 1, job.Attempt)
}

func TestQueue_WorkerNoRetry(t *testing.T) {
	// maxRetries=0 means no retry
	ctx := context.Background()
	q := NewQueue(ctx, 1, 0)
	exec := NewExecutor()
	q.Start(exec)

	job := &Job{
		RuleID: "rule-no-retry",
		Rule: &Rule{
			Actions: Actions{
				Notifications: []Notification{
					{Type: "unknown-no-retry"},
				},
			},
		},
		Interaction: map[string]interface{}{"id": "inter-no-retry"},
		Attempt:     0,
	}

	err := q.Enqueue(job)
	assert.NoError(t, err)

	// Wait a bit for processing
	time.Sleep(50 * time.Millisecond)
	q.Stop()

	// Attempt should still be 0 because no retry happened
	assert.Equal(t, 0, job.Attempt)
}

func TestQueue_MultipleWorkers(t *testing.T) {
	ctx := context.Background()
	q := NewQueue(ctx, 3, 0)
	exec := NewExecutor()
	q.Start(exec)

	// Enqueue several jobs
	for i := 0; i < 5; i++ {
		job := &Job{
			RuleID:      "rule-multi",
			Rule:        &Rule{Actions: Actions{}},
			Interaction: map[string]interface{}{"id": "inter"},
		}
		err := q.Enqueue(job)
		assert.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)
	q.Stop()
}
