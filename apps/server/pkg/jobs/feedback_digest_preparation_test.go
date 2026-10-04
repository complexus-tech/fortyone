package jobs

import (
	"context"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type feedbackPreparationStore struct {
	recordingFeedbackDigestStore
	items []feedback.CoreDigestItem
	query feedback.CoreDigestItemsQuery
}

func (s *feedbackPreparationStore) ListDigestItems(_ context.Context, query feedback.CoreDigestItemsQuery) ([]feedback.CoreDigestItem, error) {
	s.query = query
	return s.items, nil
}

func TestPrepareFeedbackDigestKeepsBoardCursorsPendingUntilSharedDelivery(t *testing.T) {
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	recipient := feedback.CoreDigestRecipient{UserID: uuid.New(), WorkspaceID: uuid.New(), WorkspaceSlug: "product", Timezone: "Africa/Harare"}
	board := uuid.New()
	store := &feedbackPreparationStore{
		recordingFeedbackDigestStore: recordingFeedbackDigestStore{
			deliveryID: uuid.New(), subscriptions: []feedback.CoreDigestSubscription{{BoardID: board, TeamID: uuid.New(), EmailFrequency: "weekly", CreatedAt: now.AddDate(0, 0, -8)}},
		},
		items: []feedback.CoreDigestItem{{ID: uuid.New(), TeamID: uuid.New(), Title: "Improve search", Description: "<p>Please add filters &amp; sorting.</p>", AuthorName: "Sam", TeamName: "Product", Status: "pending", TotalCount: 1}},
	}
	batch, err := PrepareFeedbackDigest(t.Context(), store, recipient, now)
	require.NoError(t, err)
	require.NotNil(t, batch)
	require.Zero(t, store.completion.DeliveryID)
	require.Equal(t, []uuid.UUID{board}, batch.Completion.BoardIDs)
	require.Equal(t, now.Add(-5*time.Minute), batch.Completion.WindowEnd)
	require.Equal(t, "2026-10-05", store.claim.LocalDate.Format(time.DateOnly))
	require.Equal(t, "Please add filters & sorting.", batch.Section.Rows[0].Detail)
	require.Equal(t, "feedback", batch.Targets[0].Kind)
}

func TestPrepareFeedbackDigestDoesNotClaimBeforeRecipientsScheduledHour(t *testing.T) {
	now := time.Date(2026, 10, 5, 6, 59, 0, 0, time.UTC)
	store := &feedbackPreparationStore{recordingFeedbackDigestStore: recordingFeedbackDigestStore{
		deliveryID: uuid.New(), subscriptions: []feedback.CoreDigestSubscription{{BoardID: uuid.New(), EmailFrequency: "weekly", CreatedAt: now.AddDate(0, 0, -8)}},
	}}
	batch, err := PrepareFeedbackDigest(t.Context(), store, feedback.CoreDigestRecipient{UserID: uuid.New(), WorkspaceID: uuid.New(), Timezone: "Africa/Harare"}, now)
	require.NoError(t, err)
	require.Nil(t, batch)
	require.Zero(t, store.claim.WorkspaceID)
}

func TestPrepareFeedbackDigestCombinesBoardWindowsWithoutLosingLastWeeksLateItems(t *testing.T) {
	previousDelivery := time.Date(2026, time.October, 5, 8, 0, 0, 0, time.UTC)
	dailyCursor := previousDelivery.Add(-feedbackDigestConsistencyLag)
	weeklyCursor := dailyCursor.AddDate(0, 0, -7)
	for _, now := range []time.Time{
		previousDelivery.AddDate(0, 0, 7), // Scheduled Monday after local 09:00.
		previousDelivery.AddDate(0, 0, 9), // Catch up an unsent week on Wednesday.
	} {
		t.Run(now.Weekday().String(), func(t *testing.T) {
			recipient := feedback.CoreDigestRecipient{UserID: uuid.New(), WorkspaceID: uuid.New(), WorkspaceSlug: "product", Timezone: "Africa/Harare"}
			dailyBoard, weeklyBoard := uuid.New(), uuid.New()
			lateItem := feedback.CoreDigestItem{
				ID: uuid.New(), TeamID: uuid.New(), Title: "Submitted after the previous cutoff",
				CreatedAt: previousDelivery.Add(-time.Minute), TotalCount: 1,
			}
			store := &feedbackPreparationStore{
				recordingFeedbackDigestStore: recordingFeedbackDigestStore{
					deliveryID: uuid.New(),
					subscriptions: []feedback.CoreDigestSubscription{
						{BoardID: dailyBoard, EmailFrequency: "daily", CreatedAt: weeklyCursor, LastDigestSentAt: &previousDelivery, LastDigestCursorAt: &dailyCursor},
						{BoardID: weeklyBoard, EmailFrequency: "weekly", CreatedAt: weeklyCursor, LastDigestSentAt: &previousDelivery, LastDigestCursorAt: &weeklyCursor},
						{BoardID: uuid.New(), EmailFrequency: "off", CreatedAt: weeklyCursor},
					},
				},
				items: []feedback.CoreDigestItem{lateItem},
			}

			batch, err := PrepareFeedbackDigest(t.Context(), store, recipient, now)
			require.NoError(t, err)
			require.NotNil(t, batch)
			require.Equal(t, []uuid.UUID{dailyBoard, weeklyBoard}, store.query.BoardIDs)
			require.Equal(t, []time.Time{dailyCursor, weeklyCursor}, store.query.WindowStarts)
			require.Equal(t, now.Add(-feedbackDigestConsistencyLag), store.query.WindowEnd)
			require.Equal(t, weeklyCursor, store.claim.WindowStart)
			require.Equal(t, now.Format(time.DateOnly), store.claim.LocalDate.Format(time.DateOnly))
			require.Equal(t, []uuid.UUID{dailyBoard, weeklyBoard}, batch.Completion.BoardIDs)
			require.Equal(t, feedback.DigestDeliverySent, batch.Completion.Status)
			require.Equal(t, lateItem.ID, batch.Targets[0].ID)
			require.Zero(t, store.completion.DeliveryID, "preparation leaves cursors pending until the combined send commits")
		})
	}
}
