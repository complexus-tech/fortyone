package slack

import (
	"context"
	"testing"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type channelAudienceRepoStub struct {
	*mockRepo
	channels   []slackrepository.SlackChannelRecord
	access     []slackrepository.ChannelTeamAccessRecord
	authorized []uuid.UUID
	assistant  slackrepository.AssistantChannelTeamScope
	agent      slackrepository.AgentSettingsRecord
	replaced   struct {
		workspaceID      uuid.UUID
		slackWorkspaceID uuid.UUID
		slackChannelID   string
		isConfigured     bool
		teamIDs          []uuid.UUID
		actorID          uuid.UUID
	}
}

func (r *channelAudienceRepoStub) GetAgentSettings(context.Context, uuid.UUID) (slackrepository.AgentSettingsRecord, error) {
	return r.agent, nil
}

func (r *channelAudienceRepoStub) GetAgentSettingsForAdmin(
	_ context.Context,
	_ slackdomain.WorkspaceActorQuery,
) (slackrepository.AgentSettingsRecord, error) {
	return r.agent, nil
}

func (r *channelAudienceRepoStub) UpsertAgentSettingsForAdmin(
	_ context.Context,
	command slackdomain.UpdateAgentSettingsCommand,
) (slackrepository.AgentSettingsRecord, error) {
	r.agent = slackrepository.AgentSettingsRecord{
		WorkspaceID: command.WorkspaceID,
		Guidance:    command.Guidance,
	}
	return r.agent, nil
}

func (r *channelAudienceRepoStub) ListChannelsForMember(context.Context, slackdomain.WorkspaceActorQuery) ([]slackrepository.SlackChannelRecord, error) {
	return append([]slackrepository.SlackChannelRecord(nil), r.channels...), nil
}

func (r *channelAudienceRepoStub) ListAssistantChannelTeamAccessForAdmin(context.Context, slackdomain.WorkspaceActorQuery) ([]slackrepository.ChannelTeamAccessRecord, error) {
	return append([]slackrepository.ChannelTeamAccessRecord(nil), r.access...), nil
}

func (r *channelAudienceRepoStub) ReplaceAssistantChannelTeamAccess(
	_ context.Context,
	command slackdomain.ReplaceChannelAudienceCommand,
) error {
	r.replaced.workspaceID = command.WorkspaceID
	r.replaced.slackWorkspaceID = command.InstallationID
	r.replaced.slackChannelID = command.SlackChannelID
	r.replaced.isConfigured = command.Configured
	r.replaced.teamIDs = append([]uuid.UUID(nil), command.TeamIDs...)
	r.replaced.actorID = command.ActorID
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

func (r *channelAudienceRepoStub) ListInstallationAuthorizedChannelTeamIDs(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	string,
) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), r.authorized...), nil
}

func (r *channelAudienceRepoStub) GetAuthorizedAssistantChannelTeamScope(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	string,
	uuid.UUID,
) (slackrepository.AssistantChannelTeamScope, error) {
	return slackrepository.AssistantChannelTeamScope{
		AllowedTeamIDs: append([]uuid.UUID(nil), r.assistant.AllowedTeamIDs...),
		SharedTeamIDs:  append([]uuid.UUID(nil), r.assistant.SharedTeamIDs...),
	}, nil
}

func TestListChannelAudiencesIncludesConfiguredAndUnconfiguredChannels(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	installationID := uuid.New()
	teamID := uuid.New()
	staleTeamID := uuid.New()
	repo := &channelAudienceRepoStub{
		mockRepo: &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:          installationID,
			WorkspaceID: workspaceID,
			SlackTeamID: "T1",
		}},
		channels: []slackrepository.SlackChannelRecord{
			{SlackChannelID: "C1", Name: "general", IsActive: true},
			{
				SlackChannelID:        "C2",
				Name:                  "product",
				IsActive:              true,
				IsAssistantConfigured: true,
			},
		},
		access: []slackrepository.ChannelTeamAccessRecord{
			{SlackChannelID: "C1", TeamID: staleTeamID},
			{SlackChannelID: "C2", TeamID: teamID},
		},
	}
	service := New(nil, repo, nil, nil, Config{})

	audiences, err := service.ListChannelAudiences(context.Background(), workspaceID, uuid.New())

	require.NoError(t, err)
	require.Len(t, audiences, 2)
	require.Equal(t, "C1", audiences[0].Channel.SlackChannelID)
	require.False(t, audiences[0].IsConfigured)
	require.Equal(t, []uuid.UUID{staleTeamID}, audiences[0].TeamIDs)
	require.Equal(t, "C2", audiences[1].Channel.SlackChannelID)
	require.True(t, audiences[1].IsConfigured)
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

	err := service.UpdateChannelAudience(context.Background(), workspaceID, actorID, " C1 ", true, teamIDs)

	require.NoError(t, err)
	require.Equal(t, workspaceID, repo.replaced.workspaceID)
	require.Equal(t, installationID, repo.replaced.slackWorkspaceID)
	require.Equal(t, "C1", repo.replaced.slackChannelID)
	require.True(t, repo.replaced.isConfigured)
	require.Equal(t, teamIDs, repo.replaced.teamIDs)
	require.Equal(t, actorID, repo.replaced.actorID)
}

func TestUpdateChannelAudiencePersistsConfiguredPersonalOnlyChannel(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	installationID := uuid.New()
	actorID := uuid.New()
	repo := &channelAudienceRepoStub{mockRepo: &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:          installationID,
			WorkspaceID: workspaceID,
			SlackTeamID: "T1",
		},
	}}
	service := New(nil, repo, nil, nil, Config{})

	err := service.UpdateChannelAudience(context.Background(), workspaceID, actorID, "C1", true, nil)

	require.NoError(t, err)
	require.True(t, repo.replaced.isConfigured)
	require.Empty(t, repo.replaced.teamIDs)
}

func TestUpdateChannelAudienceDelegatesUnconfigureWithoutDiscardingTeamState(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	installationID := uuid.New()
	actorID := uuid.New()
	repo := &channelAudienceRepoStub{mockRepo: &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:          installationID,
			WorkspaceID: workspaceID,
			SlackTeamID: "T1",
		},
	}}
	service := New(nil, repo, nil, nil, Config{})

	teamIDs := []uuid.UUID{uuid.New()}
	err := service.UpdateChannelAudience(
		context.Background(),
		workspaceID,
		actorID,
		"C1",
		false,
		teamIDs,
	)

	require.NoError(t, err)
	require.False(t, repo.replaced.isConfigured)
	require.Equal(t, teamIDs, repo.replaced.teamIDs)
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

func TestAuthorizedAssistantChannelTeamScopeDelegatesSafeScope(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	installationID := uuid.New()
	actorID := uuid.New()
	want := slackrepository.AssistantChannelTeamScope{
		AllowedTeamIDs: []uuid.UUID{uuid.New(), uuid.New()},
		SharedTeamIDs:  []uuid.UUID{uuid.New()},
	}
	repo := &channelAudienceRepoStub{
		mockRepo:  &mockRepo{},
		assistant: want,
	}
	service := New(nil, repo, nil, nil, Config{})

	got, err := service.authorizedAssistantChannelTeamScope(
		context.Background(),
		workspaceID,
		installationID,
		" C1 ",
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
	actorID := uuid.New()
	repo := &channelAudienceRepoStub{mockRepo: &mockRepo{}}
	service := New(nil, repo, nil, nil, Config{})
	input := CoreSlackAgentSettings{Guidance: "Use customer-facing language."}
	want := CoreSlackAgentSettings{Guidance: input.Guidance}

	updated, err := service.UpdateAgentSettings(context.Background(), workspaceID, actorID, input)

	require.NoError(t, err)
	require.Equal(t, want, updated)
	loaded, err := service.GetAgentSettings(context.Background(), workspaceID, actorID)
	require.NoError(t, err)
	require.Equal(t, want, loaded)
}

func TestGetAgentSettingsReturnsGuidance(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actorID := uuid.New()
	repo := &channelAudienceRepoStub{
		mockRepo: &mockRepo{},
		agent: slackrepository.AgentSettingsRecord{
			WorkspaceID: workspaceID,
			Guidance:    "Keep responses concise.",
		},
	}
	service := New(nil, repo, nil, nil, Config{})

	settings, err := service.GetAgentSettings(context.Background(), workspaceID, actorID)

	require.NoError(t, err)
	require.Equal(t, "Keep responses concise.", settings.Guidance)
}
