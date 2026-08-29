package notifications

import (
	"context"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

const meaningfulScheduleShiftMinutes = 60

var scheduleActivityNamespace = uuid.MustParse("9c70549c-74bb-5b68-9d6e-58db2e45d451")

func (r *Rules) handleScheduleTransition(
	ctx context.Context,
	payload events.StoryUpdatedPayload,
	actorID uuid.UUID,
	excludedRecipients map[uuid.UUID]struct{},
) []CoreNewNotification {
	if !shouldNotifyScheduleTransition(payload.Schedule) {
		return nil
	}

	transition := payload.Schedule
	recipientID := transition.UserID
	if recipientID == uuid.Nil || !shouldNotify(recipientID, actorID) {
		return nil
	}
	if _, excluded := excludedRecipients[recipientID]; excluded {
		return nil
	}

	timezone := r.getUserTimezone(ctx, recipientID, transition.Timezone)
	message := scheduleTransitionMessage(payload, timezone)
	storyTitle := r.getStoryTitle(ctx, payload.StoryID, payload.WorkspaceID)
	return []CoreNewNotification{r.createNotification(
		recipientID,
		payload,
		actorID,
		"story_update",
		storyTitle,
		message,
	)}
}

// RecordScheduleTransitionActivity appends one human-readable story activity
// for schedule changes that are significant enough to surface to a user.
func (r *Rules) RecordScheduleTransitionActivity(
	ctx context.Context,
	payload events.StoryUpdatedPayload,
	actorID uuid.UUID,
	eventTimestamp time.Time,
) error {
	transition := payload.Schedule
	if r.stories == nil || !shouldRecordScheduleTransition(transition) {
		return nil
	}

	timezone := r.getUserTimezone(ctx, transition.UserID, transition.Timezone)
	field, currentValue, oldValue, newValue := scheduleTransitionActivityValues(transition, timezone)
	reason := normalizedMayaReason(payload)
	var activityReason *string
	if reason != "" {
		activityReason = &reason
	}
	return r.stories.RecordActivity(ctx, storydomain.Activity{
		ID:           scheduleTransitionActivityID(payload, actorID, eventTimestamp),
		StoryID:      payload.StoryID,
		UserID:       actorID,
		Type:         "update",
		Field:        field,
		CurrentValue: currentValue,
		OldValue:     oldValue,
		NewValue:     newValue,
		Reason:       activityReason,
		WorkspaceID:  payload.WorkspaceID,
	})
}

func scheduleTransitionActivityID(payload events.StoryUpdatedPayload, actorID uuid.UUID, eventTimestamp time.Time) uuid.UUID {
	if eventTimestamp.IsZero() {
		return uuid.Nil
	}
	parts := []string{
		payload.WorkspaceID.String(),
		payload.StoryID.String(),
		actorID.String(),
		eventTimestamp.UTC().Format(time.RFC3339Nano),
	}
	if payload.Schedule != nil {
		parts = append(parts,
			payload.Schedule.UserID.String(),
			string(payload.Schedule.Kind),
			string(payload.Schedule.State),
		)
	}
	return uuid.NewSHA1(scheduleActivityNamespace, []byte(strings.Join(parts, ":")))
}

func shouldNotifyScheduleTransition(transition *events.StoryScheduleTransition) bool {
	if transition == nil {
		return false
	}
	if transition.Kind == events.StoryScheduleTransitionLocked ||
		transition.Kind == events.StoryScheduleTransitionUnlocked {
		return false
	}
	if transition.State == events.StoryScheduleStateNeedsTime ||
		transition.State == events.StoryScheduleStateAtRisk ||
		transition.State == events.StoryScheduleStateCannotFit {
		return true
	}
	if transition.Kind == events.StoryScheduleTransitionFirstSchedule ||
		transition.Kind == events.StoryScheduleTransitionDayChanged {
		return true
	}
	if transition.PreviousLocalDate != "" && transition.LocalDate != "" &&
		transition.PreviousLocalDate != transition.LocalDate {
		return true
	}
	return scheduleShiftMinutes(transition) >= meaningfulScheduleShiftMinutes
}

func shouldRecordScheduleTransition(transition *events.StoryScheduleTransition) bool {
	return shouldNotifyScheduleTransition(transition)
}

func scheduleShiftMinutes(transition *events.StoryScheduleTransition) int {
	if transition == nil {
		return 0
	}
	minutes := transition.ShiftMinutes
	if minutes < 0 {
		minutes = -minutes
	}
	if minutes > 0 || transition.PreviousStartAt == nil || transition.StartAt == nil {
		return minutes
	}
	minutes = int(transition.StartAt.Sub(*transition.PreviousStartAt).Minutes())
	if minutes < 0 {
		minutes = -minutes
	}
	return minutes
}

func scheduleTransitionMessage(payload events.StoryUpdatedPayload, timezone string) NotificationMessage {
	transition := payload.Schedule
	reason := normalizedMayaReason(payload)
	variables := map[string]Variable{}
	if reason != "" {
		variables["reason"] = Variable{Value: safeNotificationText(reason), Type: "value"}
	}

	template := "updated this task's schedule"
	switch {
	case transition == nil:
	case transition.State == events.StoryScheduleStateNeedsTime:
		template = "needs a time estimate before scheduling this task"
	case transition.State == events.StoryScheduleStateAtRisk:
		template = "flagged this task's schedule as at risk"
	case transition.State == events.StoryScheduleStateCannotFit:
		template = "could not fit this task into the current schedule"
	case transition.Kind == events.StoryScheduleTransitionFirstSchedule:
		template = "scheduled this task"
	case transition.Kind == events.StoryScheduleTransitionDayChanged || scheduleShiftMinutes(transition) >= meaningfulScheduleShiftMinutes:
		template = "moved this task"
	}

	if transition != nil {
		if scheduledFor := formatScheduleTransitionTime(transition, timezone); scheduledFor != "" {
			variables["scheduled_for"] = Variable{Value: scheduledFor, Type: "date"}
			template += " to {scheduled_for}"
		}
	}
	if reason != "" {
		template += ": {reason}"
	}
	return NotificationMessage{Template: template, Variables: variables}
}

func scheduleTransitionActivityValues(transition *events.StoryScheduleTransition, timezone string) (string, string, any, any) {
	if transition == nil {
		return "auto_scheduling_status", "Updated", nil, nil
	}
	if transition.StartAt != nil && (transition.Kind == events.StoryScheduleTransitionFirstSchedule ||
		transition.Kind == events.StoryScheduleTransitionDayChanged ||
		transition.Kind == events.StoryScheduleTransitionMoved) {
		return "auto_scheduling_time", formatScheduleTransitionTime(transition, timezone), transition.PreviousStartAt, transition.StartAt
	}
	if transition.Kind == events.StoryScheduleTransitionStateChanged && transition.State != "" {
		return "auto_scheduling_status", scheduleStateLabel(transition.State), transition.PreviousState, transition.State
	}
	if transition.State != transition.PreviousState && transition.State != "" {
		return "auto_scheduling_status", scheduleStateLabel(transition.State), transition.PreviousState, transition.State
	}
	return "auto_scheduling_time", formatScheduleTransitionTime(transition, timezone), transition.PreviousStartAt, transition.StartAt
}

func scheduleStateLabel(state events.StoryScheduleState) string {
	switch state {
	case events.StoryScheduleStateNeedsOwner:
		return "Needs owner"
	case events.StoryScheduleStateNeedsTime:
		return "Needs time"
	case events.StoryScheduleStatePlanning:
		return "Planning"
	case events.StoryScheduleStateScheduled:
		return "Scheduled"
	case events.StoryScheduleStateAtRisk:
		return "At risk"
	case events.StoryScheduleStateCannotFit:
		return "Cannot fit"
	case events.StoryScheduleStateLocked:
		return "Locked"
	default:
		return "Off"
	}
}

func formatScheduleTransitionTime(transition *events.StoryScheduleTransition, timezone string) string {
	if transition == nil || transition.StartAt == nil {
		return ""
	}
	value := transition.StartAt.UTC()
	if timezone := strings.TrimSpace(timezone); timezone != "" {
		if location, err := time.LoadLocation(timezone); err == nil {
			value = value.In(location)
		}
	}
	return value.Format("2 Jan 2006 at 15:04")
}

func normalizedMayaReason(payload events.StoryUpdatedPayload) string {
	if payload.Source != events.StoryUpdateSourceMaya {
		return ""
	}
	const maximumReasonRunes = 180
	reason := strings.Join(strings.Fields(payload.Reason), " ")
	runes := []rune(reason)
	if len(runes) > maximumReasonRunes {
		reason = strings.TrimSpace(string(runes[:maximumReasonRunes-3])) + "..."
	}
	return reason
}

func safeNotificationText(value string) string {
	return strings.NewReplacer("<", "‹", ">", "›").Replace(value)
}
