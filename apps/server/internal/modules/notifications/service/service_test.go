package notifications

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	created     notificationsdomain.Notification
	inserted    bool
	createErr   error
	mutations   []notificationsdomain.NotificationMutation
	portalQuery notificationsdomain.PortalListQuery
	preference  notificationsdomain.UpdatePreference
}

func (stub *repositoryStub) Create(context.Context, notificationsdomain.NewNotification) (notificationsdomain.Notification, bool, error) {
	return stub.created, stub.inserted, stub.createErr
}

func (*repositoryStub) List(context.Context, notificationsdomain.ListQuery) ([]notificationsdomain.Notification, error) {
	return nil, nil
}

func (*repositoryStub) CountUnread(context.Context, notificationsdomain.WorkspaceAccess) (int, error) {
	return 0, nil
}

func (stub *repositoryStub) Mutate(_ context.Context, command notificationsdomain.NotificationMutation) error {
	stub.mutations = append(stub.mutations, command)
	return nil
}

func (*repositoryStub) MutateAll(context.Context, notificationsdomain.WorkspaceMutation) (int, error) {
	return 0, nil
}

func (*repositoryStub) GetPreferences(context.Context, notificationsdomain.WorkspaceAccess) (notificationsdomain.Preferences, error) {
	return notificationsdomain.Preferences{}, nil
}

func (stub *repositoryStub) UpdatePreference(_ context.Context, command notificationsdomain.UpdatePreference) (notificationsdomain.Preferences, error) {
	stub.preference = command
	return notificationsdomain.Preferences{}, nil
}

func (stub *repositoryStub) ListPortalFeedback(_ context.Context, query notificationsdomain.PortalListQuery) ([]notificationsdomain.PortalNotification, error) {
	stub.portalQuery = query
	return nil, nil
}

func (*repositoryStub) CountUnreadPortalFeedback(context.Context, notificationsdomain.PortalAccess) (int, error) {
	return 0, nil
}

func (*repositoryStub) MarkPortalFeedbackRead(context.Context, notificationsdomain.PortalNotificationMutation) error {
	return nil
}

func (*repositoryStub) MarkAllPortalFeedbackRead(context.Context, notificationsdomain.PortalMutation) (int, error) {
	return 0, nil
}

func (*repositoryStub) ListKeyResultAudience(context.Context, notificationsdomain.KeyResultAudienceQuery) ([]notificationsdomain.KeyResultAudienceMember, error) {
	return nil, nil
}

func (*repositoryStub) GetEmailDelivery(context.Context, notificationsdomain.EmailNotificationQuery) (*notificationsdomain.EmailNotification, error) {
	return nil, nil
}

func (*repositoryStub) ListEmailDigest(context.Context, notificationsdomain.DeliveryScope) (*notificationsdomain.EmailDigest, error) {
	return nil, nil
}

func (*repositoryStub) ListDeliveryTeamIDs(context.Context, notificationsdomain.DeliveryScope) ([]uuid.UUID, error) {
	return nil, nil
}

func (*repositoryStub) MarkEmailSent(context.Context, notificationsdomain.MarkEmailSent) error {
	return nil
}

type taskStub struct {
	payloads []tasks.NotificationEmailDigestPayload
	err      error
}

func (stub *taskStub) EnqueueNotificationEmailDigest(payload tasks.NotificationEmailDigestPayload, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	stub.payloads = append(stub.payloads, payload)
	return &asynq.TaskInfo{ID: "digest"}, stub.err
}

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

func validNotification() (notificationsdomain.NewNotification, notificationsdomain.Notification) {
	input := notificationsdomain.NewNotification{
		DedupeKey:   "event:story-updated:recipient",
		RecipientID: uuid.New(), WorkspaceID: uuid.New(),
		Type:       notificationsdomain.NotificationTypeStoryUpdate,
		EntityType: notificationsdomain.EntityTypeStory,
		EntityID:   uuid.New(), ActorID: uuid.New(), Title: "Story updated",
		Message: notificationsdomain.NotificationMessage{Template: "{actor} updated the story", Variables: map[string]notificationsdomain.Variable{}},
	}
	created := notificationsdomain.Notification{
		ID: uuid.New(), RecipientID: input.RecipientID, WorkspaceID: input.WorkspaceID,
		Type: input.Type, EntityType: input.EntityType, EntityID: input.EntityID,
		ActorID: input.ActorID, Title: input.Title, Message: input.Message,
		InAppEnabled: true, CreatedAt: time.Now(),
	}
	return input, created
}

func newTestService(repository Repository, tasks TasksService) *Service {
	return New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repository, nil, tasks)
}

func TestCreateDuplicateReplayRequeuesDigestWithoutRealtimeFanout(t *testing.T) {
	t.Parallel()
	input, created := validNotification()
	repository := &repositoryStub{created: created, inserted: false}
	queue := &taskStub{}
	service := newTestService(repository, queue)
	published := 0
	service.publishRealtime = func(context.Context, CoreNotification) error { published++; return nil }

	result, err := service.Create(context.Background(), input)

	require.NoError(t, err)
	require.Equal(t, created.ID, result.ID)
	require.Zero(t, published)
	require.Equal(t, []tasks.NotificationEmailDigestPayload{{RecipientID: input.RecipientID, WorkspaceID: input.WorkspaceID}}, queue.payloads)
}

func TestCreateReturnsQueueFailureSoDurableIntentCanBeRetried(t *testing.T) {
	t.Parallel()
	input, created := validNotification()
	repository := &repositoryStub{created: created, inserted: true}
	queue := &taskStub{err: errors.New("queue unavailable")}
	service := newTestService(repository, queue)

	result, err := service.Create(context.Background(), input)

	require.ErrorContains(t, err, "enqueue notification email digest")
	require.Equal(t, created.ID, result.ID)
}

func TestCreateRealtimeFailureRecoversDigestWithoutDuplicateFanout(t *testing.T) {
	t.Parallel()
	input, created := validNotification()
	repository := &repositoryStub{created: created, inserted: true}
	queue := &taskStub{}
	service := newTestService(repository, queue)
	realtimeErr := errors.New("realtime unavailable")
	published := 0
	service.publishRealtime = func(context.Context, CoreNotification) error {
		published++
		return realtimeErr
	}

	result, err := service.Create(context.Background(), input)
	require.ErrorIs(t, err, realtimeErr)
	require.Equal(t, created.ID, result.ID)
	require.Equal(t, 1, published)
	require.Empty(t, queue.payloads)

	repository.inserted = false
	service.publishRealtime = func(context.Context, CoreNotification) error {
		published++
		return nil
	}
	result, err = service.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, created.ID, result.ID)
	require.Equal(t, 1, published, "an exact replay must not duplicate realtime fanout")
	require.Equal(t, []tasks.NotificationEmailDigestPayload{{
		RecipientID: input.RecipientID,
		WorkspaceID: input.WorkspaceID,
	}}, queue.payloads)
}

func TestCreateSkipsRealtimeForPortalAndEmailOnlyRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entityType notificationsdomain.EntityType
		inApp      bool
	}{
		{name: "portal feedback", entityType: notificationsdomain.EntityTypeFeedback, inApp: true},
		{name: "email-only", entityType: notificationsdomain.EntityTypeStrategy, inApp: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			input, created := validNotification()
			created.EntityType = test.entityType
			created.InAppEnabled = test.inApp
			repository := &repositoryStub{created: created, inserted: true}
			queue := &taskStub{}
			service := newTestService(repository, queue)
			published := 0
			service.publishRealtime = func(context.Context, CoreNotification) error {
				published++
				return nil
			}

			_, err := service.Create(context.Background(), input)
			require.NoError(t, err)
			require.Zero(t, published)
			require.Len(t, queue.payloads, 1)
		})
	}
}

func TestWorkspaceMutationCarriesActorWorkspaceAndClock(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repository, nil, nil, WithClock(fixedClock{now: now}))
	notificationID, actorID, workspaceID := uuid.New(), uuid.New(), uuid.New()

	require.NoError(t, service.MarkAsRead(context.Background(), notificationID, actorID, workspaceID))
	require.Equal(t, notificationsdomain.NotificationMutation{
		Access:         notificationsdomain.WorkspaceAccess{ActorID: actorID, WorkspaceID: workspaceID},
		NotificationID: notificationID, Kind: notificationsdomain.NotificationMutationRead, At: now,
	}, repository.mutations[0])
}

func TestUpdatePreferenceUsesTypedPresencePatch(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repository, nil, nil, WithClock(fixedClock{now: now}))
	actorID, workspaceID := uuid.New(), uuid.New()

	_, err := service.UpdatePreference(context.Background(), actorID, workspaceID, notificationsdomain.PreferenceTypeMention, notificationsdomain.ChannelPatch{Email: platformpatch.Set(false)})

	require.NoError(t, err)
	require.Equal(t, notificationsdomain.PreferenceTypeMention, repository.preference.Type)
	value, specified := repository.preference.Patch.Email.Value()
	require.True(t, specified)
	require.NotNil(t, value)
	require.False(t, *value)
}

func TestPortalListPreservesNormalizedTypedScope(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	service := newTestService(repository, nil)
	actorID := uuid.New()

	_, err := service.ListPortalFeedback(context.Background(), actorID, " city-roads ", true, 11, 20)

	require.NoError(t, err)
	require.Equal(t, notificationsdomain.PortalListQuery{
		Access:     notificationsdomain.PortalAccess{ActorID: actorID, PortalSlug: " city-roads "},
		UnreadOnly: true, Limit: 11, Offset: 20,
	}, repository.portalQuery)
}
