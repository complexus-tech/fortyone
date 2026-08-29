package notifications

import (
	"fmt"
	"strings"
)

// NotificationType is the finite set persisted by PostgreSQL's
// notification_type enum. Keeping this type in the domain prevents arbitrary
// strings from crossing the service/repository boundary.
type NotificationType string

const (
	NotificationTypeStoryUpdate             NotificationType = "story_update"
	NotificationTypeStoryComment            NotificationType = "story_comment"
	NotificationTypeCommentReply            NotificationType = "comment_reply"
	NotificationTypeObjectiveUpdate         NotificationType = "objective_update"
	NotificationTypeKeyResultUpdate         NotificationType = "key_result_update"
	NotificationTypeMention                 NotificationType = "mention"
	NotificationTypeFeedbackComment         NotificationType = "feedback_comment"
	NotificationTypeFeedbackStatusUpdate    NotificationType = "feedback_status_update"
	NotificationTypeFeedbackUpdatePublished NotificationType = "feedback_update_published"
	NotificationTypeFeedbackItemMerged      NotificationType = "feedback_item_merged"
	NotificationTypeStrategyUpdate          NotificationType = "strategy_update"
)

func ParseNotificationType(value string) (NotificationType, error) {
	typeValue := NotificationType(strings.TrimSpace(value))
	if !typeValue.Valid() {
		return "", fmt.Errorf("%w: unsupported notification type %q", ErrInvalid, value)
	}
	return typeValue, nil
}

func (notificationType NotificationType) Valid() bool {
	switch notificationType {
	case NotificationTypeStoryUpdate,
		NotificationTypeStoryComment,
		NotificationTypeCommentReply,
		NotificationTypeObjectiveUpdate,
		NotificationTypeKeyResultUpdate,
		NotificationTypeMention,
		NotificationTypeFeedbackComment,
		NotificationTypeFeedbackStatusUpdate,
		NotificationTypeFeedbackUpdatePublished,
		NotificationTypeFeedbackItemMerged,
		NotificationTypeStrategyUpdate:
		return true
	default:
		return false
	}
}

// EntityType is the finite set persisted by PostgreSQL's entity_type enum.
type EntityType string

const (
	EntityTypeStory     EntityType = "story"
	EntityTypeComment   EntityType = "comment"
	EntityTypeObjective EntityType = "objective"
	EntityTypeKeyResult EntityType = "key_result"
	EntityTypeFeedback  EntityType = "feedback"
	EntityTypeStrategy  EntityType = "strategy"
)

func ParseEntityType(value string) (EntityType, error) {
	entityType := EntityType(strings.TrimSpace(value))
	if !entityType.Valid() {
		return "", fmt.Errorf("%w: unsupported notification entity type %q", ErrInvalid, value)
	}
	return entityType, nil
}

func (entityType EntityType) Valid() bool {
	switch entityType {
	case EntityTypeStory, EntityTypeComment, EntityTypeObjective,
		EntityTypeKeyResult, EntityTypeFeedback, EntityTypeStrategy:
		return true
	default:
		return false
	}
}

func SupportsInAppDelivery(notificationType NotificationType) bool {
	return notificationType != NotificationTypeStrategyUpdate
}

func (notificationType NotificationType) SupportsEntity(entityType EntityType) bool {
	switch notificationType {
	case NotificationTypeStoryUpdate,
		NotificationTypeStoryComment,
		NotificationTypeCommentReply,
		NotificationTypeMention:
		return entityType == EntityTypeStory || entityType == EntityTypeComment
	case NotificationTypeObjectiveUpdate:
		return entityType == EntityTypeObjective
	case NotificationTypeKeyResultUpdate:
		return entityType == EntityTypeKeyResult
	case NotificationTypeFeedbackComment,
		NotificationTypeFeedbackStatusUpdate,
		NotificationTypeFeedbackUpdatePublished,
		NotificationTypeFeedbackItemMerged:
		return entityType == EntityTypeFeedback
	case NotificationTypeStrategyUpdate:
		return entityType == EntityTypeStrategy
	default:
		return false
	}
}
