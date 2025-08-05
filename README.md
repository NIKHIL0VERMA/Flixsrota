# 🎬 Flixsrota

A modular, high-performance video processing backend service that interfaces with FFmpeg using gRPC APIs. Flixsrota supports pluggable queue and storage systems and is designed to run as a backend service/daemon.

## 🚀 Features

- **gRPC API**: High-performance video processing APIs
- **FFmpeg Integration**: Direct FFmpeg process management
- **Modular Architecture**: Pluggable queue and storage adapters
- **Cross-Platform**: Builds for Linux, macOS, and Windows (x86_64 & ARM64)
- **Resource Monitoring**: CPU, memory, and queue depth monitoring
- **Backpressure Control**: Intelligent job admission based on system load
- **CLI Management**: Interactive configuration wizard and service management

## 📋 Prerequisites

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **FFmpeg**: [Install FFmpeg](https://ffmpeg.org/download.html)
- **Redis** (optional): For queue backend
- **Protobuf Compiler** (optional): For development

### Installing FFmpeg

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ffmpeg
```

**macOS:**
```bash
brew install ffmpeg
```

**Windows:**
Download from [FFmpeg official website](https://ffmpeg.org/download.html#build-windows)

## 🛠 Installation

### From Source

1. **Clone the repository:**
```bash
git clone https://github.com/flixsrota/flixsrota.git
cd flixsrota
```

2. **Install dependencies:**
```bash
make deps
```

3. **Build for your platform:**
```bash
make build
```

### Cross-Platform Build

Build for all platforms (Linux, macOS, Windows):
```bash
make build-all
```

Build for specific platform:
```bash
# Linux (x86_64 & ARM64)
make build-linux

# macOS (x86_64 & ARM64)
make build-darwin

# Windows (x86_64 & ARM64)
make build-windows
```

## 🚀 Quick Start

### 1. Initialize Configuration

Run the interactive configuration wizard:
```bash
./build/flixsrota config init
```

This will create a configuration file at `~/.flixsrota.yaml`.

### 2. Start the Server

```bash
./build/flixsrota serve
```

The server will start on the configured gRPC port (default: 50051).

### 3. Test the API

You can test the API using a gRPC client like `grpcurl`:

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:50051 list

# Get system metrics
grpcurl -plaintext localhost:50051 flixsrota.SystemMetrics/GetMetrics
```

## 📁 Configuration

Flixsrota uses YAML configuration files. The default location is `~/.flixsrota.yaml`.

### Configuration Structure

```yaml
grpc:
  address: "0.0.0.0"
  port: 50051
  max_concurrent: 100
  enable_reflection: true

queue:
  adapter: "redis"
  redis:
    address: "localhost:6379"
    password: ""
    db: 0
    pool_size: 10

storage:
  adapter: "local"
  local:
    base_path: "/tmp/flixsrota"
    temp_path: "/tmp/flixsrota/temp"

ffmpeg:
  executable_path: "ffmpeg"
  default_args: ["-y"]
  presets:
    h264: "-c:v libx264 -preset medium -crf 23"
    h265: "-c:v libx265 -preset medium -crf 28"
    webm: "-c:v libvpx-vp9 -crf 30 -b:v 0"
  timeout: 3600

worker:
  min_workers: 2
  max_workers: 10
  queue_size: 100
  idle_timeout: 300

metrics:
  enabled: true
  port: 9090
  path: "/metrics"
  collect_interval: 30

logging:
  level: "info"
  format: "json"
  output_path: ""
```

### Environment Variables

You can override configuration values using environment variables:

```bash
export FLIXSROTA_GRPC_PORT=50052
export FLIXSROTA_QUEUE_ADAPTER=redis
export FLIXSROTA_QUEUE_REDIS_ADDRESS=localhost:6379
```

## 🔧 CLI Commands

### Configuration Management

```bash
# Initialize configuration
flixsrota config init

# Validate configuration
flixsrota config validate
```

### Server Management

```bash
# Start the server
flixsrota serve

# Start with custom config file
flixsrota serve --config /path/to/config.yaml

# Start with debug logging
flixsrota serve --log-level debug
```

## 🏗 Architecture

```
flixsrota/
├── cmd/flixsrota/          # CLI entry point
├── internal/
│   ├── core/              # Job processor, FFmpeg executor
│   ├── grpc/              # gRPC server and services
│   ├── queue/             # Queue interfaces and adapters
│   ├── storage/           # Storage interfaces and adapters
│   ├── config/            # Configuration management
│   └── metrics/           # System metrics collection
├── plugins/               # External plugins (future)
├── proto/                 # Protobuf definitions
└── pkg/utils/             # Shared utilities
```

## 🔌 Queue Adapters

### Redis (Default)

Redis provides a reliable, in-memory queue with persistence:

```yaml
queue:
  adapter: "redis"
  redis:
    address: "localhost:6379"
    password: ""
    db: 0
    pool_size: 10
```

### Kafka (Planned)

```yaml
queue:
  adapter: "kafka"
  kafka:
    brokers: ["localhost:9092"]
    topic: "flixsrota-jobs"
    group_id: "flixsrota-workers"
```

### AWS SQS (Planned)

```yaml
queue:
  adapter: "sqs"
  sqs:
    region: "us-east-1"
    queue_url: "https://sqs.us-east-1.amazonaws.com/..."
```

## 💾 Storage Adapters

### Local Storage (Default)

```yaml
storage:
  adapter: "local"
  local:
    base_path: "/tmp/flixsrota"
    temp_path: "/tmp/flixsrota/temp"
```

### AWS S3 (Planned)

```yaml
storage:
  adapter: "s3"
  s3:
    region: "us-east-1"
    bucket: "my-video-bucket"
    access_key_id: "AKIA..."
    secret_access_key: "..."
```

### Google Cloud Storage (Planned)

```yaml
storage:
  adapter: "gcs"
  gcs:
    project_id: "my-project"
    bucket: "my-video-bucket"
    credentials_file: "/path/to/service-account.json"
```

## 📊 Monitoring

### Metrics Endpoint

When enabled, Flixsrota exposes Prometheus metrics at `/metrics`:

```bash
curl http://localhost:9090/metrics
```

### System Metrics

The gRPC API provides real-time system metrics:

```bash
grpcurl -plaintext localhost:50051 flixsrota.SystemMetrics/GetMetrics
```

## 🧪 Development

### Prerequisites

1. **Install development tools:**
```bash
# Install protobuf compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Install linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Development Commands

```bash
# Install dependencies
make deps

# Generate protobuf code
make proto

# Run tests
make test

# Run tests with coverage
make test-coverage

# Lint code
make lint

# Format code
make fmt

# Build and run
make run

# Run configuration wizard
make run-config
```

### Adding New Queue Adapters

1. Implement the `Queue` interface in `internal/queue/`
2. Add configuration options in `internal/config/config.go`
3. Register the adapter in `internal/core/server.go`

### Adding New Storage Adapters

1. Implement the `Storage` interface in `internal/storage/`
2. Add configuration options in `internal/config/config.go`
3. Register the adapter in `internal/core/server.go`

## 🐳 Docker

### Build Docker Image

```bash
docker build -t flixsrota .
```

### Run with Docker

```bash
docker run -p 50051:50051 -p 9090:9090 flixsrota
```

## 📝 API Reference

### Video Processing

```protobuf
service VideoProcessor {
  rpc ProcessVideo(ProcessVideoRequest) returns (ProcessVideoResponse);
  rpc GetJobStatus(GetJobStatusRequest) returns (GetJobStatusResponse);
  rpc CancelJob(CancelJobRequest) returns (CancelJobResponse);
  rpc ListJobs(ListJobsRequest) returns (ListJobsResponse);
}
```

### System Metrics

```protobuf
service SystemMetrics {
  rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);
  rpc StreamMetrics(StreamMetricsRequest) returns (stream StreamMetricsResponse);
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new functionality
- Update documentation for API changes
- Use conventional commit messages

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/flixsrota/flixsrota/issues)
- **Discussions**: [GitHub Discussions](https://github.com/flixsrota/flixsrota/discussions)
- **Documentation**: [Wiki](https://github.com/flixsrota/flixsrota/wiki)

## 🗺 Roadmap

- [ ] Kafka queue adapter
- [ ] AWS SQS queue adapter
- [ ] AWS S3 storage adapter
- [ ] Google Cloud Storage adapter
- [ ] REST API gateway
- [ ] Web UI dashboard
- [ ] Kubernetes operator
- [ ] Helm charts
- [ ] Prometheus exporter
- [ ] Grafana dashboards 