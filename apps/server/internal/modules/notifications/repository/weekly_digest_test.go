package notificationsrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type weeklyDigestQueriesStub struct {
	notificationssql.Querier
	recipientParams notificationssql.ListWeeklyDigestRecipientsParams
	recipients      []notificationssql.ListWeeklyDigestRecipientsRow
	recipientErr    error
	statsParams     notificationssql.GetWeeklyDigestStatsParams
	stats           notificationssql.GetWeeklyDigestStatsRow
	statsErr        error
}

func (stub *weeklyDigestQueriesStub) ListWeeklyDigestRecipients(
	_ context.Context,
	params notificationssql.ListWeeklyDigestRecipientsParams,
) ([]notificationssql.ListWeeklyDigestRecipientsRow, error) {
	stub.recipientParams = params
	return stub.recipients, stub.recipientErr
}

func (stub *weeklyDigestQueriesStub) GetWeeklyDigestStats(
	_ context.Context,
	params notificationssql.GetWeeklyDigestStatsParams,
) (notificationssql.GetWeeklyDigestStatsRow, error) {
	stub.statsParams = params
	return stub.stats, stub.statsErr
}

func TestListWeeklyDigestRecipientsMapsCompositeCursorAndRows(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	workspaceID := uuid.New()
	stub := &weeklyDigestQueriesStub{recipients: []notificationssql.ListWeeklyDigestRecipientsRow{{
		UserID:        userID,
		UserEmail:     " owner@example.com ",
		UserName:      "Workspace owner",
		WorkspaceID:   workspaceID,
		WorkspaceName: "Product",
		WorkspaceSlug: " product ",
	}}}
	repository := &Repository{queries: stub}

	recipients, err := repository.ListWeeklyDigestRecipients(
		context.Background(),
		&notificationsdomain.WeeklyDigestCursor{WorkspaceID: workspaceID, UserID: userID},
		100,
	)
	require.NoError(t, err)
	require.Equal(t, notificationssql.ListWeeklyDigestRecipientsParams{
		HasCursor: true, AfterWorkspaceID: workspaceID, AfterUserID: userID, ResultLimit: 100,
	}, stub.recipientParams)
	require.Equal(t, []notificationsdomain.WeeklyDigestRecipient{{
		UserID: userID, UserEmail: "owner@example.com", UserName: "Workspace owner",
		WorkspaceID: workspaceID, WorkspaceName: "Product", WorkspaceSlug: "product",
	}}, recipients)
}

func TestListWeeklyDigestRecipientsRejectsInvalidBoundsAndRows(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		cursor *notificationsdomain.WeeklyDigestCursor
		limit  int
	}{
		"zero limit":      {limit: 0},
		"oversized limit": {limit: maximumWeeklyDigestRecipientBatchSize + 1},
		"partial cursor":  {cursor: &notificationsdomain.WeeklyDigestCursor{WorkspaceID: uuid.New()}, limit: 10},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := (&Repository{queries: &weeklyDigestQueriesStub{}}).ListWeeklyDigestRecipients(context.Background(), test.cursor, test.limit)
			require.ErrorIs(t, err, notificationsdomain.ErrInvalid)
		})
	}

	stub := &weeklyDigestQueriesStub{recipients: []notificationssql.ListWeeklyDigestRecipientsRow{{
		UserID: uuid.New(), WorkspaceID: uuid.New(), WorkspaceSlug: "product",
	}}}
	_, err := (&Repository{queries: stub}).ListWeeklyDigestRecipients(context.Background(), nil, 10)
	require.ErrorIs(t, err, errInvalidWeeklyDigestRecipient)
}

func TestGetWeeklyDigestStatsMapsUTCAsOfAndCounts(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	workspaceID := uuid.New()
	asOf := time.Date(2026, time.August, 31, 10, 15, 0, 0, time.FixedZone("test", 2*60*60))
	stub := &weeklyDigestQueriesStub{stats: notificationssql.GetWeeklyDigestStatsRow{
		UnreadNotifications: 9, UnreadPriorityNotifications: 2, OverdueStories: 3,
		DueThisWeekStories: 4, ObjectiveRisks: 1, TeamComments: 8,
	}}
	repository := &Repository{queries: stub}

	stats, err := repository.GetWeeklyDigestStats(context.Background(), notificationsdomain.WeeklyDigestStatsQuery{
		UserID: userID, WorkspaceID: workspaceID, AsOf: asOf,
	})
	require.NoError(t, err)
	require.Equal(t, asOf.UTC(), stub.statsParams.AsOf)
	require.Equal(t, userID, stub.statsParams.UserID)
	require.Equal(t, workspaceID, stub.statsParams.WorkspaceID)
	require.Equal(t, notificationsdomain.WeeklyDigestStats{
		UnreadNotifications: 9, UnreadPriorityNotifications: 2, OverdueStories: 3,
		DueThisWeekStories: 4, ObjectiveRisks: 1, TeamComments: 8,
	}, stats)
}

func TestWeeklyDigestRepositoryPropagatesTypedQueryErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("database unavailable")
	repository := &Repository{queries: &weeklyDigestQueriesStub{recipientErr: wantErr, statsErr: wantErr}}
	_, err := repository.ListWeeklyDigestRecipients(context.Background(), nil, 10)
	require.ErrorIs(t, err, wantErr)
	_, err = repository.GetWeeklyDigestStats(context.Background(), notificationsdomain.WeeklyDigestStatsQuery{
		UserID: uuid.New(), WorkspaceID: uuid.New(), AsOf: time.Now(),
	})
	require.ErrorIs(t, err, wantErr)
}
