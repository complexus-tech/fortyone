//go:build integration

package feedbackrepository

import (
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFeedbackDigestRechecksDisabledBoardAfterSubscriptionDiscovery(t *testing.T) {
	ctx := t.Context()
	postgres := testkit.NewPostgres(t)
	f := newFeedbackIntegrationFixture(t, ctx, postgres.Pool)
	repo := New(nil, postgres.Pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	windowStart := now.AddDate(0, 0, -7)
	secondTeam, secondBoard, secondItem := uuid.New(), uuid.New(), uuid.New()
	feedbackIntegrationExec(t, ctx, postgres.Pool, `
		INSERT INTO teams(team_id,workspace_id,name,code,color)
		VALUES ($1,$2,'Enabled feedback team','FE','#000000')`, secondTeam, f.workspaceA)
	feedbackIntegrationExec(t, ctx, postgres.Pool, `INSERT INTO team_members(team_id,user_id) VALUES ($1,$2)`, secondTeam, f.actorA)
	feedbackIntegrationExec(t, ctx, postgres.Pool, `
		INSERT INTO feedback_boards(id,workspace_id,portal_id,team_id,name,slug)
		VALUES ($1,$2,$3,$4,'Enabled board','enabled-board')`, secondBoard, f.workspaceA, f.portalA, secondTeam)
	feedbackIntegrationExec(t, ctx, postgres.Pool, `
		INSERT INTO feedback_items(id,workspace_id,portal_id,board_id,contributor_id,author_id,title,slug,submission_source,created_at)
		SELECT $1,workspace_id,portal_id,$2,contributor_id,author_id,'Enabled item','enabled-item','portal',$3
		FROM feedback_items WHERE id=$4`, secondItem, secondBoard, now.Add(-time.Hour), f.itemA)
	feedbackIntegrationExec(t, ctx, postgres.Pool, `UPDATE feedback_items SET submission_source='portal',created_at=$2 WHERE id=$1`, f.itemA, now.Add(-time.Hour))
	for _, board := range []uuid.UUID{f.boardA, secondBoard} {
		_, err := repo.SetBoardReviewer(ctx, feedback.CoreBoardReviewerInput{WorkspaceID: f.workspaceA, BoardID: board, UserID: f.actorA, EmailFrequency: feedback.EmailFrequencyWeekly})
		require.NoError(t, err)
	}
	subscriptions, err := repo.ListDigestSubscriptions(ctx, f.actorA, f.workspaceA)
	require.NoError(t, err)
	require.Len(t, subscriptions, 2)
	claimID, claimed, err := repo.ClaimDigestDelivery(ctx, feedback.CoreDigestDeliveryClaim{
		WorkspaceID: f.workspaceA, RecipientID: f.actorA, LocalDate: now,
		WindowStart: windowStart, WindowEnd: now, StaleBefore: now.Add(-time.Hour),
	})
	require.NoError(t, err)
	require.True(t, claimed)

	// The off preference removes its subscription, rather than storing "off".
	_, err = repo.SetBoardReviewer(ctx, feedback.CoreBoardReviewerInput{WorkspaceID: f.workspaceA, BoardID: f.boardA, UserID: f.actorA, EmailFrequency: feedback.EmailFrequencyOff})
	require.NoError(t, err)
	items, err := repo.ListDigestItems(ctx, feedback.CoreDigestItemsQuery{
		RecipientID: f.actorA, WorkspaceID: f.workspaceA,
		BoardIDs: []uuid.UUID{f.boardA, secondBoard}, WindowStarts: []time.Time{windowStart, windowStart}, WindowEnd: now, Limit: 20,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, secondItem, items[0].ID)

	require.NoError(t, repo.CompleteDigestDelivery(ctx, feedback.CoreDigestDeliveryCompletion{
		DeliveryID: claimID, RecipientID: f.actorA, WorkspaceID: f.workspaceA,
		BoardIDs: []uuid.UUID{f.boardA, secondBoard}, DeliveredAt: now, WindowEnd: now, Status: feedback.DigestDeliverySent, ItemCount: 1,
	}))
	var disabledCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM feedback_board_subscriptions WHERE board_id=$1 AND user_id=$2`, f.boardA, f.actorA).Scan(&disabledCount))
	require.Zero(t, disabledCount, "completion cannot recreate a disabled board subscription")
	var enabledCursor *time.Time
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT last_digest_cursor_at FROM feedback_board_subscriptions WHERE board_id=$1 AND user_id=$2`, secondBoard, f.actorA).Scan(&enabledCursor))
	require.NotNil(t, enabledCursor)
	require.True(t, now.Equal(*enabledCursor))
}

func TestFeedbackDigestRechecksTeamAccessBeforeListingAndAdvancingDiscoveredBoard(t *testing.T) {
	ctx := t.Context()
	postgres := testkit.NewPostgres(t)
	f := newFeedbackIntegrationFixture(t, ctx, postgres.Pool)
	repo := New(nil, postgres.Pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	windowStart := now.AddDate(0, 0, -7)
	_, err := repo.SetBoardReviewer(ctx, feedback.CoreBoardReviewerInput{WorkspaceID: f.workspaceA, BoardID: f.boardA, UserID: f.actorA, EmailFrequency: feedback.EmailFrequencyDaily})
	require.NoError(t, err)
	feedbackIntegrationExec(t, ctx, postgres.Pool, `UPDATE feedback_items SET submission_source='portal',created_at=$2 WHERE id=$1`, f.itemA, now.Add(-time.Hour))
	subscriptions, err := repo.ListDigestSubscriptions(ctx, f.actorA, f.workspaceA)
	require.NoError(t, err)
	require.Len(t, subscriptions, 1)
	claimID, claimed, err := repo.ClaimDigestDelivery(ctx, feedback.CoreDigestDeliveryClaim{
		WorkspaceID: f.workspaceA, RecipientID: f.actorA, LocalDate: now,
		WindowStart: windowStart, WindowEnd: now, StaleBefore: now.Add(-time.Hour),
	})
	require.NoError(t, err)
	require.True(t, claimed)
	feedbackIntegrationExec(t, ctx, postgres.Pool, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.teamA, f.actorA)
	items, err := repo.ListDigestItems(ctx, feedback.CoreDigestItemsQuery{
		RecipientID: f.actorA, WorkspaceID: f.workspaceA, BoardIDs: []uuid.UUID{f.boardA},
		WindowStarts: []time.Time{windowStart}, WindowEnd: now, Limit: 20,
	})
	require.NoError(t, err)
	require.Empty(t, items)
	require.ErrorIs(t, repo.CompleteDigestDelivery(ctx, feedback.CoreDigestDeliveryCompletion{
		DeliveryID: claimID, RecipientID: f.actorA, WorkspaceID: f.workspaceA,
		BoardIDs: []uuid.UUID{f.boardA}, DeliveredAt: now, WindowEnd: now, Status: feedback.DigestDeliverySkipped,
	}), feedback.ErrNotFound)
	var cursor *time.Time
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT last_digest_cursor_at FROM feedback_board_subscriptions WHERE board_id=$1 AND user_id=$2`, f.boardA, f.actorA).Scan(&cursor))
	require.Nil(t, cursor, "revoked board access cannot consume the saved delivery window")
}
