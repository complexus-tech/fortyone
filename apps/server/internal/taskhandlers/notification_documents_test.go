package taskhandlers

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDocumentNotificationEmailRetainsReceiptAndDocumentDestination(t *testing.T) {
	documentID, notificationID := uuid.New(), uuid.New()
	destination, label := notificationEmailDestination("document", documentID, "", notificationID, "https://product.fortyone.app/example")
	parsed, err := url.Parse(destination)
	require.NoError(t, err)
	require.Equal(t, "/example/notifications/"+notificationID.String(), parsed.Path)
	require.Equal(t, documentID.String(), parsed.Query().Get("entityId"))
	require.Equal(t, "document", parsed.Query().Get("entityType"))
	require.Equal(t, "document", label)
}
