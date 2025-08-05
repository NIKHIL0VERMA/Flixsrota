package storage

import (
	"context"
	"os"
)

// Storage interface defines the methods for storage operations
type Storage interface {
	// Upload uploads a file to storage
	Upload(ctx context.Context, localPath, remotePath string) error
	
	// Download downloads a file from storage
	Download(ctx context.Context, remotePath, localPath string) error
	
	// Delete deletes a file from storage
	Delete(ctx context.Context, remotePath string) error
	
	// Exists checks if a file exists in storage
	Exists(ctx context.Context, remotePath string) (bool, error)
	
	// GetURL returns a URL for accessing a file
	GetURL(ctx context.Context, remotePath string) (string, error)
	
	// ListFiles lists files in a directory
	ListFiles(ctx context.Context, prefix string) ([]string, error)
	
	// CreateTempFile creates a temporary file
	CreateTempFile(ctx context.Context, suffix string) (*os.File, error)
	
	// Close closes the storage connection
	Close() error
}

// StorageMetrics contains storage performance metrics
type StorageMetrics struct {
	TotalSize     int64   `json:"total_size_bytes"`
	UsedSize      int64   `json:"used_size_bytes"`
	FreeSize      int64   `json:"free_size_bytes"`
	UploadCount   int64   `json:"upload_count"`
	DownloadCount int64   `json:"download_count"`
	ErrorCount    int64   `json:"error_count"`
} 