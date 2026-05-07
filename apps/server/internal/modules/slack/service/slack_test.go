package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

type mockRepo struct {
	workspace        slackrepository.WorkspaceRecord
	team             slackrepository.TeamRecord
	teams            []slackrepository.TeamRecord
	statuses         []slackrepository.StatusRecord
	statusesByTeam   map[uuid.UUID][]slackrepository.StatusRecord
	teamMembers      []slackrepository.TeamMemberRecord
	membersByTeam    map[uuid.UUID][]slackrepository.TeamMemberRecord
	labels           []slackrepository.LabelRecord
	labelsByTeam     map[uuid.UUID][]slackrepository.LabelRecord
	objectives       []slackrepository.ObjectiveRecord
	objectivesByTeam map[uuid.UUID][]slackrepository.ObjectiveRecord
	slackWorkspace   slackrepository.SlackWorkspaceRecord
	err              error
	disconnected     bool
	lastStoryLink    struct {
		storyID uuid.UUID
		title   string
		url     string
	}
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
	return slackrepository.ObjectiveRecord{}, errors.New("objective not found")
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
	if m.err != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.err
	}
	return m.slackWorkspace, nil
}

func (m *mockRepo) GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackWorkspaceRecord, error) {
	if m.err != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.err
	}
	return m.slackWorkspace, nil
}

func (m *mockRepo) GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackrepository.SlackWorkspaceRecord, error) {
	if m.err != nil {
		return slackrepository.SlackWorkspaceRecord{}, m.err
	}
	return m.slackWorkspace, nil
}

func (m *mockRepo) DisconnectSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	m.disconnected = true
	m.slackWorkspace = slackrepository.SlackWorkspaceRecord{}
	return nil
}

func (m *mockRepo) UpsertChannels(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, channels []slackrepository.SlackChannelPayload) error {
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

func (m *mockRepo) CreateStoryLink(ctx context.Context, storyID uuid.UUID, title, linkURL string) error {
	if m.err != nil {
		return m.err
	}
	m.lastStoryLink.storyID = storyID
	m.lastStoryLink.title = title
	m.lastStoryLink.url = linkURL
	return nil
}

func (m *mockRepo) InsertRequestLog(ctx context.Context, entry slackrepository.SlackRequestLogInsert) error {
	return m.err
}

func (m *mockRepo) ListRequestLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]slackrepository.SlackRequestLogRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []slackrepository.SlackRequestLogRecord{}, nil
}

type mockRequestStore struct {
	last integrationrequests.CoreUpsertRequestInput
}

func (m *mockRequestStore) UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error) {
	m.last = input
	id := uuid.New()
	return integrationrequests.CoreIntegrationRequest{ID: id}, nil
}

type mockStoryService struct {
	lastActorID   uuid.UUID
	lastWorkspace uuid.UUID
	lastStory     stories.CoreNewStory
	lastLabelIDs  []uuid.UUID
}

func (m *mockStoryService) CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	m.lastActorID = actorID
	m.lastWorkspace = workspaceID
	m.lastStory = ns
	return stories.CoreSingleStory{ID: uuid.New(), Title: ns.Title}, nil
}

func (m *mockStoryService) UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error {
	m.lastLabelIDs = append([]uuid.UUID(nil), labels...)
	return nil
}

func newTestService(repo Repository, requests RequestStore, storyService StoryService, cfg Config) *Service {
	testLogger := logger.NewWithJSON(io.Discard, slog.LevelError, "test")
	service := New(testLogger, repo, requests, storyService, cfg)
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

func TestDisconnectWorkspaceDeletesSlackWorkspace(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			IsActive:    true,
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	err := service.DisconnectWorkspace(context.Background(), workspaceID)

	require.NoError(t, err)
	require.True(t, repo.disconnected)
	require.Equal(t, slackrepository.SlackWorkspaceRecord{}, repo.slackWorkspace)
}

func TestHandleViewSubmissionCreatesSlackRequestWhenRequestStatusSelected(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installeBy := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:  []slackrepository.StatusRecord{{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installeBy,
		},
	}
	requests := &mockRequestStore{}
	service := newTestService(repo, requests, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme","slack_channel_id":"C123","slack_message_ts":"171234.000100"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": teamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Fix login bug"}},
					"description": map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "User cannot log in from iOS"}},
					"status":      map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": slackRequestStatusValue}}},
					"priority":    map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": "High"}}},
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
	require.NotNil(t, requests.last.SourceURL)
	require.True(t, strings.Contains(*requests.last.SourceURL, "acme.slack.com/archives/C123"))
}

func TestHandleViewSubmissionCreatesStoryWhenNonTriageStatusSelected(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	triageStatusID := uuid.New()
	doneStatusID := uuid.New()
	installedBy := uuid.New()
	assigneeID := uuid.New()
	labelID := uuid.New()
	objectiveID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses: []slackrepository.StatusRecord{
			{ID: triageStatusID, Name: "Triage", Category: "unstarted"},
			{ID: doneStatusID, Name: "Done", Category: "completed"},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: assigneeID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		labels:      []slackrepository.LabelRecord{{ID: labelID, Name: "Bug"}},
		objectives:  []slackrepository.ObjectiveRecord{{ID: objectiveID, Name: "Improve reliability"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
		},
	}
	requests := &mockRequestStore{}
	storyService := &mockStoryService{}
	service := newTestService(repo, requests, storyService, Config{WebsiteURL: "https://app.example.com"})

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme","slack_channel_id":"C123","slack_message_ts":"171234.000100"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": teamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Ship release"}},
					"description": map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Ready to ship"}},
					"status":      map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": doneStatusID.String()}}},
					"assignee":    map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": assigneeID.String()}}},
					"labels":      map[string]any{"value": map[string]any{"type": "multi_static_select", "selected_options": []map[string]any{{"value": labelID.String()}}}},
					"objective":   map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": objectiveID.String()}}},
					"priority":    map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": "Urgent"}}},
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
	require.Equal(t, teamID, storyService.lastStory.Team)
	require.NotNil(t, storyService.lastStory.Status)
	require.Equal(t, doneStatusID, *storyService.lastStory.Status)
	require.NotNil(t, storyService.lastStory.Assignee)
	require.Equal(t, assigneeID, *storyService.lastStory.Assignee)
	require.NotNil(t, storyService.lastStory.Objective)
	require.Equal(t, objectiveID, *storyService.lastStory.Objective)
	require.Equal(t, "Urgent", storyService.lastStory.Priority)
	require.Equal(t, []uuid.UUID{labelID}, storyService.lastLabelIDs)
	require.Contains(t, repo.lastStoryLink.url, "https://acme.slack.com/archives/C123/")
}

func TestBuildCreateTaskModalViewRefreshesTeamDependentFields(t *testing.T) {
	workspaceID := uuid.New()
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
			teamTwoID: {{UserID: teamTwoAssigneeID, Username: "ops-user", FullName: "Ops User", Email: "ops@example.com"}},
		},
		labelsByTeam: map[uuid.UUID][]slackrepository.LabelRecord{
			teamTwoID: {{ID: teamTwoLabelID, Name: "Operations"}},
		},
		objectivesByTeam: map[uuid.UUID][]slackrepository.ObjectiveRecord{
			teamTwoID: {{ID: teamTwoObjectiveID, Name: "Ship reliability"}},
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

func TestBuildCreateTaskModalViewShowsRequestAsFirstSyntheticStatus(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	toDoStatusID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamID, Code: "ENG", Name: "Engineering"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: toDoStatusID, Name: "To Do", Category: "unstarted"}},
		},
	}

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
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
}

func TestBuildCreateTaskModalViewRendersExternalOptionalSelects(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamID, Code: "ENG", Name: "Engineering"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
	}

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
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

func TestHandleCommandRespondsEvenWhenOpeningModalFails(t *testing.T) {
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
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("slack api unavailable")
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

	resp, err := service.HandleCommand(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, "ephemeral", resp.ResponseType)
	require.Contains(t, resp.Text, "Opening")
}

func TestParseCommandTitleSupportsCreateTaskPrefix(t *testing.T) {
	title := parseCommandTitle("create task Improve onboarding")
	require.Equal(t, "Improve onboarding", title)

	title = parseCommandTitle("Improve onboarding")
	require.Equal(t, "Improve onboarding", title)

	title = parseCommandTitle("")
	require.Equal(t, "New task", title)
}

func TestHandleViewSubmissionDefaultsClearedStatusToRequest(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installedBy := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
		},
	}

	requests := &mockRequestStore{}
	service := newTestService(repo, requests, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123", "domain": "acme"},
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

		taskURL := buildTaskURL("https://fortyone.app", "acme", "story-123")
		require.Equal(t, "https://acme.fortyone.app/story/story-123", taskURL)

		requestURL := buildRequestURL("https://fortyone.app", "acme", "team-1", "req-1")
		require.Equal(t, "https://acme.fortyone.app/teams/team-1/requests/req-1", requestURL)
	})

	t.Run("localhost_url_uses_workspace_path_prefix", func(t *testing.T) {
		integrationURL := buildWorkspaceURL("http://localhost:3000", "acme", "settings", "workspace", "integrations", "slack")
		require.Equal(t, "http://localhost:3000/acme/settings/workspace/integrations/slack", integrationURL)
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
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
		Selection:   createTaskModalSelection{TeamID: teamID},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	descriptionBlock := findBlock(blocks, modalBlockDescription)
	require.Equal(t, true, descriptionBlock["optional"])
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
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"block_id":  modalBlockAssignee,
		"value":     "jo",
		"team":      map[string]any{"id": "T123", "domain": "acme"},
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

func slackSignature(secret, timestamp string, body []byte) string {
	base := "v0:" + timestamp + ":" + string(body)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(base))
	return "v0=" + hex.EncodeToString(h.Sum(nil))
}

func findBlockElement(blocks []map[string]any, blockID string) map[string]any {
	for _, block := range blocks {
		if fmt.Sprint(block["block_id"]) == blockID {
			return block["element"].(map[string]any)
		}
	}
	return map[string]any{}
}

func findBlock(blocks []map[string]any, blockID string) map[string]any {
	for _, block := range blocks {
		if fmt.Sprint(block["block_id"]) == blockID {
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
