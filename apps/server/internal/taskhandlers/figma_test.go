package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type figmaWebhookProcessorStub struct {
	provider  integrations.ProviderKey
	inboxID   uuid.UUID
	recovered int
}

func (processor *figmaWebhookProcessorStub) ProcessWebhook(
	_ context.Context,
	provider integrations.ProviderKey,
	inboxID uuid.UUID,
) error {
	processor.provider = provider
	processor.inboxID = inboxID
	return nil
}

func (processor *figmaWebhookProcessorStub) RecoverPendingWebhooks(context.Context) (int, error) {
	return processor.recovered, nil
}

func TestHandleFigmaWebhookUsesInboxIdentityOnly(t *testing.T) {
	t.Parallel()

	inboxID := uuid.New()
	processor := &figmaWebhookProcessorStub{}
	handler := NewWorkerHandlers(WorkerHandlerDependencies{FigmaWebhooks: processor})
	payload, err := json.Marshal(tasks.FigmaWebhookPayload{InboxID: inboxID})
	require.NoError(t, err)

	err = handler.HandleFigmaWebhook(t.Context(), asynq.NewTask(tasks.TypeFigmaWebhook, payload))
	require.NoError(t, err)
	require.Equal(t, figmaprovider.ProviderKey, processor.provider)
	require.Equal(t, inboxID, processor.inboxID)
}

func TestHandleFigmaWebhookRejectsInvalidIdentityWithoutRetry(t *testing.T) {
	t.Parallel()

	handler := NewWorkerHandlers(WorkerHandlerDependencies{
		FigmaWebhooks: &figmaWebhookProcessorStub{},
	})
	payload, err := json.Marshal(tasks.FigmaWebhookPayload{})
	require.NoError(t, err)

	err = handler.HandleFigmaWebhook(t.Context(), asynq.NewTask(tasks.TypeFigmaWebhook, payload))
	require.Error(t, err)
	require.True(t, errors.Is(err, asynq.SkipRetry))
}

func TestHandleFigmaWebhookRecoveryUsesSharedInboxRuntime(t *testing.T) {
	t.Parallel()

	processor := &figmaWebhookProcessorStub{recovered: 2}
	handler := NewWorkerHandlers(WorkerHandlerDependencies{FigmaWebhooks: processor})

	require.NoError(t, handler.HandleFigmaWebhookRecovery(
		t.Context(),
		asynq.NewTask(tasks.TypeFigmaWebhookRecovery, nil),
	))
}
