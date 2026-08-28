package tasks

import (
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	asynqClient *asynq.Client
	log         *logger.Logger
}

// New creates a task-enqueueing service that borrows existingRdb.
//
// The caller retains ownership of the Redis client and must close it after the
// service is no longer used. Asynq deliberately prevents closing clients
// created from a shared Redis connection.
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
