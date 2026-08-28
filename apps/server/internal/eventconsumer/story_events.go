package eventconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func (c *Consumer) handleStoryUpdated(ctx context.Context, event events.Event) error {
	var payload events.StoryUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleStoryUpdated", "story_id", payload.StoryID, "workspace_id", payload.WorkspaceID, "updates", payload.Updates)
	if err := c.notificationRules.RecordScheduleTransitionActivity(ctx, payload, event.ActorID, event.Timestamp); err != nil {
		return fmt.Errorf("record story schedule transition activity: %w", err)
	}

	// Use notification rules to process the story update
	notifications, err := c.notificationRules.ProcessStoryUpdate(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process story update notifications", "error", err)
		return err
	}

	if err := c.createStoryUpdateNotifications(ctx, event, notifications); err != nil {
		return fmt.Errorf("create story update notifications: %w", err)
	}
	if shouldBridgeFeedbackStatus(payload) && c.feedbackStatuses != nil {
		if err := c.feedbackStatuses.NotifyLinkedStoryStatusTransition(ctx, payload.WorkspaceID, payload.StoryID, event.ActorID, event.Timestamp); err != nil {
			return fmt.Errorf("bridge linked story feedback status: %w", err)
		}
	}
	if c.scheduleReconcile != nil {
		switch storyScheduleDispatch(payload.Updates) {
		case storyScheduleReconcileImmediate:
			if err := c.scheduleReconcile.EnqueueStoryScheduleReconcile(ctx, payload.WorkspaceID, payload.StoryID); err != nil {
				return fmt.Errorf("enqueue story schedule reconciliation: %w", err)
			}
		case storyScheduleReconcileBatch:
			if err := c.scheduleReconcile.EnqueueCalendarWorkspaceScheduleBatch(ctx, payload.WorkspaceID); err != nil {
				return fmt.Errorf("enqueue workspace schedule batch: %w", err)
			}
		}
	}

	// Workspace broadcasting
	if c.hasSignificantChanges(payload.Updates) {
		c.broadcastToWorkspace(ctx, payload, event.ActorID)
	}

	return nil
}

func (c *Consumer) createStoryUpdateNotifications(
	ctx context.Context,
	event events.Event,
	batch []notifications.CoreNewNotification,
) error {
	var createErrors []error
	for index, notification := range batch {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err, "recipient_id", notification.RecipientID)
			createErrors = append(createErrors, fmt.Errorf("recipient %s: %w", notification.RecipientID, err))
		}
	}
	return errors.Join(createErrors...)
}

func shouldReconcileStorySchedule(updates map[string]any) bool {
	return storyScheduleDispatch(updates) != storyScheduleReconcileNone
}

func storyScheduleDispatch(updates map[string]any) storyScheduleReconcileDispatch {
	relevant := false
	for _, field := range []string{
		"title",
		"estimated_duration_minutes",
		"minimum_focus_block_minutes",
		"assignee_id",
		"start_date",
		"end_date",
		"priority",
		"status_id",
		"sprint_id",
		"completed_at",
		"archived_at",
		"auto_scheduling_enabled",
		"auto_scheduling_locked",
	} {
		if _, changed := updates[field]; changed {
			relevant = true
			break
		}
	}
	if !relevant {
		return storyScheduleReconcileNone
	}
	if status, ok := updates["auto_scheduling_status"].(string); ok && status == "off" {
		return storyScheduleReconcileImmediate
	}
	if enabled, ok := updates["auto_scheduling_enabled"].(bool); ok && !enabled {
		return storyScheduleReconcileImmediate
	}
	if assignee, changed := updates["assignee_id"]; changed && storyAssigneeWasCleared(assignee) {
		return storyScheduleReconcileImmediate
	}
	if completedAt, changed := updates["completed_at"]; changed && completedAt != nil {
		return storyScheduleReconcileImmediate
	}
	if archivedAt, changed := updates["archived_at"]; changed && archivedAt != nil {
		return storyScheduleReconcileImmediate
	}
	return storyScheduleReconcileBatch
}

func storyAssigneeWasCleared(value any) bool {
	switch assignee := value.(type) {
	case nil:
		return true
	case uuid.UUID:
		return assignee == uuid.Nil
	case *uuid.UUID:
		return assignee == nil || *assignee == uuid.Nil
	case string:
		return assignee == uuid.Nil.String()
	default:
		return false
	}
}

func shouldBridgeFeedbackStatus(payload events.StoryUpdatedPayload) bool {
	_, statusChanged := payload.Updates["status_id"]
	return statusChanged && payload.PreviousStatusID != nil
}

// handleStoryCreated processes story creation events
func (c *Consumer) handleStoryCreated(ctx context.Context, event events.Event) error {
	var payload events.StoryCreatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleStoryCreated", "story_id", payload.StoryID, "workspace_id", payload.WorkspaceID, "assignee_id", payload.AssigneeID)

	// Use notification rules to process the story creation
	notifications, err := c.notificationRules.ProcessStoryCreated(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process story creation notifications", "error", err)
		return err
	}

	// Create all notifications
	for index, notification := range notifications {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err, "recipient_id", notification.RecipientID)
			// Continue with other notifications even if one fails
		}
	}

	return nil
}

// NEW: Check if updates contain workspace-worthy changes
func (c *Consumer) hasSignificantChanges(updates map[string]any) bool {
	significantFields := map[string]bool{
		"status_id":                  true,
		"completed_at":               true,
		"assignee_id":                true,
		"collaborator_ids":           true,
		"priority":                   true,
		"start_date":                 true,
		"end_date":                   true,
		"sprint_id":                  true,
		"estimate_unit":              true,
		"title":                      true,
		"auto_scheduling_enabled":    true,
		"auto_scheduling_locked":     true,
		"auto_scheduling_status":     true,
		"auto_scheduling_reason":     true,
		"auto_scheduling_updated_at": true,
	}

	for field := range updates {
		if significantFields[field] {
			return true
		}
	}
	return false
}

// Broadcast to workspace with frontend-friendly field names
func (c *Consumer) broadcastToWorkspace(ctx context.Context, payload events.StoryUpdatedPayload, actorID uuid.UUID) {
	frontendChanges := frontendStoryChanges(payload.Updates)

	// Skip if no significant changes
	if len(frontendChanges) == 0 {
		return
	}

	// Get actor name
	actorName := "Someone"
	if actor, err := c.users.GetUser(ctx, actorID); err == nil {
		actorName = actor.Username
	}

	// Create frontend-friendly workspace update
	workspaceUpdate := map[string]any{
		"type":        "story.workspace_update",
		"storyId":     payload.StoryID,     // camelCase
		"workspaceId": payload.WorkspaceID, // camelCase
		"changes":     frontendChanges,     // camelCase field names
		"actorId":     actorID,             // camelCase
		"actorName":   actorName,           // Actor name for display
		"timestamp":   time.Now().Unix(),
	}

	// Publish to workspace channel
	data, err := json.Marshal(workspaceUpdate)
	if err != nil {
		c.log.Error(ctx, "failed to marshal workspace update", "error", err)
		return
	}

	channelName := fmt.Sprintf("workspace-updates:%s", payload.WorkspaceID.String())
	if err := c.redis.Publish(ctx, channelName, data).Err(); err != nil {
		c.log.Error(ctx, "failed to publish workspace update", "error", err)
		return
	}

	c.log.Debug(ctx, "workspace update broadcasted", "storyId", payload.StoryID, "changes", len(frontendChanges), "actorName", actorName)
}

func frontendStoryChanges(updates map[string]any) map[string]any {
	dbToFrontendFields := map[string]string{
		"status_id":                  "statusId",
		"completed_at":               "completedAt",
		"assignee_id":                "assigneeId",
		"priority":                   "priority",
		"title":                      "title",
		"auto_scheduling_enabled":    "autoSchedulingEnabled",
		"auto_scheduling_locked":     "autoSchedulingLocked",
		"auto_scheduling_status":     "autoSchedulingStatus",
		"auto_scheduling_reason":     "autoSchedulingReason",
		"auto_scheduling_updated_at": "autoSchedulingUpdatedAt",
	}

	frontendChanges := make(map[string]any)
	for dbField, value := range updates {
		if frontendField, ok := dbToFrontendFields[dbField]; ok {
			frontendChanges[frontendField] = value
		}
	}
	return frontendChanges
}
