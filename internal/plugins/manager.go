package plugins

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"runtime"
	"time"

	"github.com/flixsrota/flixsrota/internal/queue"
	"github.com/flixsrota/flixsrota/internal/storage"
	"go.uber.org/zap"
)

// PluginManager manages dynamic loading of queue and storage adapters
type PluginManager struct {
	logger     *zap.Logger
	pluginDir  string
	downloaded map[string]string // adapter name -> plugin path
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(logger *zap.Logger) *PluginManager {
	pluginDir := filepath.Join(os.TempDir(), "flixsrota", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		logger.Warn("Failed to create plugin directory", zap.Error(err))
	}

	return &PluginManager{
		logger:     logger,
		pluginDir:  pluginDir,
		downloaded: make(map[string]string),
	}
}

// DownloadQueueAdapter downloads a queue adapter plugin
func (pm *PluginManager) DownloadQueueAdapter(ctx context.Context, adapterName, downloadURL string) error {
	pm.logger.Info("Downloading queue adapter",
		zap.String("adapter", adapterName),
		zap.String("url", downloadURL))

	pluginPath := filepath.Join(pm.pluginDir, fmt.Sprintf("queue_%s.so", adapterName))

	if err := pm.downloadPlugin(ctx, downloadURL, pluginPath); err != nil {
		return fmt.Errorf("failed to download queue adapter %s: %w", adapterName, err)
	}

	pm.downloaded[fmt.Sprintf("queue_%s", adapterName)] = pluginPath
	pm.logger.Info("Queue adapter downloaded", zap.String("path", pluginPath))

	return nil
}

// DownloadStorageAdapter downloads a storage adapter plugin
func (pm *PluginManager) DownloadStorageAdapter(ctx context.Context, adapterName, downloadURL string) error {
	pm.logger.Info("Downloading storage adapter",
		zap.String("adapter", adapterName),
		zap.String("url", downloadURL))

	pluginPath := filepath.Join(pm.pluginDir, fmt.Sprintf("storage_%s.so", adapterName))

	if err := pm.downloadPlugin(ctx, downloadURL, pluginPath); err != nil {
		return fmt.Errorf("failed to download storage adapter %s: %w", adapterName, err)
	}

	pm.downloaded[fmt.Sprintf("storage_%s", adapterName)] = pluginPath
	pm.logger.Info("Storage adapter downloaded", zap.String("path", pluginPath))

	return nil
}

// LoadQueueAdapter loads a queue adapter plugin
func (pm *PluginManager) LoadQueueAdapter(adapterName string) (queue.Queue, error) {
	pluginPath := pm.downloaded[fmt.Sprintf("queue_%s", adapterName)]
	if pluginPath == "" {
		return nil, fmt.Errorf("queue adapter %s not downloaded", adapterName)
	}

	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open queue plugin %s: %w", adapterName, err)
	}

	// Look for the NewQueue function
	newQueueSym, err := p.Lookup("NewQueue")
	if err != nil {
		return nil, fmt.Errorf("queue plugin %s missing NewQueue function: %w", adapterName, err)
	}

	newQueue, ok := newQueueSym.(func() (queue.Queue, error))
	if !ok {
		return nil, fmt.Errorf("queue plugin %s NewQueue function has wrong signature", adapterName)
	}

	return newQueue()
}

// LoadStorageAdapter loads a storage adapter plugin
func (pm *PluginManager) LoadStorageAdapter(adapterName string) (storage.Storage, error) {
	pluginPath := pm.downloaded[fmt.Sprintf("storage_%s", adapterName)]
	if pluginPath == "" {
		return nil, fmt.Errorf("storage adapter %s not downloaded", adapterName)
	}

	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage plugin %s: %w", adapterName, err)
	}

	// Look for the NewStorage function
	newStorageSym, err := p.Lookup("NewStorage")
	if err != nil {
		return nil, fmt.Errorf("storage plugin %s missing NewStorage function: %w", adapterName, err)
	}

	newStorage, ok := newStorageSym.(func() (storage.Storage, error))
	if !ok {
		return nil, fmt.Errorf("storage plugin %s NewStorage function has wrong signature", adapterName)
	}

	return newStorage()
}

// downloadPlugin downloads a plugin from URL
func (pm *PluginManager) downloadPlugin(ctx context.Context, downloadURL, pluginPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	file, err := os.Create(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to create plugin file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write plugin file: %w", err)
	}

	// Make the plugin executable
	if err := os.Chmod(pluginPath, 0755); err != nil {
		return fmt.Errorf("failed to make plugin executable: %w", err)
	}

	return nil
}

// GetAvailableAdapters returns a list of available adapters
func (pm *PluginManager) GetAvailableAdapters() map[string][]string {
	return map[string][]string{
		"queue": {
			"redis",
			"kafka",
			"sqs",
			"rabbitmq",
		},
		"storage": {
			"local",
			"s3",
			"gcs",
			"azure",
			"minio",
		},
	}
}

// GetAdapterDownloadURL returns the download URL for an adapter
func (pm *PluginManager) GetAdapterDownloadURL(adapterType, adapterName string) string {
	baseURL := "https://github.com/flixsrota/flixsrota-plugins/releases/latest/download"
	return fmt.Sprintf("%s/%s_%s_%s_%s.so", baseURL, adapterType, adapterName, runtime.GOOS, runtime.GOARCH)
}
