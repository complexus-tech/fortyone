package workerbootstrap

import (
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/hibiken/asynq"
)

func integrationRetryDelay(retryCount int, err error, task *asynq.Task) time.Duration {
	if retryAfter, ok := messagingrepository.LeaseRetryAfter(err); ok {
		return retryAfter
	}
	if retryAfter, ok := slack.SlackRetryAfter(err); ok {
		return retryAfter
	}
	return asynq.DefaultRetryDelayFunc(retryCount, err, task)
}
