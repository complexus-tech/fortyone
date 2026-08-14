package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

type mockNonceStore struct {
	records map[string]messagingrepository.NonceRecord
}

func newMockNonceStore() *mockNonceStore {
	return &mockNonceStore{records: make(map[string]messagingrepository.NonceRecord)}
}

func nonceStoreKey(provider, purpose string, digest []byte) string {
	return provider + ":" + purpose + ":" + hex.EncodeToString(digest)
}

func (m *mockNonceStore) CreateNonce(_ context.Context, input messagingrepository.NonceInput) error {
	if m.records == nil {
		m.records = make(map[string]messagingrepository.NonceRecord)
	}
	var externalWorkspaceID *string
	if input.ExternalWorkspaceID != "" {
		value := input.ExternalWorkspaceID
		externalWorkspaceID = &value
	}
	var externalUserID *string
	if input.ExternalUserID != "" {
		value := input.ExternalUserID
		externalUserID = &value
	}
	m.records[nonceStoreKey(input.Provider, input.Purpose, input.NonceHash)] = messagingrepository.NonceRecord{
		ID:                  uuid.New(),
		Provider:            input.Provider,
		Purpose:             input.Purpose,
		WorkspaceID:         input.WorkspaceID,
		UserID:              input.UserID,
		ExternalWorkspaceID: externalWorkspaceID,
		ExternalUserID:      externalUserID,
		Payload:             append([]byte(nil), input.Payload...),
		ExpiresAt:           input.ExpiresAt,
	}
	return nil
}

func (m *mockNonceStore) ConsumeNonce(_ context.Context, input messagingrepository.NonceConsumeInput) (messagingrepository.NonceRecord, error) {
	key := nonceStoreKey(input.Provider, input.Purpose, input.NonceHash)
	record, ok := m.records[key]
	if !ok || record.ConsumedAt != nil || !input.Now.Before(record.ExpiresAt) {
		return messagingrepository.NonceRecord{}, messagingrepository.ErrNotFound
	}
	if input.WorkspaceID != nil && record.WorkspaceID != *input.WorkspaceID {
		return messagingrepository.NonceRecord{}, messagingrepository.ErrNotFound
	}
	if input.UserID != nil && record.UserID != nil && *record.UserID != *input.UserID {
		return messagingrepository.NonceRecord{}, messagingrepository.ErrNotFound
	}
	if record.UserID == nil && input.UserID != nil {
		boundUserID := *input.UserID
		record.UserID = &boundUserID
	}
	consumedAt := input.Now
	record.ConsumedAt = &consumedAt
	m.records[key] = record
	return record, nil
}

type mockRepo struct {
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
	registrations       int
	inputs              []messagingrepository.InboundEventInput
	conversation        messagingrepository.ConversationRecord
	conversationErr     error
	conversationLookups []messagingrepository.ConversationInput
}

func (s *eventInboxCapture) RegisterInboundEvent(_ context.Context, input messagingrepository.InboundEventInput) (messagingrepository.InboundEventRecord, bool, error) {
	s.registrations++
	s.inputs = append(s.inputs, input)
	return messagingrepository.InboundEventRecord{ID: uuid.New()}, true, nil
}

func (s *eventInboxCapture) MarkInboundEventQueued(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *eventInboxCapture) FindConversation(_ context.Context, input messagingrepository.ConversationInput) (messagingrepository.ConversationRecord, error) {
	s.conversationLookups = append(s.conversationLookups, input)
	if s.conversationErr != nil {
		return messagingrepository.ConversationRecord{}, s.conversationErr
	}
	if s.conversation.ID == uuid.Nil {
		return messagingrepository.ConversationRecord{}, messagingrepository.ErrNotFound
	}
	return s.conversation, nil
}

func (s *eventInboxCapture) FindChannelConversation(ctx context.Context, input messagingrepository.ConversationInput) (messagingrepository.ConversationRecord, error) {
	input.AudienceScope = messagingrepository.ConversationAudienceChannel
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

func (r *credentialUpgradeRepo) UpgradeSlackCredential(_ context.Context, _ uuid.UUID, encrypted string, version int) error {
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
	return slackrepository.ObjectiveRecord{}, sql.ErrNoRows
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
	return m.slackWorkspace, nil
}

func (m *mockRepo) GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackWorkspaceRecord, error) {
	if m.getSlackWorkspaceErr != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.getSlackWorkspaceErr
	}
	if m.err != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.err
	}
	if m.slackWorkspace.WorkspaceID == uuid.Nil && strings.TrimSpace(m.slackWorkspace.SlackTeamID) == "" {
		return slackrepository.SlackWorkspaceRecord{}, sql.ErrNoRows
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
		return slackrepository.SlackWorkspaceRecord{}, sql.ErrNoRows
	}
	return m.slackWorkspace, nil
}

func (m *mockRepo) DisconnectSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackUninstallRecord, error) {
	if m.err != nil {
		return slackrepository.SlackUninstallRecord{}, m.err
	}
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

func (m *mockRepo) UpsertChannels(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, channels []slackrepository.SlackChannelPayload) error {
	m.upsertChannels++
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

type mockRequestStore struct {
	last         integrationrequests.CoreUpsertRequestInput
	lastCreated  integrationrequests.CoreIntegrationRequest
	request      integrationrequests.CoreIntegrationRequest
	requestErr   error
	lastThread   integrationrequests.CoreBindProviderThreadInput
	calls        int
	bindCalls    int
	threadMatch  bool
	thread       integrationrequests.CoreProviderThread
	threadLookup integrationrequests.CoreProviderThreadLookupInput
}

func (m *mockRequestStore) GetForUser(_ context.Context, _, requestID, _ uuid.UUID) (integrationrequests.CoreIntegrationRequest, error) {
	if m.requestErr != nil {
		return integrationrequests.CoreIntegrationRequest{}, m.requestErr
	}
	if m.request.ID == requestID {
		return m.request, nil
	}
	return integrationrequests.CoreIntegrationRequest{}, sql.ErrNoRows
}

func (m *mockRequestStore) BindProviderThread(_ context.Context, input integrationrequests.CoreBindProviderThreadInput) (integrationrequests.CoreProviderThread, error) {
	m.bindCalls++
	m.lastThread = input
	return integrationrequests.CoreProviderThread{ID: uuid.New()}, nil
}

func (m *mockRequestStore) HasAuthorizedProviderThread(_ context.Context, _ integrationrequests.CoreProviderThreadMatchInput) (bool, error) {
	return m.threadMatch, nil
}

func (m *mockRequestStore) HasCurrentProviderThread(_ context.Context, input integrationrequests.CoreProviderThreadLookupInput) (bool, error) {
	m.threadLookup = input
	return m.threadMatch, nil
}

func (m *mockRequestStore) FindProviderThread(_ context.Context, _, _ uuid.UUID, _ string) (integrationrequests.CoreProviderThread, error) {
	if m.thread.ID == uuid.Nil {
		return integrationrequests.CoreProviderThread{}, integrationrequests.ErrProviderThreadNotFound
	}
	return m.thread, nil
}

func (m *mockRequestStore) UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error) {
	m.calls++
	m.last = input
	m.lastCreated = integrationrequests.CoreIntegrationRequest{
		ID:               uuid.New(),
		WorkspaceID:      input.WorkspaceID,
		TeamID:           input.TeamID,
		Provider:         input.Provider,
		SourceType:       input.SourceType,
		SourceExternalID: input.SourceExternalID,
		SourceURL:        input.SourceURL,
		Title:            input.Title,
		Description:      input.Description,
		Priority:         input.Priority,
		AssigneeID:       input.AssigneeID,
		EndDate:          input.EndDate,
		CreatedByUserID:  input.CreatedByUserID,
		Status:           integrationrequests.StatusPending,
		CreatedAt:        time.Unix(1_700_000_000, 0),
		UpdatedAt:        time.Unix(1_700_000_000, 0),
	}
	return m.lastCreated, nil
}

type mockStoryService struct {
	lastActorID   uuid.UUID
	lastWorkspace uuid.UUID
	lastStory     stories.CoreNewStory
	createCalls   int
	sequenceID    int
}

type mutationConfirmerStub struct {
	result       messaging.StoryMutationResult
	cancelResult messaging.StoryMutationCancellationResult
	err          error
	scopes       []messaging.ToolScope
	tokens       []string
}

func (s *mutationConfirmerStub) CancelStoryMutation(_ context.Context, scope messaging.ToolScope, token string) (messaging.StoryMutationCancellationResult, error) {
	s.scopes = append(s.scopes, scope)
	s.tokens = append(s.tokens, token)
	return s.cancelResult, s.err
}

func (s *mutationConfirmerStub) ConfirmStoryMutation(_ context.Context, scope messaging.ToolScope, token string) (messaging.StoryMutationResult, error) {
	s.scopes = append(s.scopes, scope)
	s.tokens = append(s.tokens, token)
	return s.result, s.err
}

func (m *mockStoryService) Create(ctx context.Context, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	m.createCalls++
	if ns.Reporter != nil {
		m.lastActorID = *ns.Reporter
	}
	m.lastWorkspace = workspaceID
	m.lastStory = ns
	return stories.CoreSingleStory{
		ID:          uuid.New(),
		SequenceID:  m.sequenceID,
		Title:       ns.Title,
		Description: ns.Description,
		Status:      ns.Status,
		Assignee:    ns.Assignee,
		Reporter:    ns.Reporter,
		Priority:    ns.Priority,
		Team:        ns.Team,
		Workspace:   workspaceID,
		CreatedAt:   time.Unix(1_700_000_000, 0),
		UpdatedAt:   time.Unix(1_700_000_000, 0),
		CreatedNow:  true,
	}, nil
}

func newTestService(repo Repository, requests RequestStore, storyService StoryService, cfg Config) *Service {
	testLogger := logger.NewWithJSON(io.Discard, slog.LevelError, "test")
	service := New(testLogger, repo, requests, storyService, cfg, WithNonceStore(newMockNonceStore()))
	service.clock = fixedClock{now: time.Unix(1_700_000_000, 0)}
	return service
}

func TestVerifyRequest(t *testing.T) {
	secret := "secret"
	service := newTestService(nil, nil, nil, Config{SigningSecret: secret})
	body := []byte("payload=test")
	timestamp := "1700000000"

	headers := http.Header{}
	headers.Set("X-Slack-Request-Timestamp", timestamp)
	headers.Set("X-Slack-Signature", slackSignature(secret, timestamp, body))

	err := service.VerifyRequest(body, headers)
	require.NoError(t, err)
}

func TestHandleEventsDropsUnknownSlackInstallationBeforePersistence(t *testing.T) {
	repo := &mockRepo{getSlackTeamErr: sql.ErrNoRows}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(directMessageEvent("Ev-unknown-team", "private message")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Zero(t, inbox.registrations)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsDropsUnrelatedChannelMessagesBeforePersistence(t *testing.T) {
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       uuid.New(),
		SlackTeamID:       "T1",
		InstallGeneration: uuid.New(),
	}}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{SecretKey: "event-secret"})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(`{"type":"event_callback","team_id":"T1","event_id":"Ev-root","event":{"type":"message","channel_type":"channel","user":"U1","channel":"C1","ts":"10.1","text":"unrelated channel message"}}`))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Zero(t, inbox.registrations)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsPersistsCandidateChannelThreadReply(t *testing.T) {
	workspaceID := uuid.New()
	installGeneration := uuid.New()
	linkedUserID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: installGeneration,
		AuthorizedAt:      time.Unix(1_700_000_000, 0).UTC(),
	}, slackUserLinks: map[string]uuid.UUID{"T1:U1": linkedUserID}}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{conversation: messagingrepository.ConversationRecord{
		ID:        uuid.New(),
		UpdatedAt: time.Unix(1_700_000_000, 0).UTC(),
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{SecretKey: "event-secret"})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(channelThreadEvent("Ev-thread", "U1", "what about urgent work?")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Equal(t, 1, inbox.registrations)
	require.Len(t, inbox.inputs, 1)
	require.Equal(t, workspaceID, *inbox.inputs[0].WorkspaceID)
	require.Equal(t, installGeneration, *inbox.inputs[0].InstallGeneration)
	require.Equal(t, "message", inbox.inputs[0].EventType)
	require.NotEmpty(t, inbox.inputs[0].PayloadEncrypted)
	require.Len(t, inbox.conversationLookups, 1)
	require.Equal(t, linkedUserID, inbox.conversationLookups[0].UserID)
	require.Equal(t, "10.1", inbox.conversationLookups[0].ExternalThreadID)
	require.Len(t, queue.payloads, 1)
	require.Equal(t, "Ev-thread", queue.payloads[0].EventID)
}

func TestHandleEventsPersistsExactRequestThreadReplyFromUnlinkedSlackActor(t *testing.T) {
	workspaceID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: installGeneration,
		AuthorizedAt:      time.Unix(1_700_000_000, 0).UTC(),
	}}
	requests := &mockRequestStore{threadMatch: true}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, requests, &mockStoryService{}, Config{SecretKey: "event-secret"})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(channelThreadEvent("Ev-unlinked-request-thread", "U-external", "Customer confirmed")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Empty(t, inbox.conversationLookups)
	require.Equal(t, 1, inbox.registrations)
	require.Len(t, queue.payloads, 1)
	require.Equal(t, workspaceID, requests.threadLookup.WorkspaceID)
	require.Equal(t, installGeneration, requests.threadLookup.InstallationGeneration)
	require.Equal(t, "T1", requests.threadLookup.ExternalWorkspaceID)
	require.Equal(t, "C1", requests.threadLookup.ExternalChannelID)
	require.Equal(t, "10.1", requests.threadLookup.ExternalThreadID)
}

func TestHandleEventsDropsUnsubscribedChannelThreadBeforePersistence(t *testing.T) {
	workspaceID := uuid.New()
	linkedUserID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: uuid.New(),
		AuthorizedAt:      time.Unix(1_700_000_000, 0).UTC(),
	}, slackUserLinks: map[string]uuid.UUID{"T1:U1": linkedUserID}}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{SecretKey: "event-secret"})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(channelThreadEvent("Ev-unsubscribed", "U1", "unrelated reply")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Len(t, inbox.conversationLookups, 1)
	require.Zero(t, inbox.registrations)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsDropsBroadDuplicateOfAppMentionBeforePersistence(t *testing.T) {
	workspaceID := uuid.New()
	botUserID := "B1"
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		BotUserID:         &botUserID,
		InstallGeneration: uuid.New(),
		AuthorizedAt:      time.Unix(1_700_000_000, 0).UTC(),
	}}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{SecretKey: "event-secret"})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(channelThreadEvent("Ev-mention-copy", "U1", "<@B1> show my work")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Empty(t, inbox.conversationLookups)
	require.Zero(t, inbox.registrations)
	require.Empty(t, queue.payloads)
}

func TestCreateInstallSessionStoresOpaqueBoundStateWithCoreScopes(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	service := newTestService(&mockRepo{}, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		SecretKey:    "encryption-secret",
	})

	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	require.Equal(t, slackBotOAuthScopeValue(), installURL.Query().Get("scope"))
	state := installURL.Query().Get("state")
	require.NotEmpty(t, state)
	require.NotContains(t, state, ".")

	nonce, err := base64.RawURLEncoding.DecodeString(state)
	require.NoError(t, err)
	require.Len(t, nonce, slackOpaqueNonceSize)
	digest := sha256.Sum256(nonce)
	store := service.nonces.(*mockNonceStore)
	record, ok := store.records[nonceStoreKey(slackProviderMessaging, slackNoncePurposeOAuth, digest[:])]
	require.True(t, ok)
	require.Equal(t, workspaceID, record.WorkspaceID)
	require.NotNil(t, record.UserID)
	require.Equal(t, userID, *record.UserID)
	require.True(t, record.ExpiresAt.Equal(service.clock.Now().Add(slackOAuthNonceTTL)))
	require.JSONEq(t, `{"workspace_slug":"acme"}`, string(record.Payload))
}

func TestCreateAccountLinkSessionBindsDashboardUserAndConnectedSlackTeam(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		SlackTeamID: "T123",
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://fortyone.app",
		SecretKey:    "encryption-secret",
	})

	session, err := service.CreateAccountLinkSession(
		context.Background(), workspaceID, userID, "acme", "https://acme.fortyone.app/teams/team-1/requests/request-1",
	)
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	require.False(t, session.Linked)
	require.True(t, session.CanLink)
	require.Equal(t, "T123", installURL.Query().Get("team"))
	require.Equal(t, slackBotOAuthScopeValue(), installURL.Query().Get("scope"))

	nonce, err := base64.RawURLEncoding.DecodeString(installURL.Query().Get("state"))
	require.NoError(t, err)
	digest := sha256.Sum256(nonce)
	record := service.nonces.(*mockNonceStore).records[nonceStoreKey(slackProviderMessaging, slackNoncePurposeAccount, digest[:])]
	require.Equal(t, workspaceID, record.WorkspaceID)
	require.Equal(t, "T123", valueOrEmpty(record.ExternalWorkspaceID))
	require.Equal(t, userID, *record.UserID)
	require.JSONEq(t, `{"workspace_slug":"acme","return_url":"https://acme.fortyone.app/teams/team-1/requests/request-1"}`, string(record.Payload))
}

func TestHandleSetupLinksDashboardOAuthUserAndReturnsToRequest(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		SlackTeamID: "T123",
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://fortyone.app",
		SecretKey:    "encryption-secret",
	})
	session, err := service.CreateAccountLinkSession(
		context.Background(), workspaceID, userID, "acme", "https://acme.fortyone.app/teams/team-1/requests/request-1",
	)
	require.NoError(t, err)
	state, err := url.Parse(session.InstallURL)
	require.NoError(t, err)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/oauth.v2.access", request.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"access_token":"xoxb-token","team":{"id":"T123","name":"Acme","domain":"acme"},"authed_user":{"id":"U123"}}`))
	}))
	defer provider.Close()
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	redirectURL, err := service.HandleSetup(context.Background(), "oauth-code", state.Query().Get("state"), "")
	require.NoError(t, err)
	require.Equal(t, "https://acme.fortyone.app/teams/team-1/requests/request-1?slack_link_status=success", redirectURL)
	require.Equal(t, userID, repo.slackUserLinks["T123:U123"])
}

func TestHandleSetupConsumesStateAndStoresEncryptedInstallationMetadata(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{ID: uuid.New(), WorkspaceID: workspaceID}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://fortyone.app",
		SecretKey:    "encryption-secret",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	state := installURL.Query().Get("state")

	requests := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests++
		require.Equal(t, "/oauth.v2.access", req.URL.Path)
		require.NoError(t, req.ParseForm())
		require.Equal(t, "oauth-code", req.Form.Get("code"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"ok": true,
			"access_token": "xoxb-secret",
			"refresh_token": "xoxe-refresh",
			"expires_in": 3600,
			"bot_user_id": "UBOT",
			"app_id": "A123",
			"scope": "app_mentions:read,channels:history,channels:read,chat:write,chat:write.public,commands,groups:history,groups:read,im:history,links:read,links:write",
			"team": {"id": "T123", "name": "Acme", "domain": "acme"},
			"enterprise": {"id": "E123"},
			"authed_user": {"id": "UADMIN"}
		}`))
	}))
	defer testServer.Close()
	service.client = testServer.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = testServer.URL

	redirectURL, err := service.HandleSetup(context.Background(), "oauth-code", state, "")
	require.NoError(t, err)
	require.Equal(t, "https://acme.fortyone.app/settings/workspace/integrations/slack", redirectURL)
	require.Equal(t, userID, repo.lastOAuthUserID)
	require.NotEqual(t, "xoxb-secret", repo.lastOAuthInstall.BotAccessToken)
	require.Equal(t, "xoxb-secret", repo.lastOAuthInstall.LegacyAccessToken)
	require.Positive(t, repo.lastOAuthInstall.CredentialVersion)
	require.Equal(t, "A123", valueOrEmpty(repo.lastOAuthInstall.SlackAppID))
	require.Equal(t, "E123", valueOrEmpty(repo.lastOAuthInstall.EnterpriseID))
	require.Equal(t, "UADMIN", valueOrEmpty(repo.lastOAuthInstall.AuthedUserID))
	credential, version, err := service.credentials.open(repo.lastOAuthInstall.BotAccessToken)
	require.NoError(t, err)
	require.Equal(t, repo.lastOAuthInstall.CredentialVersion, version)
	require.Equal(t, "xoxb-secret", credential.AccessToken)
	require.Equal(t, "xoxe-refresh", credential.RefreshToken)
	require.NotNil(t, credential.ExpiresAt)
	require.True(t, credential.ExpiresAt.Equal(service.clock.Now().Add(time.Hour)))
	require.Zero(t, repo.upsertChannels)
	require.Equal(t, 1, requests)

	_, err = service.HandleSetup(context.Background(), "oauth-code", state, "")
	require.Error(t, err)
	require.Equal(t, 1, requests)
}

func TestHandleSetupRejectsExpiredStateBeforeOAuthExchange(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	service := newTestService(&mockRepo{}, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		SecretKey:    "encryption-secret",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	service.clock = fixedClock{now: service.clock.Now().Add(slackOAuthNonceTTL + time.Second)}
	apiCalls := 0
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		apiCalls++
		return nil, errors.New("OAuth exchange must not run")
	})}

	_, err = service.HandleSetup(context.Background(), "oauth-code", installURL.Query().Get("state"), "")

	require.ErrorContains(t, err, "invalid or expired")
	require.Zero(t, apiCalls)
}

func TestHandleSetupCleansUpOnlyConclusiveUnownedWorkspaceConflict(t *testing.T) {
	tests := []struct {
		name               string
		upsertErr          error
		currentSlackTeamID string
		teamLookupErr      error
		wantUninstall      bool
	}{
		{
			name:               "workspace connected to another team and selected team unowned",
			upsertErr:          fmt.Errorf("%w: %w", slackrepository.ErrActiveInstallationConflict, slackrepository.ErrWorkspaceAlreadyConnected),
			currentSlackTeamID: "T-OLD",
			teamLookupErr:      sql.ErrNoRows,
			wantUninstall:      true,
		},
		{
			name:               "selected team is legitimately owned elsewhere",
			upsertErr:          fmt.Errorf("%w: %w", slackrepository.ErrActiveInstallationConflict, slackrepository.ErrSlackTeamAlreadyConnected),
			currentSlackTeamID: "T-NEW",
			wantUninstall:      false,
		},
		{
			name:               "uncertain commit is visible on reread",
			upsertErr:          fmt.Errorf("%w: %w", slackrepository.ErrActiveInstallationConflict, slackrepository.ErrWorkspaceAlreadyConnected),
			currentSlackTeamID: "T-NEW",
			wantUninstall:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspaceID := uuid.New()
			userID := uuid.New()
			repo := &mockRepo{
				upsertSlackWorkspaceErr: test.upsertErr,
				getSlackTeamErr:         test.teamLookupErr,
				slackWorkspace: slackrepository.SlackWorkspaceRecord{
					ID:                uuid.New(),
					WorkspaceID:       workspaceID,
					SlackTeamID:       test.currentSlackTeamID,
					InstallGeneration: uuid.New(),
					IsActive:          true,
				},
			}
			service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "https://api.example.com/integrations/slack/setup",
				WebsiteURL:   "https://app.example.com",
				SecretKey:    "encryption-secret",
			})
			session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
			require.NoError(t, err)
			installURL, err := url.Parse(session.InstallURL)
			require.NoError(t, err)
			uninstallCalls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/oauth.v2.access":
					_, _ = w.Write([]byte(`{"ok":true,"access_token":"xoxb-new","bot_user_id":"B1","team":{"id":"T-NEW","name":"New"},"authed_user":{"id":"U1"}}`))
				case "/apps.uninstall":
					uninstallCalls++
					require.NoError(t, req.ParseForm())
					require.Equal(t, "xoxb-new", req.Form.Get("token"))
					_, _ = w.Write([]byte(`{"ok":true}`))
				default:
					http.NotFound(w, req)
				}
			}))
			defer server.Close()
			service.client = server.Client()
			service.webClient = newSlackWebClient(service.client)
			service.webClient.baseURL = server.URL

			_, err = service.HandleSetup(context.Background(), "oauth-code", installURL.Query().Get("state"), "")

			require.Error(t, err)
			require.Equal(t, test.wantUninstall, uninstallCalls == 1)
			if test.wantUninstall {
				require.Len(t, repo.uninstallInputs, 1)
				require.Equal(t, "orphaned_oauth", repo.uninstallInputs[0].UninstallKind)
				require.NotEqual(t, "xoxb-new", repo.uninstallInputs[0].CredentialPayload)
			} else {
				require.Empty(t, repo.uninstallInputs)
			}
		})
	}
}

func TestHandleSetupDoesNotUninstallAfterUncertainCleanupPersistenceFailure(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{
		upsertSlackWorkspaceErr: fmt.Errorf("%w: %w", slackrepository.ErrActiveInstallationConflict, slackrepository.ErrWorkspaceAlreadyConnected),
		getSlackTeamErr:         sql.ErrNoRows,
		enqueueUninstallErr:     errors.New("database commit outcome is unknown"),
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T-OLD",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://app.example.com",
		SecretKey:    "encryption-secret",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	uninstallCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/oauth.v2.access":
			_, _ = w.Write([]byte(`{"ok":true,"access_token":"xoxb-new","bot_user_id":"B1","team":{"id":"T-NEW","name":"New"},"authed_user":{"id":"U1"}}`))
		case "/apps.uninstall":
			uninstallCalls++
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()
	service.client = server.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = server.URL

	_, err = service.HandleSetup(context.Background(), "oauth-code", installURL.Query().Get("state"), "")

	require.Error(t, err)
	require.Zero(t, uninstallCalls, "an uncertain enqueue must never trigger an unguarded apps.uninstall call")
	require.Len(t, repo.uninstallInputs, 1)
}

func TestBotTokenLazilyUpgradesLegacyCredential(t *testing.T) {
	workspaceID := uuid.New()
	slackWorkspaceID := uuid.New()
	repo := &credentialUpgradeRepo{mockRepo: &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:                slackWorkspaceID,
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    "xoxb-legacy",
		CredentialVersion: 0,
	}}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{SecretKey: "encryption-secret"})

	token, err := service.botToken(context.Background(), repo.slackWorkspace)
	require.NoError(t, err)
	require.Equal(t, "xoxb-legacy", token)
	require.Equal(t, 1, repo.upgradeCalls)
	require.Equal(t, "xoxb-legacy", repo.slackWorkspace.BotAccessToken)
	require.NotEqual(t, "xoxb-legacy", repo.encrypted)
	require.Positive(t, repo.version)
	credential, version, err := service.credentials.open(repo.encrypted)
	require.NoError(t, err)
	require.Equal(t, repo.version, version)
	require.Equal(t, "xoxb-legacy", credential.AccessToken)
}

func TestRecordRequestLogPersistsOnlyStructuredNonSensitiveFields(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{workspace: slackrepository.WorkspaceRecord{ID: workspaceID}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	body := []byte(url.Values{
		"team_id":    {"T123"},
		"user_id":    {"U123"},
		"channel_id": {"C123"},
		"command":    {"/fortyone"},
		"trigger_id": {"sensitive-trigger"},
		"text":       {"customer-sensitive content"},
	}.Encode())

	service.RecordRequestLog(context.Background(), CoreRequestLogInput{
		RequestType: "commands",
		Endpoint:    "/integrations/slack/commands",
		RawBody:     body,
		Headers: map[string]string{
			"X-Slack-Signature":    "sensitive-signature",
			"X-Slack-Retry-Num":    "2",
			"X-Slack-Retry-Reason": "http_timeout",
		},
		Outcome: "processed",
	})

	require.Equal(t, &workspaceID, repo.lastRequestLog.WorkspaceID)
	require.Equal(t, "T123", valueOrEmpty(repo.lastRequestLog.SlackTeamID))
	require.Equal(t, "U123", valueOrEmpty(repo.lastRequestLog.SlackUserID))
	require.Equal(t, "/fortyone", valueOrEmpty(repo.lastRequestLog.Command))
	require.Nil(t, repo.lastRequestLog.TriggerID)
	require.Nil(t, repo.lastRequestLog.RequestBody)
	require.JSONEq(t, `{"X-Slack-Retry-Num":"2","X-Slack-Retry-Reason":"http_timeout"}`, string(repo.lastRequestLog.Headers))
}

func TestDisconnectWorkspaceDeletesSlackWorkspace(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{ClientID: "client", ClientSecret: "secret", SecretKey: "encryption-secret"})
	credentialPayload, credentialVersion, err := service.credentials.seal(slackCredential{AccessToken: "xoxb-token"})
	require.NoError(t, err)
	repo.slackWorkspace = slackrepository.SlackWorkspaceRecord{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    credentialPayload,
		CredentialVersion: credentialVersion,
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/apps.uninstall", req.URL.String())
		require.NoError(t, req.ParseForm())
		require.Equal(t, "client", req.Form.Get("client_id"))
		require.Equal(t, "secret", req.Form.Get("client_secret"))
		require.Equal(t, "xoxb-token", req.Form.Get("token"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
		}, nil
	})}

	err = service.DisconnectWorkspace(context.Background(), workspaceID)

	require.NoError(t, err)
	require.True(t, repo.disconnected)
	require.Equal(t, slackrepository.SlackWorkspaceRecord{}, repo.slackWorkspace)
	require.Len(t, repo.completedUninstalls, 1)
	require.Empty(t, repo.uninstalls[repo.completedUninstalls[0]].CredentialPayload)
}

func TestDisconnectWorkspaceRevokesLocalInstallWhenSlackUninstallFails(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{ClientID: "client", ClientSecret: "secret", SecretKey: "encryption-secret"})
	credentialPayload, credentialVersion, err := service.credentials.seal(slackCredential{AccessToken: "xoxb-token"})
	require.NoError(t, err)
	repo.slackWorkspace = slackrepository.SlackWorkspaceRecord{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    credentialPayload,
		CredentialVersion: credentialVersion,
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":false,"error":"invalid_auth"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	err = service.DisconnectWorkspace(context.Background(), workspaceID)

	require.NoError(t, err)
	require.True(t, repo.disconnected)
	require.Equal(t, slackrepository.SlackWorkspaceRecord{}, repo.slackWorkspace)
	require.Len(t, repo.completedUninstalls, 1)
}

func TestDisconnectWorkspaceRetainsEncryptedCredentialForTransientUninstallRetry(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{ClientID: "client", ClientSecret: "secret", SecretKey: "encryption-secret"})
	credentialPayload, credentialVersion, err := service.credentials.seal(slackCredential{AccessToken: "xoxb-token"})
	require.NoError(t, err)
	repo.slackWorkspace = slackrepository.SlackWorkspaceRecord{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    credentialPayload,
		CredentialVersion: credentialVersion,
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"ok":false,"error":"internal_error"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	err = service.DisconnectWorkspace(context.Background(), workspaceID)

	require.NoError(t, err)
	require.True(t, repo.disconnected)
	require.Len(t, repo.failedUninstalls, 1)
	record := repo.uninstalls[repo.failedUninstalls[0]]
	require.Equal(t, "failed", record.Status)
	require.Equal(t, credentialPayload, record.CredentialPayload)
	require.NotNil(t, record.NextAttemptAt)
}

func TestHandleViewSubmissionCreatesSlackRequestWhenRequestStatusSelected(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installeBy := uuid.New()
	actorID := uuid.New()
	labelID := uuid.New()
	objectiveID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:  []slackrepository.StatusRecord{{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: actorID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"},
		},
		labels:     []slackrepository.LabelRecord{{ID: labelID, Name: "Bug"}},
		objectives: []slackrepository.ObjectiveRecord{{ID: objectiveID, Name: "Improve reliability"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installeBy,
			InstallGeneration: installGeneration,
			IsActive:          true,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	requests := &mockRequestStore{}
	service := newTestService(repo, requests, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	store := newEventStoreStub()
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/chat.postMessage", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"user": map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme","slack_channel_id":"C123","slack_message_ts":"171234.000100"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": teamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Fix login bug"}},
					"description": map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "User cannot log in from iOS"}},
					modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
						modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"type": "static_select", "selected_option": map[string]any{"value": slackRequestStatusValue}},
					},
					modalTeamScopedID(modalBlockAssignee, teamID): map[string]any{
						modalTeamScopedID(modalActionAssigneeSelect, teamID): map[string]any{"type": "external_select", "selected_option": map[string]any{"value": actorID.String()}},
					},
					modalTeamScopedID(modalBlockLabels, teamID): map[string]any{
						modalTeamScopedID(modalActionLabelsMultiSelect, teamID): map[string]any{"type": "multi_external_select", "selected_options": []map[string]any{{"value": labelID.String()}}},
					},
					modalTeamScopedID(modalBlockObjective, teamID): map[string]any{
						modalTeamScopedID(modalActionObjectiveSelect, teamID): map[string]any{"type": "external_select", "selected_option": map[string]any{"value": objectiveID.String()}},
					},
					"priority": map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": "High"}}},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), `"response_action":"clear"`)

	require.Equal(t, integrationrequests.ProviderSlack, requests.last.Provider)
	require.Equal(t, SourceTypeSlackMessage, requests.last.SourceType)
	require.Equal(t, "Fix login bug", requests.last.Title)
	require.Equal(t, workspaceID, requests.last.WorkspaceID)
	require.Equal(t, teamID, requests.last.TeamID)
	require.Equal(t, "High", requests.last.Priority)
	require.Equal(t, &actorID, requests.last.AssigneeID)
	require.Equal(t, &objectiveID, requests.last.ObjectiveID)
	require.Equal(t, []uuid.UUID{labelID}, requests.last.LabelIDs)
	require.Equal(t, &actorID, requests.last.CreatedByUserID)
	require.Equal(t, []string{labelID.String()}, requests.last.Metadata["label_ids"])
	require.NotNil(t, requests.last.SourceURL)
	require.True(t, strings.Contains(*requests.last.SourceURL, "acme.slack.com/archives/C123"))
	require.Equal(t, 1, requests.bindCalls)
	require.Equal(t, "T123", requests.lastThread.ExternalWorkspaceID)
	require.Equal(t, "C123", requests.lastThread.ExternalChannelID)
	require.Equal(t, "171.200", requests.lastThread.ExternalThreadID)
	require.Len(t, store.outboundInputs, 1)
	require.Equal(
		t,
		fmt.Sprintf("Joseph Mukorivo <https://acme.fortyone.app/teams/%s/requests/%s|opened a request>", teamID, requests.lastCreated.ID),
		store.outboundInputs[0].Content,
	)
	providerPayload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.NotNil(t, providerPayload.Metadata)
	require.Len(t, providerPayload.Metadata.Entities, 1)
	requestEntity := providerPayload.Metadata.Entities[0]
	require.Equal(t, slackRequestExternalRefType, requestEntity.ExternalRef.Type)
	require.Equal(t, fmt.Sprintf("https://acme.fortyone.app/teams/%s/requests/%s", teamID, requests.lastCreated.ID), requestEntity.URL)
	require.Equal(t, "Fix login bug", requestEntity.EntityPayload.Attributes.Title.Text)
	require.Equal(t, "U123", requestEntity.EntityPayload.Fields["created_by"].User.UserID)
	require.Equal(t, "Pending", requestEntity.EntityPayload.Fields["status"].Value)
	require.Nil(t, requestEntity.EntityPayload.Attributes.Title.Edit)
	require.NotNil(t, providerPayload.UnfurlLinks)
	require.False(t, *providerPayload.UnfurlLinks)
	require.NotNil(t, providerPayload.UnfurlMedia)
	require.False(t, *providerPayload.UnfurlMedia)
}

func TestHandleViewSubmissionCreatesStoryWhenNonTriageStatusSelected(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	triageStatusID := uuid.New()
	doneStatusID := uuid.New()
	installedBy := uuid.New()
	mappedActorID := uuid.New()
	assigneeID := uuid.New()
	labelID := uuid.New()
	objectiveID := uuid.New()
	installGeneration := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses: []slackrepository.StatusRecord{
			{ID: triageStatusID, Name: "Triage", Category: "unstarted"},
			{ID: doneStatusID, Name: "Done", Category: "completed"},
		},
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: mappedActorID, Username: "actor", FullName: "Slack Actor", Email: "actor@example.com"},
			{UserID: assigneeID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"},
		},
		labels:     []slackrepository.LabelRecord{{ID: labelID, Name: "Bug"}},
		objectives: []slackrepository.ObjectiveRecord{{ID: objectiveID, Name: "Improve reliability"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
			InstallGeneration: installGeneration,
			IsActive:          true,
		},
		slackUserLinks: map[string]uuid.UUID{
			"T123:U999": mappedActorID,
		},
	}
	requests := &mockRequestStore{}
	storyService := &mockStoryService{sequenceID: 123}
	service := newTestService(repo, requests, storyService, Config{WebsiteURL: "https://fortyone.app"})
	store := newEventStoreStub()
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/chat.postMessage", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"user": map[string]any{"id": "U999", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme","slack_channel_id":"C123","slack_message_ts":"171234.000100","slack_user_id":"U999","slack_username":"joseph"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": teamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Ship release"}},
					"description": map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Ready to ship"}},
					modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
						modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"type": "static_select", "selected_option": map[string]any{"value": doneStatusID.String()}},
					},
					modalTeamScopedID(modalBlockAssignee, teamID): map[string]any{
						modalTeamScopedID(modalActionAssigneeSelect, teamID): map[string]any{"type": "external_select", "selected_option": map[string]any{"value": assigneeID.String()}},
					},
					modalTeamScopedID(modalBlockLabels, teamID): map[string]any{
						modalTeamScopedID(modalActionLabelsMultiSelect, teamID): map[string]any{"type": "multi_external_select", "selected_options": []map[string]any{{"value": labelID.String()}}},
					},
					modalTeamScopedID(modalBlockObjective, teamID): map[string]any{
						modalTeamScopedID(modalActionObjectiveSelect, teamID): map[string]any{"type": "external_select", "selected_option": map[string]any{"value": objectiveID.String()}},
					},
					"priority": map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": "Urgent"}}},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), `"response_action":"clear"`)

	require.Equal(t, "", requests.last.Provider)
	require.Equal(t, mappedActorID, storyService.lastActorID)
	require.Equal(t, teamID, storyService.lastStory.Team)
	require.NotNil(t, storyService.lastStory.Status)
	require.Equal(t, doneStatusID, *storyService.lastStory.Status)
	require.NotNil(t, storyService.lastStory.Assignee)
	require.Equal(t, assigneeID, *storyService.lastStory.Assignee)
	require.NotNil(t, storyService.lastStory.Objective)
	require.Equal(t, objectiveID, *storyService.lastStory.Objective)
	require.Equal(t, "Urgent", storyService.lastStory.Priority)
	require.Equal(t, []uuid.UUID{labelID}, storyService.lastStory.LabelIDs)
	require.Empty(t, repo.lastStoryLink.url, "direct stories must not create a Slack source association")
	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "Slack Actor created <https://acme.fortyone.app/work/ENG-123|ENG-123>", store.outboundInputs[0].Content)
	require.Equal(t, "C123", store.outboundInputs[0].ExternalChannelID)
	require.Empty(t, store.outboundInputs[0].ExternalThreadID)
	providerPayload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.Len(t, providerPayload.Metadata.Entities, 1)
	require.Equal(t, "https://acme.fortyone.app/work/ENG-123", providerPayload.Metadata.Entities[0].URL)
	require.False(t, *providerPayload.UnfurlLinks)
	require.False(t, *providerPayload.UnfurlMedia)
}

func TestHandleViewSubmissionRejectsTeamOutsideActorsMembership(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	allowedTeamID := uuid.New()
	blockedTeamID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams: []slackrepository.TeamRecord{
			{ID: allowedTeamID, Code: "ENG", Name: "Engineering"},
			{ID: blockedTeamID, Code: "OPS", Name: "Operations"},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			allowedTeamID: {{UserID: actorID}},
			blockedTeamID: {{UserID: uuid.New()}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	requests := &mockRequestStore{}
	storyService := &mockStoryService{}
	service := newTestService(repo, requests, storyService, Config{})

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"}}`,
			"state": map[string]any{"values": map[string]any{
				"team":     map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": blockedTeamID.String()}}},
				"title":    map[string]any{"value": map[string]any{"value": "Unauthorized task"}},
				"status":   map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": slackRequestStatusValue}}},
				"priority": map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": "High"}}},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), modalBlockTeam)
	require.Contains(t, string(resp.Body), "no longer available to you")
	require.Zero(t, requests.calls)
	require.Zero(t, storyService.createCalls)
}

func TestHandleViewSubmissionValidatesLabelsBeforeCreatingStory(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statusID := uuid.New()
	staleLabelID := uuid.New()
	repo := &mockRepo{
		workspace:   slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:       []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:    []slackrepository.StatusRecord{{ID: statusID, Name: "To Do", Category: "unstarted"}},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		labels:      []slackrepository.LabelRecord{{ID: uuid.New(), Name: "Current label"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	storyService := &mockStoryService{}
	service := newTestService(repo, &mockRequestStore{}, storyService, Config{})

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"}}`,
			"state": map[string]any{"values": map[string]any{
				"team":     map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": teamID.String()}}},
				"title":    map[string]any{"value": map[string]any{"value": "Task with stale label"}},
				"status":   map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": statusID.String()}}},
				"labels":   map[string]any{"value": map[string]any{"selected_options": []map[string]any{{"value": staleLabelID.String()}}}},
				"priority": map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": "High"}}},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), modalBlockLabels)
	require.Zero(t, storyService.createCalls)
}

func TestHandleViewSubmissionRejectsUnavailableScopedAssigneeAndObjective(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statusID := uuid.New()

	tests := []struct {
		name          string
		blockID       string
		actionID      string
		selectedValue uuid.UUID
		errorText     string
	}{
		{
			name:          "assignee",
			blockID:       modalBlockAssignee,
			actionID:      modalActionAssigneeSelect,
			selectedValue: uuid.New(),
			errorText:     "Selected assignee is no longer available",
		},
		{
			name:          "objective",
			blockID:       modalBlockObjective,
			actionID:      modalActionObjectiveSelect,
			selectedValue: uuid.New(),
			errorText:     "Selected objective is no longer available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{
				workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
				teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
				statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
					teamID: {{ID: statusID, Name: "To Do", Category: "unstarted"}},
				},
				teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
				slackWorkspace: slackrepository.SlackWorkspaceRecord{
					WorkspaceID:    workspaceID,
					SlackTeamID:    "T123",
					BotAccessToken: "xoxb-token",
				},
				slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
			}
			storyService := &mockStoryService{}
			service := newTestService(repo, &mockRequestStore{}, storyService, Config{})
			selectedBlockID := modalTeamScopedID(tt.blockID, teamID)
			selectedActionID := modalTeamScopedID(tt.actionID, teamID)
			interaction := map[string]any{
				"type": "view_submission",
				"team": map[string]any{"id": "T123"},
				"user": map[string]any{"id": "U123"},
				"view": map[string]any{
					"callback_id":      "fortyone_create_task",
					"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` + teamID.String() + `"}`,
					"state": map[string]any{"values": map[string]any{
						modalBlockTeam: map[string]any{
							modalActionTeamSelect: map[string]any{"selected_option": map[string]any{"value": teamID.String()}},
						},
						modalBlockTitle: map[string]any{
							modalActionTitleInput: map[string]any{"value": "Do not coerce selection"},
						},
						modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
							modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"selected_option": map[string]any{"value": statusID.String()}},
						},
						selectedBlockID: map[string]any{
							selectedActionID: map[string]any{"selected_option": map[string]any{"value": tt.selectedValue.String()}},
						},
					}},
				},
			}
			payloadBytes, err := json.Marshal(interaction)
			require.NoError(t, err)
			form := url.Values{}
			form.Set("payload", string(payloadBytes))

			response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			var responseBody struct {
				ResponseAction string            `json:"response_action"`
				Errors         map[string]string `json:"errors"`
			}
			require.NoError(t, json.Unmarshal(response.Body, &responseBody))
			require.Equal(t, "errors", responseBody.ResponseAction)
			require.Equal(t, map[string]string{selectedBlockID: tt.errorText}, responseBody.Errors)
			require.Zero(t, storyService.createCalls)
		})
	}
}

func TestBuildCreateTaskModalViewRefreshesTeamDependentFields(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	teamOneID := uuid.New()
	teamTwoID := uuid.New()
	teamTwoStatusID := uuid.New()
	teamTwoAssigneeID := uuid.New()
	teamTwoLabelID := uuid.New()
	teamTwoObjectiveID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamOneID, Code: "ENG", Name: "Engineering"},
			{ID: teamTwoID, Code: "OPS", Name: "Operations"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamOneID: {{ID: uuid.New(), Name: "Triage", Category: "unstarted"}},
			teamTwoID: {{ID: teamTwoStatusID, Name: "Done", Category: "completed"}},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamOneID: {{UserID: actorID, Username: "actor", FullName: "Slack Actor", Email: "actor@example.com"}},
			teamTwoID: {
				{UserID: actorID, Username: "actor", FullName: "Slack Actor", Email: "actor@example.com"},
				{UserID: teamTwoAssigneeID, Username: "ops-user", FullName: "Ops User", Email: "ops@example.com"},
			},
		},
		labelsByTeam: map[uuid.UUID][]slackrepository.LabelRecord{
			teamTwoID: {{ID: teamTwoLabelID, Name: "Operations"}},
		},
		objectivesByTeam: map[uuid.UUID][]slackrepository.ObjectiveRecord{
			teamTwoID: {{ID: teamTwoObjectiveID, Name: "Ship reliability"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
	}

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Ship release",
		Description: "Ready to ship",
		Source: requestSourceContext{
			SlackTeamID:     "T123",
			SlackTeamDomain: "acme",
			SlackChannelID:  "C123",
			SlackMessageTS:  "171234.000100",
		},
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection: createTaskModalSelection{
			TeamID:      teamTwoID,
			Priority:    "High",
			AssigneeID:  &teamTwoAssigneeID,
			LabelIDs:    []uuid.UUID{teamTwoLabelID},
			ObjectiveID: &teamTwoObjectiveID,
		},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	statusElement := findBlockElement(blocks, modalBlockStatus)
	statusOptions := statusElement["options"].([]map[string]any)
	require.Len(t, statusOptions, 2)
	require.Equal(t, slackRequestStatusValue, selectedOptionValue(t, statusOptions[0]))
	require.Equal(t, teamTwoStatusID.String(), selectedOptionValue(t, statusOptions[1]))

	assigneeElement := findBlockElement(blocks, modalBlockAssignee)
	require.Equal(t, "external_select", fmt.Sprint(assigneeElement["type"]))
	require.Equal(t, "2", fmt.Sprint(assigneeElement["min_query_length"]))
	initialAssignee := assigneeElement["initial_option"].(map[string]any)
	require.Equal(t, teamTwoAssigneeID.String(), selectedOptionValue(t, initialAssignee))

	labelsElement := findBlockElement(blocks, modalBlockLabels)
	require.Equal(t, "multi_external_select", fmt.Sprint(labelsElement["type"]))
	require.Equal(t, "2", fmt.Sprint(labelsElement["min_query_length"]))
	initialLabels := labelsElement["initial_options"].([]map[string]any)
	require.Len(t, initialLabels, 1)
	require.Equal(t, teamTwoLabelID.String(), selectedOptionValue(t, initialLabels[0]))

	objectiveElement := findBlockElement(blocks, modalBlockObjective)
	require.Equal(t, "external_select", fmt.Sprint(objectiveElement["type"]))
	initialObjective := objectiveElement["initial_option"].(map[string]any)
	require.Equal(t, teamTwoObjectiveID.String(), selectedOptionValue(t, initialObjective))
}

func TestBuildCreateTaskModalViewVersionsOnlyTeamDependentIDs(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	teamOneID := uuid.New()
	teamTwoID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamOneID, Code: "ENG", Name: "Engineering"},
			{ID: teamTwoID, Code: "OPS", Name: "Operations"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamOneID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
			teamTwoID: {{ID: uuid.New(), Name: "Triage", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	build := func(teamID uuid.UUID) []map[string]any {
		view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
			Title:       "Preserved title",
			Description: "Preserved description",
			WorkspaceID: workspaceID,
			ActorID:     actorID,
			Selection:   createTaskModalSelection{TeamID: teamID, Priority: "High"},
		})
		require.NoError(t, err)
		return view["blocks"].([]map[string]any)
	}

	teamOneBlocks := build(teamOneID)
	teamTwoBlocks := build(teamTwoID)
	teamTwoBlocksAgain := build(teamTwoID)

	for _, stableBlockID := range []string{
		modalBlockTeam,
		modalBlockTitle,
		modalBlockDescription,
		modalBlockPriority,
	} {
		teamOneBlock := findBlock(teamOneBlocks, stableBlockID)
		teamTwoBlock := findBlock(teamTwoBlocks, stableBlockID)
		require.Equal(t, stableBlockID, fmt.Sprint(teamOneBlock["block_id"]))
		require.Equal(t, fmt.Sprint(teamOneBlock["block_id"]), fmt.Sprint(teamTwoBlock["block_id"]))
		require.Equal(t, fmt.Sprint(teamOneBlock["element"].(map[string]any)["action_id"]), fmt.Sprint(teamTwoBlock["element"].(map[string]any)["action_id"]))
	}

	dependentIDs := []struct {
		blockID  string
		actionID string
	}{
		{blockID: modalBlockStatus, actionID: modalActionStatusSelect},
		{blockID: modalBlockAssignee, actionID: modalActionAssigneeSelect},
		{blockID: modalBlockLabels, actionID: modalActionLabelsMultiSelect},
		{blockID: modalBlockObjective, actionID: modalActionObjectiveSelect},
	}
	for _, dependent := range dependentIDs {
		teamOneBlock := findBlock(teamOneBlocks, dependent.blockID)
		teamTwoBlock := findBlock(teamTwoBlocks, dependent.blockID)
		teamTwoBlockAgain := findBlock(teamTwoBlocksAgain, dependent.blockID)

		teamOneBlockID := fmt.Sprint(teamOneBlock["block_id"])
		teamTwoBlockID := fmt.Sprint(teamTwoBlock["block_id"])
		teamOneActionID := fmt.Sprint(teamOneBlock["element"].(map[string]any)["action_id"])
		teamTwoActionID := fmt.Sprint(teamTwoBlock["element"].(map[string]any)["action_id"])

		require.Equal(t, modalTeamScopedID(dependent.blockID, teamOneID), teamOneBlockID)
		require.Equal(t, modalTeamScopedID(dependent.blockID, teamTwoID), teamTwoBlockID)
		require.Equal(t, modalTeamScopedID(dependent.actionID, teamOneID), teamOneActionID)
		require.Equal(t, modalTeamScopedID(dependent.actionID, teamTwoID), teamTwoActionID)
		require.NotEqual(t, teamOneBlockID, teamTwoBlockID)
		require.NotEqual(t, teamOneActionID, teamTwoActionID)
		require.Equal(t, teamTwoBlockID, fmt.Sprint(teamTwoBlockAgain["block_id"]))
		require.Equal(t, teamTwoActionID, fmt.Sprint(teamTwoBlockAgain["element"].(map[string]any)["action_id"]))
		require.LessOrEqual(t, len(teamTwoBlockID), modalElementIDMaxBytes)
		require.LessOrEqual(t, len(teamTwoActionID), modalElementIDMaxBytes)
	}
}

func TestModalTeamScopedIDIsDeterministicAndWithinSlackLimit(t *testing.T) {
	teamID := uuid.New()
	base := strings.Repeat("a", modalElementIDMaxBytes+50)

	first := modalTeamScopedID(base, teamID)
	second := modalTeamScopedID(base, teamID)

	require.Equal(t, first, second)
	require.Len(t, first, modalElementIDMaxBytes)
	require.True(t, strings.HasSuffix(first, modalTeamScopedIDSeparator+strings.ReplaceAll(teamID.String(), "-", "")))
}

func TestSlackOptionsRespectProviderLimits(t *testing.T) {
	option := toSlackOption(
		strings.Repeat("界", slackOptionTextMaxRunes+10),
		strings.Repeat("v", slackOptionValueMaxRunes+10),
	)
	require.Len(t, []rune(optionText(t, option)), slackOptionTextMaxRunes)
	require.Len(t, []rune(selectedOptionValue(t, option)), slackOptionValueMaxRunes)

	options := make([]map[string]any, 0, slackSelectMaxOptions+20)
	for index := 0; index < slackSelectMaxOptions+20; index++ {
		options = append(options, toSlackOption(fmt.Sprintf("Option %d", index), fmt.Sprintf("value-%d", index)))
	}
	block := selectInputBlock("provider_limit", "provider_limit_select", "Provider limit", options, nil, false, false)
	renderedOptions := block["element"].(map[string]any)["options"].([]map[string]any)
	require.Len(t, renderedOptions, slackSelectMaxOptions)

	response, err := interactionOptionsResponse(options)
	require.NoError(t, err)
	var responseBody struct {
		Options []map[string]any `json:"options"`
	}
	require.NoError(t, json.Unmarshal(response.Body, &responseBody))
	require.Len(t, responseBody.Options, slackSelectMaxOptions)
}

func TestBuildCreateTaskModalViewListsOnlyActorsTeams(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	allowedTeamID := uuid.New()
	blockedTeamID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: allowedTeamID, Code: "ENG", Name: "Engineering"},
			{ID: blockedTeamID, Code: "OPS", Name: "Operations"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			allowedTeamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			allowedTeamID: {{UserID: actorID}},
			blockedTeamID: {{UserID: uuid.New()}},
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection:   createTaskModalSelection{TeamID: blockedTeamID},
	})
	require.NoError(t, err)

	teamElement := findBlockElement(view["blocks"].([]map[string]any), modalBlockTeam)
	require.Equal(t, "external_select", teamElement["type"])
	require.NotContains(t, teamElement, "options")
	require.Equal(t, slackExternalSearchMinRunes, teamElement["min_query_length"])
	require.Equal(t, allowedTeamID.String(), selectedOptionValue(t, teamElement["initial_option"]))
}

func TestBuildCreateTaskModalViewKeepsAuthorizedTeamsReachableBeyondStaticLimit(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	teams := make([]slackrepository.TeamRecord, 0, slackSelectMaxOptions+25)
	for index := 0; index < slackSelectMaxOptions+25; index++ {
		teams = append(teams, slackrepository.TeamRecord{
			ID:   uuid.New(),
			Code: fmt.Sprintf("T%03d", index),
			Name: fmt.Sprintf("Authorized Team %03d", index),
		})
	}
	selectedTeam := teams[len(teams)-1]
	repo := &mockRepo{
		teams:       teams,
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection:   createTaskModalSelection{TeamID: selectedTeam.ID},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	teamBlock := findBlock(blocks, modalBlockTeam)
	teamElement := teamBlock["element"].(map[string]any)
	require.Equal(t, "external_select", teamElement["type"])
	require.Equal(t, slackExternalSearchMinRunes, teamElement["min_query_length"])
	require.NotContains(t, teamElement, "options")
	require.Equal(t, true, teamBlock["dispatch_action"])
	require.Equal(t, selectedTeam.ID.String(), selectedOptionValue(t, teamElement["initial_option"]))
}

func TestBuildCreateTaskModalViewShowsRequestAsFirstSyntheticStatus(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	toDoStatusID := uuid.New()
	actorID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamID, Code: "ENG", Name: "Engineering"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: toDoStatusID, Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection: createTaskModalSelection{
			TeamID: teamID,
		},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	statusElement := findBlockElement(blocks, modalBlockStatus)
	statusOptions := statusElement["options"].([]map[string]any)
	require.Len(t, statusOptions, 2)
	require.Equal(t, slackRequestStatusValue, selectedOptionValue(t, statusOptions[0]))
	require.Equal(t, "Request", optionText(t, statusOptions[0]))
	require.Equal(t, toDoStatusID.String(), selectedOptionValue(t, statusOptions[1]))
	require.Equal(t, "To Do", optionText(t, statusOptions[1]))
	require.Equal(t, toDoStatusID.String(), selectedOptionValue(t, statusElement["initial_option"]))
}

func TestBuildCreateTaskModalViewUsesSearchableStatusWhenStaticLimitExceeded(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statuses := make([]slackrepository.StatusRecord, 0, slackSelectMaxOptions)
	for index := 0; index < slackSelectMaxOptions; index++ {
		statuses = append(statuses, slackrepository.StatusRecord{
			ID:       uuid.New(),
			Name:     fmt.Sprintf("Workflow Status %03d", index),
			Category: "unstarted",
		})
	}
	selectedStatus := statuses[len(statuses)-1]
	repo := &mockRepo{
		teams:       []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:    statuses,
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection: createTaskModalSelection{
			TeamID:     teamID,
			StatusKind: slackStatusKindStory,
			StatusID:   &selectedStatus.ID,
		},
	})
	require.NoError(t, err)

	statusElement := findBlockElement(view["blocks"].([]map[string]any), modalBlockStatus)
	require.Equal(t, "external_select", statusElement["type"])
	require.Equal(t, slackExternalSearchMinRunes, statusElement["min_query_length"])
	require.NotContains(t, statusElement, "options")
	require.Equal(t, selectedStatus.ID.String(), selectedOptionValue(t, statusElement["initial_option"]))
}

func TestBuildCreateTaskModalViewUsesStaticStatusAtProviderBoundary(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statuses := make([]slackrepository.StatusRecord, 0, slackSelectMaxOptions-1)
	for index := 0; index < slackSelectMaxOptions-1; index++ {
		statuses = append(statuses, slackrepository.StatusRecord{
			ID:       uuid.New(),
			Name:     fmt.Sprintf("Workflow Status %03d", index),
			Category: "unstarted",
		})
	}
	repo := &mockRepo{
		teams:       []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:    statuses,
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection:   createTaskModalSelection{TeamID: teamID},
	})
	require.NoError(t, err)

	statusElement := findBlockElement(view["blocks"].([]map[string]any), modalBlockStatus)
	require.Equal(t, "static_select", statusElement["type"])
	require.Len(t, statusElement["options"], slackSelectMaxOptions)
	require.Equal(t, statuses[0].ID.String(), selectedOptionValue(t, statusElement["initial_option"]))
}

func TestBuildCreateTaskModalViewRendersExternalOptionalSelects(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamID, Code: "ENG", Name: "Engineering"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection: createTaskModalSelection{
			TeamID: teamID,
		},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	assigneeElement := findBlockElement(blocks, modalBlockAssignee)
	require.Equal(t, "external_select", fmt.Sprint(assigneeElement["type"]))

	labelsElement := findBlockElement(blocks, modalBlockLabels)
	require.Equal(t, "multi_external_select", fmt.Sprint(labelsElement["type"]))

	objectiveElement := findBlockElement(blocks, modalBlockObjective)
	require.Equal(t, "external_select", fmt.Sprint(objectiveElement["type"]))
}

func TestHandleBlockActionsResetsTeamDependentSelections(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	oldTeamID := uuid.New()
	newTeamID := uuid.New()
	oldStatusID := uuid.New()
	newStatusID := uuid.New()
	oldAssigneeID := uuid.New()
	oldLabelID := uuid.New()
	oldObjectiveID := uuid.New()

	baseRepo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: oldTeamID, Code: "ENG", Name: "Engineering"},
			{ID: newTeamID, Code: "OPS", Name: "Operations"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			oldTeamID: {{ID: oldStatusID, Name: "In Progress", Category: "started"}},
			newTeamID: {{ID: newStatusID, Name: "To Do", Category: "unstarted"}},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			oldTeamID: {
				{UserID: actorID},
				{UserID: oldAssigneeID},
			},
			newTeamID: {{UserID: actorID}},
		},
		labelsByTeam: map[uuid.UUID][]slackrepository.LabelRecord{
			oldTeamID: {{ID: oldLabelID, Name: "Bug"}},
		},
		objectivesByTeam: map[uuid.UUID][]slackrepository.ObjectiveRecord{
			oldTeamID: {{ID: oldObjectiveID, Name: "Reliability"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	repo := &blockingSlackWorkspaceRepo{
		mockRepo: baseRepo,
		started:  make(chan struct{}, 1),
		release:  make(chan struct{}),
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	var updateRequest struct {
		View struct {
			Blocks []map[string]any `json:"blocks"`
		} `json:"view"`
	}
	updateResult := make(chan error, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://slack.com/api/views.update" {
			updateResult <- fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		} else {
			updateResult <- json.NewDecoder(req.Body).Decode(&updateRequest)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
		}, nil
	})}

	metadata, err := json.Marshal(slackModalPrivateMetadata{
		Source: requestSourceContext{
			SlackTeamID: "T123",
			SlackUserID: "U123",
		},
		SelectedTeamID: oldTeamID.String(),
	})
	require.NoError(t, err)
	interaction := map[string]any{
		"type": "block_actions",
		"team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"id":               "V123",
			"hash":             "hash",
			"callback_id":      "fortyone_create_task",
			"private_metadata": string(metadata),
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": newTeamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"value": "Preserved title"}},
					"description": map[string]any{"value": map[string]any{"value": "Preserved description"}},
					modalTeamScopedID(modalBlockStatus, oldTeamID): map[string]any{
						modalTeamScopedID(modalActionStatusSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": oldStatusID.String()}},
					},
					"priority": map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": "High"}}},
					modalTeamScopedID(modalBlockAssignee, oldTeamID): map[string]any{
						modalTeamScopedID(modalActionAssigneeSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": oldAssigneeID.String()}},
					},
					modalTeamScopedID(modalBlockLabels, oldTeamID): map[string]any{
						modalTeamScopedID(modalActionLabelsMultiSelect, oldTeamID): map[string]any{"selected_options": []map[string]any{{"value": oldLabelID.String()}}},
					},
					modalTeamScopedID(modalBlockObjective, oldTeamID): map[string]any{
						modalTeamScopedID(modalActionObjectiveSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": oldObjectiveID.String()}},
					},
				},
			},
		},
		"actions": []map[string]any{{
			"block_id":  modalBlockTeam,
			"action_id": modalActionTeamSelect,
			"selected_option": map[string]any{
				"value": newTeamID.String(),
			},
		}},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	requestCtx, cancelRequest := context.WithCancel(context.Background())
	resp, err := service.HandleInteractivity(requestCtx, []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for asynchronous Slack block action")
	}
	cancelRequest()
	close(repo.release)
	select {
	case updateErr := <-updateResult:
		require.NoError(t, updateErr)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for asynchronous Slack view update")
	}

	require.Equal(t, "Preserved title", findBlockElement(updateRequest.View.Blocks, modalBlockTitle)["initial_value"])
	require.Equal(t, "Preserved description", findBlockElement(updateRequest.View.Blocks, modalBlockDescription)["initial_value"])
	require.Equal(t, "High", selectedOptionValue(t, findBlockElement(updateRequest.View.Blocks, modalBlockPriority)["initial_option"]))
	require.Equal(t, newTeamID.String(), selectedOptionValue(t, findBlockElement(updateRequest.View.Blocks, modalBlockTeam)["initial_option"]))
	require.Equal(t, newStatusID.String(), selectedOptionValue(t, findBlockElement(updateRequest.View.Blocks, modalBlockStatus)["initial_option"]))
	require.NotContains(t, findBlockElement(updateRequest.View.Blocks, modalBlockAssignee), "initial_option")
	require.NotContains(t, findBlockElement(updateRequest.View.Blocks, modalBlockLabels), "initial_options")
	require.NotContains(t, findBlockElement(updateRequest.View.Blocks, modalBlockObjective), "initial_option")
	for _, dependentBlockID := range []string{modalBlockStatus, modalBlockAssignee, modalBlockLabels, modalBlockObjective} {
		updatedBlockID := fmt.Sprint(findBlock(updateRequest.View.Blocks, dependentBlockID)["block_id"])
		require.Equal(t, modalTeamScopedID(dependentBlockID, newTeamID), updatedBlockID)
		require.NotEqual(t, modalTeamScopedID(dependentBlockID, oldTeamID), updatedBlockID)
	}
}

func TestHandleMessageActionAcknowledgesBeforeWorkAndSurvivesRequestCancellation(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	baseRepo := &mockRepo{
		teams: []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	repo := &blockingSlackWorkspaceRepo{
		mockRepo: baseRepo,
		started:  make(chan struct{}, 1),
		release:  make(chan struct{}),
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	modalOpened := make(chan error, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://slack.com/api/views.open":
			modalOpened <- req.Context().Err()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"view":{"id":"V123"}}`)),
				Header:     make(http.Header),
			}, nil
		case "https://slack.com/api/views.update":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
			}, nil
		default:
			modalOpened <- fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
	})}

	interaction := map[string]any{
		"type":         "message_action",
		"trigger_id":   "trigger",
		"response_url": "https://hooks.slack.com/actions/T123/message",
		"team":         map[string]any{"id": "T123", "domain": "acme"},
		"channel":      map[string]any{"id": "C123", "name": "general"},
		"user":         map[string]any{"id": "U123", "username": "joseph"},
		"message":      map[string]any{"user": "U456", "text": "Ship the Slack integration", "ts": "171234.000100"},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	type interactionResult struct {
		response InteractionResponse
		err      error
	}
	result := make(chan interactionResult, 1)
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	go func() {
		response, err := service.HandleInteractivity(requestCtx, []byte(form.Encode()))
		result <- interactionResult{response: response, err: err}
	}()
	select {
	case interactionResponse := <-result:
		require.NoError(t, interactionResponse.err)
		require.Equal(t, http.StatusOK, interactionResponse.response.StatusCode)
	case <-time.After(250 * time.Millisecond):
		close(repo.release)
		t.Fatal("message action was not acknowledged before workspace lookup completed")
	}
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("timed out waiting for asynchronous message action")
	}
	cancelRequest()
	close(repo.release)

	select {
	case openErr := <-modalOpened:
		require.NoError(t, openErr)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message-action modal after request cancellation")
	}
}

func TestOpenCreateTaskModalConsumesTriggerBeforeTeamLookups(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &blockingTeamListRepo{
		mockRepo: &mockRepo{
			teams: []slackrepository.TeamRecord{{ID: teamID, Code: "WEB", Name: "Web"}},
			statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
				teamID: {{ID: uuid.New(), Name: "Todo", Category: "unstarted"}},
			},
			teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		},
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	modalOpened := make(chan struct{}, 1)
	modalUpdated := make(chan struct{}, 1)
	requestErrors := make(chan error, 2)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var payload map[string]any
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			requestErrors <- err
		}
		switch req.URL.String() {
		case "https://slack.com/api/views.open":
			view, _ := payload["view"].(map[string]any)
			if _, hasSubmit := view["submit"]; hasSubmit {
				requestErrors <- errors.New("loading modal must not expose submit before hydration")
			}
			modalOpened <- struct{}{}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"view":{"id":"V-loading"}}`)),
				Header:     make(http.Header),
			}, nil
		case "https://slack.com/api/views.update":
			if payload["view_id"] != "V-loading" {
				requestErrors <- fmt.Errorf("updated view id = %#v", payload["view_id"])
			}
			view, _ := payload["view"].(map[string]any)
			if _, hasSubmit := view["submit"]; !hasSubmit {
				requestErrors <- errors.New("hydrated modal is missing submit")
			}
			modalUpdated <- struct{}{}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
			}, nil
		default:
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
	})}

	result := make(chan error, 1)
	go func() {
		result <- service.openCreateTaskModal(
			context.Background(),
			"trigger",
			"Ship it",
			"Created from Slack",
			requestSourceContext{SlackTeamID: "T123", SlackUserID: "U123"},
			workspaceID,
			actorID,
			"xoxb-token",
		)
	}()

	select {
	case <-modalOpened:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("loading modal was not opened before team lookup")
	}
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("team lookup did not start after opening the loading modal")
	}
	close(repo.release)
	select {
	case err := <-result:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timed out hydrating create story modal")
	}
	select {
	case <-modalUpdated:
	default:
		t.Fatal("create story modal was not hydrated")
	}
	close(requestErrors)
	for err := range requestErrors {
		require.NoError(t, err)
	}
}

func TestHandleMessageActionPostsPrivateFailureFeedback(t *testing.T) {
	repo := &mockRepo{err: errors.New("database unavailable")}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	feedback := make(chan CommandResponse, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://hooks.slack.com/actions/T123/message" {
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
		var response CommandResponse
		if err := json.NewDecoder(req.Body).Decode(&response); err != nil {
			return nil, err
		}
		feedback <- response
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})}

	interaction := map[string]any{
		"type":         "message_action",
		"trigger_id":   "trigger",
		"response_url": "https://hooks.slack.com/actions/T123/message",
		"team":         map[string]any{"id": "T123"},
		"channel":      map[string]any{"id": "C123"},
		"user":         map[string]any{"id": "U123"},
		"message":      map[string]any{"user": "U456", "text": "Create a task", "ts": "171234.000100"},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	select {
	case failure := <-feedback:
		require.Equal(t, "ephemeral", failure.ResponseType)
		require.Contains(t, failure.Text, "could not update")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for private message-action failure feedback")
	}
}

func TestHandleBlockActionsRejectsTeamOutsideActorsMembership(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	allowedTeamID := uuid.New()
	blockedTeamID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: allowedTeamID, Code: "ENG", Name: "Engineering"},
			{ID: blockedTeamID, Code: "OPS", Name: "Operations"},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			allowedTeamID: {{UserID: actorID}},
			blockedTeamID: {{UserID: uuid.New()}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	apiCalls := 0
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		apiCalls++
		return nil, errors.New("views.update must not be called")
	})}

	metadata, err := json.Marshal(slackModalPrivateMetadata{
		Source:         requestSourceContext{SlackTeamID: "T123", SlackUserID: "U123"},
		SelectedTeamID: allowedTeamID.String(),
	})
	require.NoError(t, err)
	interaction := map[string]any{
		"type": "block_actions",
		"team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": string(metadata),
			"state": map[string]any{"values": map[string]any{
				"team":  map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": blockedTeamID.String()}}},
				"title": map[string]any{"value": map[string]any{"value": "Unauthorized task"}},
			}},
		},
		"actions": []map[string]any{{
			"block_id":        modalBlockTeam,
			"action_id":       modalActionTeamSelect,
			"selected_option": map[string]any{"value": blockedTeamID.String()},
		}},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	var payload interactionPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &payload))
	resp, err := service.handleBlockActions(context.Background(), payload)
	require.ErrorIs(t, err, ErrSlackTeamNotAvailable)
	require.Zero(t, resp.StatusCode)
	require.Zero(t, apiCalls)
}

func TestHandleCommandRespondsEvenWhenOpeningModalFails(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks: map[string]uuid.UUID{
			"T123:U123": actorID,
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	failureResponse := make(chan CommandResponse, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == "https://slack.com/api/views.open" {
			return nil, errors.New("slack api unavailable")
		}
		if req.URL.String() != "https://hooks.slack.com/actions/T123/response" {
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
		var response CommandResponse
		if err := json.NewDecoder(req.Body).Decode(&response); err != nil {
			return nil, err
		}
		failureResponse <- response
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})}

	form := url.Values{}
	form.Set("team_id", "T123")
	form.Set("team_domain", "acme")
	form.Set("channel_id", "C123")
	form.Set("channel_name", "general")
	form.Set("user_id", "U123")
	form.Set("user_name", "joseph")
	form.Set("trigger_id", "trigger")
	form.Set("text", "create task Ship it")
	form.Set("response_url", "https://hooks.slack.com/actions/T123/response")

	resp, err := service.HandleCommand(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Empty(t, resp.ResponseType)
	require.Empty(t, resp.Text)
	select {
	case failure := <-failureResponse:
		require.Equal(t, "ephemeral", failure.ResponseType)
		require.Contains(t, failure.Text, "Unable to open")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for asynchronous slash-command failure feedback")
	}
}

func TestHandleCommandAcknowledgesBeforeWorkAndSurvivesRequestCancellation(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	baseRepo := &mockRepo{
		teams: []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	repo := &blockingSlackWorkspaceRepo{
		mockRepo: baseRepo,
		started:  make(chan struct{}, 1),
		release:  make(chan struct{}),
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	modalOpened := make(chan error, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://slack.com/api/views.open":
			modalOpened <- req.Context().Err()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"view":{"id":"V123"}}`)),
				Header:     make(http.Header),
			}, nil
		case "https://slack.com/api/views.update":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
			}, nil
		default:
			modalOpened <- fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
	})}

	form := url.Values{}
	form.Set("team_id", "T123")
	form.Set("channel_id", "C123")
	form.Set("user_id", "U123")
	form.Set("trigger_id", "trigger")
	form.Set("text", "create task Ship it")
	form.Set("response_url", "https://hooks.slack.com/actions/T123/response")

	type commandResult struct {
		response CommandResponse
		err      error
	}
	result := make(chan commandResult, 1)
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	go func() {
		response, err := service.HandleCommand(requestCtx, []byte(form.Encode()))
		result <- commandResult{response: response, err: err}
	}()

	select {
	case command := <-result:
		require.NoError(t, command.err)
		require.Empty(t, command.response.ResponseType)
		require.Empty(t, command.response.Text)
	case <-time.After(250 * time.Millisecond):
		close(repo.release)
		t.Fatal("slash command was not acknowledged before workspace lookup completed")
	}
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("timed out waiting for asynchronous slash-command work")
	}
	cancelRequest()
	close(repo.release)

	select {
	case openErr := <-modalOpened:
		require.NoError(t, openErr)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for modal open after request cancellation")
	}
}

func TestHandleCommandPromptsAccountLinkWhenSlackUserIsUnmapped(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		WebsiteURL: "https://fortyone.app",
		SecretKey:  "test-secret",
	})
	type connectResult struct {
		response CommandResponse
		err      error
	}
	connectResponse := make(chan connectResult, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var response CommandResponse
		var resultErr error
		if req.URL.String() != "https://hooks.slack.com/actions/T123/connect" {
			resultErr = fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		} else {
			resultErr = json.NewDecoder(req.Body).Decode(&response)
		}
		connectResponse <- connectResult{response: response, err: resultErr}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})}

	form := url.Values{}
	form.Set("team_id", "T123")
	form.Set("team_domain", "acme")
	form.Set("channel_id", "C123")
	form.Set("channel_name", "general")
	form.Set("user_id", "U456")
	form.Set("user_name", "joseph")
	form.Set("trigger_id", "trigger")
	form.Set("text", "create task Ship it")
	form.Set("response_url", "https://hooks.slack.com/actions/T123/connect")

	resp, err := service.HandleCommand(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Empty(t, resp.ResponseType)
	require.Empty(t, resp.Text)
	select {
	case connect := <-connectResponse:
		require.NoError(t, connect.err)
		require.Equal(t, "ephemeral", connect.response.ResponseType)
		require.Contains(t, connect.response.Text, "Connect FortyOne account")
		require.Contains(t, connect.response.Text, "slack_link_token=")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for asynchronous Slack account-link prompt")
	}
}

func TestPostSlackCreationAckPersistsSlackTeamBinding(t *testing.T) {
	store := newEventStoreStub()
	workspaceID := uuid.New()
	installGeneration := uuid.New()
	service := newTestService(&mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: installGeneration,
		IsActive:          true,
	}}, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://slack.com/api/chat.postMessage" {
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	messageTS := service.postSlackCreationAck(
		context.Background(),
		workspaceID,
		installGeneration,
		"slack:view:V1:confirmation",
		requestSourceContext{SlackTeamID: "T1", SlackChannelID: "C1", SlackMessageTS: "171.100"},
		"xoxb-token",
		"Task created in FortyOne.",
	)

	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "171.200", messageTS)
	require.Equal(t, "T1", store.outboundInputs[0].ExternalWorkspaceID)
	require.Equal(t, installGeneration, *store.outboundInputs[0].InstallGeneration)
	require.Len(t, store.completedDeliveries, 1)
}

func TestPostSlackCreationAckRoutesToSourceConversation(t *testing.T) {
	tests := []struct {
		name         string
		source       requestSourceContext
		wantChannel  string
		wantThreadTS string
	}{
		{
			name: "source channel root",
			source: requestSourceContext{
				SlackTeamID:    "T1",
				SlackChannelID: "C1",
				SlackMessageTS: "171.100",
				SlackUserID:    "U1",
			},
			wantChannel: "C1",
		},
		{
			name: "existing source thread",
			source: requestSourceContext{
				SlackTeamID:    "T1",
				SlackChannelID: "C1",
				SlackMessageTS: "171.200",
				SlackThreadTS:  "171.100",
				SlackUserID:    "U1",
			},
			wantChannel:  "C1",
			wantThreadTS: "171.100",
		},
		{
			name: "user fallback without source channel",
			source: requestSourceContext{
				SlackTeamID:   "T1",
				SlackThreadTS: "171.100",
				SlackUserID:   "U1",
			},
			wantChannel: "U1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newEventStoreStub()
			workspaceID := uuid.New()
			installGeneration := uuid.New()
			service := newTestService(&mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
				WorkspaceID:       workspaceID,
				SlackTeamID:       "T1",
				InstallGeneration: installGeneration,
				IsActive:          true,
			}}, &mockRequestStore{}, &mockStoryService{}, Config{})
			service.outbound = store
			service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, "https://slack.com/api/chat.postMessage", req.URL.String())
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.300"}`)),
					Header:     make(http.Header),
				}, nil
			})}

			service.postSlackCreationAck(
				context.Background(),
				workspaceID,
				installGeneration,
				"slack:view:V1:confirmation:"+test.name,
				test.source,
				"xoxb-token",
				"Joseph Mukorivo created WEB-123",
			)

			require.Len(t, store.outboundInputs, 1)
			require.Equal(t, test.wantChannel, store.outboundInputs[0].ExternalChannelID)
			require.Equal(t, test.wantThreadTS, store.outboundInputs[0].ExternalThreadID)
		})
	}
}

func TestBuildSlackStoryCreatedText(t *testing.T) {
	storyCode := buildStoryCode(" web ", 123)
	require.Equal(t, "WEB-123", storyCode)
	require.Equal(
		t,
		"Joseph Mukorivo created <https://acme.fortyone.app/work/WEB-123|WEB-123>",
		buildSlackStoryCreatedText("Joseph Mukorivo", storyCode, "https://acme.fortyone.app/work/WEB-123"),
	)
	require.Equal(
		t,
		"Joseph &lt;@U123&gt; &amp; Team created WEB-123",
		buildSlackStoryCreatedText("Joseph <@U123> & Team", storyCode, ""),
	)
	fallback := buildSlackStoryCreatedText(" ", "", "https://acme.fortyone.app/work/story-id")
	require.Equal(t, "A team member created <https://acme.fortyone.app/work/story-id|a story>", fallback)
	require.NotContains(t, fallback, "✅")
	require.NotContains(t, fallback, "in FortyOne")
}

func TestBuildSlackRequestLifecycleText(t *testing.T) {
	require.Equal(
		t,
		"Joseph Mukorivo <https://acme.fortyone.app/teams/6ba7b812-9dad-11d1-80b4-00c04fd430c8/requests/6ba7b813-9dad-11d1-80b4-00c04fd430c8|opened a request>",
		buildSlackRequestOpenedText("Joseph Mukorivo", "https://acme.fortyone.app/teams/6ba7b812-9dad-11d1-80b4-00c04fd430c8/requests/6ba7b813-9dad-11d1-80b4-00c04fd430c8"),
	)
	require.Equal(
		t,
		"Joseph &lt;@U123&gt; &amp; Team opened a request",
		buildSlackRequestOpenedText("Joseph <@U123> & Team", ""),
	)
	require.Equal(
		t,
		"Joseph Mukorivo linked a request to <https://acme.fortyone.app/work/WEB-123|WEB-123>",
		buildSlackStoryLinkedRequestText("Joseph Mukorivo", "WEB-123", "https://acme.fortyone.app/work/WEB-123"),
	)
}

func TestHandleMutationActionRechecksScopeAndReplacesButtonsWithLinkedReceipt(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T1",
			BotAccessToken:    "xoxb-token",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			FullName: "Joseph Mukorivo",
		}},
	}
	confirmer := &mutationConfirmerStub{result: messaging.StoryMutationResult{
		Status:    "applied",
		Operation: messaging.StoryMutationCreate,
		StoryID:   uuid.New(),
		Reference: "web-123",
		TeamID:    teamID,
		Title:     "Fix login",
	}}
	var providerRequest map[string]any
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/chat.update", request.URL.Path)
		require.NoError(t, json.NewDecoder(request.Body).Decode(&providerRequest))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	WithMutationConfirmer(confirmer)(service)
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions",
		"team":{"id":"T1"},
		"user":{"id":"U1","name":"joseph"},
		"channel":{"id":"C1"},
		"message":{"ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"opaque-token"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "opaque-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	response, err := service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, []string{"opaque-token"}, confirmer.tokens)
	require.Len(t, confirmer.scopes, 1)
	require.Equal(t, workspaceID, confirmer.scopes[0].WorkspaceID)
	require.Equal(t, actorID, confirmer.scopes[0].UserID)
	require.Equal(t, []uuid.UUID{teamID}, confirmer.scopes[0].AllowedTeamIDs)
	require.True(t, confirmer.scopes[0].AllowMutations)
	require.Equal(t, "Joseph Mukorivo created <https://acme.fortyone.app/work/WEB-123|WEB-123>", providerRequest["text"])
	require.Empty(t, providerRequest["blocks"])
}

func TestHandleMutationActionRendersItemizedBatchReceipt(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:             uuid.New(),
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T1",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			FullName: "Joseph Mukorivo",
		}},
	}
	confirmer := &mutationConfirmerStub{result: messaging.StoryMutationResult{
		Status:    "applied",
		Operation: messaging.StoryMutationCreateBatch,
		TeamID:    teamID,
		Items: []messaging.StoryMutationItemResult{
			{Index: 0, Status: "applied", StoryID: uuid.New(), Reference: "WEB-123", TeamID: teamID, Title: "Add Microsoft auth", Priority: "High"},
			{Index: 1, Status: "applied", StoryID: uuid.New(), Reference: "WEB-124", TeamID: teamID, Title: "Add TikTok", Priority: "No Priority"},
		},
	}}
	var providerRequest map[string]any
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/chat.update", request.URL.Path)
		require.NoError(t, json.NewDecoder(request.Body).Decode(&providerRequest))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	WithMutationConfirmer(confirmer)(service)
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions","team":{"id":"T1"},"user":{"id":"U1","name":"joseph"},
		"channel":{"id":"C1"},"message":{"ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"placeholder"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "opaque-batch-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	response, err := service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "Joseph Mukorivo created 2 stories:\n• <https://acme.fortyone.app/work/WEB-123|WEB-123> — Add Microsoft auth\n• <https://acme.fortyone.app/work/WEB-124|WEB-124> — Add TikTok", providerRequest["text"])
	require.Empty(t, providerRequest["blocks"])
}

func TestHandleMutationActionKeepsRetryControlsAndShowsPartialBatchProgress(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:             uuid.New(),
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T1",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			FullName: "Joseph Mukorivo",
		}},
	}
	confirmer := &mutationConfirmerStub{
		result: messaging.StoryMutationResult{
			Status:    "partial",
			Operation: messaging.StoryMutationCreateBatch,
			TeamID:    teamID,
			Items: []messaging.StoryMutationItemResult{{
				Index: 0, Status: "applied", StoryID: uuid.New(), Reference: "WEB-123",
				TeamID: teamID, Title: "Add Microsoft auth", Priority: "High",
			}},
		},
		err: errors.New("temporary story provider failure"),
	}
	var providerRequest map[string]any
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/chat.update", request.URL.Path)
		require.NoError(t, json.NewDecoder(request.Body).Decode(&providerRequest))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	WithMutationConfirmer(confirmer)(service)
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions","team":{"id":"T1"},"user":{"id":"U1","name":"joseph"},
		"channel":{"id":"C1"},"message":{"ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"placeholder"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "opaque-batch-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	response, err := service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(
		t,
		"Joseph Mukorivo created 1 of the proposed stories before FortyOne hit an error:\n• <https://acme.fortyone.app/work/WEB-123|WEB-123> — Add Microsoft auth\nSelect *Retry remaining* to try again. Already-created stories will not be duplicated.",
		providerRequest["text"],
	)
	blocks, ok := providerRequest["blocks"].([]any)
	require.True(t, ok)
	require.Len(t, blocks, 2, "partial progress must retain the retry and cancel controls")
	actions, ok := blocks[1].(map[string]any)
	require.True(t, ok)
	elements, ok := actions["elements"].([]any)
	require.True(t, ok)
	require.Len(t, elements, 1)
	confirm, ok := elements[0].(map[string]any)
	require.True(t, ok)
	buttonText, ok := confirm["text"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Retry remaining", buttonText["text"])
	encodedValue, ok := confirm["value"].(string)
	require.True(t, ok)
	decodedValue, err := decodeSlackMutationActionValue(encodedValue)
	require.NoError(t, err)
	require.Equal(t, "opaque-batch-token", decodedValue.Token)
}

func TestHandleMutationActionIgnoresLegacyDisabledSettings(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T1",
			BotAccessToken:    "xoxb-token",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			FullName: "Joseph Mukorivo",
		}},
		agentSettings: slackrepository.AgentSettingsRecord{
			Guidance: "configured",
		},
	}
	confirmer := &mutationConfirmerStub{result: messaging.StoryMutationResult{
		Status:    "applied",
		Operation: messaging.StoryMutationCreate,
		StoryID:   uuid.New(),
		Reference: "WEB-123",
		TeamID:    teamID,
		Title:     "Fix login",
	}}
	var providerRequest map[string]any
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.NoError(t, json.NewDecoder(request.Body).Decode(&providerRequest))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	WithMutationConfirmer(confirmer)(service)
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions","team":{"id":"T1"},"user":{"id":"U1"},
		"channel":{"id":"C1"},"message":{"ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"opaque-token"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "opaque-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	_, err = service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, []string{"opaque-token"}, confirmer.tokens)
	require.Contains(t, providerRequest["text"], "created")
	require.Empty(t, providerRequest["blocks"])
}

func TestAcceptIntegrationRequestPostsLinkerAndCanonicalStoryCode(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			Username: "joseph",
			FullName: "Joseph Mukorivo",
			Email:    "joseph@example.com",
		}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T1",
			BotAccessToken:    "xoxb-token",
			InstallGeneration: installGeneration,
			IsActive:          true,
		},
	}
	store := newEventStoreStub()
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/chat.postMessage", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	err := service.AcceptIntegrationRequest(context.Background(), integrationrequests.CoreIntegrationRequest{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		TeamID:      teamID,
		Provider:    integrationrequests.ProviderSlack,
		Metadata: map[string]any{
			"slack_channel_id": "C1",
			"slack_message_ts": "171.100",
			"slack_team_id":    "T1",
			"workspace_slug":   "acme",
			"team_code":        "web",
		},
	}, stories.CoreSingleStory{
		ID:         uuid.New(),
		SequenceID: 123,
		Title:      "Fix login bug",
		Team:       teamID,
		Reporter:   &actorID,
		CreatedAt:  time.Unix(1_700_000_000, 0),
		UpdatedAt:  time.Unix(1_700_000_000, 0),
	})

	require.NoError(t, err)
	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "Joseph Mukorivo linked a request to <https://acme.fortyone.app/work/WEB-123|WEB-123>", store.outboundInputs[0].Content)
	require.Equal(t, "C1", store.outboundInputs[0].ExternalChannelID)
	require.Equal(t, "171.100", store.outboundInputs[0].ExternalThreadID)
	providerPayload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.Len(t, providerPayload.Metadata.Entities, 1)
	entity := providerPayload.Metadata.Entities[0]
	require.Equal(t, "https://acme.fortyone.app/work/WEB-123", entity.URL)
	require.NotContains(t, entity.EntityPayload.Fields, "description")
	require.NotContains(t, entity.EntityPayload.Fields, "created_by")
	require.NotContains(t, entity.EntityPayload.Fields, "date_created")
	require.NotContains(t, entity.EntityPayload.Fields, "date_updated")
	require.False(t, *providerPayload.UnfurlLinks)
	require.False(t, *providerPayload.UnfurlMedia)
}

func TestPostSlackTaskAckFallbackKeepsLifecycleCopyAndSuppressesClassicUnfurls(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	reporterID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: installGeneration,
		IsActive:          true,
	}}
	store := newEventStoreStub()
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	service.postSlackTaskAck(
		context.Background(),
		workspaceID,
		installGeneration,
		"slack:request:fallback",
		requestSourceContext{SlackTeamID: "T1", SlackChannelID: "C1", SlackMessageTS: "171.100"},
		"xoxb-token",
		"acme",
		"WEB",
		"Joseph Mukorivo",
		"",
		slackStoryReceiptActionLinkedRequest,
		stories.CoreSingleStory{ID: uuid.New(), SequenceID: 123, Team: teamID, Reporter: &reporterID},
	)

	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "Joseph Mukorivo linked a request to <https://acme.app.example.com/work/WEB-123|WEB-123>", store.outboundInputs[0].Content)
	providerPayload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.Nil(t, providerPayload.Metadata)
	require.NotNil(t, providerPayload.UnfurlLinks)
	require.False(t, *providerPayload.UnfurlLinks)
	require.NotNil(t, providerPayload.UnfurlMedia)
	require.False(t, *providerPayload.UnfurlMedia)
}

func TestPostSlackCreationAckCancelsWhenInstallationGenerationChanges(t *testing.T) {
	store := newEventStoreStub()
	workspaceID := uuid.New()
	originalGeneration := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("stale creation acknowledgement reached Slack")
		return nil, errors.New("unexpected provider call")
	})}

	service.postSlackCreationAck(
		context.Background(),
		workspaceID,
		originalGeneration,
		"slack:view:V1:confirmation",
		requestSourceContext{SlackTeamID: "T1", SlackChannelID: "C1", SlackMessageTS: "171.100"},
		"xoxb-old-token",
		"Task created in FortyOne.",
	)

	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, originalGeneration, *store.outboundInputs[0].InstallGeneration)
	require.Len(t, store.cancelledDeliveries, 1)
	require.Empty(t, store.completedDeliveries)
}

func TestLinkSlackAccountCreatesManualMapping(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{
			ID:   workspaceID,
			Slug: "acme",
			Name: "Acme",
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		WebsiteURL: "https://fortyone.app",
	})

	link, err := service.buildSlackUserLinkURL(context.Background(), workspaceID, "T123", "U999")
	require.NoError(t, err)
	parsedLink, err := url.Parse(link)
	require.NoError(t, err)
	require.Equal(t, "acme.fortyone.app", parsedLink.Host)
	require.Equal(t, "/settings/integrations/slack", parsedLink.Path)
	token := parsedLink.Query().Get("slack_link_token")
	require.NotEmpty(t, token)
	require.NotContains(t, token, ".")
	nonce, err := base64.RawURLEncoding.DecodeString(token)
	require.NoError(t, err)
	digest := sha256.Sum256(nonce)
	storeKey := nonceStoreKey(slackProviderMessaging, slackNoncePurposeAccount, digest[:])
	store := service.nonces.(*mockNonceStore)

	err = service.LinkSlackAccount(context.Background(), workspaceID, userID, token)
	require.NoError(t, err)

	require.NotNil(t, repo.slackUserLinks)
	require.Equal(t, userID, repo.slackUserLinks["T123:U999"])
	require.NotNil(t, store.records[storeKey].UserID)
	require.Equal(t, userID, *store.records[storeKey].UserID)
	require.Error(t, service.LinkSlackAccount(context.Background(), workspaceID, userID, token))
}

func TestParseCommandTitleSupportsCreateTaskPrefix(t *testing.T) {
	title := parseCommandTitle("create task Improve onboarding")
	require.Equal(t, "Improve onboarding", title)

	title = parseCommandTitle("Improve onboarding")
	require.Equal(t, "Improve onboarding", title)

	title = parseCommandTitle("")
	require.Equal(t, "New task", title)
}

func TestParseViewSubmissionReadsTeamScopedDependentIDs(t *testing.T) {
	teamID := uuid.New()
	statusID := uuid.New()
	assigneeID := uuid.New()
	labelID := uuid.New()
	objectiveID := uuid.New()
	interaction := map[string]any{
		"view": map[string]any{
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` + teamID.String() + `"}`,
			"state": map[string]any{"values": map[string]any{
				modalBlockTeam: map[string]any{
					modalActionTeamSelect: map[string]any{"selected_option": map[string]any{"value": teamID.String()}},
				},
				modalBlockTitle: map[string]any{
					modalActionTitleInput: map[string]any{"value": "Scoped task"},
				},
				modalBlockDescription: map[string]any{
					modalActionDescriptionInput: map[string]any{"value": "Scoped description"},
				},
				modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
					modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"selected_option": map[string]any{"value": statusID.String()}},
				},
				modalTeamScopedID(modalBlockAssignee, teamID): map[string]any{
					modalTeamScopedID(modalActionAssigneeSelect, teamID): map[string]any{"selected_option": map[string]any{"value": assigneeID.String()}},
				},
				modalTeamScopedID(modalBlockLabels, teamID): map[string]any{
					modalTeamScopedID(modalActionLabelsMultiSelect, teamID): map[string]any{"selected_options": []map[string]any{{"value": labelID.String()}}},
				},
				modalTeamScopedID(modalBlockObjective, teamID): map[string]any{
					modalTeamScopedID(modalActionObjectiveSelect, teamID): map[string]any{"selected_option": map[string]any{"value": objectiveID.String()}},
				},
				modalBlockPriority: map[string]any{
					modalActionPrioritySelect: map[string]any{"selected_option": map[string]any{"value": "High"}},
				},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	var payload interactionPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &payload))

	submission, err := parseViewSubmission(payload)
	require.NoError(t, err)
	require.Equal(t, "Scoped task", submission.Title)
	require.Equal(t, "Scoped description", submission.Description)
	require.Equal(t, teamID, submission.TeamID)
	require.Equal(t, slackStatusKindStory, submission.StatusKind)
	require.Equal(t, statusID, *submission.StatusID)
	require.Equal(t, assigneeID, *submission.AssigneeID)
	require.Equal(t, []uuid.UUID{labelID}, submission.LabelIDs)
	require.Equal(t, objectiveID, *submission.ObjectiveID)
	require.Equal(t, "High", submission.Priority)
	require.Equal(t, modalTeamScopedID(modalBlockStatus, teamID), submission.BlockIDs.Status)
	require.Equal(t, modalTeamScopedID(modalBlockAssignee, teamID), submission.BlockIDs.Assignee)
	require.Equal(t, modalTeamScopedID(modalBlockLabels, teamID), submission.BlockIDs.Labels)
	require.Equal(t, modalTeamScopedID(modalBlockObjective, teamID), submission.BlockIDs.Objective)
}

func TestParseViewSubmissionIgnoresStaleDependentIDsFromPreviousTeam(t *testing.T) {
	oldTeamID := uuid.New()
	newTeamID := uuid.New()
	interaction := map[string]any{
		"view": map[string]any{
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` + oldTeamID.String() + `"}`,
			"state": map[string]any{"values": map[string]any{
				modalBlockTeam: map[string]any{
					modalActionTeamSelect: map[string]any{"selected_option": map[string]any{"value": newTeamID.String()}},
				},
				modalBlockTitle: map[string]any{
					modalActionTitleInput: map[string]any{"value": "Preserved title"},
				},
				modalTeamScopedID(modalBlockStatus, oldTeamID): map[string]any{
					modalTeamScopedID(modalActionStatusSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": uuid.NewString()}},
				},
				modalTeamScopedID(modalBlockAssignee, oldTeamID): map[string]any{
					modalTeamScopedID(modalActionAssigneeSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": uuid.NewString()}},
				},
				modalTeamScopedID(modalBlockLabels, oldTeamID): map[string]any{
					modalTeamScopedID(modalActionLabelsMultiSelect, oldTeamID): map[string]any{"selected_options": []map[string]any{{"value": uuid.NewString()}}},
				},
				modalTeamScopedID(modalBlockObjective, oldTeamID): map[string]any{
					modalTeamScopedID(modalActionObjectiveSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": uuid.NewString()}},
				},
				modalBlockPriority: map[string]any{
					modalActionPrioritySelect: map[string]any{"selected_option": map[string]any{"value": "Urgent"}},
				},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	var payload interactionPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &payload))

	submission, err := parseViewSubmission(payload)
	require.NoError(t, err)
	require.Equal(t, newTeamID, submission.TeamID)
	require.Equal(t, "Preserved title", submission.Title)
	require.Equal(t, "Urgent", submission.Priority)
	require.Equal(t, slackStatusKindRequest, submission.StatusKind)
	require.Nil(t, submission.StatusID)
	require.Nil(t, submission.AssigneeID)
	require.Empty(t, submission.LabelIDs)
	require.Nil(t, submission.ObjectiveID)
	require.Equal(t, modalTeamScopedID(modalBlockStatus, newTeamID), submission.BlockIDs.Status)
	require.Equal(t, modalTeamScopedID(modalBlockAssignee, newTeamID), submission.BlockIDs.Assignee)
	require.Equal(t, modalTeamScopedID(modalBlockLabels, newTeamID), submission.BlockIDs.Labels)
	require.Equal(t, modalTeamScopedID(modalBlockObjective, newTeamID), submission.BlockIDs.Objective)
}

func TestHandleViewSubmissionDefaultsClearedStatusToRequest(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installedBy := uuid.New()
	actorID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: actorID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}

	requests := &mockRequestStore{}
	service := newTestService(repo, requests, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"user": map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme","slack_channel_id":"C123","slack_message_ts":"171234.000100"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": teamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Fix login bug"}},
					"description": map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "User cannot log in from iOS"}},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), `"response_action":"clear"`)
	require.Equal(t, integrationrequests.ProviderSlack, requests.last.Provider)
	require.Equal(t, SourceTypeSlackMessage, requests.last.SourceType)
	require.Equal(t, teamID, requests.last.TeamID)
}

func TestBuildWorkspaceURLSupportsSubdomainsAndLocalhost(t *testing.T) {
	t.Run("hosted_url_uses_workspace_subdomain", func(t *testing.T) {
		integrationURL := buildWorkspaceURL("https://fortyone.app", "acme", "settings", "workspace", "integrations", "slack")
		require.Equal(t, "https://acme.fortyone.app/settings/workspace/integrations/slack", integrationURL)

		accountIntegrationURL := buildWorkspaceURL("https://fortyone.app", "acme", "settings", "integrations", "slack")
		require.Equal(t, "https://acme.fortyone.app/settings/integrations/slack", accountIntegrationURL)

		taskURL := buildTaskURL("https://fortyone.app", "acme", "PRD-571")
		require.Equal(t, "https://acme.fortyone.app/work/PRD-571", taskURL)

		requestURL := buildRequestURL("https://fortyone.app", "acme", "team-1", "req-1")
		require.Equal(t, "https://acme.fortyone.app/teams/team-1/requests/req-1", requestURL)
	})

	t.Run("localhost_url_uses_workspace_path_prefix", func(t *testing.T) {
		integrationURL := buildWorkspaceURL("http://localhost:3000", "acme", "settings", "workspace", "integrations", "slack")
		require.Equal(t, "http://localhost:3000/acme/settings/workspace/integrations/slack", integrationURL)

		accountIntegrationURL := buildWorkspaceURL("http://localhost:3000", "acme", "settings", "integrations", "slack")
		require.Equal(t, "http://localhost:3000/acme/settings/integrations/slack", accountIntegrationURL)
	})
}

func TestBuildPrefilledDescriptionUsesLinearStyleFormat(t *testing.T) {
	description := buildPrefilledDescription(requestSourceContext{
		SlackUserID:   "U12345",
		SlackUsername: "joseph",
		SlackText:     "hey",
	})
	require.Equal(t, "@[joseph](U12345) said:\n> hey", description)
}

func TestBuildCreateTaskModalViewMarksDescriptionOptional(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection:   createTaskModalSelection{TeamID: teamID},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	descriptionBlock := findBlock(blocks, modalBlockDescription)
	require.Equal(t, true, descriptionBlock["optional"])
}

func TestBuildCreateTaskModalViewPreservesPersonalTeamOrdering(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	firstTeam := slackrepository.TeamRecord{ID: uuid.New(), Code: "WEB", Name: "Web"}
	secondTeam := slackrepository.TeamRecord{ID: uuid.New(), Code: "ENG", Name: "Engineering"}
	repo := &mockRepo{
		// ListWorkspaceTeamsForUser returns the repository's personal order.
		// Keep it deliberately different from alphabetical order.
		teams:       []slackrepository.TeamRecord{firstTeam, secondTeam},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			firstTeam.ID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Ordered story",
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Source: requestSourceContext{
			SlackTeamID:    "T123",
			SlackChannelID: "C123",
		},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	teamElement := findBlockElement(blocks, modalBlockTeam)
	require.Equal(t, firstTeam.ID.String(), selectedOptionValue(t, teamElement["initial_option"]))

	options := slackTeamSuggestionOptions([]slackrepository.TeamRecord{firstTeam, secondTeam}, "")
	require.Len(t, options, 2)
	require.Equal(t, firstTeam.ID.String(), selectedOptionValue(t, options[0]))
	require.Equal(t, secondTeam.ID.String(), selectedOptionValue(t, options[1]))
}

func TestHandleInteractivityBlockSuggestionSearchesAllAuthorizedTeams(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	blockedTeamID := uuid.New()
	teams := make([]slackrepository.TeamRecord, 0, slackSelectMaxOptions+22)
	for index := 0; index < slackSelectMaxOptions+21; index++ {
		teams = append(teams, slackrepository.TeamRecord{
			ID:   uuid.New(),
			Code: fmt.Sprintf("T%03d", index),
			Name: fmt.Sprintf("Authorized Team %03d", index),
		})
	}
	targetTeam := teams[len(teams)-1]
	teams = append(teams, slackrepository.TeamRecord{ID: blockedTeamID, Code: "PRIVATE", Name: "Restricted Team"})
	repo := &mockRepo{
		teams:       teams,
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			blockedTeamID: {{UserID: uuid.New()}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	search := func(query string) []map[string]any {
		t.Helper()
		interaction := map[string]any{
			"type":      "block_suggestion",
			"action_id": modalActionTeamSelect,
			"block_id":  modalBlockTeam,
			"value":     query,
			"team":      map[string]any{"id": "T123"},
			"user":      map[string]any{"id": "U123"},
			"view": map[string]any{
				"callback_id":      "fortyone_create_task",
				"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"}}`,
			},
		}
		payloadBytes, err := json.Marshal(interaction)
		require.NoError(t, err)
		form := url.Values{}
		form.Set("payload", string(payloadBytes))
		response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		var responseBody struct {
			Options []map[string]any `json:"options"`
		}
		require.NoError(t, json.Unmarshal(response.Body, &responseBody))
		return responseBody.Options
	}

	require.Len(t, search(""), slackSelectMaxOptions)
	targetOptions := search(targetTeam.Code)
	require.Len(t, targetOptions, 1)
	require.Equal(t, targetTeam.ID.String(), selectedOptionValue(t, targetOptions[0]))
	require.Empty(t, search("Restricted Team"))
}

func TestHandleInteractivityBlockSuggestionSearchesOverflowStatuses(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statuses := make([]slackrepository.StatusRecord, 0, slackSelectMaxOptions+20)
	for index := 0; index < slackSelectMaxOptions+20; index++ {
		statuses = append(statuses, slackrepository.StatusRecord{
			ID:       uuid.New(),
			Name:     fmt.Sprintf("Workflow Status %03d", index),
			Category: "unstarted",
		})
	}
	targetStatus := statuses[len(statuses)-1]
	repo := &mockRepo{
		teams:    []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses: statuses,
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: actorID},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	search := func(query string) []map[string]any {
		t.Helper()
		interaction := map[string]any{
			"type":      "block_suggestion",
			"action_id": modalTeamScopedID(modalActionStatusSelect, teamID),
			"block_id":  modalTeamScopedID(modalBlockStatus, teamID),
			"value":     query,
			"team":      map[string]any{"id": "T123"},
			"user":      map[string]any{"id": "U123"},
			"view": map[string]any{
				"callback_id": "fortyone_create_task",
				"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` +
					teamID.String() + `"}`,
			},
		}
		payloadBytes, err := json.Marshal(interaction)
		require.NoError(t, err)
		form := url.Values{}
		form.Set("payload", string(payloadBytes))
		response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		var responseBody struct {
			Options []map[string]any `json:"options"`
		}
		require.NoError(t, json.Unmarshal(response.Body, &responseBody))
		return responseBody.Options
	}

	initialOptions := search("")
	require.Len(t, initialOptions, slackSelectMaxOptions)
	require.Equal(t, slackRequestStatusValue, selectedOptionValue(t, initialOptions[0]))
	targetOptions := search(targetStatus.Name)
	require.Len(t, targetOptions, 1)
	require.Equal(t, targetStatus.ID.String(), selectedOptionValue(t, targetOptions[0]))
}

func TestHandleInteractivityBlockSuggestionReturnsTeamScopedAssigneeOptions(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installedBy := uuid.New()
	memberID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamID: {{UserID: memberID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": memberID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalTeamScopedID(modalActionAssigneeSelect, teamID),
		"block_id":  modalTeamScopedID(modalBlockAssignee, teamID),
		"value":     "jo",
		"team":      map[string]any{"id": "T123", "domain": "acme"},
		"user":      map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team": map[string]any{
						"value": map[string]any{
							"type":            "static_select",
							"selected_option": map[string]any{"value": teamID.String()},
						},
					},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.ContentType)
	require.Contains(t, string(resp.Body), memberID.String())
}

func TestHandleInteractivityBlockSuggestionUsesScopedActionTeamWhenMetadataIsStale(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	oldTeamID := uuid.New()
	selectedTeamID := uuid.New()
	selectedMemberID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: oldTeamID, Code: "OLD", Name: "Old Team"},
			{ID: selectedTeamID, Code: "NEW", Name: "Selected Team"},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			oldTeamID: {{UserID: actorID, FullName: "Slack Actor"}},
			selectedTeamID: {
				{UserID: actorID, FullName: "Slack Actor"},
				{UserID: selectedMemberID, FullName: "Joseph Selected"},
			},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalTeamScopedID(modalActionAssigneeSelect, selectedTeamID),
		"block_id":  modalTeamScopedID(modalBlockAssignee, selectedTeamID),
		"value":     "jo",
		"team":      map[string]any{"id": "T123"},
		"user":      map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id": "fortyone_create_task",
			// Slack can preserve metadata from the prior view while dispatching
			// from the newly rendered, team-scoped input.
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` +
				oldTeamID.String() + `"}`,
			"state": map[string]any{"values": map[string]any{}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Contains(t, string(response.Body), selectedMemberID.String())
}

func TestHandleInteractivityBlockSuggestionRejectsTeamOutsideActorsMembership(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	allowedTeamID := uuid.New()
	blockedTeamID := uuid.New()
	blockedMemberID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: allowedTeamID, Code: "ENG", Name: "Engineering"},
			{ID: blockedTeamID, Code: "OPS", Name: "Operations"},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			allowedTeamID: {{UserID: actorID, FullName: "Slack Actor"}},
			blockedTeamID: {{UserID: blockedMemberID, FullName: "Joseph Blocked"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"value":     "jo",
		"team":      map[string]any{"id": "T123"},
		"user":      map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"}}`,
			"state": map[string]any{"values": map[string]any{
				"team": map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": blockedTeamID.String()}}},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.JSONEq(t, `{"options":[]}`, string(resp.Body))
	require.NotContains(t, string(resp.Body), blockedMemberID.String())
}

func TestHandleInteractivityBlockSuggestionUsesViewInitialTeamWhenStateEmpty(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installedBy := uuid.New()
	memberID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamID: {{UserID: memberID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": memberID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"block_id":  modalBlockAssignee,
		"value":     "jo",
		"team":      map[string]any{"id": "T123", "domain": "acme"},
		"user":      map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme"}`,
			"blocks": []map[string]any{
				{
					"block_id": modalBlockTeam,
					"element": map[string]any{
						"type":      "static_select",
						"action_id": modalActionTeamSelect,
						"initial_option": map[string]any{
							"value": teamID.String(),
						},
					},
				},
			},
			"state": map[string]any{
				"values": map[string]any{},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), memberID.String())
}

func TestHandleInteractivityBlockSuggestionReturnsNoOptionsBeforeTwoCharacters(t *testing.T) {
	teamID := uuid.New()
	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"value":     "j",
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team": map[string]any{
						"value": map[string]any{
							"type":            "static_select",
							"selected_option": map[string]any{"value": teamID.String()},
						},
					},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	service := newTestService(&mockRepo{}, &mockRequestStore{}, &mockStoryService{}, Config{})
	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.JSONEq(t, `{"options":[]}`, string(resp.Body))
}

func TestHandleInteractivityBlockSuggestionUsesActionFallbackFromActionsArray(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	memberID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamID: {{UserID: memberID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:     workspaceID,
			SlackTeamID:     "T123",
			SlackTeamDomain: "acme",
			BotAccessToken:  "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": memberID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type": "block_suggestion",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"user": map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme"}`,
			"blocks": []map[string]any{
				{
					"block_id": modalBlockTeam,
					"element": map[string]any{
						"type":      "static_select",
						"action_id": modalActionTeamSelect,
						"initial_option": map[string]any{
							"value": teamID.String(),
						},
					},
				},
			},
			"state": map[string]any{"values": map[string]any{}},
		},
		"actions": []map[string]any{
			{
				"action_id": modalActionAssigneeSelect,
				"value":     "jo",
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), memberID.String())
}

func TestHandleInteractivityBlockSuggestionUsesSelectedTeamFromPrivateMetadata(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	memberID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamID: {{UserID: memberID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:     workspaceID,
			SlackTeamID:     "T123",
			SlackTeamDomain: "acme",
			BotAccessToken:  "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": memberID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"value":     "jo",
		"team":      map[string]any{"id": "T123", "domain": "acme"},
		"user":      map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id": "fortyone_create_task",
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_team_domain":"acme"},"selected_team_id":"` +
				teamID.String() + `"}`,
			"state": map[string]any{"values": map[string]any{}},
		},
	}

	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), memberID.String())
}

func slackSignature(secret, timestamp string, body []byte) string {
	base := "v0:" + timestamp + ":" + string(body)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(base))
	return "v0=" + hex.EncodeToString(h.Sum(nil))
}

func findBlockElement(blocks []map[string]any, blockID string) map[string]any {
	for _, block := range blocks {
		actualBlockID := fmt.Sprint(block["block_id"])
		if actualBlockID == blockID || strings.HasPrefix(actualBlockID, blockID+modalTeamScopedIDSeparator) {
			return block["element"].(map[string]any)
		}
	}
	return map[string]any{}
}

func findBlock(blocks []map[string]any, blockID string) map[string]any {
	for _, block := range blocks {
		actualBlockID := fmt.Sprint(block["block_id"])
		if actualBlockID == blockID || strings.HasPrefix(actualBlockID, blockID+modalTeamScopedIDSeparator) {
			return block
		}
	}
	return map[string]any{}
}

func selectedOptionValue(t *testing.T, raw any) string {
	t.Helper()
	option := raw.(map[string]any)
	return fmt.Sprint(option["value"])
}

func optionText(t *testing.T, raw any) string {
	t.Helper()
	option := raw.(map[string]any)
	switch text := option["text"].(type) {
	case map[string]any:
		return fmt.Sprint(text["text"])
	case map[string]string:
		return text["text"]
	default:
		return fmt.Sprint(option["text"])
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var _ http.RoundTripper = roundTripFunc(nil)
