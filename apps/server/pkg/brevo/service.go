package brevo

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	brv "github.com/getbrevo/brevo-go/lib"
)

type Service struct {
	client  *brv.APIClient
	log     *logger.Logger
	enabled bool
}

type Config struct {
	APIKey string
}

// NewService creates a new Brevo service instance with the provided configuration and logger.
func NewService(cfg Config, log *logger.Logger) (*Service, error) {
	if cfg.APIKey == "" {
		log.Info(context.Background(), "Brevo disabled: APP_BREVO_API_KEY not set")
		return &Service{
			client:  nil,
			log:     log,
			enabled: false,
		}, nil
	}

	config := brv.NewConfiguration()
	config.AddDefaultHeader("api-key", cfg.APIKey)
	client := brv.NewAPIClient(config)

	s := &Service{
		client:  client,
		log:     log,
		enabled: true,
	}
	log.Info(context.Background(), "Brevo service initialized successfully")
	return s, nil
}

func (s *Service) isEnabled(ctx context.Context) bool {
	if s == nil || !s.enabled || s.client == nil {
		if s != nil && s.log != nil {
			s.log.Info(ctx, "Brevo disabled: skipping request")
		}
		return false
	}

	return true
}
