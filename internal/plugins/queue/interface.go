package queue

import (
	"context"
	"time"
)

// Job represents a video processing job
type Job struct {
	ID           string            `json:"id"`
	InputPath    string            `json:"input_path"`
	OutputPath   string            `json:"output_path"`
	FFmpegArgs   string            `json:"ffmpeg_args"`
	Priority     int               `json:"priority"`
	Status       JobStatus         `json:"status"`
	Progress     float64           `json:"progress"`
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
	StorageAdapter string          `json:"storage_adapter"`
	QueueAdapter  string           `json:"queue_adapter"`
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// Queue interface defines the methods for queue operations
type Queue interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, job *Job) error
	
	// Dequeue retrieves and removes a job from the queue
	Dequeue(ctx context.Context) (*Job, error)
	
	// Acknowledge marks a job as processed
	Acknowledge(ctx context.Context, jobID string) error
	
	// GetJob retrieves a job by ID without removing it
	GetJob(ctx context.Context, jobID string) (*Job, error)
	
	// UpdateJob updates a job's status and progress
	UpdateJob(ctx context.Context, job *Job) error
	
	// ListJobs lists jobs with optional filtering
	ListJobs(ctx context.Context, status JobStatus, limit, offset int) ([]*Job, int, error)
	
	// CancelJob cancels a job
	CancelJob(ctx context.Context, jobID string) error
	
	// GetQueueDepth returns the number of jobs in the queue
	GetQueueDepth(ctx context.Context) (int, error)
	
	// Close closes the queue connection
	Close() error
}

// QueueMetrics contains queue performance metrics
type QueueMetrics struct {
	Depth           int     `json:"depth"`
	Throughput      float64 `json:"throughput_jobs_per_second"`
	AverageWaitTime float64 `json:"average_wait_time_seconds"`
} 