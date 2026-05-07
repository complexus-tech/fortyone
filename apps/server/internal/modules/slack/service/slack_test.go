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
	workspace      slackrepository.WorkspaceRecord
	team           slackrepository.TeamRecord
	teams          []slackrepository.TeamRecord
	statuses       []slackrepository.StatusRecord
	statusesByTeam map[uuid.UUID][]slackrepository.StatusRecord
	teamMembers    []slackrepository.TeamMemberRecord
	membersByTeam  map[uuid.UUID][]slackrepository.TeamMemberRecord
	labels         []slackrepository.LabelRecord
	labelsByTeam   map[uuid.UUID][]slackrepository.LabelRecord
	slackWorkspace slackrepository.SlackWorkspaceRecord
	err            error
	disconnected   bool
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
	require.Equal(t, "Urgent", storyService.lastStory.Priority)
	require.Equal(t, []uuid.UUID{labelID}, storyService.lastLabelIDs)
}

func TestBuildCreateTaskModalViewRefreshesTeamDependentFields(t *testing.T) {
	workspaceID := uuid.New()
	teamOneID := uuid.New()
	teamTwoID := uuid.New()
	teamTwoStatusID := uuid.New()
	teamTwoAssigneeID := uuid.New()
	teamTwoLabelID := uuid.New()

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
			TeamID:   teamTwoID,
			Priority: "High",
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
	assigneeOptions := assigneeElement["options"].([]map[string]any)
	require.Len(t, assigneeOptions, 1)
	require.Equal(t, teamTwoAssigneeID.String(), selectedOptionValue(t, assigneeOptions[0]))

	labelsElement := findBlockElement(blocks, modalBlockLabels)
	labelOptions := labelsElement["options"].([]map[string]any)
	require.Len(t, labelOptions, 1)
	require.Equal(t, teamTwoLabelID.String(), selectedOptionValue(t, labelOptions[0]))
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

func TestBuildCreateTaskModalViewOmitsEmptyOptionalSelects(t *testing.T) {
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
	require.Empty(t, findBlockElement(blocks, modalBlockAssignee))
	require.Empty(t, findBlockElement(blocks, modalBlockLabels))
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
