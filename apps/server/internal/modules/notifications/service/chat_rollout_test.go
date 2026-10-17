package notifications

import (
	"context"
	"errors"
	domain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type rolloutDeliveryRepository struct {
	Repository
	PushRepository
	workspace uuid.UUID
}

func (r rolloutDeliveryRepository) GetEmailDelivery(context.Context, domain.EmailNotificationQuery) (*domain.EmailNotification, error) {
	return &domain.EmailNotification{WorkspaceID: r.workspace, EntityType: domain.EntityTypeChatMessage}, nil
}
func (r rolloutDeliveryRepository) ListEmailDigest(context.Context, domain.DeliveryScope) (*domain.EmailDigest, error) {
	return &domain.EmailDigest{WorkspaceID: r.workspace, Items: []domain.EmailDigestItem{{EntityType: domain.EntityTypeChatMessage}, {EntityType: domain.EntityTypeStory}}}, nil
}
func (r rolloutDeliveryRepository) GetPushDelivery(context.Context, uuid.UUID) (*domain.PushDelivery, error) {
	return &domain.PushDelivery{WorkspaceID: r.workspace, EntityType: domain.EntityTypeChatMessage}, nil
}

func TestRolloutSuppressesAlreadyQueuedChatNotifications(t *testing.T) {
	workspace := uuid.New()
	for _, enabled := range []bool{false, true} {
		repo := rolloutDeliveryRepository{workspace: workspace}
		service := New(nil, repo, nil, nil, WithChatSettings(chatrollout.New(chatrollout.ReaderFunc(func(_ context.Context, id uuid.UUID) (bool, error) { return enabled && id == workspace, nil }))))
		push, err := service.GetPushDelivery(t.Context(), uuid.New())
		require.NoError(t, err)
		require.Equal(t, enabled, push != nil)
		email, err := service.GetEmailDelivery(t.Context(), domain.EmailNotificationQuery{})
		require.NoError(t, err)
		require.Equal(t, enabled, email != nil)
		digest, err := service.ListEmailDigest(t.Context(), domain.DeliveryScope{WorkspaceID: workspace})
		require.NoError(t, err)
		if enabled {
			require.Len(t, digest.Items, 2)
		} else {
			require.Len(t, digest.Items, 1)
			require.Equal(t, domain.EntityTypeStory, digest.Items[0].EntityType)
		}
	}
}

func TestSettingLookupFailureDoesNotAcknowledgeNotificationDelivery(t *testing.T) {
	failure := errors.New("database unavailable")
	workspace := uuid.New()
	service := New(nil, rolloutDeliveryRepository{workspace: workspace}, nil, nil, WithChatSettings(chatrollout.New(chatrollout.ReaderFunc(func(context.Context, uuid.UUID) (bool, error) { return false, failure }))))
	push, err := service.GetPushDelivery(t.Context(), uuid.New())
	require.ErrorIs(t, err, failure)
	require.Nil(t, push)
	email, err := service.GetEmailDelivery(t.Context(), domain.EmailNotificationQuery{})
	require.ErrorIs(t, err, failure)
	require.Nil(t, email)
	digest, err := service.ListEmailDigest(t.Context(), domain.DeliveryScope{WorkspaceID: workspace})
	require.ErrorIs(t, err, failure)
	require.Nil(t, digest)
}
