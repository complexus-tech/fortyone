//go:build integration

package notificationsrepository

import (
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRoutineWeeklyLimitIncludesLegacyDailyDeliveries(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	monday := time.Date(2026, time.December, 28, 9, 0, 0, 0, time.UTC)
	claim := notifications.RoutineClaim{
		RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID,
		Key: "routine:2026-12-28", Kind: "activity", LocalDate: monday, Now: monday,
	}
	legacy, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	// Legacy senders completed daily claims without a send-start marker.
	require.NoError(t, f.repo.CompleteRoutine(ctx, notifications.RoutineCompletion{
		ID: legacy, Scope: scope, Sent: true, Now: monday,
	}))
	claim.Now = monday.Add(4 * 24 * time.Hour)
	claim.LocalDate = claim.Now
	claim.Key = "weekly:2026-12-28"
	retry, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.Equal(t, uuid.Nil, retry, "a new calendar year does not start a new ISO week or erase legacy coverage")

	claim.Now = monday.Add(7 * 24 * time.Hour)
	claim.LocalDate = claim.Now
	claim.Key = "weekly:2027-01-04"
	next, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, next)
	require.NoError(t, f.repo.BeginRoutineSend(ctx, next, scope, claim.Now))
}

func TestRoutineWeeklyLimitSpacesSundayCatchUpFromNextMonday(t *testing.T) {
	for _, begun := range []bool{false, true} {
		name := "legacy_sent"
		if begun {
			name = "uncertain_begun"
		}
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			f := newNotificationIntegrationFixture(t, ctx)
			scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
			sunday := time.Date(2026, time.September, 6, 9, 0, 0, 0, time.UTC)
			claim := notifications.RoutineClaim{
				RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID,
				Key: "weekly:2026-08-31", Kind: "activity", LocalDate: sunday, Now: sunday,
			}
			id, err := f.repo.ClaimRoutine(ctx, claim)
			require.NoError(t, err)
			if begun {
				require.NoError(t, f.repo.BeginRoutineSend(ctx, id, scope, sunday))
			} else {
				require.NoError(t, f.repo.CompleteRoutine(ctx, notifications.RoutineCompletion{
					ID: id, Scope: scope, Sent: true, Now: sunday,
				}))
			}
			claim.Key = "weekly:2026-09-07"
			for _, later := range []time.Time{sunday.Add(24 * time.Hour), sunday.Add(7*24*time.Hour - time.Microsecond)} {
				claim.Now, claim.LocalDate = later, later
				retry, err := f.repo.ClaimRoutine(ctx, claim)
				require.NoError(t, err)
				require.Equal(t, uuid.Nil, retry, "a new local week must still wait a full seven elapsed days")
			}
			claim.Now = sunday.Add(7 * 24 * time.Hour)
			claim.LocalDate = claim.Now
			next, err := f.repo.ClaimRoutine(ctx, claim)
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, next, "next Sunday is eligible at exactly seven days")
			require.NoError(t, f.repo.BeginRoutineSend(ctx, next, scope, claim.Now))
		})
	}
}

func TestRoutineWeeklySendStartRejectsClaimsOutsideCurrentCoverage(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	sunday := time.Date(2026, time.September, 6, 9, 0, 0, 0, time.UTC)
	claim := notifications.RoutineClaim{
		RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID,
		Key: "weekly:2026-08-31", Kind: "activity", LocalDate: sunday, Now: sunday.Add(-11 * time.Minute),
	}
	old, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	claim.Now, claim.LocalDate = sunday.Add(24*time.Hour), sunday.Add(24*time.Hour)
	claim.Key = "weekly:2026-09-07"
	monday, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, monday)
	// The prior owner begins late, after Monday's work was already claimed.
	// The send-start check must enforce spacing independently of claim checks.
	require.NoError(t, f.repo.BeginRoutineSend(ctx, old, scope, claim.Now))
	require.Error(t, f.repo.BeginRoutineSend(ctx, monday, scope, claim.Now))
}
