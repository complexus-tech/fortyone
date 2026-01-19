package storage

import (
	"github.com/complexus-tech/projects-api/pkg/aws"
	"github.com/complexus-tech/projects-api/pkg/azure"
)

// Config holds storage configuration for all providers.
type Config struct {
	Provider string
	Azure    azure.Config
	AWS      aws.Config
	MinIO    aws.Config
}
