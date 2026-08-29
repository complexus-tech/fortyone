package workerbootstrap

import (
	"errors"
	"fmt"
	"testing"
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestIntegrationRetryDelayHonorsMessagingLease(t *testing.T) {
	t.Parallel()

	delay := integrationRetryDelay(3, fmt.Errorf("process Slack event: %w", &messagingrepository.LeaseBusyError{
		Resource:   "messaging inbound event",
		RetryAfter: 37 * time.Second,
	}), asynq.NewTask("test", nil))

	require.Equal(t, 37*time.Second, delay)
}

func TestIntegrationRetryDelayHonorsSlackRetryAfter(t *testing.T) {
	t.Parallel()

	delay := integrationRetryDelay(3, &slack.RateLimitError{
		Method:     "chat.postMessage",
		RetryAfter: 17 * time.Second,
	}, asynq.NewTask("test", nil))

	require.Equal(t, 17*time.Second, delay)
}

func TestIntegrationRetryDelayFallsBackToAsynq(t *testing.T) {
	t.Parallel()

	err := errors.New("network unavailable")
	task := asynq.NewTask("test", nil)
	delay := integrationRetryDelay(1, err, task)

	require.GreaterOrEqual(t, delay, 16*time.Second)
	require.LessOrEqual(t, delay, 74*time.Second)
}
