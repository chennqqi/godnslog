package workflow

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

// ActionLog records the result of a single action execution
type ActionLog struct {
	WorkflowID  string
	ActionID    string
	ActionType  string
	Success     bool
	Error       string
	Duration    time.Duration
	Attempt     int
	ExecutedAt  time.Time
}

// QueueJob represents a workflow execution job
type QueueJob struct {
	WorkflowID  string
	Action      models.Action
	Interaction *models.Interaction
	Attempt     int
}

// Queue is an async workflow action execution queue with retry support
type Queue struct {
	jobs       chan *QueueJob
	workers    int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	maxRetries int
	service    *Service
	logs       []ActionLog
	logsMu     sync.Mutex
}

// NewQueue creates a new async workflow queue
func NewQueue(ctx context.Context, service *Service, workers, maxRetries int) *Queue {
	ctx, cancel := context.WithCancel(ctx)
	return &Queue{
		jobs:       make(chan *QueueJob, 100),
		workers:    workers,
		ctx:        ctx,
		cancel:     cancel,
		maxRetries: maxRetries,
		service:    service,
	}
}

// Start launches the queue workers
func (q *Queue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

// Stop shuts down the queue and waits for workers to finish
func (q *Queue) Stop() {
	q.cancel()
	q.wg.Wait()
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(job *QueueJob) error {
	select {
	case q.jobs <- job:
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("queue is stopped")
	default:
		return fmt.Errorf("queue is full")
	}
}

// GetLogs returns all recorded action logs
func (q *Queue) GetLogs() []ActionLog {
	q.logsMu.Lock()
	defer q.logsMu.Unlock()
	logs := make([]ActionLog, len(q.logs))
	copy(logs, q.logs)
	return logs
}

// worker processes jobs from the queue
func (q *Queue) worker() {
	defer q.wg.Done()

	for {
		select {
		case job := <-q.jobs:
			q.processJob(job)
		case <-q.ctx.Done():
			return
		}
	}
}

// processJob processes a single job with retry logic
func (q *Queue) processJob(job *QueueJob) {
	start := time.Now()
	err := q.service.executeAction(job.Action, job.Interaction)
	duration := time.Since(start)

	// Record log
	actionLog := ActionLog{
		WorkflowID: job.WorkflowID,
		ActionID:   job.Action.ID,
		ActionType: job.Action.Type,
		Success:    err == nil,
		Duration:   duration,
		Attempt:    job.Attempt + 1,
		ExecutedAt: time.Now(),
	}
	if err != nil {
		actionLog.Error = err.Error()
	}
	q.logsMu.Lock()
	q.logs = append(q.logs, actionLog)
	q.logsMu.Unlock()

	if err != nil {
		if job.Attempt < q.maxRetries {
			job.Attempt++
			backoff := time.Duration(job.Attempt*job.Attempt) * time.Second
			log.Printf("[workflow-queue] Job failed (attempt %d), retrying in %v: %v", job.Attempt, backoff, err)
			time.Sleep(backoff)
			if enqueueErr := q.Enqueue(job); enqueueErr != nil {
				log.Printf("[workflow-queue] Failed to re-enqueue job: %v", enqueueErr)
			}
		} else {
			log.Printf("[workflow-queue] Job failed after %d attempts, giving up: %v", job.Attempt+1, err)
		}
	}
}

// EnqueueWorkflow enqueues all enabled actions from a workflow
func (q *Queue) EnqueueWorkflow(workflowID string, actions models.Actions, interaction *models.Interaction) error {
	for _, action := range actions {
		if !action.Enabled {
			continue
		}
		job := &QueueJob{
			WorkflowID:  workflowID,
			Action:      action,
			Interaction: interaction,
		}
		if err := q.Enqueue(job); err != nil {
			return fmt.Errorf("failed to enqueue action %s: %w", action.Type, err)
		}
	}
	return nil
}
