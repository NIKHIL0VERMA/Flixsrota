package core

import (
	"context"
	"sync"
	"time"

	"github.com/nikhil0verma/flixsrota/internal/config"
	"github.com/nikhil0verma/flixsrota/internal/queue"
	"github.com/nikhil0verma/flixsrota/internal/storage"
	"go.uber.org/zap"
)

// JobProcessor manages video processing jobs
type JobProcessor struct {
	config   config.WorkerConfig
	queue    queue.Queue
	storage  storage.Storage
	executor *FFmpegExecutor
	logger   *zap.Logger

	workers    []*Worker
	workerPool chan *Worker
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewJobProcessor creates a new job processor
func NewJobProcessor(
	config config.WorkerConfig,
	queue queue.Queue,
	storage storage.Storage,
	executor *FFmpegExecutor,
	logger *zap.Logger,
) *JobProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &JobProcessor{
		config:     config,
		queue:      queue,
		storage:    storage,
		executor:   executor,
		logger:     logger,
		workerPool: make(chan *Worker, config.MaxWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the job processor
func (jp *JobProcessor) Start() {
	jp.logger.Info("Starting job processor",
		zap.Int("min_workers", jp.config.MinWorkers),
		zap.Int("max_workers", jp.config.MaxWorkers))

	// Start minimum number of workers
	for i := 0; i < jp.config.MinWorkers; i++ {
		worker := NewWorker(jp.queue, jp.storage, jp.executor, jp.logger)
		jp.workers = append(jp.workers, worker)
		jp.workerPool <- worker
		go worker.Start(jp.ctx)
	}

	// Start job processing loop
	jp.wg.Add(1)
	go jp.processJobs()
}

// Stop stops the job processor
func (jp *JobProcessor) Stop() {
	jp.logger.Info("Stopping job processor")
	jp.cancel()
	jp.wg.Wait()

	// Stop all workers
	for _, worker := range jp.workers {
		worker.Stop()
	}

	jp.logger.Info("Job processor stopped")
}

// processJobs continuously processes jobs from the queue
func (jp *JobProcessor) processJobs() {
	defer jp.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-jp.ctx.Done():
			return
		case <-ticker.C:
			// Try to get a job from the queue
			job, err := jp.queue.Dequeue(jp.ctx)
			if err != nil {
				jp.logger.Error("Failed to dequeue job", zap.Error(err))
				continue
			}

			if job == nil {
				// No jobs in queue, continue
				continue
			}

			// Get a worker from the pool
			select {
			case worker := <-jp.workerPool:
				// Process job in worker
				go func(w *Worker, j *queue.Job) {
					w.ProcessJob(j)
					jp.workerPool <- w
				}(worker, job)
			default:
				// No available workers, put job back in queue
				jp.logger.Warn("No available workers, requeuing job", zap.String("job_id", job.ID))
				if err := jp.queue.Enqueue(jp.ctx, job); err != nil {
					jp.logger.Error("Failed to requeue job", zap.Error(err))
				}
			}
		}
	}
}
