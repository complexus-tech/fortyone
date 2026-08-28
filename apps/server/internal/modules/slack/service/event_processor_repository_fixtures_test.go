package slack

import (
	"context"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
)

const (
	testSlackCredentialSecret     = "test-slack-credential-secret"
	testSlackWebhookPayloadSecret = "test-slack-webhook-payload-secret"
)

var (
	testSlackWorkspaceID    = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testWorkspaceID         = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	testLinkedUserID        = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	testAllowedTeamID       = uuid.MustParse("88888888-8888-4888-8888-888888888888")
	testInboundReceiptID    = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	testConversationID      = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	testOutboundDeliveryID  = uuid.MustParse("66666666-6666-4666-8666-666666666666")
	testInstallGeneration   = uuid.MustParse("77777777-7777-4777-8777-777777777777")
	testSlackBotUserID      = "B1"
	testSlackBotAccessToken = "xoxb-test-token"
	testSlackAuthorizedAt   = time.Unix(1_700_000_000, 0).UTC()
)

type eventRepositoryStub struct {
	installation      slackrepository.SlackWorkspaceRecord
	workspace         slackrepository.WorkspaceRecord
	linkedUserID      *uuid.UUID
	agentSettings     slackrepository.AgentSettingsRecord
	authorizedTeamIDs []uuid.UUID
	sharedTeamIDs     []uuid.UUID

	getInstallationCalls       int
	requestedTeamIDs           []string
	findWorkspaceCalls         int
	findLinkedUserCalls        int
	credentialUpgrades         int
	legacyCredentials          []slackrepository.LegacySlackCredentialRecord
	versionedLegacyCredentials int
	deactivatedTeamIDs         []string
	deactivatedGenerations     []uuid.UUID
	installationErr            error
	recoverableUninstalls      []slackrepository.SlackUninstallRecord
	completedUninstalls        []uuid.UUID
	failedUninstalls           []uuid.UUID
}

func (r *eventRepositoryStub) GetAgentSettings(_ context.Context, _ uuid.UUID) (slackrepository.AgentSettingsRecord, error) {
	return r.agentSettings, nil
}

func (r *eventRepositoryStub) ListAuthorizedChannelTeamIDs(_ context.Context, _, _ uuid.UUID, _ string, _ uuid.UUID) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), r.authorizedTeamIDs...), nil
}

func (r *eventRepositoryStub) GetAuthorizedAssistantChannelTeamScope(
	_ context.Context,
	_, _ uuid.UUID,
	_ string,
	_ uuid.UUID,
) (slackrepository.AssistantChannelTeamScope, error) {
	return slackrepository.AssistantChannelTeamScope{
		AllowedTeamIDs: append([]uuid.UUID(nil), r.authorizedTeamIDs...),
		SharedTeamIDs:  append([]uuid.UUID(nil), r.sharedTeamIDs...),
	}, nil
}

func (r *eventRepositoryStub) ListTeamStatuses(_ context.Context, _ uuid.UUID) ([]slackrepository.StatusRecord, error) {
	return nil, nil
}

func (r *eventRepositoryStub) ListTeamMembers(_ context.Context, _ uuid.UUID) ([]slackrepository.TeamMemberRecord, error) {
	return nil, nil
}

func (r *eventRepositoryStub) FindTeamMemberByID(_ context.Context, _, _ uuid.UUID) (slackrepository.TeamMemberRecord, error) {
	return slackrepository.TeamMemberRecord{}, slackdomain.ErrNotFound
}

func (r *eventRepositoryStub) ListWorkspaceTeamsForUser(_ context.Context, _, _ uuid.UUID) ([]slackrepository.TeamRecord, error) {
	teams := make([]slackrepository.TeamRecord, 0, len(r.authorizedTeamIDs))
	for _, teamID := range r.authorizedTeamIDs {
		teams = append(teams, slackrepository.TeamRecord{ID: teamID})
	}
	return teams, nil
}

func newEventRepositoryStub() *eventRepositoryStub {
	return &eventRepositoryStub{
		installation: slackrepository.SlackWorkspaceRecord{
			ID:                testSlackWorkspaceID,
			WorkspaceID:       testWorkspaceID,
			SlackTeamID:       "T1",
			BotUserID:         &testSlackBotUserID,
			BotAccessToken:    testSlackBotAccessToken,
			CredentialVersion: 0,
			InstallGeneration: testInstallGeneration,
			AuthorizedAt:      testSlackAuthorizedAt,
			IsActive:          true,
		},
		workspace: slackrepository.WorkspaceRecord{
			ID:   testWorkspaceID,
			Slug: "acme",
			Name: "Acme",
		},
		agentSettings: slackrepository.AgentSettingsRecord{
			Guidance: "Keep answers concise.",
		},
		authorizedTeamIDs: []uuid.UUID{testAllowedTeamID},
		sharedTeamIDs:     []uuid.UUID{testAllowedTeamID},
	}
}

func (r *eventRepositoryStub) GetSlackWorkspaceByTeamID(_ context.Context, slackTeamID string) (slackrepository.SlackWorkspaceRecord, error) {
	r.getInstallationCalls++
	r.requestedTeamIDs = append(r.requestedTeamIDs, slackTeamID)
	if r.installationErr != nil {
		return slackrepository.SlackWorkspaceRecord{}, r.installationErr
	}
	return r.installation, nil
}

func (r *eventRepositoryStub) GetSlackWorkspace(_ context.Context, _ uuid.UUID) (slackrepository.SlackWorkspaceRecord, error) {
	r.getInstallationCalls++
	return r.installation, nil
}

func (r *eventRepositoryStub) FindWorkspaceByID(_ context.Context, _ uuid.UUID) (slackrepository.WorkspaceRecord, error) {
	r.findWorkspaceCalls++
	return r.workspace, nil
}

func (r *eventRepositoryStub) FindLinkedUserIDBySlackUser(_ context.Context, _ uuid.UUID, _, _ string) (*uuid.UUID, error) {
	r.findLinkedUserCalls++
	return r.linkedUserID, nil
}

func (r *eventRepositoryStub) UpgradeSlackCredential(_ context.Context, _ slackrepository.LegacySlackCredentialRecord, encrypted string, version int) error {
	r.credentialUpgrades++
	r.installation.BotAccessToken = encrypted
	r.installation.CredentialVersion = version
	if len(r.legacyCredentials) > 0 {
		r.legacyCredentials = r.legacyCredentials[1:]
	}
	return nil
}

func (r *eventRepositoryStub) ListLegacySlackCredentials(_ context.Context, limit int) ([]slackrepository.LegacySlackCredentialRecord, error) {
	if limit > len(r.legacyCredentials) {
		limit = len(r.legacyCredentials)
	}
	return append([]slackrepository.LegacySlackCredentialRecord(nil), r.legacyCredentials[:limit]...), nil
}

func (r *eventRepositoryStub) ScrubVersionedLegacySlackCredentials(_ context.Context, limit int) (int, error) {
	if limit > r.versionedLegacyCredentials {
		limit = r.versionedLegacyCredentials
	}
	r.versionedLegacyCredentials -= limit
	return limit, nil
}

func (r *eventRepositoryStub) DeactivateSlackWorkspaceByTeamID(_ context.Context, slackTeamID string, generation uuid.UUID) error {
	r.deactivatedTeamIDs = append(r.deactivatedTeamIDs, slackTeamID)
	r.deactivatedGenerations = append(r.deactivatedGenerations, generation)
	return nil
}

func (r *eventRepositoryStub) ClaimRecoverableSlackUninstalls(_ context.Context, limit int) ([]slackrepository.SlackUninstallRecord, error) {
	if limit > len(r.recoverableUninstalls) {
		limit = len(r.recoverableUninstalls)
	}
	return append([]slackrepository.SlackUninstallRecord(nil), r.recoverableUninstalls[:limit]...), nil
}

func (r *eventRepositoryStub) CompleteSlackUninstall(_ context.Context, id uuid.UUID, _ string) error {
	r.completedUninstalls = append(r.completedUninstalls, id)
	return nil
}

func (r *eventRepositoryStub) FailSlackUninstall(_ context.Context, id uuid.UUID, _ string, _ *time.Time) error {
	r.failedUninstalls = append(r.failedUninstalls, id)
	return nil
}

type eventStoryReaderStub struct {
	story        singleStory
	workspaceIDs []uuid.UUID
	references   []string
}

type eventRequestReaderStub struct {
	request      integrationRequest
	workspaceIDs []uuid.UUID
	requestIDs   []uuid.UUID
	userIDs      []uuid.UUID
}

func (s *eventRequestReaderStub) GetForUser(_ context.Context, workspaceID, requestID, userID uuid.UUID) (integrationRequest, error) {
	s.workspaceIDs = append(s.workspaceIDs, workspaceID)
	s.requestIDs = append(s.requestIDs, requestID)
	s.userIDs = append(s.userIDs, userID)
	return s.request, nil
}

func (s *eventStoryReaderStub) QueryByRef(_ context.Context, workspaceID uuid.UUID, reference string) (singleStory, error) {
	s.workspaceIDs = append(s.workspaceIDs, workspaceID)
	s.references = append(s.references, reference)
	return s.story, nil
}

type inboundCompletion struct {
	id      uuid.UUID
	status  string
	message string
}

type appendedMessage struct {
	conversationID    uuid.UUID
	externalMessageID string
	role              string
	content           string
}

type completedDelivery struct {
	id                uuid.UUID
	externalMessageID string
}

type failedDelivery struct {
	id      uuid.UUID
	message string
}

type deliveryDestination struct {
	id        uuid.UUID
	channelID string
	threadID  string
}
