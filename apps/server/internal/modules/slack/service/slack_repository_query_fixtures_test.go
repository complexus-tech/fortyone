package slack

import (
	"context"
	"errors"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

type mockRepo struct {
	workspaceRole           authorization.WorkspaceRole
	workspaceRoleErr        error
	workspace               slackrepository.WorkspaceRecord
	team                    slackrepository.TeamRecord
	teams                   []slackrepository.TeamRecord
	statuses                []slackrepository.StatusRecord
	statusesByTeam          map[uuid.UUID][]slackrepository.StatusRecord
	teamMembers             []slackrepository.TeamMemberRecord
	membersByTeam           map[uuid.UUID][]slackrepository.TeamMemberRecord
	labels                  []slackrepository.LabelRecord
	labelsByTeam            map[uuid.UUID][]slackrepository.LabelRecord
	objectives              []slackrepository.ObjectiveRecord
	objectivesByTeam        map[uuid.UUID][]slackrepository.ObjectiveRecord
	workspaceMembers        []slackrepository.WorkspaceMemberRecord
	slackUserLinks          map[string]uuid.UUID
	slackWorkspace          slackrepository.SlackWorkspaceRecord
	lastOAuthInstall        slackrepository.OAuthInstallPayload
	lastOAuthUserID         uuid.UUID
	lastRequestLog          slackrepository.SlackRequestLogInsert
	upsertChannels          int
	lastChannelWorkspaceID  uuid.UUID
	lastChannelInstallID    uuid.UUID
	lastChannels            []slackrepository.SlackChannelPayload
	err                     error
	upsertSlackWorkspaceErr error
	getSlackWorkspaceErr    error
	getSlackTeamErr         error
	enqueueUninstallErr     error
	disconnected            bool
	uninstalls              map[uuid.UUID]slackrepository.SlackUninstallRecord
	uninstallInputs         []slackrepository.SlackUninstallInput
	completedUninstalls     []uuid.UUID
	failedUninstalls        []uuid.UUID
	agentSettings           slackrepository.AgentSettingsRecord
	authorizedTeamIDs       []uuid.UUID
	lastStoryLink           struct {
		storyID   uuid.UUID
		sourceKey string
		title     string
		url       string
	}
}

func (m *mockRepo) GetWorkspaceRole(
	_ context.Context,
	_, _ uuid.UUID,
) (authorization.WorkspaceRole, error) {
	if m.workspaceRoleErr != nil {
		return "", m.workspaceRoleErr
	}
	if m.workspaceRole == "" {
		return authorization.WorkspaceRoleAdmin, nil
	}
	return m.workspaceRole, nil
}

func (m *mockRepo) HasSlackUserOnboardingReceipt(_ context.Context, _ uuid.UUID, _, _ string) (bool, error) {
	// Existing handler tests are not first-use tests. Default to a completed
	// receipt so adding an outbound runtime does not introduce unrelated sends;
	// onboarding-specific stubs override this method explicitly.
	return true, nil
}

func (m *mockRepo) GetAgentSettings(_ context.Context, _ uuid.UUID) (slackrepository.AgentSettingsRecord, error) {
	settings := m.agentSettings
	return settings, nil
}

func (m *mockRepo) ListAuthorizedChannelTeamIDs(_ context.Context, _, _ uuid.UUID, _ string, _ uuid.UUID) ([]uuid.UUID, error) {
	if m.authorizedTeamIDs != nil {
		return append([]uuid.UUID(nil), m.authorizedTeamIDs...), nil
	}
	result := make([]uuid.UUID, 0, len(m.teams)+1)
	for _, team := range append(append([]slackrepository.TeamRecord(nil), m.teams...), m.team) {
		if team.ID != uuid.Nil {
			result = append(result, team.ID)
		}
	}
	return result, nil
}

func (m *mockRepo) ListInstallationAuthorizedChannelTeamIDs(
	_ context.Context,
	_, _ uuid.UUID,
	_ string,
) ([]uuid.UUID, error) {
	if m.authorizedTeamIDs != nil {
		return append([]uuid.UUID(nil), m.authorizedTeamIDs...), nil
	}
	result := make([]uuid.UUID, 0, len(m.teams)+1)
	for _, team := range append(append([]slackrepository.TeamRecord(nil), m.teams...), m.team) {
		if team.ID != uuid.Nil {
			result = append(result, team.ID)
		}
	}
	return result, nil
}

type blockingSlackWorkspaceRepo struct {
	*mockRepo
	started chan struct{}
	release chan struct{}
}

type blockingTeamListRepo struct {
	*mockRepo
	started chan struct{}
	release chan struct{}
}

type eventInboxCapture struct {
	conversation        conversationRecord
	conversationErr     error
	conversationLookups []conversationInput
}

func (s *eventInboxCapture) FindConversation(_ context.Context, input conversationInput) (conversationRecord, error) {
	s.conversationLookups = append(s.conversationLookups, input)
	if s.conversationErr != nil {
		return conversationRecord{}, s.conversationErr
	}
	if s.conversation.ID == uuid.Nil {
		return conversationRecord{}, errMessagingRecordNotFound
	}
	return s.conversation, nil
}

func (s *eventInboxCapture) FindChannelConversation(ctx context.Context, input conversationInput) (conversationRecord, error) {
	input.AudienceScope = conversationAudienceChannel
	return s.FindConversation(ctx, input)
}

func (r *blockingSlackWorkspaceRepo) GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackrepository.SlackWorkspaceRecord, error) {
	select {
	case r.started <- struct{}{}:
	default:
	}
	select {
	case <-r.release:
		return r.mockRepo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	case <-ctx.Done():
		return slackrepository.SlackWorkspaceRecord{}, ctx.Err()
	}
}

func (r *blockingTeamListRepo) ListWorkspaceTeamsForUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]slackrepository.TeamRecord, error) {
	select {
	case r.started <- struct{}{}:
	default:
	}
	select {
	case <-r.release:
		return r.mockRepo.ListWorkspaceTeamsForUser(ctx, workspaceID, userID)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type credentialUpgradeRepo struct {
	*mockRepo
	upgradeCalls int
	encrypted    string
	version      int
}

func (r *credentialUpgradeRepo) UpgradeSlackCredential(_ context.Context, _ slackrepository.LegacySlackCredentialRecord, encrypted string, version int) error {
	r.upgradeCalls++
	r.encrypted = encrypted
	r.version = version
	return nil
}

func (m *mockRepo) FindWorkspaceBySlug(ctx context.Context, slug string) (slackrepository.WorkspaceRecord, error) {
	if m.err != nil {
		return slackrepository.WorkspaceRecord{}, m.err
	}
	return m.workspace, nil
}

func (m *mockRepo) FindWorkspaceByID(ctx context.Context, workspaceID uuid.UUID) (slackrepository.WorkspaceRecord, error) {
	if m.err != nil {
		return slackrepository.WorkspaceRecord{}, m.err
	}
	return m.workspace, nil
}

func (m *mockRepo) FindTeamByCode(ctx context.Context, workspaceID uuid.UUID, code string) (slackrepository.TeamRecord, error) {
	if m.err != nil {
		return slackrepository.TeamRecord{}, m.err
	}
	return m.team, nil
}

func (m *mockRepo) FindTeamByID(ctx context.Context, workspaceID, teamID uuid.UUID) (slackrepository.TeamRecord, error) {
	if m.err != nil {
		return slackrepository.TeamRecord{}, m.err
	}
	for _, team := range m.teams {
		if team.ID == teamID {
			return team, nil
		}
	}
	if m.team.ID == teamID {
		return m.team, nil
	}
	if m.team.ID == uuid.Nil {
		return slackrepository.TeamRecord{}, errors.New("team not found")
	}
	return m.team, nil
}

func (m *mockRepo) ListWorkspaceTeams(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.TeamRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.teams) > 0 {
		return m.teams, nil
	}
	if m.team.ID != uuid.Nil {
		return []slackrepository.TeamRecord{m.team}, nil
	}
	return []slackrepository.TeamRecord{}, nil
}

func (m *mockRepo) ListWorkspaceTeamsForUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]slackrepository.TeamRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	teams, err := m.ListWorkspaceTeams(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	filtered := make([]slackrepository.TeamRecord, 0)
	for _, team := range teams {
		members, memberErr := m.ListTeamMembers(ctx, team.ID)
		if memberErr != nil {
			return nil, memberErr
		}
		for _, member := range members {
			if member.UserID == userID {
				filtered = append(filtered, team)
				break
			}
		}
	}
	return filtered, nil
}

func (m *mockRepo) ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]slackrepository.StatusRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.statusesByTeam) > 0 {
		if rows, ok := m.statusesByTeam[teamID]; ok {
			return rows, nil
		}
	}
	return m.statuses, nil
}

func (m *mockRepo) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]slackrepository.TeamMemberRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.membersByTeam) > 0 {
		if rows, ok := m.membersByTeam[teamID]; ok {
			return rows, nil
		}
	}
	return m.teamMembers, nil
}

func (m *mockRepo) ListTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID) ([]slackrepository.LabelRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.labelsByTeam) > 0 {
		if rows, ok := m.labelsByTeam[teamID]; ok {
			return rows, nil
		}
	}
	return m.labels, nil
}

func (m *mockRepo) FindTeamMemberByID(ctx context.Context, teamID, userID uuid.UUID) (slackrepository.TeamMemberRecord, error) {
	if m.err != nil {
		return slackrepository.TeamMemberRecord{}, m.err
	}
	members, err := m.ListTeamMembers(ctx, teamID)
	if err != nil {
		return slackrepository.TeamMemberRecord{}, err
	}
	for _, member := range members {
		if member.UserID == userID {
			return member, nil
		}
	}
	return slackrepository.TeamMemberRecord{}, errors.New("member not found")
}

func (m *mockRepo) FindTeamLabelByID(ctx context.Context, workspaceID, teamID, labelID uuid.UUID) (slackrepository.LabelRecord, error) {
	if m.err != nil {
		return slackrepository.LabelRecord{}, m.err
	}
	labels, err := m.ListTeamLabels(ctx, workspaceID, teamID)
	if err != nil {
		return slackrepository.LabelRecord{}, err
	}
	for _, label := range labels {
		if label.ID == labelID {
			return label, nil
		}
	}
	return slackrepository.LabelRecord{}, errors.New("label not found")
}

func (m *mockRepo) FindTeamObjectiveByID(ctx context.Context, workspaceID, teamID, objectiveID uuid.UUID) (slackrepository.ObjectiveRecord, error) {
	if m.err != nil {
		return slackrepository.ObjectiveRecord{}, m.err
	}
	rows := m.objectives
	if len(m.objectivesByTeam) > 0 {
		rows = m.objectivesByTeam[teamID]
	}
	for _, objective := range rows {
		if objective.ID == objectiveID {
			return objective, nil
		}
	}
	return slackrepository.ObjectiveRecord{}, slackdomain.ErrNotFound
}

func (m *mockRepo) SearchTeamMembers(ctx context.Context, teamID uuid.UUID, query string, limit int) ([]slackrepository.TeamMemberRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	members, err := m.ListTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	filtered := make([]slackrepository.TeamMemberRecord, 0)
	q := strings.ToLower(strings.TrimSpace(query))
	for _, member := range members {
		name := strings.ToLower(member.FullName)
		username := strings.ToLower(member.Username)
		email := strings.ToLower(member.Email)
		if strings.Contains(name, q) || strings.Contains(username, q) || strings.Contains(email, q) {
			filtered = append(filtered, member)
		}
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func (m *mockRepo) SearchTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]slackrepository.LabelRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	labels, err := m.ListTeamLabels(ctx, workspaceID, teamID)
	if err != nil {
		return nil, err
	}
	filtered := make([]slackrepository.LabelRecord, 0)
	q := strings.ToLower(strings.TrimSpace(query))
	for _, label := range labels {
		if strings.Contains(strings.ToLower(label.Name), q) {
			filtered = append(filtered, label)
		}
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func (m *mockRepo) SearchTeamObjectives(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]slackrepository.ObjectiveRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	rows := m.objectives
	if len(m.objectivesByTeam) > 0 {
		rows = m.objectivesByTeam[teamID]
	}
	filtered := make([]slackrepository.ObjectiveRecord, 0)
	q := strings.ToLower(strings.TrimSpace(query))
	for _, objective := range rows {
		if strings.Contains(strings.ToLower(objective.Name), q) {
			filtered = append(filtered, objective)
		}
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}
