package notifications

import (
	"context"
	"fmt"
	"time"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

// getUserName gets a user's name with fallback
func (r *Rules) getUserName(ctx context.Context, userID uuid.UUID) string {
	var userName string
	if r.users == nil {
		return userName
	}
	if user, err := r.users.GetUser(ctx, userID); err == nil {
		userName = user.Username
	}
	return userName
}

// getStoryTitle gets a story's title with fallback
func (r *Rules) getStoryTitle(ctx context.Context, storyID, workspaceID uuid.UUID) string {
	var storyTitle string
	if story, err := r.stories.Get(ctx, storyID, workspaceID); err == nil {
		storyTitle = story.Title
	}
	return storyTitle
}

// getStatus gets a status
func (r *Rules) getStatus(ctx context.Context, statusID uuid.UUID, workspaceID uuid.UUID) states.CoreState {
	status, _ := r.statuses.Get(ctx, workspaceID, statusID)
	return status
}

// createNotification creates a notification with consistent structure
func (r *Rules) createNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, notifType, title string, message NotificationMessage) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        notifType,
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       title,
		Message:     message,
	}
}

// parseDate tries to parse date strings in various formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",           // YYYY-MM-DD
		"2006-01-02T15:04:05Z", // ISO8601
		time.RFC3339,
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, dateStr); err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// shouldNotify checks if a recipient should be notified (never notify the actor)
func shouldNotify(recipientID, actorID uuid.UUID) bool {
	return recipientID != actorID
}

func storyAudience(audienceIDs []uuid.UUID, audienceResolved bool, fallbackAssignee *uuid.UUID) []uuid.UUID {
	if !audienceResolved {
		if fallbackAssignee == nil {
			return nil
		}
		return []uuid.UUID{*fallbackAssignee}
	}

	seen := make(map[uuid.UUID]struct{}, len(audienceIDs))
	result := make([]uuid.UUID, 0, len(audienceIDs))
	for _, audienceID := range audienceIDs {
		if audienceID == uuid.Nil {
			continue
		}
		if _, exists := seen[audienceID]; exists {
			continue
		}
		seen[audienceID] = struct{}{}
		result = append(result, audienceID)
	}
	return result
}

func uuidSet(values []uuid.UUID) map[uuid.UUID]struct{} {
	result := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value != uuid.Nil {
			result[value] = struct{}{}
		}
	}
	return result
}

func parseUUIDSlice(value any) []uuid.UUID {
	switch values := value.(type) {
	case []uuid.UUID:
		return values
	case []string:
		result := make([]uuid.UUID, 0, len(values))
		for _, item := range values {
			if parsed, err := uuid.Parse(item); err == nil {
				result = append(result, parsed)
			}
		}
		return result
	case []any:
		result := make([]uuid.UUID, 0, len(values))
		for _, item := range values {
			if text, ok := item.(string); ok {
				if parsed, err := uuid.Parse(text); err == nil {
					result = append(result, parsed)
				}
			}
		}
		return result
	default:
		return nil
	}
}

func addNotificationRecipients(recipients map[uuid.UUID]struct{}, notifications []CoreNewNotification) {
	for _, notification := range notifications {
		recipients[notification.RecipientID] = struct{}{}
	}
}
