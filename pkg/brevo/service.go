package brevo

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	brevo "github.com/getbrevo/brevo-go/lib"
)

type Service struct {
	client *brevo.APIClient
	log    *logger.Logger
}

type Config struct {
	APIKey string
}

// NewService creates a new Brevo service instance with the provided configuration and logger.
func NewService(cfg Config, log *logger.Logger) (*Service, error) {
	config := brevo.NewConfiguration()
	config.AddDefaultHeader("api-key", cfg.APIKey)
	client := brevo.NewAPIClient(config)

	s := &Service{
		client: client,
		log:    log,
	}
	log.Info(context.Background(), "Brevo service initialized successfully")
	return s, nil
}
