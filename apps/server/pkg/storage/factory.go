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
		service, err := aws.NewS3Service(cfg.AWS, log)
		if err != nil {
			return nil, err
		}
		if cfg.AWS.Bucket != "" {
			return newPrefixedStorageService(service, cfg.AWS.Bucket), nil
		}
		return service, nil
	case "":
		return nil, fmt.Errorf("storage provider is required")
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}
}
