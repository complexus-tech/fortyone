package apiv1http

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

type collaborationServiceStub struct {
	teamErr error

	labels     []labels.CoreLabel
	states     []states.CoreState
	sprints    []sprints.CoreSprint
	objectives []objectives.CoreObjective
	keyResults keyresults.CoreKeyResultListResponse
	comments   []stories.CoreComment
	comment    stories.CoreComment

	teamCalls        int
	labelCalls       int
	stateCalls       int
	sprintCalls      int
	objectiveCalls   int
	keyResultCalls   int
	commentListCalls int
	commentGetCalls  int

	labelFilter     labels.LabelFilters
	sprintQuery     sprintdomain.ListQuery
	objectiveQuery  objectivesdomain.ListQuery
	keyResultFilter keyresultsdomain.Filters
	commentPage     int
	commentPageSize int
}

func (stub *collaborationServiceStub) List(context.Context, uuid.UUID, uuid.UUID, ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error) {
	return []teams.CoreTeam{}, nil
}

func (stub *collaborationServiceStub) GetByID(_ context.Context, teamID, workspaceID, _ uuid.UUID) (teams.CoreTeam, error) {
	stub.teamCalls++
	if stub.teamErr != nil {
		return teams.CoreTeam{}, stub.teamErr
	}
	return teams.CoreTeam{ID: teamID, Workspace: workspaceID}, nil
}

func (stub *collaborationServiceStub) GetLabels(_ context.Context, _, _ uuid.UUID, filter labels.LabelFilters) ([]labels.CoreLabel, error) {
	stub.labelCalls++
	stub.labelFilter = filter
	return append([]labels.CoreLabel(nil), stub.labels...), nil
}

func (stub *collaborationServiceStub) TeamListForMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]states.CoreState, error) {
	stub.stateCalls++
	return append([]states.CoreState(nil), stub.states...), nil
}

func (stub *collaborationServiceStub) ListQuery(_ context.Context, query sprintdomain.ListQuery) ([]sprints.CoreSprint, error) {
	stub.sprintCalls++
	stub.sprintQuery = query
	return append([]sprints.CoreSprint(nil), stub.sprints...), nil
}

func (stub *collaborationServiceStub) ListIntent(_ context.Context, query objectivesdomain.ListQuery) ([]objectives.CoreObjective, error) {
	stub.objectiveCalls++
	stub.objectiveQuery = query
	return append([]objectives.CoreObjective(nil), stub.objectives...), nil
}

func (stub *collaborationServiceStub) ListPaginated(_ context.Context, filter keyresultsdomain.Filters) (keyresults.CoreKeyResultListResponse, error) {
	stub.keyResultCalls++
	stub.keyResultFilter = filter
	return stub.keyResults, nil
}

func (stub *collaborationServiceStub) GetComments(_ context.Context, _, _ uuid.UUID, page, pageSize int) ([]stories.CoreComment, bool, error) {
	stub.commentListCalls++
	stub.commentPage = page
	stub.commentPageSize = pageSize
	return append([]stories.CoreComment(nil), stub.comments...), false, nil
}

func (stub *collaborationServiceStub) GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (stories.CoreComment, error) {
	stub.commentGetCalls++
	if stub.comment.ID == uuid.Nil {
		return stories.CoreComment{}, stories.ErrNotFound
	}
	return stub.comment, nil
}

func (stub *collaborationServiceStub) totalCalls() int {
	return stub.teamCalls + stub.labelCalls + stub.stateCalls + stub.sprintCalls +
		stub.objectiveCalls + stub.keyResultCalls + stub.commentListCalls + stub.commentGetCalls
}

func TestPublicAPICollaborationReadsReturnTypedTenantScopedResources(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, storyID, objectiveID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	commentID, replyID := uuid.New(), uuid.New()
	stub := &collaborationServiceStub{
		labels: []labels.CoreLabel{{
			ID: uuid.New(), WorkspaceID: workspaceID, TeamID: &teamID,
			Name: "API", Color: "#123456", CreatedAt: now, UpdatedAt: now,
		}},
		states: []states.CoreState{{
			ID: uuid.New(), Workspace: workspaceID, Team: teamID, Name: "In Progress",
			Category: "started", OrderIndex: 30, Color: "#abcdef", CreatedAt: now, UpdatedAt: now,
		}},
		sprints: []sprints.CoreSprint{{
			ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Name: "Sprint 42",
			StartDate: now, EndDate: now.AddDate(0, 0, 14), CreatedAt: now, UpdatedAt: now,
		}},
		objectives: []objectives.CoreObjective{{
			ID: objectiveID, Workspace: workspaceID, Team: teamID, SequenceID: 4,
			Name: "Ship API", Status: uuid.New(), Color: "#456789",
			ScheduleStatus: objectivesdomain.ScheduleStatusNoSchedule,
			CreatedBy:      uuid.New(), CreatedAt: now, UpdatedAt: now,
		}},
		keyResults: keyresults.CoreKeyResultListResponse{KeyResults: []keyresults.CoreKeyResultWithObjective{{
			CoreKeyResult: keyresults.CoreKeyResult{
				ID: uuid.New(), ObjectiveID: objectiveID, SequenceID: 2, Name: "Five integrations",
				MeasurementType: "number", TargetValue: 5, Contributors: []uuid.UUID{},
				CreatedBy: uuid.New(), CreatedAt: now, UpdatedAt: now,
			},
			ObjectiveName: "Ship API", ObjectiveID: objectiveID, TeamID: teamID, WorkspaceID: workspaceID,
		}}},
		comments: []stories.CoreComment{{
			ID: commentID, StoryID: storyID, UserID: uuid.New(), Comment: "Top level",
			CreatedAt: now, UpdatedAt: now, SubComments: []stories.CoreComment{{
				ID: replyID, StoryID: storyID, Parent: &commentID, UserID: uuid.New(), Comment: "Reply",
				CreatedAt: now, UpdatedAt: now,
			}},
		}},
	}
	stub.comment = stub.comments[0]
	actor := testMachineActor(
		t, platformauth.PrincipalPersonalToken, workspaceID,
		platformauth.ScopeLabelsRead, platformauth.ScopeStoriesRead, platformauth.ScopeSprintsRead,
		platformauth.ScopeObjectivesRead, platformauth.ScopeCommentsRead,
	)
	app := publicCollaborationTestApp(t, actor, stub)

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "labels", path: "/api/v1/workspaces/" + workspaceID.String() + "/labels?teamId=" + teamID.String(), want: `"name":"API"`},
		{name: "workflow states", path: "/api/v1/workspaces/" + workspaceID.String() + "/states?teamId=" + teamID.String(), want: `"category":"started"`},
		{name: "sprints", path: "/api/v1/workspaces/" + workspaceID.String() + "/sprints?teamId=" + teamID.String(), want: `"name":"Sprint 42"`},
		{name: "objectives", path: "/api/v1/workspaces/" + workspaceID.String() + "/objectives?teamId=" + teamID.String(), want: `"name":"Ship API"`},
		{name: "key results", path: "/api/v1/workspaces/" + workspaceID.String() + "/key-results?teamId=" + teamID.String() + "&objectiveId=" + objectiveID.String(), want: `"name":"Five integrations"`},
		{name: "comments", path: "/api/v1/workspaces/" + workspaceID.String() + "/stories/" + storyID.String() + "/comments", want: replyID.String()},
		{name: "comment", path: "/api/v1/workspaces/" + workspaceID.String() + "/stories/" + storyID.String() + "/comments/" + commentID.String(), want: `"content":"Top level"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request.Header.Set("Authorization", "Bearer valid-machine-token")
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Body.String(), test.want)
			require.NotContains(t, recorder.Body.String(), "valid-machine-token")
		})
	}
	require.NotNil(t, stub.labelFilter.TeamID)
	require.Equal(t, teamID, *stub.labelFilter.TeamID)
	require.Equal(t, teamID, *stub.sprintQuery.Filter.TeamID)
	require.Equal(t, teamID, *stub.objectiveQuery.TeamID)
	require.Equal(t, []uuid.UUID{teamID}, stub.keyResultFilter.TeamIDs)
	require.Equal(t, []uuid.UUID{objectiveID}, stub.keyResultFilter.ObjectiveIDs)
	require.Equal(t, 1, stub.commentPage)
	require.Equal(t, defaultPageLimit, stub.commentPageSize)
}

func TestPublicAPICollaborationReadsFailBeforeUnsafeProductCalls(t *testing.T) {
	t.Parallel()

	workspaceID, allowedTeamID, deniedTeamID, storyID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name     string
		actor    platformauth.Actor
		path     string
		wantCode string
	}{
		{
			name:     "missing label scope",
			actor:    testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeStoriesRead),
			path:     "/api/v1/workspaces/" + workspaceID.String() + "/labels?teamId=" + allowedTeamID.String(),
			wantCode: "access_denied",
		},
		{
			name:     "comment requires owning story scope",
			actor:    testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeCommentsRead),
			path:     "/api/v1/workspaces/" + workspaceID.String() + "/stories/" + storyID.String() + "/comments",
			wantCode: "access_denied",
		},
		{
			name:     "service account is not silently treated as a user",
			actor:    testMachineActor(t, platformauth.PrincipalServiceAccount, workspaceID, platformauth.ScopeSprintsRead),
			path:     "/api/v1/workspaces/" + workspaceID.String() + "/sprints",
			wantCode: "principal_not_supported",
		},
		{
			name:     "credential team restriction",
			actor:    testRestrictedMachineActor(t, workspaceID, []uuid.UUID{allowedTeamID}, platformauth.ScopeLabelsRead),
			path:     "/api/v1/workspaces/" + workspaceID.String() + "/labels?teamId=" + deniedTeamID.String(),
			wantCode: "team_access_denied",
		},
		{
			name:     "multi-team credential requires stable filter",
			actor:    testRestrictedMachineActor(t, workspaceID, []uuid.UUID{allowedTeamID, deniedTeamID}, platformauth.ScopeObjectivesRead),
			path:     "/api/v1/workspaces/" + workspaceID.String() + "/objectives",
			wantCode: "team_filter_required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &collaborationServiceStub{}
			app := publicCollaborationTestApp(t, test.actor, stub)
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request.Header.Set("Authorization", "Bearer valid-machine-token")
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			require.Contains(t, []int{http.StatusBadRequest, http.StatusForbidden}, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Body.String(), `"code":"`+test.wantCode+`"`)
			require.Zero(t, stub.totalCalls())
		})
	}
}

func TestPublicAPILabelCursorIsBoundToTeamAndMembership(t *testing.T) {
	t.Parallel()

	workspaceID, firstTeamID, secondTeamID := uuid.New(), uuid.New(), uuid.New()
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	stub := &collaborationServiceStub{labels: []labels.CoreLabel{
		{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: &firstTeamID, Name: "one", CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: &firstTeamID, Name: "two", CreatedAt: now, UpdatedAt: now},
	}}
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeLabelsRead)
	app := publicCollaborationTestApp(t, actor, stub)
	first := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+workspaceID.String()+"/labels?limit=1&teamId="+firstTeamID.String(), nil)
	first.Header.Set("Authorization", "Bearer valid-machine-token")
	firstRecorder := httptest.NewRecorder()
	app.ServeHTTP(firstRecorder, first)
	require.Equal(t, http.StatusOK, firstRecorder.Code, firstRecorder.Body.String())
	cursor := decodeNextCursor(t, firstRecorder.Body.Bytes())

	second := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+workspaceID.String()+"/labels?limit=1&teamId="+secondTeamID.String()+"&cursor="+cursor, nil)
	second.Header.Set("Authorization", "Bearer valid-machine-token")
	secondRecorder := httptest.NewRecorder()
	app.ServeHTTP(secondRecorder, second)

	require.Equal(t, http.StatusBadRequest, secondRecorder.Code, secondRecorder.Body.String())
	require.Contains(t, secondRecorder.Body.String(), `"code":"invalid_cursor"`)
	require.Equal(t, 1, stub.labelCalls)
	// Membership is checked before cursor parsing so the endpoint never lets a
	// cursor become a team-existence oracle.
	require.Equal(t, 2, stub.teamCalls)
}

func TestPublicAPILabelReadRequiresCurrentTeamMembership(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	stub := &collaborationServiceStub{teamErr: teams.ErrTeamNotFound}
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeLabelsRead)
	app := publicCollaborationTestApp(t, actor, stub)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+workspaceID.String()+"/labels?teamId="+teamID.String(), nil)
	request.Header.Set("Authorization", "Bearer valid-machine-token")
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"code":"team_access_denied"`)
	require.Equal(t, 1, stub.teamCalls)
	require.Zero(t, stub.labelCalls)
}

func publicCollaborationTestApp(t *testing.T, actor platformauth.Actor, services *collaborationServiceStub) *web.App {
	t.Helper()
	log := logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, strings.ReplaceAll(t.Name(), "/", "-"))
	app := web.New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("public-collaboration-api-test"))
	app.SetLogger(log)
	Routes(Config{
		Log: log, SecretKey: strings.Repeat("test-public-api-secret", 2), Cache: &rateLimitStoreStub{count: 1},
		DeveloperCredentials: &credentialResolverStub{actor: actor}, Workspaces: &workspaceReaderStub{},
		Teams: services, Stories: storyReaderStub{}, StoryComments: services,
		Labels: services, States: services, Sprints: services, Objectives: services, KeyResults: services,
		Idempotency: idempotencyManagerStub{}, Webhooks: &webhookManagerStub{},
	}, app)
	return app
}

func testRestrictedMachineActor(
	t *testing.T,
	workspaceID uuid.UUID,
	teamIDs []uuid.UUID,
	scopes ...platformauth.Scope,
) platformauth.Actor {
	t.Helper()
	teamAccess, err := platformauth.RestrictedTeamAccess(teamIDs...)
	require.NoError(t, err)
	actor, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalPersonalToken, uuid.New(),
		platformauth.MustScopeSet(scopes...), teamAccess,
	)
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	return actor
}

func decodeNextCursor(t *testing.T, body []byte) string {
	t.Helper()
	marker := `"nextCursor":"`
	start := bytes.Index(body, []byte(marker))
	require.NotEqual(t, -1, start, string(body))
	start += len(marker)
	end := bytes.IndexByte(body[start:], '"')
	require.NotEqual(t, -1, end, string(body))
	return string(body[start : start+end])
}
