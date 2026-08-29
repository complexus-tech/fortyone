package messagingcontext

import (
	"context"
	"errors"
	"testing"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	testWorkspaceID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testUserID      = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testAllowedTeam = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testHiddenTeam  = uuid.MustParse("44444444-4444-4444-4444-444444444444")
)

func TestProviderLoadsPersistedLocalTimeIdentityTerminologyAndAuthorizedTeamOrder(t *testing.T) {
	t.Parallel()

	provider := mustProvider(t,
		&userReaderStub{user: users.CoreUser{ID: testUserID, FullName: "Joseph Mukorivo", Username: "joseph", Timezone: "Africa/Harare"}},
		&workspaceReaderStub{
			workspace: workspaces.CoreWorkspace{ID: testWorkspaceID, Name: "FortyOne", Slug: "fortyone", UserRole: "admin"},
			settings:  workspaces.CoreWorkspaceSettings{StoryTerm: "work item", SprintTerm: "cycle", ObjectiveTerm: "focus area", KeyResultTerm: "result"},
		},
		&teamReaderStub{teams: []teams.CoreTeam{
			{ID: testHiddenTeam, Workspace: testWorkspaceID, Name: "Hidden", Code: "hid"},
			{ID: testAllowedTeam, Workspace: testWorkspaceID, Name: "Web", Code: "web"},
			{ID: uuid.New(), Workspace: uuid.New(), Name: "Foreign", Code: "bad"},
		}},
	)

	runtime, err := provider.Load(
		context.Background(),
		testWorkspaceID,
		testUserID,
		[]uuid.UUID{testAllowedTeam},
		messaging.RuntimeSurfaceContext{Provider: "slack", Kind: messaging.RuntimeSurfaceThread},
		time.Date(2026, time.August, 9, 5, 30, 45, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Equal(t, "Joseph Mukorivo", runtime.Actor.DisplayName)
	require.Equal(t, "joseph", runtime.Actor.Username)
	require.Equal(t, "FortyOne", runtime.Workspace.Name)
	require.Equal(t, "admin", runtime.Workspace.Role)
	require.Equal(t, "Africa/Harare", runtime.LocalTime.Location().String())
	require.Equal(t, 7, runtime.LocalTime.Hour())
	require.Equal(t, messaging.RuntimeTerm{Singular: "work item", Plural: "work items"}, runtime.Terminology.Story)
	require.Equal(t, messaging.RuntimeTerm{Singular: "focus area", Plural: "focus areas"}, runtime.Terminology.Objective)
	require.Equal(t, []messaging.RuntimeTeamHint{{Name: "Web", Code: "web"}}, runtime.TeamHints)
	require.Equal(t, messaging.RuntimeSurfaceThread, runtime.Surface.Kind)
}

func TestProviderUsesUTCForBlankOrInvalidPersistedTimezone(t *testing.T) {
	t.Parallel()

	for _, timezone := range []string{"", "not/a-timezone"} {
		t.Run(timezone, func(t *testing.T) {
			provider := mustProvider(t,
				&userReaderStub{user: users.CoreUser{ID: testUserID, Timezone: timezone}},
				&workspaceReaderStub{workspace: workspaces.CoreWorkspace{ID: testWorkspaceID}, settings: workspaces.CoreWorkspaceSettings{}},
				&teamReaderStub{},
			)
			runtime, err := provider.Load(context.Background(), testWorkspaceID, testUserID, nil, messaging.RuntimeSurfaceContext{}, time.Date(2026, 8, 9, 1, 2, 3, 0, time.UTC))
			require.NoError(t, err)
			require.Equal(t, time.UTC, runtime.LocalTime.Location())
			require.Equal(t, messaging.RuntimeTerm{Singular: "story", Plural: "stories"}, runtime.Terminology.Story)
		})
	}
}

func TestProviderUsesDefaultTerminologyWhenLegacySettingsAreMissing(t *testing.T) {
	t.Parallel()

	provider := mustProvider(t,
		&userReaderStub{user: users.CoreUser{ID: testUserID}},
		&workspaceReaderStub{workspace: workspaces.CoreWorkspace{ID: testWorkspaceID}, settingsErr: workspaces.ErrNotFound},
		&teamReaderStub{},
	)
	runtime, err := provider.Load(context.Background(), testWorkspaceID, testUserID, nil, messaging.RuntimeSurfaceContext{}, time.Now())
	require.NoError(t, err)
	require.Equal(t, messaging.RuntimeTerm{Singular: "story", Plural: "stories"}, runtime.Terminology.Story)
	require.Equal(t, messaging.RuntimeTerm{Singular: "key result", Plural: "key results"}, runtime.Terminology.KeyResult)
}

func TestProviderPreservesExplicitEmptyAudience(t *testing.T) {
	t.Parallel()

	provider := mustProvider(t,
		&userReaderStub{user: users.CoreUser{ID: testUserID}},
		&workspaceReaderStub{workspace: workspaces.CoreWorkspace{ID: testWorkspaceID}, settings: workspaces.CoreWorkspaceSettings{}},
		&teamReaderStub{teams: []teams.CoreTeam{{ID: testAllowedTeam, Workspace: testWorkspaceID, Name: "Web"}}},
	)
	runtime, err := provider.Load(context.Background(), testWorkspaceID, testUserID, []uuid.UUID{}, messaging.RuntimeSurfaceContext{}, time.Now())
	require.NoError(t, err)
	require.Empty(t, runtime.TeamHints)
}

func TestProviderPropagatesAuthoritativeLookupFailure(t *testing.T) {
	t.Parallel()

	provider := mustProvider(t,
		&userReaderStub{err: errors.New("database unavailable")},
		&workspaceReaderStub{},
		&teamReaderStub{},
	)
	_, err := provider.Load(context.Background(), testWorkspaceID, testUserID, nil, messaging.RuntimeSurfaceContext{}, time.Now())
	require.ErrorContains(t, err, "load messaging actor context")
}

func mustProvider(t *testing.T, userReader UserReader, workspaceReader WorkspaceReader, teamReader TeamReader) *Provider {
	t.Helper()
	provider, err := New(userReader, workspaceReader, teamReader)
	require.NoError(t, err)
	return provider
}

type userReaderStub struct {
	user users.CoreUser
	err  error
}

func (s *userReaderStub) GetUser(context.Context, uuid.UUID) (users.CoreUser, error) {
	return s.user, s.err
}

type workspaceReaderStub struct {
	workspace   workspaces.CoreWorkspace
	settings    workspaces.CoreWorkspaceSettings
	err         error
	settingsErr error
}

func (s *workspaceReaderStub) Get(context.Context, uuid.UUID, uuid.UUID) (workspaces.CoreWorkspace, error) {
	return s.workspace, s.err
}

func (s *workspaceReaderStub) GetWorkspaceSettings(context.Context, uuid.UUID) (workspaces.CoreWorkspaceSettings, error) {
	return s.settings, s.settingsErr
}

type teamReaderStub struct {
	teams []teams.CoreTeam
	err   error
}

func (s *teamReaderStub) List(context.Context, uuid.UUID, uuid.UUID, ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error) {
	return s.teams, s.err
}
