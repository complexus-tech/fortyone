package storage

import (
	"github.com/complexus-tech/projects-api/pkg/aws"
	"github.com/complexus-tech/projects-api/pkg/azure"
)

// Config holds storage configuration for all providers.
type Config struct {
	Provider          string
	ProfilesBucket    string
	LogosBucket       string
	AttachmentsBucket string
	Azure             azure.Config
	AWS               aws.Config
}
