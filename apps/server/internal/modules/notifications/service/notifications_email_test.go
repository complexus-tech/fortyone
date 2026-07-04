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
	notification CoreNotification
}

func (r *notificationRepoStub) Create(context.Context, CoreNewNotification) (CoreNotification, error) {
	return r.notification, nil
}

func (r *notificationRepoStub) List(context.Context, uuid.UUID, uuid.UUID, int, int) ([]CoreNotification, error) {
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
