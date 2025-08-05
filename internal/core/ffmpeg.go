package core

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/nikhil0verma/flixsrota/internal/config"
	"github.com/nikhil0verma/flixsrota/internal/queue"
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

	// Add input file
	args = append(args, "-i", job.InputPath)

	// Build the filter_complex string dynamically
	var filterComplexParts []string
	var videoMapParts []string
	var audioMapParts []string

	// Keep track of the stream labels for video and audio (e.g., [v1out], [v2out], ...)
	var videoStreamIndex int
	// var audioStreamIndex int // TODO: handle audio stream

	// Build the filter_complex part (for splitting and scaling)
	for quality := range fe.config.Qualities {
		if fe.config.Qualities[quality] { // Only process enabled qualities
			// For each quality, add a split and scale
			var resolution string
			var bitrate string
			switch quality {
			case "360p":
				resolution = "854x480"
				bitrate = "1M"
			case "480p":
				resolution = "1280x720"
				bitrate = "1.5M"
			case "720p":
				resolution = "1280x720"
				bitrate = "3M"
			case "1080p":
				resolution = "1920x1080"
				bitrate = "5M"
			case "2K":
				resolution = "2048x1080"
				bitrate = "7M"
			case "4K":
				resolution = "3840x2160"
				bitrate = "10M"
			case "8K":
				resolution = "7680x4320"
				bitrate = "20M"
			default:
				// If an unknown quality is found, skip
				continue
			}

			// Add scale filter for this quality
			filterComplexParts = append(filterComplexParts,
				fmt.Sprintf("[%d:v]scale=w=%s:h=%s[v%dout]", 0, resolution, resolution, videoStreamIndex),
			)

			// Add video mapping for this quality
			videoMapParts = append(videoMapParts,
				fmt.Sprintf("-map [v%dout] -c:v:%d libx264 -x264-params \"nal-hrd=cbr:force-cfr=1\" -b:v:%d %s -maxrate:v:%d %s -minrate:v:%d %s -bufsize:v:%d %s -preset slow -g 48 -sc_threshold 0 -keyint_min 48",
					videoStreamIndex, videoStreamIndex, videoStreamIndex, bitrate, videoStreamIndex, bitrate, videoStreamIndex, bitrate, videoStreamIndex, bitrate),
			)

			// Increment the video stream index
			videoStreamIndex++
		}
	}

	// Add audio mappings (assuming you want to map the same audio for all streams)
	audioMapParts = append(audioMapParts,
		"-map a:0 -c:a:0 aac -b:a:0 96k -ac 2",
		"-map a:0 -c:a:1 aac -b:a:1 96k -ac 2",
		"-map a:0 -c:a:2 aac -b:a:2 48k -ac 2",
	)

	// Combine all parts together
	if len(filterComplexParts) > 0 {
		args = append(args, "-filter_complex")
		args = append(args, strings.Join(filterComplexParts, "; ")+";")
	}

	// Add video mapping parts
	args = append(args, videoMapParts...)

	// Add audio mapping parts
	args = append(args, audioMapParts...)

	// HLS-specific options
	args = append(args,
		"-f hls",
		"-hls_time 2",
		"-hls_playlist_type vod",
		"-hls_flags independent_segments",
		"-hls_segment_type mpegts",
		"-hls_segment_filename stream_%v/data%02d.ts",
		"-master_pl_name srota.m3u8",
		"-var_stream_map \"v:0,a:0 v:1,a:1 v:2,a:2 v:3,a:0 v:4,a:1 v:5,a:2 v:6,a:0 v:7,a:1\"",
		"stream_%v.m3u8",
	)

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
