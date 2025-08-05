package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RunWizard runs the interactive configuration wizard
func RunWizard(configPath string) error {
	fmt.Println("🎬 Flixsrota Configuration Wizard")
	fmt.Println("==================================")
	fmt.Println()

	cfg := DefaultConfig()

	// Get config file path
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".flixsrota.yaml")
	}

	fmt.Printf("Configuration will be saved to: %s\n", configPath)
	fmt.Println()

	// GRPC Configuration
	fmt.Println("📡 gRPC Server Configuration")
	fmt.Println("----------------------------")
	cfg.GRPC.Address = promptString("Server address", cfg.GRPC.Address)
	cfg.GRPC.Port = promptInt("Server port", cfg.GRPC.Port)
	fmt.Println()

	// Queue Configuration
	fmt.Println("📋 Queue Configuration")
	fmt.Println("----------------------")
	queueAdapter := promptChoice("Queue adapter", []string{"redis", "kafka", "sqs"}, cfg.Queue.Adapter)
	cfg.Queue.Adapter = queueAdapter

	switch queueAdapter {
	case "redis":
		cfg.Queue.Redis.Address = promptString("Redis address", cfg.Queue.Redis.Address)
		cfg.Queue.Redis.Password = promptPassword("Redis password (leave empty if none)")
	case "kafka":
		brokers := promptString("Kafka brokers (comma-separated)", "localhost:9092")
		cfg.Queue.Kafka.Brokers = strings.Split(brokers, ",")
		cfg.Queue.Kafka.Topic = promptString("Kafka topic", "flixsrota-jobs")
	case "sqs":
		cfg.Queue.SQS.Region = promptString("AWS region", "us-east-1")
		cfg.Queue.SQS.QueueURL = promptString("SQS queue URL", "")
	}
	fmt.Println()

	// Storage Configuration
	fmt.Println("💾 Storage Configuration")
	fmt.Println("------------------------")
	storageAdapter := promptChoice("Storage adapter", []string{"local", "s3", "gcs"}, cfg.Storage.Adapter)
	cfg.Storage.Adapter = storageAdapter

	switch storageAdapter {
	case "local":
		cfg.Storage.Local.BasePath = promptString("Base storage path", cfg.Storage.Local.BasePath)
		cfg.Storage.Local.TempPath = promptString("Temporary files path", cfg.Storage.Local.TempPath)
	case "s3":
		cfg.Storage.S3.Region = promptString("AWS region", "us-east-1")
		cfg.Storage.S3.Bucket = promptString("S3 bucket name", "")
	case "gcs":
		cfg.Storage.GCS.ProjectID = promptString("Google Cloud project ID", "")
		cfg.Storage.GCS.Bucket = promptString("GCS bucket name", "")
	}
	fmt.Println()

	// FFmpeg Configuration
	fmt.Println("🎥 FFmpeg Configuration")
	fmt.Println("-----------------------")
	cfg.FFmpeg.ExecutablePath = promptString("FFmpeg executable path", cfg.FFmpeg.ExecutablePath)
	cfg.FFmpeg.Timeout = promptInt("Job timeout (seconds)", cfg.FFmpeg.Timeout)

	// Video Quality Configuration
	fmt.Println("🎬 Video Quality Configuration")
	fmt.Println("------------------------------")
	enableQualityDetection := promptBool("Enable automatic quality detection", cfg.FFmpeg.Quality.EnableQualityDetection)
	cfg.FFmpeg.Quality.EnableQualityDetection = enableQualityDetection

	if enableQualityDetection {
		maxQuality := promptChoice("Maximum output quality", []string{"480p", "720p", "1080p", "4k", "8k"}, cfg.FFmpeg.Quality.MaxQuality)
		cfg.FFmpeg.Quality.MaxQuality = maxQuality

		fmt.Println("Supported resolutions (comma-separated):")
		supportedResolutions := promptString("Supported resolutions", strings.Join(cfg.FFmpeg.Quality.SupportedResolutions, ","))
		cfg.FFmpeg.Quality.SupportedResolutions = strings.Split(supportedResolutions, ",")
	}
	fmt.Println()

	// Worker Configuration
	fmt.Println("👷 Worker Configuration")
	fmt.Println("----------------------")
	cfg.Worker.MinWorkers = promptInt("Minimum workers", cfg.Worker.MinWorkers)
	cfg.Worker.MaxWorkers = promptInt("Maximum workers", cfg.Worker.MaxWorkers)
	fmt.Println()

	// Plugin Download Configuration
	fmt.Println("🔌 Plugin Configuration")
	fmt.Println("----------------------")
	downloadPlugins := promptBool("Download required plugins automatically", true)
	if downloadPlugins {
		fmt.Println("Plugins will be downloaded on first run")
	}
	fmt.Println()

	// Save configuration
	if err := Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ Configuration saved successfully!")
	fmt.Printf("📁 Location: %s\n", configPath)
	fmt.Println()
	fmt.Println("🚀 You can now start Flixsrota with:")
	fmt.Printf("   flixsrota serve --config %s\n", configPath)
	fmt.Println()

	return nil
}

// promptString prompts for a string input
func promptString(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

// promptInt prompts for an integer input
func promptInt(prompt string, defaultValue int) int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%d]: ", prompt, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	if value, err := strconv.Atoi(input); err == nil {
		return value
	}
	return defaultValue
}

// promptBool prompts for a boolean input
func promptBool(prompt string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "yes"
}

// promptChoice prompts for a choice from a list of options
func promptChoice(prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s:\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter choice [%s]: ", defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	if choice, err := strconv.Atoi(input); err == nil && choice > 0 && choice <= len(choices) {
		return choices[choice-1]
	}

	// Try to match by name
	for _, choice := range choices {
		if strings.EqualFold(input, choice) {
			return choice
		}
	}

	return defaultValue
}

// promptPassword prompts for a password input
func promptPassword(prompt string) string {
	fmt.Printf("%s: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
