package workflow

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// ActionLog records the result of a single action execution
type ActionLog struct {
	WorkflowID string
	ActionID   string
	ActionType string
	Success    bool
	Error      string
	Duration   time.Duration
	Attempt    int
	ExecutedAt time.Time
}

// QueueJob represents a workflow execution job
type QueueJob struct {
	WorkflowID  string
	Action      models.Action
	Interaction *models.Interaction
	Attempt     int
}

// Queue is an async workflow action execution queue with retry support.
// If engine is set, jobs are persisted to the database for recovery after restart.
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
	engine     *xorm.Engine
}

// NewQueue creates a new async workflow queue.
// If engine is non-nil, jobs are persisted to the database for recovery.
func NewQueue(ctx context.Context, service *Service, workers, maxRetries int, engine *xorm.Engine) *Queue {
	ctx, cancel := context.WithCancel(ctx)
	return &Queue{
		jobs:       make(chan *QueueJob, 100),
		workers:    workers,
		ctx:        ctx,
		cancel:     cancel,
		maxRetries: maxRetries,
		service:    service,
		engine:     engine,
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

// processJob processes a single job with retry logic and persists the result.
func (q *Queue) processJob(job *QueueJob) {
	start := time.Now()

	// Update persistent log status to running
	var plogID string
	if q.engine != nil {
		plogID = fmt.Sprintf("wal-%d-%s", start.UnixNano(), job.Action.ID)
		interactionID := ""
		if job.Interaction != nil {
			interactionID = job.Interaction.ID
		}
		plog := &PersistentActionLog{
			ID:            plogID,
			WorkflowID:    job.WorkflowID,
			ActionID:      job.Action.ID,
			ActionType:    job.Action.Type,
			InteractionID: interactionID,
			Status:        "running",
			Attempt:       job.Attempt,
		}
		if _, err := q.engine.Insert(plog); err != nil {
			log.Printf("[workflow-queue] failed to persist running status: %v", err)
		}
	}

	err := q.service.executeAction(job.Action, job.Interaction)
	duration := time.Since(start)

	// Record in-memory log
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

	// Update persistent log with final status
	if q.engine != nil && plogID != "" {
		status := "completed"
		errMsg := ""
		if err != nil {
			status = "failed"
			errMsg = err.Error()
		}
		_, updateErr := q.engine.ID(plogID).Update(&PersistentActionLog{
			Status:     status,
			Error:      errMsg,
			DurationMs: duration.Milliseconds(),
			FinishedAt: time.Now(),
		})
		if updateErr != nil {
			log.Printf("[workflow-queue] failed to update persistent log: %v", updateErr)
		}
	}

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

// EnqueueWorkflow enqueues all enabled actions from a workflow and persists them to the database.
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
		// Persist to database if engine is available
		if q.engine != nil {
			interactionID := ""
			if interaction != nil {
				interactionID = interaction.ID
			}
			plog := &PersistentActionLog{
				ID:            fmt.Sprintf("wal-%d", time.Now().UnixNano()),
				WorkflowID:    workflowID,
				ActionID:      action.ID,
				ActionType:    action.Type,
				InteractionID: interactionID,
				Status:        "pending",
			}
			if _, err := q.engine.Insert(plog); err != nil {
				log.Printf("[workflow-queue] failed to persist action log: %v", err)
			}
		}
		if err := q.Enqueue(job); err != nil {
			return fmt.Errorf("failed to enqueue action %s: %w", action.Type, err)
		}
	}
	return nil
}
