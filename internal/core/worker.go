package core

import (
	"context"
	"time"

	"github.com/nikhil0verma/flixsrota/internal/queue"
	"github.com/nikhil0verma/flixsrota/internal/storage"
	"go.uber.org/zap"
)

// Worker processes individual video processing jobs
type Worker struct {
	queue    queue.Queue
	storage  storage.Storage
	executor *FFmpegExecutor
	logger   *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

// NewWorker creates a new worker
func NewWorker(queue queue.Queue, storage storage.Storage, executor *FFmpegExecutor, logger *zap.Logger) *Worker {
	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		queue:    queue,
		storage:  storage,
		executor: executor,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("Worker started")
	// Worker is now ready to process jobs
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.logger.Info("Worker stopping")
	w.cancel()
}

// ProcessJob processes a video processing job
func (w *Worker) ProcessJob(job *queue.Job) {
	w.logger.Info("Processing job",
		zap.String("job_id", job.ID),
		zap.String("input_path", job.InputPath),
		zap.String("output_path", job.OutputPath))

	// Update job status to processing
	job.Status = queue.JobStatusProcessing
	now := time.Now()
	job.StartedAt = &now
	job.Progress = 0.0

	if err := w.queue.UpdateJob(w.ctx, job); err != nil {
		w.logger.Error("Failed to update job status", zap.Error(err))
		return
	}

	// Execute FFmpeg command
	err := w.executor.Execute(w.ctx, job)
	if err != nil {
		w.logger.Error("Failed to execute FFmpeg",
			zap.String("job_id", job.ID),
			zap.Error(err))

		// Update job status to failed
		job.Status = queue.JobStatusFailed
		job.Error = err.Error()
		now := time.Now()
		job.CompletedAt = &now

		if updateErr := w.queue.UpdateJob(w.ctx, job); updateErr != nil {
			w.logger.Error("Failed to update failed job", zap.Error(updateErr))
		}
		return
	}

	// Update job status to completed
	job.Status = queue.JobStatusCompleted
	job.Progress = 100.0
	now = time.Now()
	job.CompletedAt = &now

	if err := w.queue.UpdateJob(w.ctx, job); err != nil {
		w.logger.Error("Failed to update completed job", zap.Error(err))
		return
	}

	// Acknowledge job completion
	if err := w.queue.Acknowledge(w.ctx, job.ID); err != nil {
		w.logger.Error("Failed to acknowledge job", zap.Error(err))
	}

	w.logger.Info("Job completed successfully",
		zap.String("job_id", job.ID),
		zap.String("output_path", job.OutputPath))
}
