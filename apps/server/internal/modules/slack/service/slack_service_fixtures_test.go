package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type mockRequestStore struct {
	last         upsertIntegrationRequestInput
	lastCreated  integrationRequest
	request      integrationRequest
	requestErr   error
	lastThread   BindProviderThreadInput
	calls        int
	bindCalls    int
	threadMatch  bool
	thread       ProviderThread
	threadLookup providerThreadLookupInput
}

func (m *mockRequestStore) GetForUser(_ context.Context, _, requestID, _ uuid.UUID) (integrationRequest, error) {
	if m.requestErr != nil {
		return integrationRequest{}, m.requestErr
	}
	if m.request.ID == requestID {
		return m.request, nil
	}
	return integrationRequest{}, slackdomain.ErrNotFound
}

func (m *mockRequestStore) BindProviderThread(_ context.Context, input BindProviderThreadInput) (ProviderThread, error) {
	m.bindCalls++
	m.lastThread = input
	return ProviderThread{ID: uuid.New()}, nil
}

func (m *mockRequestStore) HasAuthorizedProviderThread(_ context.Context, _ ProviderThreadMatchInput) (bool, error) {
	return m.threadMatch, nil
}

func (m *mockRequestStore) HasCurrentProviderThread(_ context.Context, input providerThreadLookupInput) (bool, error) {
	m.threadLookup = input
	return m.threadMatch, nil
}

func (m *mockRequestStore) FindProviderThread(_ context.Context, _, _ uuid.UUID, _ string) (ProviderThread, error) {
	if m.thread.ID == uuid.Nil {
		return ProviderThread{}, errProviderThreadNotFound
	}
	return m.thread, nil
}

func (m *mockRequestStore) UpsertPending(ctx context.Context, input upsertIntegrationRequestInput) (integrationRequest, error) {
	m.calls++
	m.last = input
	m.lastCreated = integrationRequest{
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
		Status:           integrationRequestStatusPending,
		CreatedAt:        time.Unix(1_700_000_000, 0),
		UpdatedAt:        time.Unix(1_700_000_000, 0),
	}
	return m.lastCreated, nil
}

type mockStoryService struct {
	lastActorID   uuid.UUID
	lastWorkspace uuid.UUID
	lastStory     newStory
	createCalls   int
	sequenceID    int
}

type mutationConfirmerStub struct {
	result       StoryMutationResult
	cancelResult StoryMutationCancellationResult
	err          error
	scopes       []storyMutationScope
	tokens       []string
}

func (s *mutationConfirmerStub) CancelStoryMutation(_ context.Context, scope storyMutationScope, token string) (StoryMutationCancellationResult, error) {
	s.scopes = append(s.scopes, scope)
	s.tokens = append(s.tokens, token)
	return s.cancelResult, s.err
}

func (s *mutationConfirmerStub) ConfirmStoryMutation(_ context.Context, scope storyMutationScope, token string) (StoryMutationResult, error) {
	s.scopes = append(s.scopes, scope)
	s.tokens = append(s.tokens, token)
	return s.result, s.err
}

func (m *mockStoryService) Create(ctx context.Context, ns newStory, workspaceID uuid.UUID) (singleStory, error) {
	m.createCalls++
	if ns.Reporter != nil {
		m.lastActorID = *ns.Reporter
	}
	m.lastWorkspace = workspaceID
	m.lastStory = ns
	return singleStory{
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
	if cfg.CredentialVault == nil {
		var err error
		cfg.CredentialVault, err = buildTestCredentialVault()
		if err != nil {
			panic(err)
		}
	}
	testLogger := logger.NewWithJSON(io.Discard, slog.LevelError, "test")
	service := New(testLogger, repo, requests, storyService, cfg, WithNonceStore(newMockNonceStore()))
	sealTestServiceRepositoryCredential(service, repo)
	service.clock = fixedClock{now: time.Unix(1_700_000_000, 0)}
	return service
}

func sealTestServiceRepositoryCredential(service *Service, repo Repository) {
	if service == nil || service.credentials == nil || repo == nil {
		return
	}
	var record *slackrepository.SlackWorkspaceRecord
	switch typed := repo.(type) {
	case *mockRepo:
		record = &typed.slackWorkspace
	case *blockingSlackWorkspaceRepo:
		record = &typed.mockRepo.slackWorkspace
	case *blockingTeamListRepo:
		record = &typed.mockRepo.slackWorkspace
	case *credentialUpgradeRepo:
		record = &typed.mockRepo.slackWorkspace
	case *onboardingReceiptRepoStub:
		record = &typed.mockRepo.slackWorkspace
	case *blockingOnboardingReceiptRepo:
		record = &typed.mockRepo.slackWorkspace
	}
	if record == nil || strings.TrimSpace(record.BotAccessToken) == "" || credentialvault.IsEnvelope(record.BotAccessToken) {
		return
	}
	if record.WorkspaceID == uuid.Nil || strings.TrimSpace(record.SlackTeamID) == "" {
		return
	}
	if record.InstallGeneration == uuid.Nil {
		record.InstallGeneration = testInstallGeneration
	}
	encrypted, version, err := service.credentials.seal(slackCredentialBinding{
		WorkspaceID:       record.WorkspaceID,
		SlackTeamID:       record.SlackTeamID,
		InstallGeneration: record.InstallGeneration,
	}, slackCredential{AccessToken: record.BotAccessToken})
	if err != nil {
		panic("seal test Slack service credential: " + err.Error())
	}
	record.BotAccessToken = encrypted
	record.CredentialVersion = version
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
