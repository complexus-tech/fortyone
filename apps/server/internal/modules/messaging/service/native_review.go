package messaging

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// The review is captured with the proposal's version so later human edits do
// not change the before/after values of an approval that is already visible.
func (e *FortyOneToolExecutor) nativeMutationReview(ctx context.Context, scope ToolScope, story messagingStory, claims storyMutationClaims) (string, error) {
	updates, err := desiredStoryUpdates(story, scope.UserID, claims.Title, claims.Priority, claims.AssigneeAction, claims.StatusID, nil, nil, nil, claims.StartDate, claims.EndDate,
		storyTimeMutation{estimatedDurationAction: claims.EstimatedDurationAction, estimatedDurationMinutes: claims.EstimatedDurationMinutes, minimumFocusBlockAction: claims.MinimumFocusBlockAction, minimumFocusBlockMinutes: claims.MinimumFocusBlockMinutes},
		storyAutoSchedulingMutation{enabled: claims.AutoSchedulingEnabled, locked: claims.AutoSchedulingLocked})
	if err != nil {
		return "", err
	}
	addNativeStoryUpdates(updates, story, claims)
	before := map[string]any{
		"title": story.Title, "description": story.Description, "priority": story.Priority,
		"assignee_id": story.Assignee, "status_id": story.Status, "start_date": story.StartDate, "end_date": story.EndDate,
		"estimated_duration_minutes": story.EstimatedDurationMinutes, "minimum_focus_block_minutes": story.MinimumFocusBlockMinutes,
		"auto_scheduling_enabled": story.AutoSchedulingEnabled, "auto_scheduling_locked": story.AutoSchedulingLocked,
	}
	labels := map[string]string{"title": "Title", "description": "Description", "priority": "Priority", "assignee_id": "Assignee", "status_id": "Status", "start_date": "Start date", "end_date": "Due date", "estimated_duration_minutes": "Time needed", "minimum_focus_block_minutes": "Minimum focus block", "auto_scheduling_enabled": "Auto-scheduling", "auto_scheduling_locked": "Schedule locked"}
	var review strings.Builder
	for index, field := range claims.ChangedFields {
		label, known := labels[field]
		if !known {
			return "", fmt.Errorf("%w: field has no reviewable native representation", ErrInvalidToolArguments)
		}
		oldValue, err := e.nativeReviewValue(ctx, scope, field, before[field])
		if err != nil {
			return "", err
		}
		newValue, err := e.nativeReviewValue(ctx, scope, field, updates[field])
		if err != nil {
			return "", err
		}
		if index > 0 {
			review.WriteString("\n\n")
		}
		fmt.Fprintf(&review, "%s\nBefore:\n%s\nAfter:\n%s", label, oldValue, newValue)
	}
	if review.Len() > 64<<10 {
		return "", fmt.Errorf("%w: native proposal review is too large", ErrInvalidToolArguments)
	}
	return review.String(), nil
}

func (e *FortyOneToolExecutor) nativeReviewValue(ctx context.Context, scope ToolScope, field string, value any) (string, error) {
	if value == nil {
		return "Not set", nil
	}
	switch typed := value.(type) {
	case *string:
		if typed == nil || *typed == "" {
			return "Empty", nil
		}
		return *typed, nil
	case string:
		if typed == "" {
			return "Empty", nil
		}
		return typed, nil
	case *int:
		if typed == nil {
			return "Not set", nil
		}
		return fmt.Sprintf("%d minutes", *typed), nil
	case int:
		return fmt.Sprintf("%d minutes", typed), nil
	case bool:
		if typed {
			return "Yes", nil
		}
		return "No", nil
	case *time.Time:
		if typed == nil {
			return "Not set", nil
		}
		return typed.UTC().Format(time.RFC3339), nil
	case time.Time:
		return typed.UTC().Format(time.RFC3339), nil
	case *uuid.UUID:
		if typed == nil {
			return "Not set", nil
		}
		return e.nativeReviewReference(ctx, scope, field, *typed)
	case uuid.UUID:
		return e.nativeReviewReference(ctx, scope, field, typed)
	default:
		return "", fmt.Errorf("%w: native proposal value cannot be reviewed", ErrInvalidToolArguments)
	}
}

func (e *FortyOneToolExecutor) nativeReviewReference(ctx context.Context, scope ToolScope, field string, id uuid.UUID) (string, error) {
	switch field {
	case "status_id":
		statuses, err := e.nativeStatuses(ctx, scope)
		if err != nil {
			return "", err
		}
		for _, status := range statuses {
			if status.ID == id && strings.TrimSpace(status.Name) != "" {
				return status.Name, nil
			}
		}
	case "assignee_id":
		if e.users == nil {
			break
		}
		member, err := e.users.GetUser(ctx, id)
		if err != nil {
			return "", fmt.Errorf("resolve native approval member: %w", err)
		}
		if member.ID == id && memberDisplayName(member) != "" {
			return memberDisplayName(member), nil
		}
		// The team list is also authoritative for adapters that provide only
		// their current user through GetUser.
		active, err := e.activeTeamMemberByID(ctx, scope.WorkspaceID, scope.SourceScope.TeamID, id)
		if err != nil {
			return "", err
		}
		if active != nil && memberDisplayName(*active) != "" {
			return memberDisplayName(*active), nil
		}
	}
	return "", fmt.Errorf("%w: %s has no accessible human-readable name", ErrInvalidToolArguments, field)
}
