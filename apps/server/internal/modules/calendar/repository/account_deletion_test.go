package calendarrepository

import (
	"context"
	"errors"
	"testing"

	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAccountDeletionStagesProviderCleanupBeforeDetaching(t *testing.T) {
	userID := uuid.New()
	queries := &accountDeletionQueries{userID: userID, connectionID: uuid.New()}
	require.NoError(t, stageAccountDeletion(t.Context(), queries, userID))
	require.Equal(t, []string{"lock", "list", "pending", "reactivate", "mirrors", "supersede", "enqueue", "detach", "events", "busy", "drain", "revoked", "scrub", "scrub-outbox"}, queries.calls)
	require.Equal(t, "delete", queries.enqueued.Operation)
	require.Equal(t, "provider-event", queries.enqueued.ProviderEventID)
	require.Equal(t, userID, queries.enqueued.UserID)
	require.True(t, queries.enqueued.ReactivateTerminal)
}

func TestAccountDeletionDoesNotDetachWhenDurableCleanupCannotBeQueued(t *testing.T) {
	cause := errors.New("outbox unavailable")
	queries := &accountDeletionQueries{userID: uuid.New(), connectionID: uuid.New(), enqueueError: cause}
	require.ErrorIs(t, stageAccountDeletion(t.Context(), queries, queries.userID), cause)
	require.NotContains(t, queries.calls, "detach")
	require.NotContains(t, queries.calls, "events")
	require.NotContains(t, queries.calls, "drain")
}

func TestAccountDeletionRequiresCallerTransaction(t *testing.T) {
	repository := &Repo{}
	require.Error(t, repository.StageAccountDeletion(t.Context(), nil, uuid.New()))
	_, err := repository.PendingAccountDeletion(t.Context(), nil, uuid.New())
	require.Error(t, err)
}

type accountDeletionQueries struct {
	calendarsql.Querier
	userID, connectionID uuid.UUID
	calls                []string
	enqueued             calendarsql.EnqueueScheduleEventOutboxParams
	enqueueError         error
}

func (q *accountDeletionQueries) LockCalendarUser(context.Context, calendarsql.LockCalendarUserParams) error {
	q.calls = append(q.calls, "lock")
	return nil
}
func (q *accountDeletionQueries) ListCalendarConnectionsByUser(context.Context, calendarsql.ListCalendarConnectionsByUserParams) ([]calendarsql.CalendarConnection, error) {
	q.calls = append(q.calls, "list")
	return []calendarsql.CalendarConnection{{ConnectionID: q.connectionID, UserID: q.userID, Provider: "google"}}, nil
}
func (q *accountDeletionQueries) MarkCalendarConnectionCleanupPending(context.Context, calendarsql.MarkCalendarConnectionCleanupPendingParams) (int64, error) {
	q.calls = append(q.calls, "pending")
	return 1, nil
}
func (q *accountDeletionQueries) ReactivateCalendarOutboxForCleanup(context.Context, calendarsql.ReactivateCalendarOutboxForCleanupParams) error {
	q.calls = append(q.calls, "reactivate")
	return nil
}
func (q *accountDeletionQueries) ListMayaScheduleMirrorsForCleanup(context.Context, calendarsql.ListMayaScheduleMirrorsForCleanupParams) ([]calendarsql.ListMayaScheduleMirrorsForCleanupRow, error) {
	q.calls = append(q.calls, "mirrors")
	calendarID, eventID := "primary", "provider-event"
	return []calendarsql.ListMayaScheduleMirrorsForCleanupRow{{BlockID: uuid.New(), WorkspaceID: uuid.New(), ExternalCalendarID: &calendarID, ExternalEventID: &eventID}}, nil
}
func (q *accountDeletionQueries) SupersedeStaleScheduleEventOutbox(context.Context, calendarsql.SupersedeStaleScheduleEventOutboxParams) error {
	q.calls = append(q.calls, "supersede")
	return nil
}
func (q *accountDeletionQueries) EnqueueScheduleEventOutbox(_ context.Context, params calendarsql.EnqueueScheduleEventOutboxParams) error {
	q.calls = append(q.calls, "enqueue")
	q.enqueued = params
	return q.enqueueError
}
func (q *accountDeletionQueries) DetachMayaScheduleMirrors(context.Context, calendarsql.DetachMayaScheduleMirrorsParams) error {
	q.calls = append(q.calls, "detach")
	return nil
}
func (q *accountDeletionQueries) DeleteCalendarConnectionEvents(context.Context, calendarsql.DeleteCalendarConnectionEventsParams) error {
	q.calls = append(q.calls, "events")
	return nil
}
func (q *accountDeletionQueries) DeleteCalendarConnectionBusyWindows(context.Context, calendarsql.DeleteCalendarConnectionBusyWindowsParams) error {
	q.calls = append(q.calls, "busy")
	return nil
}
func (q *accountDeletionQueries) DeleteDrainedCalendarConnection(context.Context, calendarsql.DeleteDrainedCalendarConnectionParams) error {
	q.calls = append(q.calls, "drain")
	return nil
}
func (q *accountDeletionQueries) DeleteRevokedAccountCalendarConnections(context.Context, calendarsql.DeleteRevokedAccountCalendarConnectionsParams) error {
	q.calls = append(q.calls, "revoked")
	return nil
}
func (q *accountDeletionQueries) ScrubAccountCalendarCleanup(context.Context, calendarsql.ScrubAccountCalendarCleanupParams) error {
	q.calls = append(q.calls, "scrub")
	return nil
}
func (q *accountDeletionQueries) ScrubAccountCalendarOutbox(context.Context, calendarsql.ScrubAccountCalendarOutboxParams) error {
	q.calls = append(q.calls, "scrub-outbox")
	return nil
}
