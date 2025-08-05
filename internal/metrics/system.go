package metrics

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
)

// SystemMetricsCollector collects system resource metrics
type SystemMetricsCollector struct {
	logger *zap.Logger
	ctx    context.Context
}

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector(logger *zap.Logger) *SystemMetricsCollector {
	return &SystemMetricsCollector{
		logger: logger,
		ctx:    context.Background(),
	}
}

// SystemMetrics contains system resource information
type SystemMetrics struct {
	CPUUsagePercent      float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent   float64 `json:"memory_usage_percent"`
	DiskUsagePercent     float64 `json:"disk_usage_percent"`
	TotalMemoryBytes     uint64  `json:"total_memory_bytes"`
	AvailableMemoryBytes uint64  `json:"available_memory_bytes"`
	TotalDiskBytes       uint64  `json:"total_disk_bytes"`
	AvailableDiskBytes   uint64  `json:"available_disk_bytes"`
	ActiveWorkerCount    int     `json:"active_worker_count"`
	MaxWorkerCount       int     `json:"max_worker_count"`
	Goroutines           int     `json:"goroutines"`
	HeapAllocBytes       uint64  `json:"heap_alloc_bytes"`
	HeapSysBytes         uint64  `json:"heap_sys_bytes"`
}

// CollectMetrics collects current system metrics
func (smc *SystemMetricsCollector) CollectMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		smc.logger.Warn("Failed to get CPU usage", zap.Error(err))
		metrics.CPUUsagePercent = 0
	} else if len(cpuPercent) > 0 {
		metrics.CPUUsagePercent = cpuPercent[0]
	}

	// Memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		smc.logger.Warn("Failed to get memory info", zap.Error(err))
		metrics.MemoryUsagePercent = 0
		metrics.TotalMemoryBytes = 0
		metrics.AvailableMemoryBytes = 0
	} else {
		metrics.MemoryUsagePercent = memInfo.UsedPercent
		metrics.TotalMemoryBytes = memInfo.Total
		metrics.AvailableMemoryBytes = memInfo.Available
	}

	// Disk usage (using current directory)
	diskInfo, err := disk.Usage(".")
	if err != nil {
		smc.logger.Warn("Failed to get disk info", zap.Error(err))
		metrics.DiskUsagePercent = 0
		metrics.TotalDiskBytes = 0
		metrics.AvailableDiskBytes = 0
	} else {
		metrics.DiskUsagePercent = diskInfo.UsedPercent
		metrics.TotalDiskBytes = diskInfo.Total
		metrics.AvailableDiskBytes = diskInfo.Free
	}

	// Go runtime metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	metrics.Goroutines = runtime.NumGoroutine()
	metrics.HeapAllocBytes = m.HeapAlloc
	metrics.HeapSysBytes = m.HeapSys

	// Worker metrics (placeholder - will be set by job processor)
	metrics.ActiveWorkerCount = 0
	metrics.MaxWorkerCount = 10

	return metrics, nil
}

// GetProcessMetrics gets metrics for the current process
func (smc *SystemMetricsCollector) GetProcessMetrics() (*ProcessMetrics, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		smc.logger.Warn("Failed to get process CPU usage", zap.Error(err))
		cpuPercent = 0
	}

	memPercent, err := proc.MemoryPercent()
	if err != nil {
		smc.logger.Warn("Failed to get process memory usage", zap.Error(err))
		memPercent = 0
	}

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		smc.logger.Warn("Failed to get process memory info", zap.Error(err))
		memInfo = &process.MemoryInfoStat{}
	}

	return &ProcessMetrics{
		CPUPercent:    cpuPercent,
		MemoryPercent: memPercent,
		RSSBytes:      memInfo.RSS,
		VMSBytes:      memInfo.VMS,
	}, nil
}

// ProcessMetrics contains process-specific metrics
type ProcessMetrics struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	RSSBytes      uint64  `json:"rss_bytes"`
	VMSBytes      uint64  `json:"vms_bytes"`
}
