package grpc

import (
	"context"
	"net"
	"time"

	pb "github.com/flixsrota/flixsrota/internal/grpc/pb"
	"github.com/flixsrota/flixsrota/internal/metrics"
	"github.com/flixsrota/flixsrota/internal/queue"
	"github.com/flixsrota/flixsrota/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server represents the gRPC server
type Server struct {
	queue      queue.Queue
	storage    storage.Storage
	processor  interface{} // JobProcessor interface
	logger     *zap.Logger
	grpcServer *grpc.Server
	metrics    *metrics.SystemMetricsCollector
}

// NewServer creates a new gRPC server
func NewServer(queue queue.Queue, storage storage.Storage, processor interface{}, logger *zap.Logger) *grpc.Server {
	s := &Server{
		queue:     queue,
		storage:   storage,
		processor: processor,
		logger:    logger,
		metrics:   metrics.NewSystemMetricsCollector(logger),
	}

	grpcServer := grpc.NewServer()

	// Register services (these will be implemented when protobuf is generated)
	// pb.RegisterVideoProcessorServer(grpcServer, s)
	// pb.RegisterSystemMetricsServer(grpcServer, s)

	s.grpcServer = grpcServer
	return grpcServer
}

// Serve starts the gRPC server
func (s *Server) Serve(lis net.Listener) error {
	return s.grpcServer.Serve(lis)
}

// GracefulStop gracefully stops the gRPC server
func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

// ProcessVideo handles video processing requests
func (s *Server) ProcessVideo(ctx context.Context, req *pb.ProcessVideoRequest) (*pb.ProcessVideoResponse, error) {
	s.logger.Info("Processing video request",
		zap.String("input_path", req.InputPath),
		zap.String("output_path", req.OutputPath))

	// Create job
	job := &queue.Job{
		InputPath:      req.InputPath,
		OutputPath:     req.OutputPath,
		FFmpegArgs:     req.FfmpegArgs,
		Priority:       int(req.Priority),
		Metadata:       req.Metadata,
		StorageAdapter: req.StorageAdapter,
		QueueAdapter:   req.QueueAdapter,
	}

	// Enqueue job
	if err := s.queue.Enqueue(ctx, job); err != nil {
		s.logger.Error("Failed to enqueue job", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to enqueue job: %v", err)
	}

	return &pb.ProcessVideoResponse{
		JobId:   job.ID,
		Status:  pb.JobStatus_JOB_STATUS_QUEUED,
		Message: "Job queued successfully",
	}, nil
}

// GetJobStatus retrieves job status
func (s *Server) GetJobStatus(ctx context.Context, req *pb.GetJobStatusRequest) (*pb.GetJobStatusResponse, error) {
	job, err := s.queue.GetJob(ctx, req.JobId)
	if err != nil {
		s.logger.Error("Failed to get job", zap.String("job_id", req.JobId), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get job: %v", err)
	}

	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job not found: %s", req.JobId)
	}

	response := &pb.GetJobStatusResponse{
		JobId:      job.ID,
		Status:     convertJobStatus(job.Status),
		Progress:   float32(job.Progress),
		OutputPath: job.OutputPath,
		Metadata:   job.Metadata,
	}

	if job.StartedAt != nil {
		response.StartedAt = job.StartedAt
	}
	if job.CompletedAt != nil {
		response.CompletedAt = job.CompletedAt
	}

	return response, nil
}

// CancelJob cancels a running job
func (s *Server) CancelJob(ctx context.Context, req *pb.CancelJobRequest) (*pb.CancelJobResponse, error) {
	err := s.queue.CancelJob(ctx, req.JobId)
	if err != nil {
		s.logger.Error("Failed to cancel job", zap.String("job_id", req.JobId), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to cancel job: %v", err)
	}

	return &pb.CancelJobResponse{
		Success: true,
		Message: "Job cancelled successfully",
	}, nil
}

// ListJobs lists jobs with optional filtering
func (s *Server) ListJobs(ctx context.Context, req *pb.ListJobsRequest) (*pb.ListJobsResponse, error) {
	statusFilter := queue.JobStatus("")
	if req.StatusFilter != pb.JobStatus_JOB_STATUS_UNSPECIFIED {
		statusFilter = convertPBJobStatus(req.StatusFilter)
	}

	jobs, total, err := s.queue.ListJobs(ctx, statusFilter, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("Failed to list jobs", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list jobs: %v", err)
	}

	var jobInfos []*pb.JobInfo
	for _, job := range jobs {
		jobInfo := &pb.JobInfo{
			JobId:      job.ID,
			Status:     convertJobStatus(job.Status),
			Progress:   float32(job.Progress),
			InputPath:  job.InputPath,
			OutputPath: job.OutputPath,
			CreatedAt:  job.CreatedAt,
		}
		if job.StartedAt != nil {
			jobInfo.UpdatedAt = job.StartedAt
		}
		jobInfos = append(jobInfos, jobInfo)
	}

	return &pb.ListJobsResponse{
		Jobs:       jobInfos,
		TotalCount: int32(total),
	}, nil
}

// GetMetrics returns system metrics
func (s *Server) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	// Get queue metrics
	queueDepth, err := s.queue.GetQueueDepth(ctx)
	if err != nil {
		s.logger.Error("Failed to get queue depth", zap.Error(err))
		queueDepth = 0
	}

	// Get system metrics
	systemMetrics, err := s.metrics.CollectMetrics()
	if err != nil {
		s.logger.Error("Failed to collect system metrics", zap.Error(err))
		// Use default values if metrics collection fails
		systemMetrics = &metrics.SystemMetrics{
			CPUUsagePercent:    0.0,
			MemoryUsagePercent: 0.0,
			DiskUsagePercent:   0.0,
			ActiveWorkerCount:  0,
			MaxWorkerCount:     10,
		}
	}

	// Create metrics response
	response := &pb.GetMetricsResponse{
		SystemMetrics: &pb.SystemMetrics{
			CpuUsagePercent:      systemMetrics.CPUUsagePercent,
			MemoryUsagePercent:   systemMetrics.MemoryUsagePercent,
			DiskUsagePercent:     systemMetrics.DiskUsagePercent,
			TotalMemoryBytes:     int64(systemMetrics.TotalMemoryBytes),
			AvailableMemoryBytes: int64(systemMetrics.AvailableMemoryBytes),
			TotalDiskBytes:       int64(systemMetrics.TotalDiskBytes),
			AvailableDiskBytes:   int64(systemMetrics.AvailableDiskBytes),
			ActiveWorkerCount:    int32(systemMetrics.ActiveWorkerCount),
			MaxWorkerCount:       int32(systemMetrics.MaxWorkerCount),
		},
		JobMetrics: &pb.JobMetrics{
			// TODO: Implement job metrics collection
			TotalJobs:      0,
			QueuedJobs:     int32(queueDepth),
			ProcessingJobs: 0,
			CompletedJobs:  0,
			FailedJobs:     0,
			CancelledJobs:  0,
		},
		QueueMetrics: &pb.QueueMetrics{
			QueueDepth: int32(queueDepth),
			// TODO: Implement queue throughput metrics
		},
	}

	return response, nil
}

// StreamMetrics streams real-time metrics
func (s *Server) StreamMetrics(req *pb.StreamMetricsRequest, stream pb.SystemMetrics_StreamMetricsServer) error {
	interval := time.Duration(req.IntervalSeconds) * time.Second
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-ticker.C:
			metrics, err := s.GetMetrics(stream.Context(), &pb.GetMetricsRequest{})
			if err != nil {
				s.logger.Error("Failed to get metrics", zap.Error(err))
				continue
			}

			response := &pb.StreamMetricsResponse{
				Metrics:   metrics,
				Timestamp: time.Now(),
			}

			if err := stream.Send(response); err != nil {
				s.logger.Error("Failed to send metrics", zap.Error(err))
				return err
			}
		}
	}
}

// Helper functions to convert between internal and protobuf types
func convertJobStatus(status queue.JobStatus) pb.JobStatus {
	switch status {
	case queue.JobStatusQueued:
		return pb.JobStatus_JOB_STATUS_QUEUED
	case queue.JobStatusProcessing:
		return pb.JobStatus_JOB_STATUS_PROCESSING
	case queue.JobStatusCompleted:
		return pb.JobStatus_JOB_STATUS_COMPLETED
	case queue.JobStatusFailed:
		return pb.JobStatus_JOB_STATUS_FAILED
	case queue.JobStatusCancelled:
		return pb.JobStatus_JOB_STATUS_CANCELLED
	default:
		return pb.JobStatus_JOB_STATUS_UNSPECIFIED
	}
}

func convertPBJobStatus(status pb.JobStatus) queue.JobStatus {
	switch status {
	case pb.JobStatus_JOB_STATUS_QUEUED:
		return queue.JobStatusQueued
	case pb.JobStatus_JOB_STATUS_PROCESSING:
		return queue.JobStatusProcessing
	case pb.JobStatus_JOB_STATUS_COMPLETED:
		return queue.JobStatusCompleted
	case pb.JobStatus_JOB_STATUS_FAILED:
		return queue.JobStatusFailed
	case pb.JobStatus_JOB_STATUS_CANCELLED:
		return queue.JobStatusCancelled
	default:
		return queue.JobStatusQueued
	}
}
