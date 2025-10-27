package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	GRPC    GRPCConfig    `mapstructure:"grpc" yaml:"grpc"`
	Queue   QueueConfig   `mapstructure:"queue" yaml:"queue"`
	Storage StorageConfig `mapstructure:"storage" yaml:"storage"`
	FFmpeg  FFmpegConfig  `mapstructure:"ffmpeg" yaml:"ffmpeg"`
	Worker  WorkerConfig  `mapstructure:"worker" yaml:"worker"`
	Metrics MetricsConfig `mapstructure:"metrics" yaml:"metrics"`
	Logging LoggingConfig `mapstructure:"logging" yaml:"logging"`
}

// GRPCConfig contains gRPC server settings
type GRPCConfig struct {
	Address          string `mapstructure:"address" yaml:"address"`
	Port             int    `mapstructure:"port" yaml:"port"`
	MaxConcurrent    int    `mapstructure:"max_concurrent" yaml:"max_concurrent"`
	EnableReflection bool   `mapstructure:"enable_reflection" yaml:"enable_reflection"`
}

// QueueConfig contains queue adapter settings
type QueueConfig struct {
	Adapter string           `mapstructure:"adapter" yaml:"adapter"`
	Redis   RedisQueueConfig `mapstructure:"redis" yaml:"redis"`
	Kafka   KafkaQueueConfig `mapstructure:"kafka" yaml:"kafka"`
	SQS     SQSQueueConfig   `mapstructure:"sqs" yaml:"sqs"`
}

// RedisQueueConfig contains Redis-specific settings
type RedisQueueConfig struct {
	Address  string `mapstructure:"address" yaml:"address"`
	Password string `mapstructure:"password" yaml:"password"`
	DB       int    `mapstructure:"db" yaml:"db"`
	PoolSize int    `mapstructure:"pool_size" yaml:"pool_size"`
}

// KafkaQueueConfig contains Kafka-specific settings
type KafkaQueueConfig struct {
	Brokers []string `mapstructure:"brokers" yaml:"brokers"`
	Topic   string   `mapstructure:"topic" yaml:"topic"`
	GroupID string   `mapstructure:"group_id" yaml:"group_id"`
}

// SQSQueueConfig contains AWS SQS-specific settings
type SQSQueueConfig struct {
	Region          string `mapstructure:"region" yaml:"region"`
	QueueURL        string `mapstructure:"queue_url" yaml:"queue_url"`
	MaxMessages     int    `mapstructure:"max_messages" yaml:"max_messages"`
	WaitTimeSeconds int    `mapstructure:"wait_time_seconds" yaml:"wait_time_seconds"`
}

// StorageConfig contains storage adapter settings
type StorageConfig struct {
	Adapter string             `mapstructure:"adapter" yaml:"adapter"`
	Local   LocalStorageConfig `mapstructure:"local" yaml:"local"`
	S3      S3StorageConfig    `mapstructure:"s3" yaml:"s3"`
	GCS     GCSStorageConfig   `mapstructure:"gcs" yaml:"gcs"`
}

// LocalStorageConfig contains local file storage settings
type LocalStorageConfig struct {
	BasePath string `mapstructure:"base_path" yaml:"base_path"`
	TempPath string `mapstructure:"temp_path" yaml:"temp_path"`
}

// S3StorageConfig contains AWS S3 settings
type S3StorageConfig struct {
	Region          string `mapstructure:"region" yaml:"region"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key"`
}

// GCSStorageConfig contains Google Cloud Storage settings
type GCSStorageConfig struct {
	ProjectID       string `mapstructure:"project_id" yaml:"project_id"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
	CredentialsFile string `mapstructure:"credentials_file" yaml:"credentials_file"`
}

// FFmpegConfig contains FFmpeg execution settings
type FFmpegConfig struct {
	ExecutablePath string          `mapstructure:"executable_path" yaml:"executable_path"`
	Timeout        int             `mapstructure:"timeout" yaml:"timeout"`
	Qualities      map[string]bool `mapstructure:"qualities" yaml:"qualities"`
}

// WorkerConfig contains worker pool settings
type WorkerConfig struct {
	MinWorkers  int `mapstructure:"min_workers" yaml:"min_workers"`
	MaxWorkers  int `mapstructure:"max_workers" yaml:"max_workers"`
	QueueSize   int `mapstructure:"queue_size" yaml:"queue_size"`
	IdleTimeout int `mapstructure:"idle_timeout" yaml:"idle_timeout"`
}

// MetricsConfig contains metrics collection settings
type MetricsConfig struct {
	Enabled         bool   `mapstructure:"enabled" yaml:"enabled"`
	Port            int    `mapstructure:"port" yaml:"port"`
	Path            string `mapstructure:"path" yaml:"path"`
	CollectInterval int    `mapstructure:"collect_interval" yaml:"collect_interval"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level" yaml:"level"`
	Format     string `mapstructure:"format" yaml:"format"`
	OutputPath string `mapstructure:"output_path" yaml:"output_path"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		GRPC: GRPCConfig{
			Address:          "0.0.0.0",
			Port:             50051,
			MaxConcurrent:    100,
			EnableReflection: true,
		},
		Queue: QueueConfig{
			Adapter: "redis",
			Redis: RedisQueueConfig{
				Address:  "localhost:6379",
				Password: "",
				DB:       0,
				PoolSize: 10,
			},
		},
		Storage: StorageConfig{
			Adapter: "local",
			Local: LocalStorageConfig{
				BasePath: "/tmp/flixsrota",
				TempPath: "/tmp/flixsrota/temp",
			},
		},
		FFmpeg: FFmpegConfig{
			ExecutablePath: "ffmpeg",
			Timeout:        3600,
			Qualities: map[string]bool{
				"360p":  true,
				"480p":  true,
				"720p":  true,
				"1080p": false,
				"2160p": false,
				"4320p": false,
				"8640p": false,
			},
		},
		Worker: WorkerConfig{
			MinWorkers:  2,
			MaxWorkers:  10,
			QueueSize:   100,
			IdleTimeout: 300,
		},
		Metrics: MetricsConfig{
			Enabled:         true,
			Port:            9090,
			Path:            "/metrics",
			CollectInterval: 30,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "",
		},
	}
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	cfg := DefaultConfig()
	v.SetConfigType("yaml")

	// Set default values
	setDefaults(v, cfg)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Search for config file in common locations
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
		v.AddConfigPath("/etc/flixsrota")
		v.SetConfigName(".flixsrota")
		v.SetConfigType("yaml")
	}

	// Read environment variables
	v.SetEnvPrefix("FLIXSROTA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			v.SetConfigFile("") // Clear config file to use defaults only
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into config struct
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.GRPC.Port <= 0 || c.GRPC.Port > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", c.GRPC.Port)
	}

	if c.Worker.MinWorkers < 1 {
		return fmt.Errorf("min workers must be at least 1")
	}

	if c.Worker.MaxWorkers < c.Worker.MinWorkers {
		return fmt.Errorf("max workers must be greater than or equal to min workers")
	}

	if c.FFmpeg.Timeout <= 0 {
		return fmt.Errorf("FFmpeg timeout must be positive")
	}

	return nil
}

// setDefaults sets default values in viper
func setDefaults(v *viper.Viper, cfg *Config) {
	// GRPC defaults
	v.SetDefault("grpc.address", cfg.GRPC.Address)
	v.SetDefault("grpc.port", cfg.GRPC.Port)
	v.SetDefault("grpc.max_concurrent", cfg.GRPC.MaxConcurrent)
	v.SetDefault("grpc.enable_reflection", cfg.GRPC.EnableReflection)

	// Queue defaults
	v.SetDefault("queue.adapter", cfg.Queue.Adapter)
	v.SetDefault("queue.redis.address", cfg.Queue.Redis.Address)
	v.SetDefault("queue.redis.password", cfg.Queue.Redis.Password)
	v.SetDefault("queue.redis.db", cfg.Queue.Redis.DB)
	v.SetDefault("queue.redis.pool_size", cfg.Queue.Redis.PoolSize)

	// Storage defaults
	v.SetDefault("storage.adapter", cfg.Storage.Adapter)
	v.SetDefault("storage.local.base_path", cfg.Storage.Local.BasePath)
	v.SetDefault("storage.local.temp_path", cfg.Storage.Local.TempPath)

	// FFmpeg defaults
	v.SetDefault("ffmpeg.executable_path", cfg.FFmpeg.ExecutablePath)
	v.SetDefault("ffmpeg.timeout", cfg.FFmpeg.Timeout)
	v.SetDefault("ffmpeg.qualities", cfg.FFmpeg.Qualities)

	// Worker defaults
	v.SetDefault("worker.min_workers", cfg.Worker.MinWorkers)
	v.SetDefault("worker.max_workers", cfg.Worker.MaxWorkers)
	v.SetDefault("worker.queue_size", cfg.Worker.QueueSize)
	v.SetDefault("worker.idle_timeout", cfg.Worker.IdleTimeout)

	// Metrics defaults
	v.SetDefault("metrics.enabled", cfg.Metrics.Enabled)
	v.SetDefault("metrics.port", cfg.Metrics.Port)
	v.SetDefault("metrics.path", cfg.Metrics.Path)
	v.SetDefault("metrics.collect_interval", cfg.Metrics.CollectInterval)

	// Logging defaults
	v.SetDefault("logging.level", cfg.Logging.Level)
	v.SetDefault("logging.format", cfg.Logging.Format)
	v.SetDefault("logging.output_path", cfg.Logging.OutputPath)
}

// GetString returns a string value from environment or config
func GetString(key, defaultValue string) string {
	if value := os.Getenv("FLIXSROTA_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))); value != "" {
		return value
	}
	return defaultValue
}

// GetInt returns an integer value from environment or config
func GetInt(key string, defaultValue int) int {
	if value := os.Getenv("FLIXSROTA_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetBool returns a boolean value from environment or config
func GetBool(key string, defaultValue bool) bool {
	if value := os.Getenv("FLIXSROTA_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
