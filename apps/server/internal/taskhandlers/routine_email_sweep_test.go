package taskhandlers

import (
	"context"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type routineSweepStore struct {
	RoutineDeliveryStore
	pages   [][]notifications.RoutineRecipient
	cursors []*notifications.WeeklyDigestCursor
}

func (s *routineSweepStore) ListRoutineRecipients(_ context.Context, cursor *notifications.WeeklyDigestCursor, limit int) ([]notifications.RoutineRecipient, error) {
	if cursor == nil {
		s.cursors = append(s.cursors, nil)
	} else {
		copied := *cursor
		s.cursors = append(s.cursors, &copied)
	}
	page := len(s.cursors) - 1
	if page >= len(s.pages) {
		return nil, nil
	}
	return s.pages[page], nil
}

type routineSweepTasks struct {
	payloads []tasks.NotificationEmailDigestPayload
}

func (s *routineSweepTasks) EnqueueNotificationEmailDigest(payload tasks.NotificationEmailDigestPayload, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	s.payloads = append(s.payloads, payload)
	return &asynq.TaskInfo{}, nil
}

func TestRoutineSweepWakesPendingLocalMorningRecipientsAcrossPages(t *testing.T) {
	workspace := uuid.New()
	first := notifications.RoutineRecipient{UserID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), WorkspaceID: workspace, Timezone: "Africa/Harare"}
	second := notifications.RoutineRecipient{UserID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), WorkspaceID: workspace, Timezone: "America/New_York"}
	third := notifications.RoutineRecipient{UserID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), WorkspaceID: workspace, Timezone: "Asia/Tokyo"}
	store := &routineSweepStore{pages: [][]notifications.RoutineRecipient{{first, second}, {third}}}
	queue := &routineSweepTasks{}
	h := &handlers{routineDeliveries: store, routineTasks: queue}
	require.NoError(t, h.handleRoutineEmailSweepAt(t.Context(), time.Date(2026, 9, 8, 7, 0, 0, 0, time.UTC)))
	require.Equal(t, []tasks.NotificationEmailDigestPayload{{RecipientID: first.UserID, WorkspaceID: workspace}, {RecipientID: third.UserID, WorkspaceID: workspace}}, queue.payloads)
	require.Len(t, store.cursors, 3)
	require.Equal(t, second.UserID, store.cursors[1].UserID)
	require.Equal(t, third.UserID, store.cursors[2].UserID)
}

func TestRoutineSweepRejectsNonAdvancingRecipientCursor(t *testing.T) {
	recipient := notifications.RoutineRecipient{UserID: uuid.New(), WorkspaceID: uuid.New(), Timezone: "UTC"}
	store := &routineSweepStore{pages: [][]notifications.RoutineRecipient{{recipient}, {recipient}}}
	h := &handlers{routineDeliveries: store, routineTasks: &routineSweepTasks{}}
	require.ErrorContains(t, h.handleRoutineEmailSweepAt(t.Context(), time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC)), "cursor did not advance")
}
