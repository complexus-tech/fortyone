package integrationrequestsrepository

import (
	"os"
	"strings"
	"testing"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListParamsUsesTypedOptionalFiltersAndActorScope(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	assigneeID := uuid.New()
	params, err := listParams(workspaceID, teamID, actorID, integrationrequestdomain.ListRequestsFilter{
		Search: "  launch  ", Provider: integrationrequestdomain.ProviderSlack,
		Priority: "High", AssigneeID: &assigneeID, Page: 2, PageSize: 25,
	})
	require.NoError(t, err)
	require.Equal(t, workspaceID, params.WorkspaceID)
	require.Equal(t, teamID, params.TeamID)
	require.Equal(t, actorID, params.ActorID)
	require.Equal(t, integrationrequestdomain.StatusPending, params.RequestStatus)
	require.True(t, params.HasSearch)
	require.Equal(t, "%launch%", params.SearchPattern)
	require.True(t, params.HasProvider)
	require.True(t, params.HasPriority)
	require.True(t, params.HasAssignee)
	require.Equal(t, &assigneeID, params.AssigneeID)
	require.True(t, params.Paginated)
	require.Equal(t, int32(25), params.RowLimit)
	require.Equal(t, int32(25), params.RowOffset)
}

func TestRequestQueriesRepeatTeamMemberOrWorkspaceAdminAuthorization(t *testing.T) {
	t.Parallel()

	for _, path := range []string{"queries/requests.sql", "queries/mutations.sql", "queries/provider_threads.sql", "queries/comments.sql"} {
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		query := strings.Join(strings.Fields(string(raw)), " ")
		require.Contains(t, query, "FROM public.team_members AS request_team_member", path)
		require.Contains(t, query, "FROM public.workspace_members AS request_workspace_member", path)
		require.Contains(t, query, "request_workspace_member.role = 'admin'", path)
	}
}

func TestRequestLabelIDsRejectsNilAndDeduplicates(t *testing.T) {
	t.Parallel()

	labelID := uuid.New()
	values := []uuid.UUID{labelID, labelID}
	labels, err := requestLabelIDs(nil, &values)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{labelID}, labels)

	invalid := []uuid.UUID{uuid.Nil}
	_, err = requestLabelIDs(nil, &invalid)
	require.ErrorIs(t, err, integrationrequestdomain.ErrInvalidRequestProperty)
}
