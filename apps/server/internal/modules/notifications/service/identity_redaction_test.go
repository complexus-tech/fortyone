package notifications

import (
	"context"
	"encoding/json"
	"testing"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type identityRepository struct {
	repositoryStub
}

func (repository *identityRepository) List(context.Context, notificationsdomain.ListQuery) ([]notificationsdomain.Notification, error) {
	return []notificationsdomain.Notification{repository.created}, nil
}

func TestInternalIdentityReferencesDoNotReachInboxOrRealtime(t *testing.T) {
	t.Parallel()

	input, created := validNotification()
	assigneeID := uuid.New()
	input.Message.IdentityReferences = map[string]uuid.UUID{"assignee": assigneeID}
	created.Message = input.Message
	repository := &identityRepository{repositoryStub{created: created, inserted: true}}
	service := newTestService(repository, &taskStub{})
	var published []CoreNotification
	service.publishRealtime = func(_ context.Context, notification CoreNotification) error {
		published = append(published, notification)
		return nil
	}

	_, err := service.Create(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, published, 1)
	inbox, err := service.List(context.Background(), input.RecipientID, input.WorkspaceID, "", 20, 0)
	require.NoError(t, err)
	require.Len(t, inbox, 1)

	for _, notification := range []CoreNotification{published[0], inbox[0]} {
		encoded, err := json.Marshal(notification)
		require.NoError(t, err)
		require.Nil(t, notification.Message.IdentityReferences)
		require.NotContains(t, string(encoded), "identityReferences")
		require.NotContains(t, string(encoded), assigneeID.String())
		require.Equal(t, created.Message.Template, notification.Message.Template)
	}
	require.Equal(t, assigneeID, repository.created.Message.IdentityReferences["assignee"])
}
