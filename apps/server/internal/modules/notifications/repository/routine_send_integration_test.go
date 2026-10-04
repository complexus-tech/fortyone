//go:build integration

package notificationsrepository

import (
	"sync"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRoutineSendFenceKeepsUncertainWeekClosed(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	now := time.Date(2026, time.August, 31, 9, 0, 0, 0, time.UTC)
	claim := notifications.RoutineClaim{
		RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID,
		Key: "routine:" + now.Format(time.DateOnly), Kind: "activity", LocalDate: now, Now: now,
	}
	id, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
	require.NoError(t, f.repo.BeginRoutineSend(ctx, id, scope, now))

	// SMTP has accepted the summary, but the completion cannot commit.
	wrongScope := scope
	wrongScope.WorkspaceID = f.workspaceB
	require.Error(t, f.repo.CompleteRoutine(ctx, notifications.RoutineCompletion{
		ID: id, Scope: wrongScope, Sent: true, Now: now,
	}))
	var startedAt time.Time
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, "SELECT send_started_at FROM routine_email_deliveries WHERE id=$1", id).Scan(&startedAt))
	require.True(t, now.Equal(startedAt))
	claim.Now = now.Add(11 * time.Minute)
	for _, key := range []string{claim.Key, "legacy:another-batch"} {
		claim.Key = key
		retry, err := f.repo.ClaimRoutine(ctx, claim)
		require.NoError(t, err)
		require.Equal(t, uuid.Nil, retry, "a stale begun send must keep the same local week closed")
	}
	require.Error(t, f.repo.BeginRoutineSend(ctx, id, scope, claim.Now), "the same owner cannot begin a second send")

	claim.Now = now.Add(24 * time.Hour)
	claim.LocalDate = claim.Now
	claim.Key = "routine:" + claim.LocalDate.Format(time.DateOnly)
	retry, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.Equal(t, uuid.Nil, retry, "another day and key in the covered local week cannot send")

	claim.Now = now.Add(7 * 24 * time.Hour)
	claim.LocalDate = claim.Now
	claim.Key = "weekly:" + claim.LocalDate.Format(time.DateOnly)
	next, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, next, "the next Monday is eligible at exactly seven elapsed days")
	require.NoError(t, f.repo.BeginRoutineSend(ctx, next, scope, claim.Now))
}

func TestRoutineSendFenceAllowsExplicitFailureAndRejectsExpiredOwner(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	now := time.Now().UTC()
	claim := notifications.RoutineClaim{
		RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID,
		Key: "routine:" + now.Format(time.DateOnly), Kind: "activity", LocalDate: now, Now: now,
	}
	old, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	claim.Now = now.Add(11 * time.Minute)
	current, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, old, current, "unstarted stale work can still be recovered")
	require.Error(t, f.repo.BeginRoutineSend(ctx, old, scope, claim.Now), "an expired ID cannot authorize SMTP")
	wrongScope := scope
	wrongScope.RecipientID = f.recipientB
	require.Error(t, f.repo.BeginRoutineSend(ctx, current, wrongScope, claim.Now))
	require.NoError(t, f.repo.BeginRoutineSend(ctx, current, scope, claim.Now))
	require.NoError(t, f.repo.FailRoutine(ctx, current), "an explicit send rejection permits a retry")
	retry, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, retry)
	require.NotEqual(t, current, retry)
	var startedAt *time.Time
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, "SELECT send_started_at FROM routine_email_deliveries WHERE id=$1", retry).Scan(&startedAt))
	require.Nil(t, startedAt, "the replacement attempt starts without the failed owner's fence")
	require.Error(t, f.repo.BeginRoutineSend(ctx, current, scope, claim.Now))
	require.NoError(t, f.repo.BeginRoutineSend(ctx, retry, scope, claim.Now))
}

func TestRoutineSendFenceSerializesDistinctLegacyBatchKeys(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	now := time.Now().UTC()
	claim := notifications.RoutineClaim{
		RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID,
		Key: "legacy:batch", Kind: "activity", LocalDate: now, Now: now.Add(-11 * time.Minute),
	}
	old, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	claim.Key, claim.Now = "routine:"+now.Format(time.DateOnly), now
	current, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, old)
	require.NotEqual(t, uuid.Nil, current)
	var group sync.WaitGroup
	results := make(chan error, 2)
	for _, id := range []uuid.UUID{old, current} {
		group.Add(1)
		go func() {
			defer group.Done()
			results <- f.repo.BeginRoutineSend(ctx, id, scope, now)
		}()
	}
	group.Wait()
	close(results)
	started := 0
	for err := range results {
		if err == nil {
			started++
		}
	}
	require.Equal(t, 1, started, "only one legacy or replacement attempt can authorize a send for this week")
}
