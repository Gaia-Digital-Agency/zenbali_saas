package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/config"
)

var (
	ErrFileTooLarge    = errors.New("file too large")
	ErrInvalidFileType = errors.New("invalid file type")
)

type UploadService struct {
	config    config.UploadConfig
	gcsClient *storage.Client
}

func NewUploadService(ctx context.Context, cfg config.UploadConfig) (*UploadService, error) {
	service := &UploadService{config: cfg}

	if cfg.Backend != "gcs" {
		return service, nil
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create GCS client: %w", err)
	}

	service.gcsClient = client
	return service, nil
}

func (s *UploadService) SaveEventImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	maxSize := int64(s.config.MaxSizeMB * 1024 * 1024)
	if header.Size > maxSize {
		return "", ErrFileTooLarge
	}

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

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	if s.config.Backend == "gcs" {
		return s.saveToGCS(file, header, filename)
	}

	return s.saveToLocal(file, filename)
}

func (s *UploadService) saveToLocal(file multipart.File, filename string) (string, error) {
	filePath := filepath.Join(s.config.Dir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		_ = os.Remove(filePath)
		return "", err
	}

	return "/uploads/" + filename, nil
}

func (s *UploadService) saveToGCS(file multipart.File, header *multipart.FileHeader, filename string) (string, error) {
	if s.gcsClient == nil {
		return "", errors.New("GCS client not configured")
	}

	objectName := s.objectName(filename)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	writer := s.gcsClient.Bucket(s.config.GCSBucket).Object(objectName).NewWriter(ctx)
	writer.ContentType = header.Header.Get("Content-Type")
	writer.CacheControl = "public, max-age=31536000"

	if _, err := io.Copy(writer, file); err != nil {
		_ = writer.Close()
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	return s.gcsPublicURL(objectName), nil
}

func (s *UploadService) DeleteFile(imageURL string) error {
	if imageURL == "" {
		return nil
	}

	if strings.HasPrefix(imageURL, "/uploads/") {
		filename := strings.TrimPrefix(imageURL, "/uploads/")
		filePath := filepath.Join(s.config.Dir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return nil
		}
		return os.Remove(filePath)
	}

	objectName, ok := s.gcsObjectNameFromURL(imageURL)
	if !ok || s.gcsClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.gcsClient.Bucket(s.config.GCSBucket).Object(objectName).Delete(ctx)
	if err != nil && !errors.Is(err, storage.ErrObjectNotExist) {
		return err
	}

	return nil
}

func (s *UploadService) GetMaxSizeMB() int {
	return s.config.MaxSizeMB
}

func (s *UploadService) GetAllowedExtensions() []string {
	return s.config.AllowedExt
}

func (s *UploadService) objectName(filename string) string {
	prefix := strings.Trim(strings.TrimSpace(s.config.GCSPrefix), "/")
	if prefix == "" {
		return filename
	}
	return path.Join(prefix, filename)
}

func (s *UploadService) gcsPublicURL(objectName string) string {
	base := strings.TrimSuffix(strings.TrimSpace(s.config.GCSPublicBase), "/")
	if base == "" {
		base = "https://storage.googleapis.com/" + s.config.GCSBucket
	}
	return base + "/" + objectName
}

func (s *UploadService) gcsObjectNameFromURL(imageURL string) (string, bool) {
	publicBase := strings.TrimSuffix(strings.TrimSpace(s.config.GCSPublicBase), "/")
	defaultBase := "https://storage.googleapis.com/" + s.config.GCSBucket
	gsPrefix := "gs://" + s.config.GCSBucket + "/"

	switch {
	case publicBase != "" && strings.HasPrefix(imageURL, publicBase+"/"):
		return strings.TrimPrefix(imageURL, publicBase+"/"), true
	case strings.HasPrefix(imageURL, defaultBase+"/"):
		return strings.TrimPrefix(imageURL, defaultBase+"/"), true
	case strings.HasPrefix(imageURL, gsPrefix):
		return strings.TrimPrefix(imageURL, gsPrefix), true
	default:
		return "", false
	}
}
