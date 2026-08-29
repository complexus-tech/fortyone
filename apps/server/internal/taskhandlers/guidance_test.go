package taskhandlers

import (
	"context"
	"errors"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestGuidanceHandlersDelegateToTypedStores(t *testing.T) {
	store := &guidanceStoreStub{}
	handlers := newTestGuidanceHandlers(store)

	require.NoError(t, handlers.HandleOverdueStoriesEmail(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleObjectiveOverdueEmail(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleWeeklyDigestEmail(t.Context(), asynq.NewTask("test", nil)))
	require.Equal(t, 1, store.storyRecipientCalls)
	require.Equal(t, 1, store.objectiveRecipientCalls)
	require.Equal(t, 1, store.weeklyDigestRecipientCalls)
}

func TestGuidanceHandlersPreserveStoreFailures(t *testing.T) {
	sentinel := errors.New("guidance store unavailable")
	tests := []struct {
		name      string
		configure func(*guidanceStoreStub)
		handle    func(*GuidanceHandlers) error
	}{
		{
			name:      "stories",
			configure: func(store *guidanceStoreStub) { store.storyRecipientErr = sentinel },
			handle: func(handlers *GuidanceHandlers) error {
				return handlers.HandleOverdueStoriesEmail(context.Background(), asynq.NewTask("test", nil))
			},
		},
		{
			name:      "objectives",
			configure: func(store *guidanceStoreStub) { store.objectiveRecipientErr = sentinel },
			handle: func(handlers *GuidanceHandlers) error {
				return handlers.HandleObjectiveOverdueEmail(context.Background(), asynq.NewTask("test", nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &guidanceStoreStub{}
			test.configure(store)
			err := test.handle(newTestGuidanceHandlers(store))
			require.ErrorIs(t, err, sentinel)
		})
	}
}

func TestGuidanceHandlersRequireCoreDependencies(t *testing.T) {
	var handlers *GuidanceHandlers
	err := handlers.HandleOverdueStoriesEmail(t.Context(), nil)
	require.ErrorContains(t, err, "dependencies are required")

	handlers = NewGuidanceHandlers(GuidanceHandlerDependencies{
		Log:    testTaskLogger(),
		Mailer: guidanceMailerStub{},
	})
	err = handlers.HandleOverdueStoriesEmail(t.Context(), nil)
	require.ErrorContains(t, err, "overdue story store is required")
}

func newTestGuidanceHandlers(store *guidanceStoreStub) *GuidanceHandlers {
	return NewGuidanceHandlers(GuidanceHandlerDependencies{
		Log:          testTaskLogger(),
		Objectives:   store,
		Stories:      store,
		WeeklyDigest: store,
		Mailer:       guidanceMailerStub{},
	})
}

type guidanceStoreStub struct {
	storyRecipientCalls        int
	objectiveRecipientCalls    int
	storyRecipientErr          error
	objectiveRecipientErr      error
	weeklyDigestRecipientCalls int
}

func (store *guidanceStoreStub) ListWeeklyDigestRecipients(
	context.Context,
	*notificationsdomain.WeeklyDigestCursor,
	int,
) ([]notificationsdomain.WeeklyDigestRecipient, error) {
	store.weeklyDigestRecipientCalls++
	return nil, nil
}

func (*guidanceStoreStub) GetWeeklyDigestStats(
	context.Context,
	notificationsdomain.WeeklyDigestStatsQuery,
) (notificationsdomain.WeeklyDigestStats, error) {
	return notificationsdomain.WeeklyDigestStats{}, nil
}

func (store *guidanceStoreStub) ListOverdueStoryGuidanceRecipients(
	context.Context,
	time.Time,
	*storydomain.OverdueGuidanceCursor,
	int,
) ([]storydomain.OverdueGuidanceRecipient, error) {
	store.storyRecipientCalls++
	return nil, store.storyRecipientErr
}

func (*guidanceStoreStub) ListOverdueStoryGuidanceItems(
	context.Context,
	time.Time,
	uuid.UUID,
	uuid.UUID,
) ([]storydomain.OverdueGuidanceStory, error) {
	return nil, nil
}

func (store *guidanceStoreStub) ListOverdueObjectiveGuidanceRecipients(
	context.Context,
	time.Time,
	*objectivesdomain.OverdueGuidanceCursor,
	int,
) ([]objectivesdomain.OverdueGuidanceRecipient, error) {
	store.objectiveRecipientCalls++
	return nil, store.objectiveRecipientErr
}

func (*guidanceStoreStub) ListOverdueObjectiveGuidanceItems(
	context.Context,
	time.Time,
	uuid.UUID,
	uuid.UUID,
) ([]objectivesdomain.OverdueGuidanceObjective, error) {
	return nil, nil
}

type guidanceMailerStub struct{}

func (guidanceMailerStub) Send(context.Context, mailer.Email) error { return nil }

func (guidanceMailerStub) SendTemplated(context.Context, mailer.TemplatedEmail) error { return nil }
