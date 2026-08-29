package taskhandlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type slackEventProcessorStub struct {
	externalWorkspaceID string
	eventID             string
	provider            integrations.ProviderKey
	inboxID             uuid.UUID
	err                 error
	backfilled          int
	backfillErr         error
	recovered           int
	recoveryErr         error
}

func (s *slackEventProcessorStub) ProcessWebhook(_ context.Context, provider integrations.ProviderKey, inboxID uuid.UUID) error {
	s.provider = provider
	s.inboxID = inboxID
	return s.err
}

func (s *slackEventProcessorStub) ProcessEvent(_ context.Context, externalWorkspaceID, eventID string) error {
	s.externalWorkspaceID = externalWorkspaceID
	s.eventID = eventID
	return s.err
}

func (s *slackEventProcessorStub) BackfillLegacyCredentials(context.Context) (int, error) {
	return s.backfilled, s.backfillErr
}

func (s *slackEventProcessorStub) RecoverPendingEvents(context.Context) (int, error) {
	return s.recovered, s.recoveryErr
}

func TestHandleSlackEvent(t *testing.T) {
	t.Parallel()

	t.Run("processes inbox identity payload", func(t *testing.T) {
		t.Parallel()
		inboxID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
		processor := &slackEventProcessorStub{}
		handler := &handlers{
			log:         logger.NewWithText(io.Discard, slog.LevelError, "test"),
			slackEvents: processor,
		}
		task := asynq.NewTask("slack:event:process", []byte(`{"provider":"slack","inboxId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}`))

		require.NoError(t, handler.HandleSlackEvent(context.Background(), task))
		require.Equal(t, integrations.ProviderKey("slack"), processor.provider)
		require.Equal(t, inboxID, processor.inboxID)
	})

	t.Run("processes a legacy rollout payload through compatibility lookup", func(t *testing.T) {
		t.Parallel()
		processor := &slackEventProcessorStub{}
		handler := &handlers{
			log:         logger.NewWithText(io.Discard, slog.LevelError, "test"),
			slackEvents: processor,
		}
		task := asynq.NewTask("slack:event:process", []byte(`{"externalWorkspaceId":"T1","eventId":"Ev123","recoveryAttempt":2}`))

		require.NoError(t, handler.HandleSlackEvent(context.Background(), task))
		require.Equal(t, "Ev123", processor.eventID)
		require.Equal(t, "T1", processor.externalWorkspaceID)
	})

	t.Run("does not retry malformed payload", func(t *testing.T) {
		t.Parallel()
		handler := &handlers{
			log:         logger.NewWithText(io.Discard, slog.LevelError, "test"),
			slackEvents: &slackEventProcessorStub{},
		}

		err := handler.HandleSlackEvent(context.Background(), asynq.NewTask("slack:event:process", []byte(`{"eventId":""}`)))

		require.ErrorIs(t, err, asynq.SkipRetry)
	})

	t.Run("propagates processor failure for retry", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("provider unavailable")
		handler := &handlers{
			log:         logger.NewWithText(io.Discard, slog.LevelError, "test"),
			slackEvents: &slackEventProcessorStub{err: expected},
		}
		task := asynq.NewTask("slack:event:process", []byte(`{"provider":"slack","inboxId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}`))

		err := handler.HandleSlackEvent(context.Background(), task)

		require.ErrorIs(t, err, expected)
		require.False(t, errors.Is(err, asynq.SkipRetry))
	})
}

func TestHandleSlackCredentialBackfill(t *testing.T) {
	processor := &slackEventProcessorStub{backfilled: 3}
	handler := &handlers{
		log:              logger.NewWithText(io.Discard, slog.LevelError, "test"),
		slackCredentials: processor,
	}

	require.NoError(t, handler.HandleSlackCredentialBackfill(context.Background(), asynq.NewTask("cleanup:slack_credentials", nil)))

	expected := errors.New("database unavailable")
	processor.backfillErr = expected
	err := handler.HandleSlackCredentialBackfill(context.Background(), asynq.NewTask("cleanup:slack_credentials", nil))
	require.ErrorIs(t, err, expected)
}

func TestHandleSlackInboxRecovery(t *testing.T) {
	processor := &slackEventProcessorStub{recovered: 2}
	handler := &handlers{
		log:           logger.NewWithText(io.Discard, slog.LevelError, "test"),
		slackRecovery: processor,
	}

	require.NoError(t, handler.HandleSlackInboxRecovery(context.Background(), asynq.NewTask("cleanup:slack_inbox", nil)))

	expected := errors.New("queue unavailable")
	processor.recoveryErr = expected
	err := handler.HandleSlackInboxRecovery(context.Background(), asynq.NewTask("cleanup:slack_inbox", nil))
	require.ErrorIs(t, err, expected)
}
