package core

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikhil0verma/flixsrota/internal/config"
	"github.com/nikhil0verma/flixsrota/internal/plugins/queue"
	"github.com/nikhil0verma/flixsrota/internal/plugins/storage"
	"go.uber.org/zap"
	grpcstd "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server represents the main Flixsrota server
type Server struct {
	config     *config.Config
	logger     *zap.Logger
	grpcServer *grpcstd.Server
	processor  *JobProcessor
	queue      queue.Queue
	storage    storage.Storage
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewServer creates a new Flixsrota server instance
func NewServer(cfg *config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	logger, _ := zap.NewProduction()

	return &Server{
		config: cfg,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the server and all its components
func (s *Server) Start() error {
	s.logger.Info("Starting Flixsrota server...")

	// Initialize queue
	if err := s.initializeQueue(); err != nil {
		return fmt.Errorf("failed to initialize queue: %w", err)
	}

	// Initialize storage
	if err := s.initializeStorage(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize job processor
	if err := s.initializeJobProcessor(); err != nil {
		return fmt.Errorf("failed to initialize job processor: %w", err)
	}

	// Initialize gRPC server
	if err := s.initializeGRPCServer(); err != nil {
		return fmt.Errorf("failed to initialize gRPC server: %w", err)
	}

	// Start job processor
	go s.processor.Start()

	// Start gRPC server
	go s.startGRPCServer()

	// Wait for shutdown signal
	s.waitForShutdown()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	s.logger.Info("Shutting down Flixsrota server...")

	// Cancel context to stop all goroutines
	s.cancel()

	// Stop job processor
	if s.processor != nil {
		s.processor.Stop()
	}

	// Stop gRPC server
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	// Close queue connection
	if s.queue != nil {
		s.queue.Close()
	}

	s.logger.Info("Server stopped")
	return nil
}

// initializeQueue initializes the queue adapter
func (s *Server) initializeQueue() error {
	var err error

	switch s.config.Queue.Adapter {
	case "redis":
		s.queue, err = queue.NewRedisQueue(
			s.ctx,
			s.config.Queue.Redis.Address,
			s.config.Queue.Redis.Password,
			s.config.Queue.Redis.DB,
		)
	case "kafka":
		// TODO: Implement Kafka queue
		return fmt.Errorf("kafka queue not implemented yet")
	case "sqs":
		// TODO: Implement SQS queue
		return fmt.Errorf("sqs queue not implemented yet")
	default:
		return fmt.Errorf("unknown queue adapter: %s", s.config.Queue.Adapter)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize queue: %w", err)
	}

	s.logger.Info("Queue initialized", zap.String("adapter", s.config.Queue.Adapter))
	return nil
}

// initializeStorage initializes the storage adapter
func (s *Server) initializeStorage() error {
	var err error

	switch s.config.Storage.Adapter {
	case "local":
		s.storage, err = storage.NewLocalStorage(
			s.config.Storage.Local.BasePath,
			s.config.Storage.Local.TempPath,
		)
	case "s3":
		// TODO: Implement S3 storage
		return fmt.Errorf("s3 storage not implemented yet")
	case "gcs":
		// TODO: Implement GCS storage
		return fmt.Errorf("gcs storage not implemented yet")
	default:
		return fmt.Errorf("unknown storage adapter: %s", s.config.Storage.Adapter)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	s.logger.Info("Storage initialized", zap.String("adapter", s.config.Storage.Adapter))
	return nil
}

// initializeJobProcessor initializes the job processor
func (s *Server) initializeJobProcessor() error {
	executor := NewFFmpegExecutor(s.config.FFmpeg)

	s.processor = NewJobProcessor(
		s.config.Worker,
		s.queue,
		s.storage,
		executor,
		s.logger,
	)

	s.logger.Info("Job processor initialized")
	return nil
}

// initializeGRPCServer initializes the gRPC server
func (s *Server) initializeGRPCServer() error {
	s.grpcServer = grpcstd.NewServer()

	// TODO: Register services when protobuf is generated
	// For now, we'll just create the server without services

	return nil
}

// startGRPCServer starts the gRPC server
func (s *Server) startGRPCServer() error {
	address := fmt.Sprintf("%s:%d", s.config.GRPC.Address, s.config.GRPC.Port)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	s.logger.Info("gRPC server starting", zap.String("address", address))

	// Add reflection service if enabled
	if s.config.GRPC.EnableReflection {
		reflection.Register(s.grpcServer)
	}

	// Start serving
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// waitForShutdown waits for shutdown signals
func (s *Server) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	s.logger.Info("Received shutdown signal")

	if err := s.Stop(); err != nil {
		s.logger.Error("Error during shutdown", zap.Error(err))
	}
}
