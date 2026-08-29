package emailreplyrepository

import (
	"context"
	"errors"
	"math"
	"strconv"
	"testing"
	"time"

	emailreplydomain "github.com/complexus-tech/projects-api/internal/modules/emailreply/domain"
	emailreplysql "github.com/complexus-tech/projects-api/internal/modules/emailreply/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestNewRejectsNilPool(t *testing.T) {
	t.Parallel()

	require.Nil(t, New(nil))
}

func TestActorScopeKeepsWorkspaceAndActorBoundAcrossQueries(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	teamIDs := []uuid.UUID{uuid.New(), uuid.New()}
	queries := &contextQueryStub{
		actor: emailreplysql.GetEmailReplyActorWorkspaceRow{
			Slug: "engineering",
			Role: emailreplysql.UserRoleGuest,
		},
		teamIDs: teamIDs,
	}

	scope, found, err := newWithQueries(queries).ActorScope(context.Background(), workspaceID, userID)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, emailreplydomain.ActorScope{
		WorkspaceSlug: "engineering",
		Role:          "guest",
		TeamIDs:       teamIDs,
	}, scope)
	require.Equal(t, emailreplysql.GetEmailReplyActorWorkspaceParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	}, queries.actorParams)
	require.Equal(t, emailreplysql.ListEmailReplyActorTeamsParams{
		WorkspaceID: workspaceID,
		ActorRole:   "guest",
		UserID:      userID,
	}, queries.teamParams)

	teamIDs[0] = uuid.Nil
	require.NotEqual(t, uuid.Nil, scope.TeamIDs[0], "repository output must not alias query storage")
}

func TestActorScopeTreatsMissingActorAsUnauthorizedWithoutListingTeams(t *testing.T) {
	t.Parallel()

	queries := &contextQueryStub{actorErr: pgx.ErrNoRows}
	scope, found, err := newWithQueries(queries).ActorScope(context.Background(), uuid.New(), uuid.New())

	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, scope)
	require.Zero(t, queries.teamCalls)
}

func TestListStoryChoicesCapsAndScopesTheSecondQuery(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	statusID := uuid.New()
	firstUserID := uuid.New()
	secondUserID := uuid.New()
	queries := &contextQueryStub{
		statuses: []emailreplysql.ListEmailReplyStoryStatusesRow{{ID: statusID, Name: "In progress"}},
		assignees: []emailreplysql.ListEmailReplyStoryAssigneesRow{
			{ID: firstUserID, Name: "Ada"},
			{ID: secondUserID, Name: "Grace"},
		},
	}

	choices, err := newWithQueries(queries).ListStoryChoices(context.Background(), workspaceID, teamID, 3)

	require.NoError(t, err)
	require.Equal(t, []emailreplydomain.Choice{
		{Kind: emailreplydomain.ChoiceStoryStatus, ID: statusID, TeamID: teamID, Name: "In progress"},
		{Kind: emailreplydomain.ChoiceStoryAssignee, ID: firstUserID, TeamID: teamID, Name: "Ada"},
		{Kind: emailreplydomain.ChoiceStoryAssignee, ID: secondUserID, TeamID: teamID, Name: "Grace"},
	}, choices)
	require.Equal(t, emailreplysql.ListEmailReplyStoryStatusesParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
	}, queries.statusParams)
	require.Equal(t, emailreplysql.ListEmailReplyStoryAssigneesParams{
		TeamID:   teamID,
		RowLimit: 2,
	}, queries.assigneeParams)
}

func TestListStoryChoicesRejectsOutOfRangeLimitBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()
	if strconv.IntSize < 64 {
		t.Skip("Go int cannot represent a value above int32 on this architecture")
	}

	queries := &contextQueryStub{}
	_, err := newWithQueries(queries).ListStoryChoices(
		context.Background(),
		uuid.New(),
		uuid.New(),
		int(math.MaxInt32)+1,
	)

	require.ErrorIs(t, err, safecast.ErrOutOfRange)
	require.ErrorContains(t, err, "validate email story choice limit")
	require.Zero(t, queries.statusCalls)
}

func TestCurrentVersionNormalizesUTCAndPreservesTenantParameters(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	entityID := uuid.New()
	location := time.FixedZone("test", 2*60*60)
	queries := &contextQueryStub{storyVersion: time.Date(2026, time.August, 28, 12, 0, 0, 0, location)}

	version, found, err := newWithQueries(queries).CurrentVersion(
		context.Background(),
		workspaceID,
		emailreplydomain.ActionStoryUpdate,
		entityID,
	)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, time.Date(2026, time.August, 28, 10, 0, 0, 0, time.UTC), version)
	require.Equal(t, emailreplysql.GetStoryVersionParams{
		WorkspaceID: workspaceID,
		EntityID:    entityID,
	}, queries.storyVersionParams)
}

type contextQueryStub struct {
	emailreplysql.Querier

	actor              emailreplysql.GetEmailReplyActorWorkspaceRow
	actorErr           error
	actorParams        emailreplysql.GetEmailReplyActorWorkspaceParams
	teamIDs            []uuid.UUID
	teamParams         emailreplysql.ListEmailReplyActorTeamsParams
	teamCalls          int
	statuses           []emailreplysql.ListEmailReplyStoryStatusesRow
	statusParams       emailreplysql.ListEmailReplyStoryStatusesParams
	statusCalls        int
	assignees          []emailreplysql.ListEmailReplyStoryAssigneesRow
	assigneeParams     emailreplysql.ListEmailReplyStoryAssigneesParams
	storyVersion       time.Time
	storyVersionErr    error
	storyVersionParams emailreplysql.GetStoryVersionParams
}

func (stub *contextQueryStub) GetEmailReplyActorWorkspace(
	_ context.Context,
	params emailreplysql.GetEmailReplyActorWorkspaceParams,
) (emailreplysql.GetEmailReplyActorWorkspaceRow, error) {
	stub.actorParams = params
	return stub.actor, stub.actorErr
}

func (stub *contextQueryStub) ListEmailReplyActorTeams(
	_ context.Context,
	params emailreplysql.ListEmailReplyActorTeamsParams,
) ([]uuid.UUID, error) {
	stub.teamCalls++
	stub.teamParams = params
	return stub.teamIDs, nil
}

func (stub *contextQueryStub) ListEmailReplyStoryStatuses(
	_ context.Context,
	params emailreplysql.ListEmailReplyStoryStatusesParams,
) ([]emailreplysql.ListEmailReplyStoryStatusesRow, error) {
	stub.statusCalls++
	stub.statusParams = params
	return stub.statuses, nil
}

func (stub *contextQueryStub) ListEmailReplyStoryAssignees(
	_ context.Context,
	params emailreplysql.ListEmailReplyStoryAssigneesParams,
) ([]emailreplysql.ListEmailReplyStoryAssigneesRow, error) {
	stub.assigneeParams = params
	return stub.assignees, nil
}

func (stub *contextQueryStub) GetStoryVersion(
	_ context.Context,
	params emailreplysql.GetStoryVersionParams,
) (time.Time, error) {
	stub.storyVersionParams = params
	return stub.storyVersion, stub.storyVersionErr
}

func TestCurrentVersionMapsMissingAndQueryFailures(t *testing.T) {
	t.Parallel()

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		queries := &contextQueryStub{storyVersionErr: pgx.ErrNoRows}
		_, found, err := newWithQueries(queries).CurrentVersion(
			context.Background(), uuid.New(), emailreplydomain.ActionStoryUpdate, uuid.New(),
		)
		require.NoError(t, err)
		require.False(t, found)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		queryErr := errors.New("database unavailable")
		queries := &contextQueryStub{storyVersionErr: queryErr}
		_, found, err := newWithQueries(queries).CurrentVersion(
			context.Background(), uuid.New(), emailreplydomain.ActionStoryUpdate, uuid.New(),
		)
		require.False(t, found)
		require.ErrorIs(t, err, queryErr)
		require.ErrorContains(t, err, "read email action version")
	})
}
