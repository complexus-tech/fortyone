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
	workspace slackrepository.WorkspaceRecord
	team      slackrepository.TeamRecord
	err       error
}

func (m *mockRepo) FindWorkspaceBySlug(ctx context.Context, slug string) (slackrepository.WorkspaceRecord, error) {
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

type mockRequestStore struct {
	last integrationrequests.CoreUpsertRequestInput
}

func (m *mockRequestStore) UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error) {
	m.last = input
	id := uuid.New()
	return integrationrequests.CoreIntegrationRequest{ID: id}, nil
}

func newTestService(repo Repository, requests RequestStore, cfg Config) *Service {
	testLogger := logger.NewWithJSON(io.Discard, slog.LevelError, "test")
	service := New(testLogger, repo, requests, cfg)
	service.clock = fixedClock{now: time.Unix(1_700_000_000, 0)}
	return service
}

func TestVerifyRequest(t *testing.T) {
	secret := "secret"
	service := newTestService(nil, nil, Config{SigningSecret: secret})
	body := []byte("payload=test")
	timestamp := "1700000000"

	headers := http.Header{}
	headers.Set("X-Slack-Request-Timestamp", timestamp)
	headers.Set("X-Slack-Signature", slackSignature(secret, timestamp, body))

	err := service.VerifyRequest(body, headers)
	require.NoError(t, err)
}

func TestHandleViewSubmissionCreatesSlackIntegrationRequest(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
	}
	requests := &mockRequestStore{}
	service := newTestService(repo, requests, Config{WebsiteURL: "https://app.example.com"})

	interaction := map[string]any{
		"type": "view_submission",
		"view": map[string]any{
			"callback_id":      "fortyone_create_request",
			"private_metadata": `{"slack_team_domain":"acme","slack_channel_id":"C123","slack_message_ts":"171234.000100"}`,
			"state": map[string]any{
				"values": map[string]any{
					"workspace_slug": map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "acme"}},
					"team_code":      map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "ENG"}},
					"title":          map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Fix login bug"}},
					"description":    map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "User cannot log in from iOS"}},
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

func TestAcceptIntegrationRequestNoChannelIsNoop(t *testing.T) {
	service := newTestService(nil, nil, Config{})
	err := service.AcceptIntegrationRequest(
		context.Background(),
		integrationrequests.CoreIntegrationRequest{Provider: integrationrequests.ProviderSlack, Metadata: map[string]any{}},
		stories.CoreSingleStory{ID: uuid.New(), Title: "Story"},
	)
	require.NoError(t, err)
}

func TestParseCommandTextSupportsCreateStoryPrefix(t *testing.T) {
	workspace, team, title, err := parseCommandText("create story acme ENG Improve onboarding")
	require.NoError(t, err)
	require.Equal(t, "acme", workspace)
	require.Equal(t, "ENG", team)
	require.Equal(t, "Improve onboarding", title)

	workspace, team, title, err = parseCommandText("acme ENG Improve onboarding")
	require.NoError(t, err)
	require.Equal(t, "acme", workspace)
	require.Equal(t, "ENG", team)
	require.Equal(t, "Improve onboarding", title)
}

func slackSignature(secret, timestamp string, body []byte) string {
	base := "v0:" + timestamp + ":" + string(body)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(base))
	return "v0=" + hex.EncodeToString(h.Sum(nil))
}
