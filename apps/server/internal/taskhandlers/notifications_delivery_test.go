package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type notificationDeliveryStoreStub struct {
	emailQuery  notificationsdomain.EmailNotificationQuery
	email       *notificationsdomain.EmailNotification
	digestScope notificationsdomain.DeliveryScope
	digest      *notificationsdomain.EmailDigest
	teamScope   notificationsdomain.DeliveryScope
	teamIDs     []uuid.UUID
	markScope   notificationsdomain.DeliveryScope
	markedIDs   []uuid.UUID
	err         error
}

func (stub *notificationDeliveryStoreStub) GetEmailDelivery(
	_ context.Context,
	query notificationsdomain.EmailNotificationQuery,
) (*notificationsdomain.EmailNotification, error) {
	stub.emailQuery = query
	return stub.email, stub.err
}

func (stub *notificationDeliveryStoreStub) ListEmailDigest(
	_ context.Context,
	scope notificationsdomain.DeliveryScope,
) (*notificationsdomain.EmailDigest, error) {
	stub.digestScope = scope
	return stub.digest, stub.err
}

func (stub *notificationDeliveryStoreStub) ListDeliveryTeamIDs(
	_ context.Context,
	scope notificationsdomain.DeliveryScope,
) ([]uuid.UUID, error) {
	stub.teamScope = scope
	return stub.teamIDs, stub.err
}

func (stub *notificationDeliveryStoreStub) MarkEmailSent(
	_ context.Context,
	scope notificationsdomain.DeliveryScope,
	notificationIDs []uuid.UUID,
) error {
	stub.markScope = scope
	stub.markedIDs = append([]uuid.UUID(nil), notificationIDs...)
	return stub.err
}

func TestNotificationTaskHandlerUsesScopedDeliveryPort(t *testing.T) {
	t.Parallel()

	recipientID, workspaceID, notificationID, entityID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	scope := notificationsdomain.DeliveryScope{RecipientID: recipientID, WorkspaceID: workspaceID}
	message := json.RawMessage(`{"template":"updated","variables":{}}`)
	store := &notificationDeliveryStoreStub{email: &notificationsdomain.EmailNotification{
		NotificationID: notificationID, RecipientID: recipientID, WorkspaceID: workspaceID,
		NotificationType: notificationsdomain.NotificationTypeStoryUpdate,
		EntityType:       notificationsdomain.EntityTypeStory, EntityID: entityID,
		Title: "Story updated", Message: message, UserEmail: "recipient@example.com",
		UserName: "Recipient", ActorName: "Actor", WorkspaceName: "Workspace",
		WorkspaceSlug: "workspace", WorkspaceRole: "member", EmailEnabled: true,
	}}
	handler := &handlers{
		log:                    logger.NewWithText(io.Discard, slog.LevelError, "notification-delivery-test"),
		notificationDeliveries: store,
	}

	data, err := handler.getNotificationEmailData(t.Context(), notificationsdomain.EmailNotificationQuery{
		Scope: scope, NotificationID: notificationID,
	})
	require.NoError(t, err)
	require.Equal(t, notificationsdomain.EmailNotificationQuery{Scope: scope, NotificationID: notificationID}, store.emailQuery)
	require.Equal(t, notificationID, data.NotificationID)
	require.Equal(t, string(notificationsdomain.NotificationTypeStoryUpdate), data.NotificationType)
	require.Equal(t, string(notificationsdomain.EntityTypeStory), data.EntityType)
	require.Equal(t, message, data.Message)

	require.NoError(t, handler.markNotificationsEmailSent(t.Context(), scope, []uuid.UUID{notificationID}))
	require.Equal(t, scope, store.markScope)
	require.Equal(t, []uuid.UUID{notificationID}, store.markedIDs)
}

func TestNotificationTaskHandlerDigestAndStrategyTeamReadsStayScoped(t *testing.T) {
	t.Parallel()

	recipientID, workspaceID, notificationID, teamID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	scope := notificationsdomain.DeliveryScope{RecipientID: recipientID, WorkspaceID: workspaceID}
	strategyMessage, err := json.Marshal(NotificationMessage{Strategy: &strategyNotificationSnapshot{
		Version: 1, Kind: "weekly_check_in",
		WeeklyCheckIn: &strategyWeeklyCheckInSnapshot{
			TeamCounts: []strategyWeeklyCheckInTeamCountsSnapshot{{TeamID: teamID}},
			Objectives: []strategyObjectiveSnapshot{{ID: uuid.New(), TeamID: teamID, Name: "Objective"}},
		},
	}})
	require.NoError(t, err)
	store := &notificationDeliveryStoreStub{
		digest: &notificationsdomain.EmailDigest{
			RecipientID: recipientID, WorkspaceID: workspaceID,
			UserEmail: "recipient@example.com", UserName: "Recipient",
			WorkspaceName: "Workspace", WorkspaceSlug: "workspace", WorkspaceRole: "member",
			Items: []notificationsdomain.EmailDigestItem{{
				NotificationID:   notificationID,
				NotificationType: notificationsdomain.NotificationTypeStrategyUpdate,
				EntityType:       notificationsdomain.EntityTypeStrategy,
				EntityID:         workspaceID, Message: strategyMessage, CreatedAt: time.Now(),
			}},
		},
		teamIDs: []uuid.UUID{teamID},
	}
	handler := &handlers{notificationDeliveries: store}

	data, err := handler.getNotificationEmailDigestData(t.Context(), recipientID, workspaceID)
	require.NoError(t, err)
	require.Equal(t, scope, store.digestScope)
	suppressed, err := handler.filterStrategyDigestForCurrentAccess(t.Context(), data)
	require.NoError(t, err)
	require.Empty(t, suppressed)
	require.Equal(t, scope, store.teamScope)
	require.Len(t, data.Items, 1)
}

func TestNotificationTaskHandlerPropagatesDeliveryPortErrors(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("delivery store unavailable")
	store := &notificationDeliveryStoreStub{err: storeErr}
	handler := &handlers{
		log:                    logger.NewWithText(io.Discard, slog.LevelError, "notification-delivery-test"),
		notificationDeliveries: store,
	}
	scope := notificationsdomain.DeliveryScope{RecipientID: uuid.New(), WorkspaceID: uuid.New()}

	_, err := handler.getNotificationEmailData(t.Context(), notificationsdomain.EmailNotificationQuery{
		Scope: scope, NotificationID: uuid.New(),
	})
	require.ErrorIs(t, err, storeErr)
	_, err = handler.getNotificationEmailDigestData(t.Context(), scope.RecipientID, scope.WorkspaceID)
	require.ErrorIs(t, err, storeErr)
	require.ErrorIs(t, handler.markNotificationsEmailSent(t.Context(), scope, []uuid.UUID{uuid.New()}), storeErr)
}
