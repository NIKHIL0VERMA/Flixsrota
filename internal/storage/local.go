package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements the Storage interface using local file system
type LocalStorage struct {
	basePath string
	tempPath string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath, tempPath string) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		tempPath: tempPath,
	}, nil
}

// Upload uploads a file to local storage
func (ls *LocalStorage) Upload(ctx context.Context, localPath, remotePath string) error {
	// Ensure the target directory exists
	targetPath := filepath.Join(ls.basePath, remotePath)
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Open source file
	source, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	// Create target file
	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer target.Close()

	// Copy file content
	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Download downloads a file from local storage
func (ls *LocalStorage) Download(ctx context.Context, remotePath, localPath string) error {
	// Ensure the target directory exists
	targetDir := filepath.Dir(localPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Open source file
	sourcePath := filepath.Join(ls.basePath, remotePath)
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	// Create target file
	target, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer target.Close()

	// Copy file content
	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Delete deletes a file from local storage
func (ls *LocalStorage) Delete(ctx context.Context, remotePath string) error {
	targetPath := filepath.Join(ls.basePath, remotePath)
	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Exists checks if a file exists in local storage
func (ls *LocalStorage) Exists(ctx context.Context, remotePath string) (bool, error) {
	targetPath := filepath.Join(ls.basePath, remotePath)
	_, err := os.Stat(targetPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check file existence: %w", err)
}

// GetURL returns a file URL (for local storage, this is just the file path)
func (ls *LocalStorage) GetURL(ctx context.Context, remotePath string) (string, error) {
	targetPath := filepath.Join(ls.basePath, remotePath)
	if _, err := os.Stat(targetPath); err != nil {
		return "", fmt.Errorf("file does not exist: %w", err)
	}
	return "file://" + targetPath, nil
}

// ListFiles lists files in a directory
func (ls *LocalStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	searchPath := filepath.Join(ls.basePath, prefix)
	
	var files []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			// Convert to relative path
			relPath, err := filepath.Rel(ls.basePath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	
	return files, nil
}

// CreateTempFile creates a temporary file
func (ls *LocalStorage) CreateTempFile(ctx context.Context, suffix string) (*os.File, error) {
	return os.CreateTemp(ls.tempPath, "flixsrota-*"+suffix)
}

// Close closes the storage connection (no-op for local storage)
func (ls *LocalStorage) Close() error {
	return nil
} 