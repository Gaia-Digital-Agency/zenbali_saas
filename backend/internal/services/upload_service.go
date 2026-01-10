package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/config"
)

var (
	ErrFileTooLarge     = errors.New("file too large")
	ErrInvalidFileType  = errors.New("invalid file type")
)

type UploadService struct {
	config config.UploadConfig
}

func NewUploadService(cfg config.UploadConfig) *UploadService {
	return &UploadService{config: cfg}
}

func (s *UploadService) SaveEventImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Check file size
	maxSize := int64(s.config.MaxSizeMB * 1024 * 1024)
	if header.Size > maxSize {
		return "", ErrFileTooLarge
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExt := false
	for _, allowed := range s.config.AllowedExt {
		if ext == allowed {
			validExt = true
			break
		}
	}
	if !validExt {
		return "", ErrInvalidFileType
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.config.Dir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // Clean up on error
		return "", err
	}

	// Return relative URL path
	return "/uploads/" + filename, nil
}

func (s *UploadService) DeleteFile(imageURL string) error {
	if imageURL == "" {
		return nil
	}

	// Extract filename from URL
	filename := strings.TrimPrefix(imageURL, "/uploads/")
	filePath := filepath.Join(s.config.Dir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(filePath)
}

func (s *UploadService) GetMaxSizeMB() int {
	return s.config.MaxSizeMB
}

func (s *UploadService) GetAllowedExtensions() []string {
	return s.config.AllowedExt
}
