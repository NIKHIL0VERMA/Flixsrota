package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisQueue implements the Queue interface using Redis
type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisQueue creates a new Redis queue instance
func NewRedisQueue(ctx context.Context, address, password string, db int) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisQueue{
		client: client,
		ctx:    ctx,
	}, nil
}

// Enqueue adds a job to the queue
func (rq *RedisQueue) Enqueue(ctx context.Context, job *Job) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	
	job.CreatedAt = time.Now()
	job.Status = JobStatusQueued
	job.Progress = 0

	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add to queue with priority
	pipe := rq.client.Pipeline()
	pipe.ZAdd(ctx, "flixsrota:queue", &redis.Z{
		Score:  float64(job.Priority),
		Member: job.ID,
	})
	pipe.Set(ctx, fmt.Sprintf("flixsrota:job:%s", job.ID), jobData, 24*time.Hour)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

// Dequeue retrieves and removes a job from the queue
func (rq *RedisQueue) Dequeue(ctx context.Context) (*Job, error) {
	// Get job with highest priority
	result, err := rq.client.ZPopMax(ctx, "flixsrota:queue").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs in queue
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}

	jobID := result[0].Member.(string)
	
	// Get job data
	jobData, err := rq.client.Get(ctx, fmt.Sprintf("flixsrota:job:%s", jobID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Update status to processing
	now := time.Now()
	job.Status = JobStatusProcessing
	job.StartedAt = &now

	// Update job in Redis
	if err := rq.UpdateJob(ctx, &job); err != nil {
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	return &job, nil
}

// Acknowledge marks a job as processed
func (rq *RedisQueue) Acknowledge(ctx context.Context, jobID string) error {
	// Remove job from Redis
	pipe := rq.client.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("flixsrota:job:%s", jobID))
	pipe.ZRem(ctx, "flixsrota:queue", jobID)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to acknowledge job: %w", err)
	}

	return nil
}

// GetJob retrieves a job by ID without removing it
func (rq *RedisQueue) GetJob(ctx context.Context, jobID string) (*Job, error) {
	jobData, err := rq.client.Get(ctx, fmt.Sprintf("flixsrota:job:%s", jobID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Job not found
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// UpdateJob updates a job's status and progress
func (rq *RedisQueue) UpdateJob(ctx context.Context, job *Job) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	err = rq.client.Set(ctx, fmt.Sprintf("flixsrota:job:%s", job.ID), jobData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// ListJobs lists jobs with optional filtering
func (rq *RedisQueue) ListJobs(ctx context.Context, status JobStatus, limit, offset int) ([]*Job, int, error) {
	// Get all job IDs
	jobIDs, err := rq.client.ZRange(ctx, "flixsrota:queue", int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get job IDs: %w", err)
	}

	var jobs []*Job
	for _, jobID := range jobIDs {
		job, err := rq.GetJob(ctx, jobID)
		if err != nil {
			continue // Skip jobs that can't be retrieved
		}
		
		if status == "" || job.Status == status {
			jobs = append(jobs, job)
		}
	}

	// Get total count
	total, err := rq.client.ZCard(ctx, "flixsrota:queue").Result()
	if err != nil {
		return jobs, len(jobs), fmt.Errorf("failed to get queue count: %w", err)
	}

	return jobs, int(total), nil
}

// CancelJob cancels a job
func (rq *RedisQueue) CancelJob(ctx context.Context, jobID string) error {
	job, err := rq.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = JobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now

	return rq.UpdateJob(ctx, job)
}

// GetQueueDepth returns the number of jobs in the queue
func (rq *RedisQueue) GetQueueDepth(ctx context.Context) (int, error) {
	count, err := rq.client.ZCard(ctx, "flixsrota:queue").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue depth: %w", err)
	}

	return int(count), nil
}

// Close closes the queue connection
func (rq *RedisQueue) Close() error {
	return rq.client.Close()
} 