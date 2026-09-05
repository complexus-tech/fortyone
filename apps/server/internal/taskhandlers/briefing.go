package taskhandlers

import (
	"context"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type RoutineDeliveryStore interface {
	ListRoutineRecipients(context.Context, *notifications.WeeklyDigestCursor, int) ([]notifications.RoutineRecipient, error)
	ClaimRoutine(context.Context, notifications.RoutineClaim) (uuid.UUID, error)
	CompleteRoutine(context.Context, notifications.RoutineCompletion) error
	FailRoutine(context.Context, uuid.UUID) error
}

type RoutineGuidanceStore interface {
	GetRoutineRecipient(context.Context, notifications.DeliveryScope) (*notifications.RoutineRecipient, error)
	HasRoutineGuidance(context.Context, notifications.DeliveryScope, time.Time) (bool, error)
}

// Retired scheduled task types remain registered so queued jobs are consumed
// without sending a second email. Guidance now joins unread activity batches.
func (h *handlers) HandleMorningBriefing(ctx context.Context, task *asynq.Task) error {
	return nil
}

func guidanceDate(now time.Time, timezone string) (time.Time, bool) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location = time.UTC
	}
	local := now.In(location)
	date := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
	return date, local.Weekday() != time.Saturday && local.Weekday() != time.Sunday
}

// Called after claiming the activity batch. The existing recipient claim also
// serializes the daily guidance check; completion covers both in one transaction.
func (h *handlers) activityGuidance(ctx context.Context, scope notifications.DeliveryScope, now time.Time) (jobs.BriefingContent, *time.Time, error) {
	var empty jobs.BriefingContent
	store, ok := h.routineDeliveries.(RoutineGuidanceStore)
	if !ok {
		return empty, nil, nil
	}
	recipient, err := store.GetRoutineRecipient(ctx, scope)
	if err != nil || recipient == nil {
		return empty, nil, err
	}
	date, eligible := guidanceDate(now, recipient.Timezone)
	if !eligible {
		return empty, nil, nil
	}
	covered, err := store.HasRoutineGuidance(ctx, scope, date)
	if err != nil || covered {
		return empty, nil, err
	}
	content, err := h.briefingSources.BuildBriefing(ctx, *recipient, date)
	if err != nil {
		return empty, nil, err
	}
	if len(content.Sections) == 0 {
		return empty, nil, nil
	}
	return content, &date, nil
}

func (h *handlers) completeRoutine(ctx context.Context, completion notifications.RoutineCompletion) error {
	stateCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	return h.routineDeliveries.CompleteRoutine(stateCtx, completion)
}
func (h *handlers) failRoutine(ctx context.Context, id uuid.UUID) error {
	stateCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	return h.routineDeliveries.FailRoutine(stateCtx, id)
}
