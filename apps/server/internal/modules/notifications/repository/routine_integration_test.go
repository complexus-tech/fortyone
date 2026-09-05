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

func TestRoutineEmailClaimsSerializeAndFenceDeliveryAttempts(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	notification, _, err := f.repo.Create(ctx, f.storyNotification(f.recipientA, notificationDedupeKey("routine")))
	require.NoError(t, err)
	now := time.Now().UTC()
	claim := notifications.RoutineClaim{RecipientID: f.recipientA, WorkspaceID: f.workspaceA, Key: "briefing:today", Kind: "briefing", LocalDate: now, Now: now}
	var lock sync.Mutex
	var winners []uuid.UUID
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			id, err := f.repo.ClaimRoutine(ctx, claim)
			if err == nil && id != uuid.Nil {
				lock.Lock()
				winners = append(winners, id)
				lock.Unlock()
			}
		}()
	}
	group.Wait()
	require.Len(t, winners, 1, "only one worker can own a person's send window")
	id := winners[0]
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	completion := notifications.RoutineCompletion{ID: id, Scope: scope, NotificationIDs: []uuid.UUID{notification.ID}, GuidanceDate: &now, Sent: true, Now: now}
	wrongScope := completion
	wrongScope.Scope.WorkspaceID = f.workspaceB
	require.Error(t, f.repo.CompleteRoutine(ctx, wrongScope))
	require.NoError(t, f.repo.CompleteRoutine(ctx, completion))
	covered, err := f.repo.HasRoutineGuidance(ctx, scope, now)
	require.NoError(t, err)
	require.True(t, covered, "activity and guidance coverage commit together")
	digest, err := f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.Nil(t, digest, "completion covers notifications in the same transaction")
	replay, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.Equal(t, uuid.Nil, replay)
	claim.Key, claim.Kind = "activity:new", "activity"
	old, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NoError(t, f.repo.FailRoutine(ctx, old))
	current, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, old, current, "a retry must fence the old owner")
	completion.ID = old
	require.Error(t, f.repo.CompleteRoutine(ctx, completion))
	claim.Now = now.Add(11 * time.Minute)
	reclaimed, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, current, reclaimed)
	completion.ID = current
	require.Error(t, f.repo.CompleteRoutine(ctx, completion))
	completion.ID = reclaimed
	require.NoError(t, f.repo.CompleteRoutine(ctx, completion))
}
