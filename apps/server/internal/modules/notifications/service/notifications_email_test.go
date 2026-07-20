package notifications

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type notificationRepoStub struct {
	notification    CoreNotification
	insertedResults []bool
	createCalls     int
}

func (r *notificationRepoStub) Create(context.Context, CoreNewNotification) (CoreNotification, bool, error) {
	inserted := true
	if r.createCalls < len(r.insertedResults) {
		inserted = r.insertedResults[r.createCalls]
	}
	r.createCalls++
	return r.notification, inserted, nil
}

func (r *notificationRepoStub) List(context.Context, uuid.UUID, uuid.UUID, string, int, int) ([]CoreNotification, error) {
	return nil, nil
}

func (r *notificationRepoStub) GetUnreadCount(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return 0, nil
}

func (r *notificationRepoStub) MarkAsRead(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (r *notificationRepoStub) MarkAllAsRead(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (r *notificationRepoStub) GetPreferences(context.Context, uuid.UUID, uuid.UUID) (CoreNotificationPreferences, error) {
	return CoreNotificationPreferences{}, nil
}

func (r *notificationRepoStub) UpdatePreference(context.Context, uuid.UUID, uuid.UUID, string, map[string]any) error {
	return nil
}

func (r *notificationRepoStub) DeleteNotification(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (r *notificationRepoStub) DeleteAllNotifications(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	return 0, nil
}

func (r *notificationRepoStub) DeleteReadNotifications(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	return 0, nil
}

func (r *notificationRepoStub) MarkAsUnread(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (r *notificationRepoStub) ListPortalFeedback(context.Context, uuid.UUID, string, int, int) ([]CorePortalNotification, error) {
	return nil, nil
}

func (r *notificationRepoStub) GetPortalFeedbackUnreadCount(context.Context, uuid.UUID, string) (int, error) {
	return 0, nil
}

func (r *notificationRepoStub) MarkPortalFeedbackAsRead(context.Context, uuid.UUID, uuid.UUID, string) error {
	return nil
}

type notificationTasksStub struct {
	digestPayloads []tasks.NotificationEmailDigestPayload
}

func (s *notificationTasksStub) EnqueueNotificationEmailDigest(payload tasks.NotificationEmailDigestPayload, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	s.digestPayloads = append(s.digestPayloads, payload)
	return &asynq.TaskInfo{ID: "digest-task"}, nil
}

func TestCreateEnqueuesRecipientWorkspaceEmailDigest(t *testing.T) {
	recipientID := uuid.New()
	workspaceID := uuid.New()
	notificationID := uuid.New()

	repo := &notificationRepoStub{
		notification: CoreNotification{
			ID:          notificationID,
			RecipientID: recipientID,
			WorkspaceID: workspaceID,
			Type:        "story_update",
			EntityType:  "story",
			EntityID:    uuid.New(),
			ActorID:     uuid.New(),
			Title:       "Story updated",
			Message: NotificationMessage{
				Template: "{actor} changed {field}",
				Variables: map[string]Variable{
					"actor": {Value: "Maya", Type: "actor"},
					"field": {Value: "status", Type: "field"},
				},
			},
			CreatedAt: time.Now(),
		},
		insertedResults: []bool{true},
	}
	taskService := &notificationTasksStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, taskService)

	_, err := service.Create(context.Background(), CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: workspaceID,
		Type:        "story_update",
		EntityType:  "story",
		EntityID:    uuid.New(),
		ActorID:     uuid.New(),
		Title:       "Story updated",
		Message: NotificationMessage{
			Template: "{actor} changed {field}",
			Variables: map[string]Variable{
				"actor": {Value: "Maya", Type: "actor"},
				"field": {Value: "status", Type: "field"},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, taskService.digestPayloads, 1)
	require.Equal(t, recipientID, taskService.digestPayloads[0].RecipientID)
	require.Equal(t, workspaceID, taskService.digestPayloads[0].WorkspaceID)
}

func TestCreateSkipsDeliveryForDuplicateReplay(t *testing.T) {
	readAt := time.Now().Add(-time.Minute)
	repo := &notificationRepoStub{
		notification: CoreNotification{
			ID:          uuid.New(),
			RecipientID: uuid.New(),
			WorkspaceID: uuid.New(),
			Type:        "feedback_comment",
			EntityType:  "feedback",
			EntityID:    uuid.New(),
			ActorID:     uuid.New(),
			ReadAt:      &readAt,
		},
		insertedResults: []bool{true, false},
	}
	taskService := &notificationTasksStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, taskService)
	input := CoreNewNotification{
		DedupeKey:   "feedback-comment:source-id",
		RecipientID: repo.notification.RecipientID,
		WorkspaceID: repo.notification.WorkspaceID,
		Type:        repo.notification.Type,
		EntityType:  repo.notification.EntityType,
		EntityID:    repo.notification.EntityID,
		ActorID:     repo.notification.ActorID,
	}

	_, err := service.Create(context.Background(), input)
	require.NoError(t, err)
	replayed, err := service.Create(context.Background(), input)

	require.NoError(t, err)
	require.Equal(t, &readAt, replayed.ReadAt)
	require.Len(t, taskService.digestPayloads, 1)
}

func TestCreatePublishesOnlyWorkspaceNotificationsToGenericRealtimeChannel(t *testing.T) {
	tests := []struct {
		name            string
		entityType      string
		expectedPublish int
	}{
		{name: "workspace story notification", entityType: "story", expectedPublish: 1},
		{name: "public feedback notification", entityType: "feedback", expectedPublish: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &notificationRepoStub{notification: CoreNotification{
				ID:          uuid.New(),
				RecipientID: uuid.New(),
				WorkspaceID: uuid.New(),
				EntityType:  tt.entityType,
			}}
			taskService := &notificationTasksStub{}
			service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, taskService)
			publishCount := 0
			service.publishRealtime = func(context.Context, CoreNotification) error {
				publishCount++
				return nil
			}

			_, err := service.Create(context.Background(), CoreNewNotification{})

			require.NoError(t, err)
			require.Equal(t, tt.expectedPublish, publishCount)
			require.Len(t, taskService.digestPayloads, 1)
		})
	}
}
