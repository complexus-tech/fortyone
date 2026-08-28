package eventconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/events"
)

func (c *Consumer) handleObjectiveUpdated(ctx context.Context, event events.Event) error {
	var payload events.ObjectiveUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleObjectiveUpdated", "objective_id", payload.ObjectiveID, "lead_id", payload.LeadID)

	// The lead owns the follow-through on meaningful objective changes.
	if payload.LeadID != nil && *payload.LeadID != event.ActorID {
		objective, err := c.objectives.Get(ctx, payload.ObjectiveID, payload.WorkspaceID)
		objectiveName := "an objective"
		if err == nil {
			objectiveName = objective.Name
		} else {
			c.log.Error(ctx, "failed to get objective details for notification", "error", err, "objective_id", payload.ObjectiveID)
		}

		actor, err := c.users.GetUser(ctx, event.ActorID)
		actorName := "Someone"
		if err == nil {
			actorName = actor.Username
		}

		title := fmt.Sprintf("Objective updated: %s", objectiveName)
		template := "{actor} updated the objective {objective}"
		if _, leadChanged := payload.Updates["lead_user_id"]; leadChanged {
			title = fmt.Sprintf("You are now leading: %s", objectiveName)
			template = "{actor} assigned you as lead for {objective}"
		}

		notification := notifications.CoreNewNotification{
			RecipientID: *payload.LeadID,
			WorkspaceID: payload.WorkspaceID,
			Type:        "objective_update",
			EntityType:  "objective",
			EntityID:    payload.ObjectiveID,
			ActorID:     event.ActorID,
			Title:       title,
			Message: notifications.NotificationMessage{
				Template: template,
				Variables: map[string]notifications.Variable{
					"actor":     {Value: actorName, Type: "actor"},
					"objective": {Value: objectiveName, Type: "value"},
				},
			},
		}

		notification = withEventDedupeKey(event, notification, 0)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("create objective update notification: %w", err)
		}
	}

	return nil
}

func (c *Consumer) handleKeyResultUpdated(ctx context.Context, event events.Event) error {
	var payload events.KeyResultUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleKeyResultUpdated", "key_result_id", payload.KeyResultID)

	if c.notificationContexts == nil {
		return errors.New("notification context reader is unavailable")
	}
	audience, err := c.notificationContexts.ListKeyResultUpdateAudience(ctx, event.ActorID, payload.WorkspaceID, payload.KeyResultID)
	if err != nil {
		return fmt.Errorf("load key result notification audience: %w", err)
	}
	if len(audience) == 0 {
		c.log.Info(ctx, "key result has no currently authorized notification audience", "key_result_id", payload.KeyResultID)
		return nil
	}

	actorName := "Someone"
	if c.users != nil {
		if actor, actorErr := c.users.GetUser(ctx, event.ActorID); actorErr == nil {
			actorName = actor.Username
		}
	}

	for index, audienceMember := range audience {
		notification := withEventDedupeKey(event, notifications.CoreNewNotification{
			RecipientID: audienceMember.RecipientID,
			WorkspaceID: payload.WorkspaceID,
			Type:        notifications.NotificationTypeKeyResultUpdate,
			EntityType:  notifications.EntityTypeKeyResult,
			EntityID:    audienceMember.KeyResultID,
			ActorID:     event.ActorID,
			Title:       fmt.Sprintf("Key result updated: %s", audienceMember.KeyResultName),
			Message: notifications.NotificationMessage{
				Template: "{actor} updated {keyResult} under {objective}",
				Variables: map[string]notifications.Variable{
					"actor":     {Value: actorName, Type: "actor"},
					"keyResult": {Value: audienceMember.KeyResultName, Type: "value"},
					"objective": {Value: audienceMember.ObjectiveName, Type: "value"},
				},
			},
		}, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("create key result update notification: %w", err)
		}
	}

	return nil
}
