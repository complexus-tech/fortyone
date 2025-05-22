package tasks

import (
	// "context" // No longer needed here if logging is minimal in New/Close
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Service is the main struct for interacting with task enqueuing.
// It holds the Asynq client and the application logger.
type Service struct {
	asynqClient *asynq.Client
	log         *logger.Logger
}

// New creates a new tasks Service instance.
func New(existingRdb redis.UniversalClient, log *logger.Logger) (*Service, error) {
	if existingRdb == nil {
		return nil, fmt.Errorf("tasks: existing Redis client (redis.UniversalClient) cannot be nil")
	}
	if log == nil {
		return nil, fmt.Errorf("tasks: logger cannot be nil")
	}

	client := asynq.NewClientFromRedisClient(existingRdb)

	if client == nil {
		return nil, fmt.Errorf("tasks: failed to create Asynq client using NewClientFromRedisClient (returned nil)")
	}

	s := &Service{
		asynqClient: client,
		log:         log,
	}
	return s, nil
}

// Close allows the application to close the underlying Asynq client during shutdown.
func (s *Service) Close() error {
	if s.asynqClient != nil {
		return s.asynqClient.Close()
	}
	return nil
}

// Task type constants and payload structs will be in their respective task files (e.g., onboarding_task.go)
// Enqueue methods for specific tasks will also be in their respective task files.
