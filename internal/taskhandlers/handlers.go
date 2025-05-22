package taskhandlers

import (
	"github.com/complexus-tech/projects-api/pkg/logger"
)

type handlers struct {
	log *logger.Logger
}

// NewWorkerHandlers initializes the central task Handlers service.
func NewWorkerHandlers(log *logger.Logger) *handlers {
	return &handlers{
		log: log,
	}
}
