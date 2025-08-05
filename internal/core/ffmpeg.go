package core

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/flixsrota/flixsrota/internal/config"
	"github.com/flixsrota/flixsrota/internal/queue"
	"go.uber.org/zap"
)

// FFmpegExecutor manages FFmpeg process execution
type FFmpegExecutor struct {
	config config.FFmpegConfig
	logger *zap.Logger
}

// NewFFmpegExecutor creates a new FFmpeg executor
func NewFFmpegExecutor(config config.FFmpegConfig) *FFmpegExecutor {
	return &FFmpegExecutor{
		config: config,
		logger: zap.NewNop(), // Will be set by caller
	}
}

// Execute runs an FFmpeg command for a job
func (fe *FFmpegExecutor) Execute(ctx context.Context, job *queue.Job) error {
	fe.logger.Info("Executing FFmpeg command",
		zap.String("job_id", job.ID),
		zap.String("input_path", job.InputPath),
		zap.String("output_path", job.OutputPath))

	// Build FFmpeg command
	args := fe.buildFFmpegArgs(job)

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(fe.config.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, fe.config.ExecutablePath, args...)

	// Set up command output capture
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fe.logger.Debug("FFmpeg command",
		zap.String("executable", fe.config.ExecutablePath),
		zap.Strings("args", args))

	// Execute command
	if err := cmd.Run(); err != nil {
		fe.logger.Error("FFmpeg execution failed",
			zap.String("job_id", job.ID),
			zap.Error(err),
			zap.String("stdout", stdout.String()),
			zap.String("stderr", stderr.String()))
		return fmt.Errorf("FFmpeg execution failed: %w (stderr: %s)", err, stderr.String())
	}

	fe.logger.Info("FFmpeg execution completed",
		zap.String("job_id", job.ID),
		zap.String("output_path", job.OutputPath))

	return nil
}

// buildFFmpegArgs builds the FFmpeg command arguments
func (fe *FFmpegExecutor) buildFFmpegArgs(job *queue.Job) []string {
	var args []string

	// Add default arguments
	args = append(args, fe.config.DefaultArgs...)

	// Add input file
	args = append(args, "-i", job.InputPath)

	// Add custom FFmpeg arguments if provided
	if job.FFmpegArgs != "" {
		// Split arguments by space (simple parsing)
		customArgs := strings.Fields(job.FFmpegArgs)
		args = append(args, customArgs...)
	}

	// Add output file
	args = append(args, job.OutputPath)

	return args
}

// Validate checks if FFmpeg is available and working
func (fe *FFmpegExecutor) Validate() error {
	cmd := exec.Command(fe.config.ExecutablePath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg not found or not executable: %w", err)
	}
	return nil
}

// GetPreset returns a preset configuration
func (fe *FFmpegExecutor) GetPreset(name string) (string, bool) {
	preset, exists := fe.config.Presets[name]
	return preset, exists
}
