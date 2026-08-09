package slack

import (
	"context"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type channelAudienceRepoStub struct {
	*mockRepo
	channels   []slackrepository.SlackChannelRecord
	access     []slackrepository.ChannelTeamAccessRecord
	authorized []uuid.UUID
	agent      slackrepository.AgentSettingsRecord
	replaced   struct {
		workspaceID      uuid.UUID
		slackWorkspaceID uuid.UUID
		slackChannelID   string
		teamIDs          []uuid.UUID
		actorID          uuid.UUID
	}
}

func (r *channelAudienceRepoStub) GetAgentSettings(context.Context, uuid.UUID) (slackrepository.AgentSettingsRecord, error) {
	return r.agent, nil
}

func (r *channelAudienceRepoStub) UpsertAgentSettings(
	_ context.Context,
	workspaceID uuid.UUID,
	input slackrepository.AgentSettingsInput,
) (slackrepository.AgentSettingsRecord, error) {
	r.agent = slackrepository.AgentSettingsRecord{
		WorkspaceID:            workspaceID,
		AssistantEnabled:       input.AssistantEnabled,
		WorkflowActionsEnabled: input.WorkflowActionsEnabled,
		Guidance:               input.Guidance,
	}
	return r.agent, nil
}

func (r *channelAudienceRepoStub) ListChannels(context.Context, uuid.UUID) ([]slackrepository.SlackChannelRecord, error) {
	return append([]slackrepository.SlackChannelRecord(nil), r.channels...), nil
}

func (r *channelAudienceRepoStub) ListChannelTeamAccess(context.Context, uuid.UUID) ([]slackrepository.ChannelTeamAccessRecord, error) {
	return append([]slackrepository.ChannelTeamAccessRecord(nil), r.access...), nil
}

func (r *channelAudienceRepoStub) ReplaceChannelTeamAccess(
	_ context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	teamIDs []uuid.UUID,
	actorID uuid.UUID,
) error {
	r.replaced.workspaceID = workspaceID
	r.replaced.slackWorkspaceID = slackWorkspaceID
	r.replaced.slackChannelID = slackChannelID
	r.replaced.teamIDs = append([]uuid.UUID(nil), teamIDs...)
	r.replaced.actorID = actorID
	return nil
}

func (r *channelAudienceRepoStub) ListAuthorizedChannelTeamIDs(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	string,
	uuid.UUID,
) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), r.authorized...), nil
}

func TestListChannelAudiencesIncludesUnconfiguredChannels(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	installationID := uuid.New()
	teamID := uuid.New()
	repo := &channelAudienceRepoStub{
		mockRepo: &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:          installationID,
			WorkspaceID: workspaceID,
			SlackTeamID: "T1",
		}},
		channels: []slackrepository.SlackChannelRecord{
			{SlackChannelID: "C1", Name: "general", IsActive: true},
			{SlackChannelID: "C2", Name: "product", IsActive: true},
		},
		access: []slackrepository.ChannelTeamAccessRecord{
			{SlackChannelID: "C2", TeamID: teamID},
		},
	}
	service := New(nil, repo, nil, nil, Config{})

	audiences, err := service.ListChannelAudiences(context.Background(), workspaceID)

	require.NoError(t, err)
	require.Len(t, audiences, 2)
	require.Equal(t, "C1", audiences[0].Channel.SlackChannelID)
	require.Empty(t, audiences[0].TeamIDs)
	require.Equal(t, "C2", audiences[1].Channel.SlackChannelID)
	require.Equal(t, []uuid.UUID{teamID}, audiences[1].TeamIDs)
}

func TestUpdateChannelAudienceBindsActiveInstallationAndActor(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	installationID := uuid.New()
	actorID := uuid.New()
	teamIDs := []uuid.UUID{uuid.New(), uuid.New()}
	repo := &channelAudienceRepoStub{mockRepo: &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:          installationID,
			WorkspaceID: workspaceID,
			SlackTeamID: "T1",
		},
	}}
	service := New(nil, repo, nil, nil, Config{})

	err := service.UpdateChannelAudience(context.Background(), workspaceID, actorID, " C1 ", teamIDs)

	require.NoError(t, err)
	require.Equal(t, workspaceID, repo.replaced.workspaceID)
	require.Equal(t, installationID, repo.replaced.slackWorkspaceID)
	require.Equal(t, "C1", repo.replaced.slackChannelID)
	require.Equal(t, teamIDs, repo.replaced.teamIDs)
	require.Equal(t, actorID, repo.replaced.actorID)
}

func TestAuthorizedChannelTeamIDsDelegatesAuthoritativeScope(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	installationID := uuid.New()
	actorID := uuid.New()
	want := []uuid.UUID{uuid.New(), uuid.New()}
	repo := &channelAudienceRepoStub{
		mockRepo:   &mockRepo{},
		authorized: want,
	}
	service := New(nil, repo, nil, nil, Config{})

	got, err := service.authorizedChannelTeamIDs(
		context.Background(),
		workspaceID,
		installationID,
		"C1",
		actorID,
	)

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestCreationDeliveryUsesCurrentActorMembershipWithoutChannelMapping(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	installationID := uuid.New()
	actorID := uuid.New()
	teamID := uuid.New()
	repo := &channelAudienceRepoStub{
		mockRepo: &mockRepo{
			teams:       []slackrepository.TeamRecord{{ID: teamID, Name: "Private team"}},
			teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
			slackWorkspace: slackrepository.SlackWorkspaceRecord{
				ID:          installationID,
				WorkspaceID: workspaceID,
				SlackTeamID: "T1",
			},
		},
		authorized: nil,
	}
	service := New(nil, repo, nil, nil, Config{})

	current, err := service.slackDeliveryAuthorizationCurrent(
		context.Background(),
		workspaceID,
		"T1",
		"C1",
		"U1",
		SlackProviderPayload{Authorization: &SlackDeliveryAuthorization{
			AllowedTeamIDs: []uuid.UUID{teamID},
			ActorUserID:    &actorID,
			Scope:          slackDeliveryAuthorizationScopeActorMembership,
		}},
	)

	require.NoError(t, err)
	require.True(t, current)

	current, err = service.slackDeliveryAuthorizationCurrent(
		context.Background(),
		workspaceID,
		"T1",
		"C1",
		"U1",
		SlackProviderPayload{Authorization: &SlackDeliveryAuthorization{
			AllowedTeamIDs: []uuid.UUID{teamID},
			ActorUserID:    &actorID,
		}},
	)

	require.NoError(t, err)
	require.False(t, current)
}

func TestUpdateAgentSettingsRoundTripsProviderNeutralConfiguration(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	repo := &channelAudienceRepoStub{mockRepo: &mockRepo{}}
	service := New(nil, repo, nil, nil, Config{})
	input := CoreSlackAgentSettings{Guidance: "Use customer-facing language."}
	want := CoreSlackAgentSettings{
		AssistantEnabled:       true,
		WorkflowActionsEnabled: true,
		Guidance:               input.Guidance,
	}

	updated, err := service.UpdateAgentSettings(context.Background(), workspaceID, input)

	require.NoError(t, err)
	require.Equal(t, want, updated)
	loaded, err := service.GetAgentSettings(context.Background(), workspaceID)
	require.NoError(t, err)
	require.Equal(t, want, loaded)
}

func TestGetAgentSettingsKeepsAssistantAndActionsAlwaysEnabled(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	repo := &channelAudienceRepoStub{
		mockRepo: &mockRepo{},
		agent: slackrepository.AgentSettingsRecord{
			WorkspaceID:            workspaceID,
			AssistantEnabled:       false,
			WorkflowActionsEnabled: false,
			Guidance:               "Keep responses concise.",
		},
	}
	service := New(nil, repo, nil, nil, Config{})

	settings, err := service.GetAgentSettings(context.Background(), workspaceID)

	require.NoError(t, err)
	require.True(t, settings.AssistantEnabled)
	require.True(t, settings.WorkflowActionsEnabled)
	require.Equal(t, "Keep responses concise.", settings.Guidance)
}
