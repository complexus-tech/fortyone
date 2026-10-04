package taskhandlers

import (
	"context"
	"fmt"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

type RoutineEmailTasks interface {
	EnqueueNotificationEmailDigest(tasks.NotificationEmailDigestPayload, ...asynq.Option) (*asynq.TaskInfo, error)
}

// HandleRoutineEmailSweep wakes deferred unread work and due reviewer feedback
// even when there has been no fresh activity since last week's summary.
func (h *handlers) HandleRoutineEmailSweep(ctx context.Context, _ *asynq.Task) error {
	return h.handleRoutineEmailSweepAt(ctx, time.Now().UTC())
}

func (h *handlers) handleRoutineEmailSweepAt(ctx context.Context, now time.Time) error {
	if h.routineDeliveries == nil || h.routineTasks == nil {
		return fmt.Errorf("routine email sweep dependencies are unavailable")
	}
	var cursor *notifications.WeeklyDigestCursor
	for {
		recipients, err := h.routineDeliveries.ListRoutineRecipients(ctx, cursor, 100)
		if err != nil {
			return fmt.Errorf("list routine email sweep recipients: %w", err)
		}
		if len(recipients) == 0 {
			return nil
		}
		for _, recipient := range recipients {
			if _, ready := routineDeliveryDate(now, recipient.Timezone); !ready {
				continue
			}
			_, err := h.routineTasks.EnqueueNotificationEmailDigest(tasks.NotificationEmailDigestPayload{
				RecipientID: recipient.UserID, WorkspaceID: recipient.WorkspaceID,
			}, asynq.ProcessIn(0))
			if err != nil {
				return fmt.Errorf("wake routine email for %s: %w", recipient.UserID, err)
			}
		}
		last := recipients[len(recipients)-1]
		if cursor != nil && (last.WorkspaceID.String() < cursor.WorkspaceID.String() ||
			(last.WorkspaceID == cursor.WorkspaceID && last.UserID.String() <= cursor.UserID.String())) {
			return fmt.Errorf("routine email recipient cursor did not advance")
		}
		cursor = &notifications.WeeklyDigestCursor{WorkspaceID: last.WorkspaceID, UserID: last.UserID}
	}
}

func routineDeliveryDate(now time.Time, timezone string) (time.Time, bool) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location = time.UTC
	}
	local := now.In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC), local.Hour() >= 9
}

func routineDeliveryKey(date time.Time) string {
	return "routine:weekly:" + routineWeekStart(date).Format(time.DateOnly)
}

// Dates represent the recipient's calendar at UTC midnight. Calendar arithmetic
// keeps week identity independent of DST and the worker's timezone.
func routineWeekStart(date time.Time) time.Time {
	daysSinceMonday := (int(date.Weekday()) + 6) % 7
	return date.AddDate(0, 0, -daysSinceMonday)
}

func (h *handlers) routineRecipient(ctx context.Context, scope notifications.DeliveryScope) (*notifications.RoutineRecipient, error) {
	store, ok := h.routineDeliveries.(RoutineGuidanceStore)
	if !ok {
		return nil, nil
	}
	return store.GetRoutineRecipient(ctx, scope)
}
