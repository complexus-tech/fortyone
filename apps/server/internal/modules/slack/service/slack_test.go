package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	slackWorkspace slackrepository.SlackWorkspaceRecord
	settings       slackrepository.SlackWorkspaceSettingsRecord
	err            error
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
	return m.team, nil
}

func (m *mockRepo) GetWorkspaceBySlackTeamID(ctx context.Context, slackTeamID string) (slackrepository.WorkspaceRecord, error) {
	if m.err != nil {
		return slackrepository.WorkspaceRecord{}, m.err
	}
	return m.workspace, nil
}

func (m *mockRepo) GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackWorkspaceSettingsRecord, error) {
	if m.err != nil {
		return slackrepository.SlackWorkspaceSettingsRecord{}, m.err
	}
	return m.settings, nil
}

func (m *mockRepo) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, defaultCreateMode string) (slackrepository.SlackWorkspaceSettingsRecord, error) {
	if m.err != nil {
		return slackrepository.SlackWorkspaceSettingsRecord{}, m.err
	}
	m.settings.DefaultCreateMode = defaultCreateMode
	return m.settings, nil
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

func (m *mockRepo) UpsertChannels(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, channels []slackrepository.SlackChannelPayload) error {
	return m.err
}

func (m *mockRepo) ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.SlackChannelRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []slackrepository.SlackChannelRecord{}, nil
}

func (m *mockRepo) UpsertChannelLink(ctx context.Context, workspaceID uuid.UUID, slackChannelID string, teamID, createdByUserID uuid.UUID) (slackrepository.SlackChannelLinkRecord, error) {
	if m.err != nil {
		return slackrepository.SlackChannelLinkRecord{}, m.err
	}
	return slackrepository.SlackChannelLinkRecord{}, nil
}

func (m *mockRepo) GetChannelLinkByID(ctx context.Context, workspaceID, linkID uuid.UUID) (slackrepository.SlackChannelLinkRecord, error) {
	if m.err != nil {
		return slackrepository.SlackChannelLinkRecord{}, m.err
	}
	return slackrepository.SlackChannelLinkRecord{}, nil
}

func (m *mockRepo) ListChannelLinks(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.SlackChannelLinkRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []slackrepository.SlackChannelLinkRecord{}, nil
}

func (m *mockRepo) DeleteChannelLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	return m.err
}

func (m *mockRepo) FindTeamByChannel(ctx context.Context, workspaceID uuid.UUID, slackChannelID string) (slackrepository.TeamRecord, error) {
	if m.err != nil {
		return slackrepository.TeamRecord{}, m.err
	}
	return m.team, nil
}

func (m *mockRepo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

type mockRequestStore struct {
	last integrationrequests.CoreUpsertRequestInput
}

func (m *mockRequestStore) UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error) {
	m.last = input
	id := uuid.New()
	return integrationrequests.CoreIntegrationRequest{ID: id}, nil
}

type mockStoryService struct{}

func (m *mockStoryService) CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	return stories.CoreSingleStory{ID: uuid.New(), Title: ns.Title}, nil
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

func TestHandleViewSubmissionCreatesSlackRequestWhenModeIsRequests(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installeBy := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installeBy,
		},
		settings: slackrepository.SlackWorkspaceSettingsRecord{
			WorkspaceID:       workspaceID,
			DefaultCreateMode: CreateModeSendToRequests,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
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
	require.Equal(t, "Fix login bug", requests.last.Title)
	require.Equal(t, workspaceID, requests.last.WorkspaceID)
	require.Equal(t, teamID, requests.last.TeamID)
	require.NotNil(t, requests.last.SourceURL)
	require.True(t, strings.Contains(*requests.last.SourceURL, "acme.slack.com/archives/C123"))
}

func TestParseCommandTitleSupportsCreateTaskPrefix(t *testing.T) {
	title := parseCommandTitle("create task Improve onboarding")
	require.Equal(t, "Improve onboarding", title)

	title = parseCommandTitle("Improve onboarding")
	require.Equal(t, "Improve onboarding", title)

	title = parseCommandTitle("")
	require.Equal(t, "New task", title)
}

func slackSignature(secret, timestamp string, body []byte) string {
	base := "v0:" + timestamp + ":" + string(body)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(base))
	return "v0=" + hex.EncodeToString(h.Sum(nil))
}
