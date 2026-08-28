package slack

import (
	"context"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
)

func (m *mockRepo) GetWorkspaceBySlackTeamID(ctx context.Context, slackTeamID string) (slackrepository.WorkspaceRecord, error) {
	if m.err != nil {
		return slackrepository.WorkspaceRecord{}, m.err
	}
	return m.workspace, nil
}

func (m *mockRepo) UpsertSlackWorkspace(ctx context.Context, workspaceID, installedByUserID uuid.UUID, payload slackrepository.OAuthInstallPayload) (slackrepository.SlackWorkspaceRecord, error) {
	m.lastOAuthInstall = payload
	m.lastOAuthUserID = installedByUserID
	if m.upsertSlackWorkspaceErr != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.upsertSlackWorkspaceErr
	}
	if m.err != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.err
	}
	result := m.slackWorkspace
	result.WorkspaceID = workspaceID
	result.SlackTeamID = payload.SlackTeamID
	result.InstallGeneration = payload.InstallGeneration
	result.BotAccessToken = payload.BotAccessToken
	result.CredentialVersion = payload.CredentialVersion
	return result, nil
}

func (m *mockRepo) GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackWorkspaceRecord, error) {
	if m.getSlackWorkspaceErr != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.getSlackWorkspaceErr
	}
	if m.err != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.err
	}
	if m.slackWorkspace.WorkspaceID == uuid.Nil && strings.TrimSpace(m.slackWorkspace.SlackTeamID) == "" {
		return slackrepository.SlackWorkspaceRecord{}, slackdomain.ErrNotFound
	}
	return m.slackWorkspace, nil
}

func (m *mockRepo) GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackrepository.SlackWorkspaceRecord, error) {
	if m.getSlackTeamErr != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.getSlackTeamErr
	}
	if m.err != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.err
	}
	if m.slackWorkspace.WorkspaceID == uuid.Nil && strings.TrimSpace(m.slackWorkspace.SlackTeamID) == "" {
		return slackrepository.SlackWorkspaceRecord{}, slackdomain.ErrNotFound
	}
	return m.slackWorkspace, nil
}

func (m *mockRepo) DisconnectSlackWorkspace(ctx context.Context, command slackdomain.DisconnectInstallationCommand) (slackrepository.SlackUninstallRecord, error) {
	if m.err != nil {
		return slackrepository.SlackUninstallRecord{}, m.err
	}
	workspaceID := command.WorkspaceID
	record := slackrepository.SlackUninstallRecord{
		ID:                   uuid.New(),
		SlackWorkspaceID:     m.slackWorkspace.ID,
		WorkspaceID:          workspaceID,
		InstallGeneration:    m.slackWorkspace.InstallGeneration,
		SlackTeamID:          m.slackWorkspace.SlackTeamID,
		UninstallKind:        "disconnect",
		CredentialPayload:    m.slackWorkspace.BotAccessToken,
		CredentialKeyVersion: m.slackWorkspace.CredentialVersion,
		Status:               "pending",
	}
	if m.uninstalls == nil {
		m.uninstalls = make(map[uuid.UUID]slackrepository.SlackUninstallRecord)
	}
	m.uninstalls[record.ID] = record
	m.disconnected = true
	m.slackWorkspace = slackrepository.SlackWorkspaceRecord{}
	return record, nil
}

func (m *mockRepo) EnqueueSlackUninstall(_ context.Context, input slackrepository.SlackUninstallInput) (slackrepository.SlackUninstallRecord, error) {
	m.uninstallInputs = append(m.uninstallInputs, input)
	if m.enqueueUninstallErr != nil {
		return slackrepository.SlackUninstallRecord{}, m.enqueueUninstallErr
	}
	record := slackrepository.SlackUninstallRecord{
		ID:                   uuid.New(),
		SlackWorkspaceID:     input.SlackWorkspaceID,
		WorkspaceID:          input.WorkspaceID,
		InstallGeneration:    input.InstallGeneration,
		SlackTeamID:          input.SlackTeamID,
		UninstallKind:        input.UninstallKind,
		CredentialPayload:    input.CredentialPayload,
		CredentialKeyVersion: input.CredentialKeyVersion,
		Status:               "pending",
	}
	if m.uninstalls == nil {
		m.uninstalls = make(map[uuid.UUID]slackrepository.SlackUninstallRecord)
	}
	m.uninstalls[record.ID] = record
	return record, nil
}

func (m *mockRepo) ClaimSlackUninstall(_ context.Context, id uuid.UUID) (slackrepository.SlackUninstallRecord, bool, error) {
	record, ok := m.uninstalls[id]
	if !ok || record.Status == "completed" {
		return slackrepository.SlackUninstallRecord{}, false, nil
	}
	record.Status = "processing"
	record.AttemptCount++
	m.uninstalls[id] = record
	return record, true, nil
}

func (m *mockRepo) CompleteSlackUninstall(_ context.Context, id uuid.UUID, _ string) error {
	record := m.uninstalls[id]
	record.Status = "completed"
	record.CredentialPayload = ""
	m.uninstalls[id] = record
	m.completedUninstalls = append(m.completedUninstalls, id)
	return nil
}

func (m *mockRepo) FailSlackUninstall(_ context.Context, id uuid.UUID, _ string, nextAttemptAt *time.Time) error {
	record := m.uninstalls[id]
	if nextAttemptAt == nil {
		record.Status = "revocation_required"
	} else {
		record.Status = "failed"
		record.NextAttemptAt = nextAttemptAt
	}
	m.uninstalls[id] = record
	m.failedUninstalls = append(m.failedUninstalls, id)
	return nil
}

func (m *mockRepo) UpsertChannels(ctx context.Context, command slackdomain.SyncChannelsCommand) error {
	m.upsertChannels++
	m.lastChannelWorkspaceID = command.WorkspaceID
	m.lastChannelInstallID = command.InstallationID
	m.lastChannels = append([]slackrepository.SlackChannelPayload(nil), command.Channels...)
	return m.err
}

func (m *mockRepo) ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.SlackChannelRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []slackrepository.SlackChannelRecord{}, nil
}

func (m *mockRepo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

func (m *mockRepo) CreateStoryLink(ctx context.Context, storyID uuid.UUID, sourceKey, title, linkURL string) error {
	if m.err != nil {
		return m.err
	}
	m.lastStoryLink.storyID = storyID
	m.lastStoryLink.sourceKey = sourceKey
	m.lastStoryLink.title = title
	m.lastStoryLink.url = linkURL
	return nil
}

func (m *mockRepo) ListWorkspaceMembersForSlackLinking(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.WorkspaceMemberRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return append([]slackrepository.WorkspaceMemberRecord(nil), m.workspaceMembers...), nil
}

func (m *mockRepo) UpsertSlackUserLinks(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, slackTeamID string, links []slackrepository.SlackUserLinkUpsert) error {
	if m.err != nil {
		return m.err
	}
	if m.slackUserLinks == nil {
		m.slackUserLinks = make(map[string]uuid.UUID)
	}
	for _, link := range links {
		key := strings.TrimSpace(slackTeamID) + ":" + strings.TrimSpace(link.SlackUserID)
		m.slackUserLinks[key] = link.UserID
	}
	return nil
}

func (m *mockRepo) FindLinkedUserIDBySlackUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (*uuid.UUID, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.slackUserLinks == nil {
		return nil, nil
	}
	key := strings.TrimSpace(slackTeamID) + ":" + strings.TrimSpace(slackUserID)
	userID, ok := m.slackUserLinks[key]
	if !ok {
		return nil, nil
	}
	return &userID, nil
}

func (m *mockRepo) FindSlackUserLinkByUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID string, userID uuid.UUID) (*slackrepository.SlackUserLinkRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.slackUserLinks == nil {
		return nil, nil
	}
	prefix := strings.TrimSpace(slackTeamID) + ":"
	for key, linkedUserID := range m.slackUserLinks {
		if strings.HasPrefix(key, prefix) && linkedUserID == userID {
			return &slackrepository.SlackUserLinkRecord{
				SlackUserID: strings.TrimPrefix(key, prefix),
				UserID:      linkedUserID,
			}, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) DeleteSlackUserLink(_ context.Context, _ uuid.UUID, slackTeamID, slackUserID string, userID uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	key := strings.TrimSpace(slackTeamID) + ":" + strings.TrimSpace(slackUserID)
	if m.slackUserLinks == nil || m.slackUserLinks[key] != userID {
		return false, nil
	}
	delete(m.slackUserLinks, key)
	return true, nil
}

func (m *mockRepo) InsertRequestLog(ctx context.Context, entry slackrepository.SlackRequestLogInsert) error {
	m.lastRequestLog = entry
	return m.err
}

func (m *mockRepo) ListRequestLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]slackrepository.SlackRequestLogRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []slackrepository.SlackRequestLogRecord{}, nil
}
