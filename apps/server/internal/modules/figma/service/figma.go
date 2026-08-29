package figma

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/internal/platform/workspaceurl"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var ErrNotConnected = errors.New("connect Figma before linking a design")

const storyLinkRefreshInterval = 5 * time.Minute

var oauthScopes = []string{
	"current_user:read",
	"file_metadata:read",
	"file_content:read",
	"file_dev_resources:read",
	"file_dev_resources:write",
	"webhooks:read",
	"webhooks:write",
}

type Service struct {
	log             *logger.Logger
	repo            Repository
	stories         StoryService
	config          Config
	client          apiClient
	secrets         CredentialVault
	webhookGateway  *webhooks.Gateway
	webhookInbox    webhooks.Inbox
	webhookPayloads WebhookPayloadOpener
	now             func() time.Time
}

func New(log *logger.Logger, repo Repository, storyService StoryService, config Config) *Service {
	httpClient := &http.Client{Timeout: 20 * time.Second}
	return &Service{
		log: log, repo: repo, stories: storyService, config: config,
		client: apiClient{http: httpClient, config: config}, secrets: config.Credentials,
		webhookGateway: config.WebhookGateway, webhookInbox: config.WebhookInbox,
		webhookPayloads: config.WebhookPayloads,
		now:             time.Now,
	}
}

func (s *Service) configured() bool {
	return strings.TrimSpace(s.config.ClientID) != "" &&
		strings.TrimSpace(s.config.ClientSecret) != "" &&
		strings.TrimSpace(s.config.RedirectURL) != "" &&
		s.secrets != nil
}

func (s *Service) GetIntegration(ctx context.Context, workspaceID uuid.UUID) (Integration, error) {
	integration := Integration{Configured: s.configured()}
	connection, err := s.repo.GetConnection(ctx, workspaceID)
	if err != nil {
		if !errors.Is(err, figmadomain.ErrNotFound) {
			return integration, err
		}
		return integration, nil
	}
	connection.CredentialPayload = ""
	integration.Connection = &connection
	return integration, nil
}

func (s *Service) CreateInstallSession(ctx context.Context, workspaceID, userID uuid.UUID, workspaceSlug string) (string, error) {
	if !s.configured() {
		return "", errors.New("figma integration is not configured")
	}
	state, verifier, err := s.createOAuthState(ctx, workspaceID, userID, workspaceSlug)
	if err != nil {
		return "", err
	}
	challengeBytes := sha256.Sum256([]byte(verifier))
	query := url.Values{
		"client_id": {s.config.ClientID}, "redirect_uri": {s.config.RedirectURL}, "scope": {strings.Join(oauthScopes, ",")},
		"state": {state}, "response_type": {"code"}, "code_challenge": {base64.RawURLEncoding.EncodeToString(challengeBytes[:])}, "code_challenge_method": {"S256"},
	}
	return "https://www.figma.com/oauth?" + query.Encode(), nil
}

func (s *Service) CompleteOAuth(ctx context.Context, code, state string) (string, error) {
	stored, err := s.consumeOAuthState(ctx, state)
	if err != nil {
		return "", fmt.Errorf("the Figma connection session expired; try again: %w", err)
	}
	failureURL := s.integrationURL(stored.WorkspaceSlug, "figma_error=connection_failed")
	if strings.TrimSpace(code) == "" {
		return failureURL, errors.New("figma OAuth callback is missing its authorization code")
	}
	response, err := s.client.exchange(ctx, code, s.config.RedirectURL, stored.CodeVerifier)
	if err != nil {
		s.log.Error(ctx, "failed exchanging Figma OAuth code", "error", err, "workspace_id", stored.WorkspaceID)
		return failureURL, err
	}
	user, err := s.client.currentUser(ctx, response.AccessToken)
	if err != nil {
		s.log.Error(ctx, "failed loading Figma OAuth user", "error", err, "workspace_id", stored.WorkspaceID)
		return failureURL, err
	}
	expiresAt := s.now().UTC().Add(time.Duration(response.ExpiresIn) * time.Second)
	token := Token{AccessToken: response.AccessToken, RefreshToken: response.RefreshToken, TokenType: response.TokenType, ExpiresAt: expiresAt}
	connectionID := uuid.New()
	installationGeneration := uuid.New()
	payload, err := s.sealToken(stored.WorkspaceID, connectionID, installationGeneration, token)
	if err != nil {
		s.log.Error(ctx, "failed encrypting Figma OAuth token", "error", err, "workspace_id", stored.WorkspaceID)
		return failureURL, err
	}
	userID := user.ID
	if userID == "" {
		userID = response.UserIDString
	}
	// Reconnecting replaces the active installation. Best-effort cleanup keeps
	// old file-scoped webhooks from continuing to call us after the replacement.
	if previous, previousToken, previousErr := s.connectionToken(ctx, stored.WorkspaceID); previousErr == nil {
		s.cleanupWebhooks(ctx, previous, previousToken)
	}
	_, err = s.repo.UpsertConnection(ctx, Connection{
		ID:                     connectionID,
		WorkspaceID:            stored.WorkspaceID,
		FigmaUserID:            userID,
		Email:                  optional(user.Email),
		Handle:                 optional(user.Handle),
		CredentialPayload:      payload,
		CredentialVersion:      int16(credentialvault.CurrentVersion),
		InstallationGeneration: installationGeneration,
		Scopes:                 append([]string(nil), oauthScopes...),
		ExpiresAt:              expiresAt,
		ConnectedByUserID:      stored.UserID,
		IsActive:               true,
	})
	if err != nil {
		s.log.Error(ctx, "failed storing Figma OAuth connection", "error", err, "workspace_id", stored.WorkspaceID)
		return failureURL, err
	}
	return s.integrationURL(stored.WorkspaceSlug, "connected=1"), nil
}

func (s *Service) integrationURL(workspaceSlug, rawQuery string) string {
	destination := workspaceurl.Build(
		s.config.WebsiteURL,
		workspaceSlug,
		"settings",
		"workspace",
		"integrations",
		"figma",
	)
	parsed, err := url.Parse(destination)
	if err != nil || parsed.Host == "" {
		return "/"
	}
	parsed.RawQuery = rawQuery
	return parsed.String()
}

func (s *Service) Disconnect(ctx context.Context, workspaceID uuid.UUID) error {
	connection, token, err := s.connectionToken(ctx, workspaceID)
	if err == nil {
		s.cleanupWebhooks(ctx, connection, token)
	}
	return s.repo.Disconnect(ctx, workspaceID)
}

func (s *Service) ResolveLink(ctx context.Context, workspaceID uuid.UUID, rawURL string) (Artifact, error) {
	artifact, err := ParseURL(rawURL)
	if err != nil {
		return Artifact{}, err
	}
	_, token, err := s.connectionToken(ctx, workspaceID)
	if err != nil {
		return Artifact{}, err
	}
	artifact, err = s.client.resolve(ctx, token.AccessToken, artifact)
	return artifact, err
}

func (s *Service) ListStoryLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]StoryLink, error) {
	if _, err := s.stories.Get(ctx, storyID, workspaceID); err != nil {
		return nil, err
	}
	return s.repo.ListStoryLinks(ctx, workspaceID, storyID)
}

func (s *Service) ListStoryHandoffStatuses(ctx context.Context, workspaceID uuid.UUID) (map[uuid.UUID]string, error) {
	statuses, err := s.repo.ListStoryHandoffStatuses(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]string, len(statuses))
	for _, status := range statuses {
		result[status.StoryID] = status.Status
	}
	return result, nil
}

func (s *Service) LinkStory(ctx context.Context, workspaceID, actorID, storyID uuid.UUID, workspaceSlug, rawURL string) (StoryLink, error) {
	story, err := s.stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return StoryLink{}, err
	}
	artifact, err := s.ResolveLink(ctx, workspaceID, rawURL)
	if err != nil {
		return StoryLink{}, err
	}
	link, err := s.repo.UpsertStoryLink(ctx, StoryLink{WorkspaceID: workspaceID, StoryID: storyID, CreatedByUserID: actorID, Artifact: artifact})
	if err != nil {
		return StoryLink{}, err
	}
	connection, token, err := s.connectionToken(ctx, workspaceID)
	if err == nil {
		storyURL := s.storyURL(workspaceSlug, story)
		if resourceID, resourceErr := s.client.createDevResource(ctx, token.AccessToken, link, storyURL); resourceErr != nil {
			s.log.Warn(ctx, "failed to create Figma Dev Resource", "error", resourceErr, "story_id", storyID)
		} else if resourceID != nil {
			link.DevResourceID = resourceID
			if updateErr := s.repo.UpdateStoryLink(ctx, link); updateErr != nil {
				s.log.Warn(ctx, "failed to store Figma Dev Resource", "error", updateErr, "story_id", storyID)
			}
		}
		s.ensureWebhooks(ctx, connection, token, artifact.FileKey)
	}
	_ = s.stories.RecordActivity(ctx, StoryActivity{
		StoryID: storyID, ActorID: actorID, Type: "link", Field: "figma",
		Previous: deref(artifact.NodeName, artifact.FileName),
		Current:  artifact.CanonicalURL, WorkspaceID: workspaceID,
	})
	return link, nil
}

func (s *Service) CreateStoryFromLink(ctx context.Context, workspaceID, actorID uuid.UUID, input CreateStoryInput) (Story, StoryLink, error) {
	artifact, err := s.ResolveLink(ctx, workspaceID, input.URL)
	if err != nil {
		return Story{}, StoryLink{}, err
	}
	title := deref(artifact.NodeName, artifact.FileName)
	if input.Title != nil && strings.TrimSpace(*input.Title) != "" {
		title = strings.TrimSpace(*input.Title)
	}
	description := input.Description
	if description == nil {
		value := "Design: " + artifact.CanonicalURL
		description = &value
	}
	story, err := s.stories.CreateExternal(ctx, actorID, NewStory{
		Title: title, Description: description, DescriptionHTML: description,
		TeamID: input.TeamID, StatusID: input.StatusID,
		ReporterID: &actorID, Priority: "No Priority",
	}, workspaceID)
	if err != nil {
		return Story{}, StoryLink{}, err
	}
	link, err := s.LinkStory(ctx, workspaceID, actorID, story.ID, input.WorkspaceSlug, input.URL)
	return story, link, err
}

func (s *Service) DeleteStoryLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	link, err := s.repo.GetStoryLink(ctx, workspaceID, linkID)
	if err != nil {
		return err
	}
	if link.DevResourceID != nil {
		if _, token, tokenErr := s.connectionToken(ctx, workspaceID); tokenErr == nil {
			if deleteErr := s.client.deleteDevResource(ctx, token.AccessToken, link.Artifact.FileKey, *link.DevResourceID); deleteErr != nil {
				return deleteErr
			}
		}
	}
	_, err = s.repo.DeleteStoryLink(ctx, workspaceID, linkID)
	return err
}

func (s *Service) RefreshStoryLink(ctx context.Context, workspaceID, linkID uuid.UUID) (StoryLink, error) {
	link, err := s.repo.GetStoryLink(ctx, workspaceID, linkID)
	if err != nil {
		return StoryLink{}, err
	}
	if s.now().UTC().Before(link.LastSyncedAt.Add(storyLinkRefreshInterval)) {
		return link, nil
	}
	artifact, err := s.ResolveLink(ctx, workspaceID, link.Artifact.CanonicalURL)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests {
			return link, nil
		}
		now := s.now().UTC()
		link.UnavailableAt = &now
		_ = s.repo.UpdateStoryLink(ctx, link)
		return StoryLink{}, err
	}
	if artifact.ThumbnailURL == nil {
		artifact.ThumbnailURL = link.Artifact.ThumbnailURL
	}
	link.Artifact = artifact
	link.UnavailableAt = nil
	if err := s.repo.UpdateStoryLink(ctx, link); err != nil {
		return StoryLink{}, err
	}
	return link, nil
}

func (s *Service) HandleWebhook(ctx context.Context, body []byte) error {
	_, err := s.ReceiveWebhook(ctx, webhooks.SignedRequest{
		Method: "POST",
		Body:   append([]byte(nil), body...),
	})
	return err
}

func (s *Service) ensureWebhooks(ctx context.Context, connection Connection, token Token, fileKey string) {
	if strings.TrimSpace(s.config.WebhookURL) == "" {
		return
	}
	for _, eventType := range []string{EventFileUpdate, EventDevModeStatusUpdate} {
		if _, err := s.repo.FindWebhook(ctx, connection.ID, fileKey, eventType); err == nil {
			continue
		} else if !errors.Is(err, figmadomain.ErrNotFound) {
			s.log.Warn(ctx, "failed checking existing Figma webhook", "error", err)
			continue
		}
		passcode, err := randomValue(32)
		if err != nil {
			continue
		}
		id, err := s.client.createWebhook(ctx, token.AccessToken, eventType, fileKey, s.config.WebhookURL, passcode)
		if err != nil {
			s.log.Warn(ctx, "failed to create Figma webhook", "event_type", eventType, "file_key", fileKey, "error", err)
			continue
		}
		if err := s.repo.SaveWebhook(ctx, Webhook{ConnectionID: connection.ID, FileKey: fileKey, EventType: eventType, FigmaWebhookID: id, PasscodeHash: digest(passcode), IsActive: true}); err != nil {
			s.log.Warn(ctx, "failed to store Figma webhook", "error", err)
		}
	}
}

func (s *Service) cleanupWebhooks(ctx context.Context, connection Connection, token Token) {
	webhooks, err := s.repo.ListWebhooks(ctx, connection.ID)
	if err != nil {
		s.log.Warn(ctx, "failed listing Figma webhooks during cleanup", "error", err)
		return
	}
	for _, webhook := range webhooks {
		if err := s.client.deleteWebhook(ctx, token.AccessToken, webhook.FigmaWebhookID); err != nil {
			s.log.Warn(ctx, "failed deleting Figma webhook", "error", err, "webhook_id", webhook.FigmaWebhookID)
		}
		if err := s.repo.DeactivateWebhook(ctx, webhook.FigmaWebhookID); err != nil {
			s.log.Warn(ctx, "failed deactivating Figma webhook", "error", err, "webhook_id", webhook.FigmaWebhookID)
		}
	}
}

func (s *Service) storyURL(workspaceSlug string, story Story) string {
	baseURL, err := url.Parse(strings.TrimRight(strings.TrimSpace(s.config.WebsiteURL), "/"))
	if err != nil || baseURL.Hostname() == "" {
		return ""
	}
	reference := fmt.Sprintf("%s-%d", strings.ToUpper(strings.TrimSpace(story.TeamCode)), story.SequenceID)
	host := baseURL.Hostname()
	if strings.EqualFold(host, "localhost") || strings.EqualFold(host, "0.0.0.0") || net.ParseIP(host) != nil {
		baseURL.Path = path.Join("/", workspaceSlug, "work", reference)
		return baseURL.String()
	}
	baseURL.Path = path.Join("/", "work", reference)
	if !strings.HasPrefix(host, workspaceSlug+".") {
		if port := baseURL.Port(); port != "" {
			baseURL.Host = fmt.Sprintf("%s.%s:%s", workspaceSlug, host, port)
		} else {
			baseURL.Host = workspaceSlug + "." + host
		}
	}
	return baseURL.String()
}

func randomValue(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
func deref(value *string, fallback string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}
	return *value
}
