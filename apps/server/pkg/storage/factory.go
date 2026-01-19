package storage

import (
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/aws"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

// NewStorageService creates a new storage service based on the provider.
func NewStorageService(cfg Config, log *logger.Logger) (StorageService, error) {
	switch cfg.Provider {
	case "azure":
		return azure.NewStorageService(cfg.Azure, log)
	case "aws":
		return aws.NewS3Service(cfg.AWS, log)
	case "minio":
		return aws.NewS3Service(cfg.MinIO, log)
	case "":
		return nil, fmt.Errorf("storage provider is required")
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}
}
