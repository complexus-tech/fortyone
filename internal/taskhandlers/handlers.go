package taskhandlers

import (
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailerlite"
)

type handlers struct {
	log               *logger.Logger
	mailerLiteService *mailerlite.Service
	onboardingGroupID string
}

// NewWorkerHandlers initializes the central task Handlers service.
func NewWorkerHandlers(log *logger.Logger, mailerLiteService *mailerlite.Service, onboardingGroupID string) *handlers {
	return &handlers{
		log:               log,
		mailerLiteService: mailerLiteService,
		onboardingGroupID: onboardingGroupID,
	}
}
