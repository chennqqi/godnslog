package rule

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Queue represents an async queue for rule execution
type Queue struct {
	jobs       chan *Job
	workers    int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	maxRetries int
}

// Job represents a rule execution job
type Job struct {
	RuleID      string
	Interaction map[string]interface{}
	Rule        *Rule
	Attempt     int
}

// NewQueue creates a new async queue
func NewQueue(ctx context.Context, workers, maxRetries int) *Queue {
	ctx, cancel := context.WithCancel(ctx)
	return &Queue{
		jobs:       make(chan *Job, 100),
		workers:    workers,
		ctx:        ctx,
		cancel:     cancel,
		maxRetries: maxRetries,
	}
}

// Start starts the queue workers
func (q *Queue) Start(executor *Executor) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(executor)
	}
}

// Stop stops the queue
func (q *Queue) Stop() {
	q.cancel()
	q.wg.Wait()
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(job *Job) error {
	select {
	case q.jobs <- job:
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("queue is stopped")
	default:
		return fmt.Errorf("queue is full")
	}
}

// worker processes jobs from the queue
func (q *Queue) worker(executor *Executor) {
	defer q.wg.Done()

	for {
		select {
		case job := <-q.jobs:
			if err := q.processJob(executor, job); err != nil {
				fmt.Printf("Job failed: %v\n", err)
				if job.Attempt < q.maxRetries {
					job.Attempt++
					// Retry with exponential backoff
					backoff := time.Duration(job.Attempt*job.Attempt) * time.Second
					time.Sleep(backoff)
					q.Enqueue(job)
				}
			}
		case <-q.ctx.Done():
			return
		}
	}
}

// processJob processes a single job
func (q *Queue) processJob(executor *Executor, job *Job) error {
	ctx, cancel := context.WithTimeout(q.ctx, 30*time.Second)
	defer cancel()

	return executor.Execute(ctx, job.Rule, job.Interaction)
}
