package notifications

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type KeyResultAudienceQuery struct {
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
	KeyResultID uuid.UUID
}

func (query KeyResultAudienceQuery) Validate() error {
	if query.ActorID == uuid.Nil || query.WorkspaceID == uuid.Nil || query.KeyResultID == uuid.Nil {
		return fmt.Errorf("%w: actor, workspace, and key result are required", ErrInvalid)
	}
	return nil
}

type KeyResultAudienceMember struct {
	RecipientID   uuid.UUID
	KeyResultID   uuid.UUID
	ObjectiveID   uuid.UUID
	KeyResultName string
	ObjectiveName string
}

type DeliveryScope struct {
	RecipientID uuid.UUID
	WorkspaceID uuid.UUID
}

func (scope DeliveryScope) Validate() error {
	if scope.RecipientID == uuid.Nil || scope.WorkspaceID == uuid.Nil {
		return fmt.Errorf("%w: recipient and workspace are required", ErrInvalid)
	}
	return nil
}

type EmailNotificationQuery struct {
	Scope          DeliveryScope
	NotificationID uuid.UUID
}

func (query EmailNotificationQuery) Validate() error {
	if err := query.Scope.Validate(); err != nil {
		return err
	}
	if query.NotificationID == uuid.Nil {
		return fmt.Errorf("%w: notification ID is required", ErrInvalid)
	}
	return nil
}

type EmailNotification struct {
	NotificationID   uuid.UUID
	RecipientID      uuid.UUID
	WorkspaceID      uuid.UUID
	NotificationType NotificationType
	EntityType       EntityType
	EntityID         uuid.UUID
	Title            string
	Message          json.RawMessage
	UserEmail        string
	UserName         string
	ActorName        string
	WorkspaceName    string
	WorkspaceSlug    string
	WorkspaceRole    string
	EmailEnabled     bool
	FeedbackSlug     string
}

type EmailDigestItem struct {
	NotificationID   uuid.UUID
	NotificationType NotificationType
	EntityType       EntityType
	EntityID         uuid.UUID
	Title            string
	Message          json.RawMessage
	CreatedAt        time.Time
	ActorName        string
	FeedbackSlug     string
}

type EmailDigest struct {
	RecipientID   uuid.UUID
	WorkspaceID   uuid.UUID
	UserEmail     string
	UserName      string
	WorkspaceName string
	WorkspaceSlug string
	WorkspaceRole string
	Items         []EmailDigestItem
}

type MarkEmailSent struct {
	Scope           DeliveryScope
	NotificationIDs []uuid.UUID
	At              time.Time
}

func (command MarkEmailSent) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if len(command.NotificationIDs) == 0 || command.At.IsZero() {
		return fmt.Errorf("%w: notification IDs and delivery time are required", ErrInvalid)
	}
	for _, notificationID := range command.NotificationIDs {
		if notificationID == uuid.Nil {
			return fmt.Errorf("%w: notification IDs cannot contain nil", ErrInvalid)
		}
	}
	return nil
}
