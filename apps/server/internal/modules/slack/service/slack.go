package slack

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

var (
	ErrSlackNotConfigured              = errors.New("slack integration is not configured")
	ErrSlackSigningSecretNotConfigured = errors.New("slack signing secret is not configured")
	ErrSlackRequestExpired             = errors.New("slack request timestamp is too old")
	ErrSlackInvalidSignature           = errors.New("invalid slack request signature")
	ErrSlackNoWorkspaceLinked          = errors.New("slack workspace is not connected")
	ErrSlackNoTeamsAvailable           = errors.New("no teams are available in this workspace")
	ErrSlackTeamSelectionRequired      = errors.New("team selection is required")
	ErrSlackUserNotLinked              = errors.New("slack user is not linked to a fortyone user")
	ErrSlackTeamNotAvailable           = errors.New("selected team is not available to the slack user")
	ErrSlackInteractionActorMismatch   = errors.New("slack interaction actor does not match the modal source")
	ErrSlackEventRuntimeNotConfigured  = errors.New("slack event runtime is not configured")
	ErrSlackInvalidEventPayload        = errors.New("invalid slack event payload")
	slackMrkdwnTextEscaper             = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
)

const (
	slackProviderMessaging   = "slack"
	slackNoncePurposeOAuth   = "oauth_install"
	slackNoncePurposeAccount = "account_link"
	slackOAuthNonceTTL       = 10 * time.Minute
	slackAccountLinkNonceTTL = 15 * time.Minute
	slackOpaqueNonceSize     = 32
	slackPriorityNoPriority  = "No Priority"
	slackRequestStatusValue  = "__fortyone_request__"
	slackStatusKindRequest   = "request"
	slackStatusKindStory     = "story"

	modalBlockTeam        = "team"
	modalBlockTitle       = "title"
	modalBlockDescription = "description"
	modalBlockStatus      = "status"
	modalBlockPriority    = "priority"
	modalBlockAssignee    = "assignee"
	modalBlockLabels      = "labels"
	modalBlockObjective   = "objective"

	modalActionTeamSelect        = "team_select"
	modalActionTitleInput        = "title_input"
	modalActionDescriptionInput  = "description_input"
	modalActionStatusSelect      = "status_select"
	modalActionPrioritySelect    = "priority_select"
	modalActionAssigneeSelect    = "assignee_select"
	modalActionLabelsMultiSelect = "labels_multi_select"
	modalActionObjectiveSelect   = "objective_select"

	modalTitleMaxRunes       = 255
	modalDescriptionMaxRunes = 3000
	modalMetadataMaxBytes    = 3000
	modalSourceTextMaxRunes  = 1000

	slackInteractiveWorkTimeout   = 2500 * time.Millisecond
	slackWorkObjectTriggerTimeout = 2500 * time.Millisecond
	slackFailureFeedbackTimeout   = 2 * time.Second
)

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type Service struct {
	log                      *logger.Logger
	repo                     Repository
	requests                 RequestStore
	stories                  StoryService
	cfg                      Config
	client                   *http.Client
	clock                    Clock
	random                   io.Reader
	nonces                   NonceStore
	eventQueue               EventQueue
	eventInbox               EventInbox
	outbound                 OutboundStore
	credentials              *credentialCodec
	webClient                *slackWebClient
	mutationConfirmer        messaging.StoryMutationConfirmer
	workObjectTriggerTimeout time.Duration
}

type Option func(*Service)

type slackCredentialUpgrader interface {
	UpgradeSlackCredential(ctx context.Context, slackWorkspaceID uuid.UUID, encrypted string, version int) error
}

func WithEventRuntime(queue EventQueue, inbox EventInbox) Option {
	return func(service *Service) {
		service.eventQueue = queue
		service.eventInbox = inbox
		service.outbound, _ = inbox.(OutboundStore)
	}
}

func WithNonceStore(store NonceStore) Option {
	return func(service *Service) {
		service.nonces = store
	}
}

func WithMutationConfirmer(confirmer messaging.StoryMutationConfirmer) Option {
	return func(service *Service) {
		service.mutationConfirmer = confirmer
	}
}

func New(log *logger.Logger, repo Repository, requests RequestStore, stories StoryService, cfg Config, options ...Option) *Service {
	service := &Service{
		log:      log,
		repo:     repo,
		requests: requests,
		stories:  stories,
		cfg:      cfg,
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
		clock:                    realClock{},
		random:                   rand.Reader,
		workObjectTriggerTimeout: slackWorkObjectTriggerTimeout,
	}
	service.webClient = newSlackWebClient(service.client)
	if codec, err := newCredentialCodec(cfg.SecretKey); err == nil {
		service.credentials = codec
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *Service) GetIntegration(ctx context.Context, workspaceID uuid.UUID) (CoreIntegration, error) {
	integration := CoreIntegration{
		Channels: make([]CoreSlackChannel, 0),
	}

	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return integration, nil
		}
		return CoreIntegration{}, err
	}
	coreWorkspace := toCoreSlackWorkspace(slackWorkspace)
	integration.SlackWorkspace = &coreWorkspace

	channels, err := s.repo.ListChannels(ctx, workspaceID)
	if err != nil && !slackrepository.IsNotFound(err) {
		return CoreIntegration{}, err
	}
	if err == nil {
		integration.Channels = toCoreChannels(channels)
	}

	return integration, nil
}

func (s *Service) GetRequestLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]CoreRequestLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.repo.ListRequestLogs(ctx, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	logs := make([]CoreRequestLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, toCoreRequestLog(row))
	}
	return logs, nil
}

func (s *Service) CreateInstallSession(ctx context.Context, workspaceID, userID uuid.UUID, workspaceSlug string) (CoreCreateInstallSession, error) {
	if !s.canUseOAuth() {
		return CoreCreateInstallSession{}, ErrSlackNotConfigured
	}
	if workspaceID == uuid.Nil || userID == uuid.Nil || strings.TrimSpace(workspaceSlug) == "" {
		return CoreCreateInstallSession{}, errors.New("workspace, user, and workspace slug are required")
	}
	payload, err := json.Marshal(oauthInstallNoncePayload{WorkspaceSlug: strings.TrimSpace(workspaceSlug)})
	if err != nil {
		return CoreCreateInstallSession{}, err
	}
	state, digest, err := s.newOpaqueNonce()
	if err != nil {
		return CoreCreateInstallSession{}, err
	}
	boundUserID := userID
	if err := s.nonces.CreateNonce(ctx, messagingrepository.NonceInput{
		Provider:    slackProviderMessaging,
		Purpose:     slackNoncePurposeOAuth,
		NonceHash:   digest,
		WorkspaceID: workspaceID,
		UserID:      &boundUserID,
		Payload:     payload,
		ExpiresAt:   s.clock.Now().UTC().Add(slackOAuthNonceTTL),
	}); err != nil {
		return CoreCreateInstallSession{}, fmt.Errorf("create Slack install state: %w", err)
	}

	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&scope=%s&state=%s&redirect_uri=%s",
		url.QueryEscape(s.cfg.ClientID),
		url.QueryEscape(slackBotOAuthScopeValue()),
		url.QueryEscape(state),
		url.QueryEscape(s.cfg.RedirectURL),
	)

	return CoreCreateInstallSession{InstallURL: authURL}, nil
}

func (s *Service) HandleSetup(ctx context.Context, code, state, slackError string) (string, error) {
	if !s.canUseOAuth() {
		return "", ErrSlackNotConfigured
	}
	nonce, err := s.consumeNonce(ctx, slackNoncePurposeOAuth, state, nil, nil)
	if err != nil {
		return "", fmt.Errorf("invalid or expired Slack install state: %w", err)
	}
	if strings.TrimSpace(slackError) != "" {
		return "", fmt.Errorf("slack oauth failed: %s", slackError)
	}
	if strings.TrimSpace(code) == "" {
		return "", errors.New("missing slack oauth code")
	}
	if nonce.WorkspaceID == uuid.Nil || nonce.UserID == nil || *nonce.UserID == uuid.Nil {
		return "", errors.New("invalid Slack install state binding")
	}
	var noncePayload oauthInstallNoncePayload
	if err := json.Unmarshal(nonce.Payload, &noncePayload); err != nil || strings.TrimSpace(noncePayload.WorkspaceSlug) == "" {
		return "", errors.New("invalid Slack install state payload")
	}

	oauthResp, err := s.exchangeOAuthCode(ctx, code)
	if err != nil {
		return "", err
	}

	credential := slackCredential{
		AccessToken:  strings.TrimSpace(oauthResp.AccessToken),
		RefreshToken: strings.TrimSpace(oauthResp.RefreshToken),
	}
	if oauthResp.ExpiresIn > 0 {
		expiresAt := s.clock.Now().UTC().Add(time.Duration(oauthResp.ExpiresIn) * time.Second)
		credential.ExpiresAt = &expiresAt
	}
	credentialPayload, credentialVersion, err := s.credentials.seal(credential)
	if err != nil {
		return "", err
	}
	_, err = s.repo.UpsertSlackWorkspace(ctx, nonce.WorkspaceID, *nonce.UserID, slackrepository.OAuthInstallPayload{
		SlackTeamID:       strings.TrimSpace(oauthResp.Team.ID),
		SlackTeamName:     strings.TrimSpace(oauthResp.Team.Name),
		SlackTeamDomain:   strings.TrimSpace(oauthResp.Team.Domain),
		BotUserID:         optionalString(oauthResp.BotUserID),
		BotAccessToken:    credentialPayload,
		LegacyAccessToken: credential.AccessToken,
		CredentialVersion: credentialVersion,
		SlackAppID:        optionalString(oauthResp.AppID),
		EnterpriseID:      optionalString(oauthResp.Enterprise.ID),
		AuthedUserID:      optionalString(oauthResp.AuthedUser.ID),
		Scope:             optionalString(oauthResp.Scope),
	})
	if err != nil {
		if errors.Is(err, slackrepository.ErrWorkspaceAlreadyConnected) {
			if cleanupErr := s.cleanupRejectedOAuthInstallation(
				ctx,
				nonce.WorkspaceID,
				strings.TrimSpace(oauthResp.Team.ID),
				credentialPayload,
				credentialVersion,
			); cleanupErr != nil && s.log != nil {
				s.log.Error(ctx, "failed scheduling cleanup for rejected Slack OAuth installation", "error", cleanupErr, "workspace_id", nonce.WorkspaceID, "slack_team_id", strings.TrimSpace(oauthResp.Team.ID))
			}
		}
		return "", err
	}

	return s.buildWorkspaceIntegrationURL(noncePayload.WorkspaceSlug), nil
}

func (s *Service) cleanupRejectedOAuthInstallation(ctx context.Context, workspaceID uuid.UUID, slackTeamID, encryptedCredential string, credentialVersion int) error {
	_, err := s.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err == nil {
		// The token belongs to an app installation that is now legitimately
		// bound, whether by a concurrent winner or an uncertain commit.
		return nil
	}
	if !slackrepository.IsNotFound(err) {
		return fmt.Errorf("verify rejected Slack OAuth team ownership: %w", err)
	}
	workspaceInstallation, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err == nil && strings.TrimSpace(workspaceInstallation.SlackTeamID) == slackTeamID {
		return nil
	}
	if err != nil && !slackrepository.IsNotFound(err) {
		return fmt.Errorf("verify rejected Slack OAuth workspace ownership: %w", err)
	}

	attemptCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	uninstall, err := s.repo.EnqueueSlackUninstall(attemptCtx, slackrepository.SlackUninstallInput{
		SlackWorkspaceID:     uuid.New(),
		WorkspaceID:          workspaceID,
		InstallGeneration:    uuid.New(),
		SlackTeamID:          slackTeamID,
		UninstallKind:        "orphaned_oauth",
		CredentialPayload:    encryptedCredential,
		CredentialKeyVersion: credentialVersion,
	})
	if err != nil {
		// A direct provider call after an uncertain persistence failure is not
		// safe: a concurrent OAuth callback may have installed this team between
		// the ownership checks above and the failed enqueue. Retain the error and
		// let an operator retry the guarded cleanup path.
		return err
	}
	claimed, shouldAttempt, err := s.repo.ClaimSlackUninstall(attemptCtx, uninstall.ID)
	if err != nil || !shouldAttempt {
		return err
	}
	_, err = executeSlackUninstall(
		attemptCtx,
		s.repo,
		s.repo,
		s.slackClient(),
		s.credentials,
		s.cfg.ClientID,
		s.cfg.ClientSecret,
		s.clock.Now(),
		claimed,
	)
	return err
}

func (s *Service) SyncChannels(ctx context.Context, workspaceID uuid.UUID) error {
	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return err
	}
	channels, err := s.fetchChannels(ctx, botToken)
	if err != nil {
		return err
	}
	return s.repo.UpsertChannels(ctx, workspaceID, slackWorkspace.ID, channels)
}

func (s *Service) DisconnectWorkspace(ctx context.Context, workspaceID uuid.UUID) error {
	if workspaceID == uuid.Nil {
		return errors.New("workspace id is required")
	}

	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("load Slack installation for uninstall: %w", err)
	}
	if slackWorkspace.CredentialVersion <= 0 {
		if _, err := s.botToken(ctx, slackWorkspace); err != nil {
			return fmt.Errorf("encrypt Slack credential before disconnect: %w", err)
		}
	}

	uninstall, err := s.repo.DisconnectSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	attemptCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	claimed, shouldAttempt, err := s.repo.ClaimSlackUninstall(attemptCtx, uninstall.ID)
	if err != nil {
		if s.log != nil {
			s.log.Warn(ctx, "Slack was disconnected locally but the uninstall record could not be claimed", "error", err, "workspace_id", workspaceID, "slack_team_id", uninstall.SlackTeamID, "uninstall_id", uninstall.ID)
		}
		return nil
	}
	if !shouldAttempt {
		return nil
	}
	_, uninstallErr := executeSlackUninstall(
		attemptCtx,
		s.repo,
		s.repo,
		s.slackClient(),
		s.credentials,
		s.cfg.ClientID,
		s.cfg.ClientSecret,
		s.clock.Now(),
		claimed,
	)
	if uninstallErr != nil {
		if s.log != nil {
			s.log.Warn(ctx, "Slack was disconnected locally and remote uninstall was scheduled for retry", "error", uninstallErr, "workspace_id", workspaceID, "slack_team_id", uninstall.SlackTeamID, "uninstall_id", uninstall.ID, "attempt", claimed.AttemptCount)
		}
	}
	return nil
}

func (s *Service) botToken(ctx context.Context, installation slackrepository.SlackWorkspaceRecord) (string, error) {
	payload := strings.TrimSpace(installation.BotAccessToken)
	if payload == "" {
		return "", errors.New("slack installation is missing bot token")
	}
	if s.credentials == nil {
		if installation.CredentialVersion == 0 {
			return payload, nil
		}
		return "", errors.New("slack credential encryption is not configured")
	}
	credential, openedVersion, err := s.credentials.open(payload)
	if err != nil {
		return "", err
	}
	if openedVersion == 0 || installation.CredentialVersion == 0 {
		upgrader, ok := s.repo.(slackCredentialUpgrader)
		if ok {
			encrypted, version, sealErr := s.credentials.seal(credential)
			if sealErr != nil {
				return "", sealErr
			}
			if upgradeErr := upgrader.UpgradeSlackCredential(ctx, installation.ID, encrypted, version); upgradeErr != nil && !slackrepository.IsNotFound(upgradeErr) {
				return "", upgradeErr
			}
		}
	}
	return credential.AccessToken, nil
}

func (s *Service) LinkSlackAccount(ctx context.Context, workspaceID, userID uuid.UUID, token string) error {
	if workspaceID == uuid.Nil {
		return errors.New("workspace id is required")
	}
	if userID == uuid.Nil {
		return errors.New("user id is required")
	}
	nonce, err := s.consumeNonce(ctx, slackNoncePurposeAccount, token, &workspaceID, &userID)
	if err != nil {
		return fmt.Errorf("invalid or expired Slack link token: %w", err)
	}
	if nonce.WorkspaceID != workspaceID {
		return errors.New("slack link token workspace mismatch")
	}
	if nonce.UserID != nil && *nonce.UserID != userID {
		return errors.New("slack link token user mismatch")
	}

	slackTeamID := strings.TrimSpace(valueOrEmpty(nonce.ExternalWorkspaceID))
	slackUserID := strings.TrimSpace(valueOrEmpty(nonce.ExternalUserID))
	if slackTeamID == "" || slackUserID == "" {
		return errors.New("invalid slack link token")
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err != nil {
		return err
	}
	if slackWorkspace.WorkspaceID != workspaceID {
		return errors.New("slack workspace does not belong to this workspace")
	}

	return s.repo.UpsertSlackUserLinks(ctx, workspaceID, slackWorkspace.ID, slackTeamID, []slackrepository.SlackUserLinkUpsert{
		{
			SlackUserID: slackUserID,
			UserID:      userID,
			LinkedVia:   "manual_link",
		},
	})
}

func IsNotFound(err error) bool {
	return slackrepository.IsNotFound(err)
}

func (s *Service) VerifyRequest(rawBody []byte, headers http.Header) error {
	secret := strings.TrimSpace(s.cfg.SigningSecret)
	if secret == "" {
		return ErrSlackSigningSecretNotConfigured
	}

	timestamp := headers.Get("X-Slack-Request-Timestamp")
	signature := headers.Get("X-Slack-Signature")
	if timestamp == "" || signature == "" {
		return ErrSlackInvalidSignature
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return ErrSlackInvalidSignature
	}
	if math.Abs(float64(s.clock.Now().Unix()-ts)) > 300 {
		return ErrSlackRequestExpired
	}

	base := "v0:" + timestamp + ":" + string(rawBody)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(base))
	expected := "v0=" + hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrSlackInvalidSignature
	}
	return nil
}

func (s *Service) HandleEvents(ctx context.Context, rawBody []byte) (EventResponse, error) {
	payload, err := decodeSlackEvent(rawBody)
	if err != nil {
		return EventResponse{}, fmt.Errorf("%w: %v", ErrSlackInvalidEventPayload, err)
	}
	if payload.Type == "url_verification" {
		return EventResponse{Challenge: payload.Challenge}, nil
	}
	if payload.Type != slackEventCallback {
		return EventResponse{}, nil
	}
	if payload.EventID == "" || payload.TeamID == "" {
		return EventResponse{}, fmt.Errorf("%w: callback is missing event_id or team_id", ErrSlackInvalidEventPayload)
	}
	event, supported := normalizeSlackEvent(payload)
	if !supported {
		// message.channels and message.groups are intentionally broad. Discard
		// unrelated roots, bot messages, edits, and unsupported event shapes
		// before encrypting or retaining their content in the durable inbox.
		return EventResponse{}, nil
	}
	if event.Kind == slackEventKindEntityDetails {
		workCtx, cancel := s.newSlackWorkObjectTriggerContext(ctx)
		err := s.handleSlackEntityDetailsEvent(workCtx, event)
		cancel()
		if err != nil && s.log != nil {
			s.log.Error(
				context.WithoutCancel(ctx),
				"failed processing Slack entity details within the trigger window",
				"error", err,
				"terminal", isSlackEntityDetailsTerminalError(err),
				"event_id", event.EventID,
				"slack_team_id", event.TeamID,
				"slack_user_id", event.UserID,
			)
		}
		// entity_details_requested carries a single-use trigger that expires in
		// three seconds. It is intentionally never persisted, queued, or retried;
		// the user can request a fresh trigger by refreshing the flexpane.
		return EventResponse{}, nil
	}
	if s.eventQueue == nil || s.eventInbox == nil {
		return EventResponse{}, ErrSlackEventRuntimeNotConfigured
	}
	var (
		workspaceID       *uuid.UUID
		installGeneration *uuid.UUID
	)
	installation, installationErr := s.repo.GetSlackWorkspaceByTeamID(ctx, payload.TeamID)
	if installationErr == nil {
		workspace := installation.WorkspaceID
		generation := installation.InstallGeneration
		workspaceID = &workspace
		installGeneration = &generation
	} else if slackrepository.IsNotFound(installationErr) {
		// A valid Slack signature proves the sender is Slack, not that FortyOne
		// still owns this installation. Disconnected and orphaned installations
		// have no recoverable work, so do not retain their message content.
		return EventResponse{}, nil
	} else {
		return EventResponse{}, fmt.Errorf("resolve Slack installation for event receipt: %w", installationErr)
	}
	if event.Kind == slackEventKindChannelThread {
		if installation.BotUserID != nil && containsSlackUserMention(event.Text, *installation.BotUserID) {
			return EventResponse{}, nil
		}
		linkedUserID, linkErr := s.repo.FindLinkedUserIDBySlackUser(ctx, installation.WorkspaceID, event.TeamID, event.UserID)
		if linkErr != nil {
			return EventResponse{}, fmt.Errorf("resolve Slack thread actor: %w", linkErr)
		}
		subscribed := false
		if linkedUserID != nil && *linkedUserID != uuid.Nil {
			conversation, conversationErr := findSlackConversation(ctx, s.eventInbox, assistantConversationInput(installation.WorkspaceID, *linkedUserID, event))
			switch {
			case conversationErr == nil:
				subscribed = slackThreadSubscriptionIsCurrent(conversation, installation, s.clock.Now())
			case !errors.Is(conversationErr, messagingrepository.ErrNotFound):
				return EventResponse{}, fmt.Errorf("resolve Slack assistant thread subscription: %w", conversationErr)
			}
		}
		if !subscribed && s.requests != nil {
			subscribed, err = s.requests.HasCurrentProviderThread(ctx, integrationrequests.CoreProviderThreadLookupInput{
				WorkspaceID:            installation.WorkspaceID,
				Provider:               integrationrequests.ProviderSlack,
				ExternalWorkspaceID:    event.TeamID,
				InstallationGeneration: installation.InstallGeneration,
				ExternalChannelID:      event.ChannelID,
				ExternalThreadID:       event.ThreadTS,
			})
			if err != nil {
				return EventResponse{}, fmt.Errorf("resolve Slack request thread subscription: %w", err)
			}
		}
		if !subscribed {
			return EventResponse{}, nil
		}
	}
	encryptedPayload, err := s.credentials.sealPayload(rawBody)
	if err != nil {
		return EventResponse{}, err
	}
	receipt, created, err := s.eventInbox.RegisterInboundEvent(ctx, messagingrepository.InboundEventInput{
		Provider:            "slack",
		WorkspaceID:         workspaceID,
		InstallGeneration:   installGeneration,
		ExternalWorkspaceID: payload.TeamID,
		ExternalEventID:     payload.EventID,
		EventType:           payload.Event.Type,
		PayloadEncrypted:    encryptedPayload,
	})
	if err != nil {
		return EventResponse{}, fmt.Errorf("persist Slack event receipt: %w", err)
	}
	if !created && (receipt.Status == "completed" || receipt.Status == "ignored" || receipt.Status == "cancelled") {
		return EventResponse{}, nil
	}
	if err := s.eventQueue.EnqueueSlackEvent(ctx, tasks.SlackEventPayload{
		ExternalWorkspaceID: payload.TeamID,
		EventID:             payload.EventID,
	}); err != nil {
		return EventResponse{}, fmt.Errorf("enqueue Slack event: %w", err)
	}
	if err := s.eventInbox.MarkInboundEventQueued(ctx, receipt.ID); err != nil {
		return EventResponse{}, fmt.Errorf("record Slack event queue handoff: %w", err)
	}
	return EventResponse{}, nil
}

func (s *Service) HandleCommand(ctx context.Context, rawBody []byte) (CommandResponse, error) {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return CommandResponse{}, err
	}
	if values.Get("ssl_check") == "1" {
		return CommandResponse{}, nil
	}

	triggerID := strings.TrimSpace(values.Get("trigger_id"))
	if triggerID == "" {
		return CommandResponse{}, errors.New("missing trigger_id")
	}

	source := requestSourceContext{
		SlackTeamID:     strings.TrimSpace(values.Get("team_id")),
		SlackTeamDomain: strings.TrimSpace(values.Get("team_domain")),
		SlackChannelID:  strings.TrimSpace(values.Get("channel_id")),
		SlackChannel:    strings.TrimSpace(values.Get("channel_name")),
		SlackUserID:     strings.TrimSpace(values.Get("user_id")),
		SlackUsername:   strings.TrimSpace(values.Get("user_name")),
		SlackText:       strings.TrimSpace(values.Get("text")),
		ResponseURL:     strings.TrimSpace(values.Get("response_url")),
	}
	title := parseCommandTitle(values.Get("text"))
	s.dispatchCommand(ctx, triggerID, title, source)

	return CommandResponse{
		ResponseType: "ephemeral",
		Text:         "Opening FortyOne create story form...",
	}, nil
}

func (s *Service) HandleInteractivity(ctx context.Context, rawBody []byte) (InteractionResponse, error) {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return InteractionResponse{}, err
	}
	payloadText := values.Get("payload")
	if payloadText == "" {
		return InteractionResponse{}, errors.New("missing payload")
	}

	var payload interactionPayload
	if err := json.Unmarshal([]byte(payloadText), &payload); err != nil {
		return InteractionResponse{}, err
	}

	switch payload.Type {
	case "message_action":
		s.dispatchInteraction(ctx, payload.Type, payload, s.handleMessageAction)
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	case "view_submission":
		if isSlackWorkObjectEditSubmission(payload) {
			s.dispatchSlackWorkObjectEdit(ctx, payload)
			return InteractionResponse{StatusCode: http.StatusOK}, nil
		}
		response, err := s.handleViewSubmission(ctx, payload)
		if err == nil {
			return response, nil
		}
		s.log.Error(ctx, "failed processing slack view submission", "error", err, "view_id", payload.View.ID, "slack_team_id", payload.Team.ID, "slack_user_id", payload.User.ID)
		return interactionValidationErrors(map[string]string{
			modalBlockTitle: "FortyOne could not create this task. Please try again.",
		})
	case "block_actions":
		if isSlackMutationAction(payload) {
			s.dispatchInteraction(ctx, payload.Type, payload, s.handleMutationAction)
		} else {
			s.dispatchInteraction(ctx, payload.Type, payload, s.handleBlockActions)
		}
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	case "block_suggestion":
		return s.handleBlockSuggestion(ctx, payload)
	default:
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
}

func (s *Service) dispatchCommand(parent context.Context, triggerID, title string, source requestSourceContext) {
	baseCtx := context.WithoutCancel(parent)
	go func() {
		workCtx, cancel := context.WithTimeout(baseCtx, slackInteractiveWorkTimeout)
		feedback, err := s.processCommand(workCtx, triggerID, title, source)
		cancel()
		if err != nil {
			s.log.Error(baseCtx, "failed processing slack command", "error", err, "slack_team_id", source.SlackTeamID, "slack_user_id", source.SlackUserID)
			feedback = "Unable to open the FortyOne create story form. Please try again."
		}
		if strings.TrimSpace(feedback) == "" {
			return
		}
		if strings.TrimSpace(source.ResponseURL) == "" {
			s.log.Warn(baseCtx, "cannot post slack command feedback without a response URL", "slack_team_id", source.SlackTeamID, "slack_user_id", source.SlackUserID)
			return
		}

		feedbackCtx, feedbackCancel := context.WithTimeout(baseCtx, slackFailureFeedbackTimeout)
		defer feedbackCancel()
		if notifyErr := s.postCommandResponse(feedbackCtx, source.ResponseURL, feedback); notifyErr != nil {
			s.log.Error(baseCtx, "failed posting slack command feedback", "error", notifyErr, "slack_team_id", source.SlackTeamID, "slack_user_id", source.SlackUserID)
		}
	}()
}

func (s *Service) processCommand(ctx context.Context, triggerID, title string, source requestSourceContext) (string, error) {
	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return "Slack is not connected to this FortyOne workspace.", nil
		}
		return "", err
	}

	linkedUserID, connectURL, err := s.resolveLinkedSlackUser(ctx, slackWorkspace.WorkspaceID, source)
	if err != nil {
		return "", err
	}
	if linkedUserID == uuid.Nil {
		return buildConnectSlackAccountMessage(connectURL), nil
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return "", err
	}
	if err := s.openCreateTaskModal(ctx, triggerID, title, "", source, slackWorkspace.WorkspaceID, linkedUserID, botToken); err != nil {
		return "", err
	}
	return "", nil
}

type interactionHandler func(context.Context, interactionPayload) (InteractionResponse, error)

func (s *Service) dispatchInteraction(parent context.Context, interactionType string, payload interactionPayload, handler interactionHandler) {
	baseCtx := context.WithoutCancel(parent)
	go func() {
		workCtx, cancel := context.WithTimeout(baseCtx, slackInteractiveWorkTimeout)
		_, err := handler(workCtx, payload)
		cancel()
		if err == nil {
			return
		}

		s.log.Error(baseCtx, "failed processing slack interaction", "error", err, "interaction_type", interactionType, "view_id", payload.View.ID, "slack_team_id", payload.Team.ID, "slack_user_id", payload.User.ID)
		feedbackCtx, feedbackCancel := context.WithTimeout(baseCtx, slackFailureFeedbackTimeout)
		defer feedbackCancel()
		if notifyErr := s.postInteractionFailure(feedbackCtx, payload, interactionFailureMessage(err)); notifyErr != nil {
			s.log.Error(baseCtx, "failed posting slack interaction failure feedback", "error", notifyErr, "interaction_type", interactionType, "slack_team_id", payload.Team.ID, "slack_user_id", payload.User.ID)
		}
	}()
}

func (s *Service) RecordRequestLog(ctx context.Context, input CoreRequestLogInput) {
	statusCode := input.ResponseCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	requestDetails := parseRequestLogDetails(input.RequestType, input.RawBody)
	workspaceID := s.resolveWorkspaceIDFromLog(ctx, requestDetails.SlackTeamID)

	entry := slackrepository.SlackRequestLogInsert{
		RequestType:  input.RequestType,
		Endpoint:     strings.TrimSpace(input.Endpoint),
		WorkspaceID:  workspaceID,
		SlackTeamID:  optionalString(requestDetails.SlackTeamID),
		SlackUserID:  optionalString(requestDetails.SlackUserID),
		SlackChannel: optionalString(requestDetails.SlackChannelID),
		Command:      optionalString(requestDetails.Command),
		Headers:      safeRequestLogHeaders(input.Headers),
		ResponseCode: statusCode,
		Outcome:      truncateForLog(strings.TrimSpace(input.Outcome), 120),
		ErrorMessage: optionalString(truncateForLog(input.ErrorMessage, 1000)),
	}

	if err := s.repo.InsertRequestLog(ctx, entry); err != nil {
		s.log.Warn(ctx, "failed to insert slack request log", "error", err, "request_type", input.RequestType)
	}
}

func safeRequestLogHeaders(headers map[string]string) []byte {
	allowed := map[string]struct{}{
		"X-Slack-Retry-Num":    {},
		"X-Slack-Retry-Reason": {},
		"User-Agent":           {},
		"Content-Type":         {},
	}
	filtered := make(map[string]string, len(allowed))
	for key, value := range headers {
		canonicalKey := http.CanonicalHeaderKey(strings.TrimSpace(key))
		if _, ok := allowed[canonicalKey]; !ok {
			continue
		}
		if value = truncateForLog(strings.TrimSpace(value), 250); value != "" {
			filtered[canonicalKey] = value
		}
	}
	payload, err := json.Marshal(filtered)
	if err != nil {
		return []byte("{}")
	}
	return payload
}

func (s *Service) handleMessageAction(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	messageAuthorID := strings.TrimSpace(payload.Message.User)
	messageAuthor := messageAuthorID
	if strings.EqualFold(messageAuthorID, strings.TrimSpace(payload.User.ID)) && strings.TrimSpace(payload.User.Username) != "" {
		messageAuthor = strings.TrimSpace(payload.User.Username)
	}

	title := messageToTitle(payload.Message.Text)
	description := buildPrefilledDescription(requestSourceContext{
		SlackUserID:   messageAuthorID,
		SlackUsername: messageAuthor,
		SlackText:     strings.TrimSpace(payload.Message.Text),
	})
	source := requestSourceContext{
		SlackTeamID:     strings.TrimSpace(payload.Team.ID),
		SlackTeamDomain: strings.TrimSpace(payload.Team.Domain),
		SlackChannelID:  strings.TrimSpace(payload.Channel.ID),
		SlackChannel:    strings.TrimSpace(payload.Channel.Name),
		SlackMessageTS:  strings.TrimSpace(payload.Message.TS),
		SlackThreadTS:   strings.TrimSpace(payload.Message.ThreadTS),
		SlackUserID:     strings.TrimSpace(payload.User.ID),
		SlackUsername:   strings.TrimSpace(payload.User.Username),
		SlackText:       strings.TrimSpace(payload.Message.Text),
		ResponseURL:     strings.TrimSpace(payload.ResponseURL),
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return InteractionResponse{}, ErrSlackNoWorkspaceLinked
		}
		return InteractionResponse{}, err
	}

	linkedUserID, connectURL, err := s.resolveLinkedSlackUser(ctx, slackWorkspace.WorkspaceID, source)
	if err != nil {
		return InteractionResponse{}, err
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return InteractionResponse{}, err
	}
	if linkedUserID == uuid.Nil {
		message := buildConnectSlackAccountMessage(connectURL)
		if responseURL := strings.TrimSpace(payload.ResponseURL); responseURL != "" {
			if responseErr := s.postCommandResponse(ctx, responseURL, message); responseErr == nil {
				return InteractionResponse{StatusCode: http.StatusOK}, nil
			} else {
				s.log.Warn(ctx, "failed posting slack connect prompt via response_url", "error", responseErr)
			}
		}
		if responseErr := s.postEphemeralMessage(ctx, botToken, source.SlackChannelID, source.SlackUserID, message); responseErr != nil {
			return InteractionResponse{}, fmt.Errorf("post Slack account connection prompt: %w", responseErr)
		}
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}

	if err := s.openCreateTaskModal(ctx, payload.TriggerID, title, description, source, slackWorkspace.WorkspaceID, linkedUserID, botToken); err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK}, nil
}

func isSlackMutationAction(payload interactionPayload) bool {
	actionID := strings.TrimSpace(payload.ActionID)
	if len(payload.Actions) > 0 {
		actionID = strings.TrimSpace(payload.Actions[0].ActionID)
	}
	return actionID == slackConfirmMutationActionID || actionID == slackCancelMutationActionID
}

func (s *Service) handleMutationAction(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if s.mutationConfirmer == nil {
		return InteractionResponse{}, errors.New("Slack mutation confirmer is not configured")
	}
	if len(payload.Actions) == 0 {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	action := payload.Actions[0]
	channelID := strings.TrimSpace(payload.Channel.ID)
	messageTS := strings.TrimSpace(payload.Message.TS)
	if channelID == "" || messageTS == "" {
		return InteractionResponse{}, errors.New("Slack mutation action is missing its message destination")
	}
	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, strings.TrimSpace(payload.Team.ID))
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return InteractionResponse{}, ErrSlackNoWorkspaceLinked
		}
		return InteractionResponse{}, err
	}
	linkedUserID, err := s.repo.FindLinkedUserIDBySlackUser(ctx, installation.WorkspaceID, installation.SlackTeamID, strings.TrimSpace(payload.User.ID))
	if err != nil {
		return InteractionResponse{}, err
	}
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		return InteractionResponse{}, ErrSlackUserNotLinked
	}
	actionValue, err := decodeSlackMutationActionValue(action.Value)
	if err != nil {
		return InteractionResponse{}, err
	}
	if actionValue.SlackUserID != strings.TrimSpace(payload.User.ID) {
		return InteractionResponse{}, ErrSlackInteractionActorMismatch
	}
	botToken, err := s.botToken(ctx, installation)
	if err != nil {
		return InteractionResponse{}, err
	}
	if action.ActionID == slackCancelMutationActionID {
		_, cancelErr := s.mutationConfirmer.CancelStoryMutation(ctx, messaging.ToolScope{
			WorkspaceID: installation.WorkspaceID,
			UserID:      *linkedUserID,
		}, actionValue.Token)
		if cancelErr != nil {
			switch {
			case errors.Is(cancelErr, messaging.ErrAppliedConfirmation):
				if err := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, "This story change was already confirmed."); err != nil {
					return InteractionResponse{}, errors.Join(cancelErr, err)
				}
				return InteractionResponse{StatusCode: http.StatusOK}, nil
			case errors.Is(cancelErr, messaging.ErrExpiredConfirmation),
				errors.Is(cancelErr, messaging.ErrInvalidConfirmation),
				errors.Is(cancelErr, messaging.ErrCancelledConfirmation):
				if err := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, "This story change is no longer available."); err != nil {
					return InteractionResponse{}, errors.Join(cancelErr, err)
				}
				return InteractionResponse{StatusCode: http.StatusOK}, nil
			default:
				return InteractionResponse{}, cancelErr
			}
		}
		if err := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, "Cancelled. No changes were made."); err != nil {
			return InteractionResponse{}, err
		}
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	if action.ActionID != slackConfirmMutationActionID || strings.TrimSpace(action.Value) == "" {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}

	settings, err := s.GetAgentSettings(ctx, installation.WorkspaceID)
	if err != nil {
		return InteractionResponse{}, err
	}
	if !settings.AssistantEnabled || !settings.WorkflowActionsEnabled {
		if err := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, "This story change is no longer available because Slack workflow actions are disabled."); err != nil {
			return InteractionResponse{}, err
		}
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	var allowedTeamIDs []uuid.UUID
	if !strings.HasPrefix(strings.ToUpper(channelID), "D") {
		allowedTeamIDs, err = s.authorizedChannelTeamIDs(ctx, installation.WorkspaceID, installation.ID, channelID, *linkedUserID)
		if err != nil {
			return InteractionResponse{}, err
		}
		if len(allowedTeamIDs) == 0 {
			if err := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, "This story change is no longer available from this channel."); err != nil {
				return InteractionResponse{}, err
			}
			return InteractionResponse{StatusCode: http.StatusOK}, nil
		}
	}
	result, err := s.mutationConfirmer.ConfirmStoryMutation(ctx, messaging.ToolScope{
		WorkspaceID:    installation.WorkspaceID,
		UserID:         *linkedUserID,
		AllowedTeamIDs: allowedTeamIDs,
		AllowMutations: settings.WorkflowActionsEnabled,
	}, actionValue.Token)
	if err != nil {
		if errors.Is(err, messaging.ErrMutationNotAllowed) ||
			errors.Is(err, messaging.ErrInvalidConfirmation) ||
			errors.Is(err, messaging.ErrExpiredConfirmation) ||
			errors.Is(err, messaging.ErrStaleMutation) ||
			errors.Is(err, messaging.ErrTeamNotAccessible) {
			if updateErr := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, "This story change is no longer valid. Ask Maya to try again."); updateErr != nil {
				return InteractionResponse{}, errors.Join(err, updateErr)
			}
			return InteractionResponse{StatusCode: http.StatusOK}, nil
		}
		return InteractionResponse{}, err
	}
	creatorName := ""
	if member, memberErr := s.repo.FindTeamMemberByID(ctx, result.TeamID, *linkedUserID); memberErr == nil {
		creatorName = slackMemberDisplayName(member)
	}
	if creatorName == "" {
		creatorName = strings.TrimSpace(payload.User.Name)
	}
	if creatorName == "" {
		creatorName = strings.TrimSpace(payload.User.Username)
	}
	reference := strings.ToUpper(strings.TrimSpace(result.Reference))
	workspace, err := s.repo.FindWorkspaceByID(ctx, installation.WorkspaceID)
	if err != nil {
		return InteractionResponse{}, err
	}
	storyURL := buildTaskURL(s.cfg.WebsiteURL, workspace.Slug, reference)
	text := buildSlackStoryMutationReceiptText(creatorName, reference, storyURL, result.Operation)
	if err := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, text); err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK}, nil
}

func buildSlackStoryMutationReceiptText(creatorName, reference, storyURL string, operation messaging.StoryMutationOperation) string {
	creatorLabel := slackMrkdwnTextEscaper.Replace(strings.TrimSpace(creatorName))
	if creatorLabel == "" {
		creatorLabel = "A team member"
	}
	storyLabel := strings.ToUpper(strings.TrimSpace(reference))
	if storyLabel == "" {
		storyLabel = "a story"
	}
	if storyURL = strings.TrimSpace(storyURL); storyURL != "" {
		storyLabel = fmt.Sprintf("<%s|%s>", storyURL, storyLabel)
	}
	verb := "updated"
	if operation == messaging.StoryMutationCreate {
		verb = "created"
	}
	return fmt.Sprintf("%s %s %s", creatorLabel, verb, storyLabel)
}

func (s *Service) updateSlackInteractiveMessage(ctx context.Context, botToken, channelID, messageTS, text string) error {
	payload := map[string]any{
		"channel": strings.TrimSpace(channelID),
		"ts":      strings.TrimSpace(messageTS),
		"text":    truncateSlackText(text),
		"blocks":  []any{},
	}
	return s.slackClient().callJSON(ctx, botToken, "chat.update", payload, nil)
}

func (s *Service) handleBlockActions(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if payload.View.CallbackID != "fortyone_create_task" {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	if len(payload.Actions) == 0 {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	firstAction := payload.Actions[0]
	if firstAction.BlockID != modalBlockTeam || firstAction.ActionID != modalActionTeamSelect {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}

	submission, err := parseViewSubmission(payload)
	if err != nil {
		s.log.Error(ctx, "failed parsing slack block actions payload", "error", err)
		return InteractionResponse{}, fmt.Errorf("parse Slack block action: %w", err)
	}
	selectedTeamIDRaw := strings.TrimSpace(firstAction.SelectedOption.Value)
	if selectedTeamIDRaw != "" {
		selectedTeamID, parseErr := uuid.Parse(selectedTeamIDRaw)
		if parseErr != nil || selectedTeamID == uuid.Nil {
			s.log.Warn(ctx, "slack team change contained an invalid team id", "team_id", selectedTeamIDRaw)
			return InteractionResponse{}, ErrSlackTeamSelectionRequired
		}
		submission.TeamID = selectedTeamID
	}

	metadata, err := parseSlackModalPrivateMetadata(payload.View.PrivateMetadata)
	if err != nil {
		s.log.Error(ctx, "failed parsing slack modal metadata for team change", "error", err)
		return InteractionResponse{}, fmt.Errorf("parse Slack modal metadata: %w", err)
	}
	submission.Source, err = interactionSourceForPayload(payload, submission.Source)
	if err != nil {
		s.log.Warn(ctx, "rejected slack team change with mismatched actor", "error", err)
		return InteractionResponse{}, ErrSlackInteractionActorMismatch
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, submission.Source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return InteractionResponse{}, ErrSlackNoWorkspaceLinked
		}
		return InteractionResponse{}, err
	}
	actorID, err := s.findLinkedInteractionActor(ctx, slackWorkspace.WorkspaceID, submission.Source)
	if err != nil {
		if errors.Is(err, ErrSlackUserNotLinked) {
			s.log.Warn(ctx, "rejected slack team change from unlinked user", "slack_user_id", submission.Source.SlackUserID)
			return InteractionResponse{}, ErrSlackUserNotLinked
		}
		return InteractionResponse{}, err
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, slackWorkspace.WorkspaceID, actorID, submission.TeamID, submission.Source); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			s.log.Warn(ctx, "rejected slack team change outside actor membership", "actor_id", actorID, "team_id", submission.TeamID)
			return InteractionResponse{}, ErrSlackTeamNotAvailable
		}
		return InteractionResponse{}, err
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return InteractionResponse{}, err
	}

	selection := createTaskModalSelection{
		StatusKind:  submission.StatusKind,
		TeamID:      submission.TeamID,
		StatusID:    submission.StatusID,
		Priority:    submission.Priority,
		AssigneeID:  submission.AssigneeID,
		LabelIDs:    submission.LabelIDs,
		ObjectiveID: submission.ObjectiveID,
	}
	previousTeamID, _ := uuid.Parse(strings.TrimSpace(metadata.SelectedTeamID))
	if previousTeamID != uuid.Nil && previousTeamID != submission.TeamID {
		selection = createTaskModalSelection{
			TeamID:   submission.TeamID,
			Priority: submission.Priority,
		}
	}

	view, err := s.buildCreateTaskModalView(ctx, createTaskModalViewInput{
		Title:       submission.Title,
		Description: submission.Description,
		Source:      submission.Source,
		WorkspaceID: slackWorkspace.WorkspaceID,
		ActorID:     actorID,
		Selection:   selection,
	})
	if err != nil {
		return InteractionResponse{}, err
	}

	updatePayload := map[string]any{
		"view_id": payload.View.ID,
		"hash":    payload.View.Hash,
		"view":    view,
	}
	if err := s.callSlackAPI(ctx, botToken, "https://slack.com/api/views.update", updatePayload, nil); err != nil {
		return InteractionResponse{}, err
	}

	return InteractionResponse{StatusCode: http.StatusOK}, nil
}

func (s *Service) handleViewSubmission(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if payload.View.CallbackID != "fortyone_create_task" {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}

	submission, err := parseViewSubmission(payload)
	if err != nil {
		s.log.Error(ctx, "failed parsing slack view submission payload", "error", err)
		return interactionValidationErrors(map[string]string{
			"title": interactionErrorMessage(err),
		})
	}

	errorsByBlock := map[string]string{}
	if submission.Title == "" {
		errorsByBlock["title"] = "Title is required"
	}
	if len([]rune(submission.Title)) > modalTitleMaxRunes {
		errorsByBlock[modalBlockTitle] = "Title must be 255 characters or fewer"
	}
	if len([]rune(submission.Description)) > modalDescriptionMaxRunes {
		errorsByBlock[modalBlockDescription] = "Description must be 3000 characters or fewer"
	}
	if len(errorsByBlock) > 0 {
		return interactionValidationErrors(errorsByBlock)
	}
	submission.Source, err = interactionSourceForPayload(payload, submission.Source)
	if err != nil {
		s.log.Warn(ctx, "rejected slack view submission with mismatched actor", "error", err)
		return interactionValidationErrors(map[string]string{"title": "This Slack form no longer belongs to the current user. Open it again and retry."})
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, submission.Source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return interactionValidationErrors(map[string]string{"title": "Slack workspace is not connected"})
		}
		s.log.Error(ctx, "failed loading slack workspace by team id", "error", err, "slack_team_id", submission.Source.SlackTeamID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	actorID, err := s.findLinkedInteractionActor(ctx, slackWorkspace.WorkspaceID, submission.Source)
	if err != nil {
		if errors.Is(err, ErrSlackUserNotLinked) {
			return interactionValidationErrors(map[string]string{"title": "Connect your FortyOne account first, then submit again."})
		}
		s.log.Error(ctx, "failed resolving slack actor for view submission", "error", err, "workspace_id", slackWorkspace.WorkspaceID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		s.log.Error(ctx, "failed loading Slack credential for view submission", "error", err, "workspace_id", slackWorkspace.WorkspaceID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}

	workspace, err := s.repo.FindWorkspaceByID(ctx, slackWorkspace.WorkspaceID)
	if err != nil {
		s.log.Error(ctx, "failed loading workspace for slack submission", "error", err, "workspace_id", slackWorkspace.WorkspaceID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	if submission.TeamID == uuid.Nil {
		return interactionValidationErrors(map[string]string{modalBlockTeam: "Team is required"})
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, workspace.ID, actorID, submission.TeamID, submission.Source); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return interactionValidationErrors(map[string]string{modalBlockTeam: "Selected team is not available from this Slack channel"})
		}
		s.log.Error(ctx, "failed validating Slack channel team audience", "error", err, "workspace_id", workspace.ID, "team_id", submission.TeamID, "actor_id", actorID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	team, err := s.findTeamForActor(ctx, workspace.ID, actorID, submission.TeamID)
	if err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return interactionValidationErrors(map[string]string{modalBlockTeam: "Selected team is no longer available to you"})
		}
		s.log.Error(ctx, "failed validating selected team for slack submission", "error", err, "workspace_id", workspace.ID, "team_id", submission.TeamID, "actor_id", actorID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}

	sourceURL := permalinkFromContext(submission.Source)
	sourceExternalID := buildSourceExternalID(submission.Source)
	if viewID := strings.TrimSpace(payload.View.ID); viewID != "" {
		sourceExternalID = fmt.Sprintf("view:%s:%s", viewID, actorID)
	}
	if sourceExternalID == "" {
		sourceExternalID = fmt.Sprintf("slack:%d", s.clock.Now().UnixNano())
	}
	creationKey := fmt.Sprintf("slack:%s:%s", workspace.ID, sourceExternalID)

	description := strings.TrimSpace(submission.Description)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	metadata := map[string]any{
		"workspace_slug":    workspace.Slug,
		"workspace_name":    workspace.Name,
		"team_code":         team.Code,
		"team_name":         team.Name,
		"slack_team_id":     submission.Source.SlackTeamID,
		"slack_team_domain": submission.Source.SlackTeamDomain,
		"slack_channel_id":  submission.Source.SlackChannelID,
		"slack_channel":     submission.Source.SlackChannel,
		"slack_message_ts":  submission.Source.SlackMessageTS,
		"slack_thread_ts":   submission.Source.SlackThreadTS,
		"slack_user_id":     submission.Source.SlackUserID,
		"slack_username":    submission.Source.SlackUsername,
		"slack_text":        submission.Source.SlackText,
	}

	sendToRequests := submission.StatusKind == slackStatusKindRequest

	if submission.StatusID != nil {
		statuses, statusErr := s.repo.ListTeamStatuses(ctx, team.ID)
		if statusErr != nil {
			s.log.Error(ctx, "failed loading team statuses for slack submission", "error", statusErr, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(statusErr)})
		}
		_, found := findStatusByID(statuses, *submission.StatusID)
		if found {
			sendToRequests = false
		} else {
			return interactionValidationErrors(map[string]string{submission.BlockIDs.Status: "Selected status is no longer available"})
		}
	}
	if submission.StatusKind == slackStatusKindStory && submission.StatusID == nil {
		return interactionValidationErrors(map[string]string{submission.BlockIDs.Status: "Selected status is no longer available"})
	}

	var assigneeID *uuid.UUID
	if submission.AssigneeID != nil {
		members, membersErr := s.repo.ListTeamMembers(ctx, team.ID)
		if membersErr != nil {
			s.log.Error(ctx, "failed loading team members for slack submission", "error", membersErr, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(membersErr)})
		}
		if teamMemberExists(members, *submission.AssigneeID) {
			assigneeID = submission.AssigneeID
		} else {
			return interactionValidationErrors(map[string]string{submission.BlockIDs.Assignee: "Selected assignee is no longer available"})
		}
	}

	var objectiveID *uuid.UUID
	if submission.ObjectiveID != nil {
		if _, objectiveErr := s.repo.FindTeamObjectiveByID(ctx, workspace.ID, team.ID, *submission.ObjectiveID); objectiveErr != nil {
			if !slackrepository.IsNotFound(objectiveErr) {
				s.log.Error(ctx, "failed validating objective for slack submission", "error", objectiveErr, "workspace_id", workspace.ID, "team_id", team.ID)
				return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(objectiveErr)})
			}
			return interactionValidationErrors(map[string]string{submission.BlockIDs.Objective: "Selected objective is no longer available"})
		} else {
			objectiveID = submission.ObjectiveID
		}
	}

	priority := normalizeSlackPriority(submission.Priority)
	var labelIDs []uuid.UUID
	if len(submission.LabelIDs) > 0 {
		labels, labelsErr := s.repo.ListTeamLabels(ctx, workspace.ID, team.ID)
		if labelsErr != nil {
			s.log.Error(ctx, "failed loading team labels for slack submission", "error", labelsErr, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(labelsErr)})
		}
		labelIDs = filterValidLabelIDs(labels, submission.LabelIDs)
		if len(labelIDs) != len(submission.LabelIDs) {
			return interactionValidationErrors(map[string]string{submission.BlockIDs.Labels: "One or more selected labels are no longer available"})
		}
	}

	if sendToRequests {
		metadata["label_ids"] = uuidStrings(labelIDs)
		requestInput := integrationrequests.CoreUpsertRequestInput{
			WorkspaceID:      workspace.ID,
			TeamID:           team.ID,
			Provider:         integrationrequests.ProviderSlack,
			SourceType:       SourceTypeSlackMessage,
			SourceExternalID: sourceExternalID,
			Title:            submission.Title,
			Description:      descriptionPtr,
			Priority:         priority,
			AssigneeID:       assigneeID,
			ObjectiveID:      objectiveID,
			LabelIDs:         labelIDs,
			Metadata:         metadata,
			CreatedByUserID:  &actorID,
		}
		if sourceURL != "" {
			requestInput.SourceURL = &sourceURL
		}
		request, err := s.requests.UpsertPending(ctx, requestInput)
		if err != nil {
			s.log.Error(ctx, "failed creating slack integration request", "error", err, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
		}
		ackKey := fmt.Sprintf("slack:%s:request:%s:confirmation", workspace.ID, sourceExternalID)
		s.postSlackRequestAck(ctx, workspace.ID, slackWorkspace.InstallGeneration, ackKey, submission.Source, botToken, workspace.Slug, team.ID, request.ID, actorID)
		return interactionClearResponse()
	}

	creator, err := s.repo.FindTeamMemberByID(ctx, team.ID, actorID)
	if err != nil {
		s.log.Error(ctx, "failed loading story creator for Slack confirmation", "error", err, "workspace_id", workspace.ID, "team_id", team.ID, "actor_id", actorID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	creatorName := storyCreatorDisplayName(creator)

	var statusID *uuid.UUID
	if submission.StatusID != nil {
		statusID = submission.StatusID
	} else {
		statusID, err = s.repo.FindFirstStatusByCategory(ctx, team.ID, "unstarted")
		if err != nil {
			s.log.Error(ctx, "failed loading unstarted status for slack task creation", "error", err, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
		}
	}

	story, err := s.stories.Create(ctx, stories.CoreNewStory{
		Title:       submission.Title,
		Description: descriptionPtr,
		Status:      statusID,
		Assignee:    assigneeID,
		Objective:   objectiveID,
		Team:        team.ID,
		Reporter:    &actorID,
		Priority:    priority,
		LabelIDs:    labelIDs,
		CreationKey: &creationKey,
	}, workspace.ID)
	if err != nil {
		s.log.Error(ctx, "failed creating story from slack submission", "error", err, "workspace_id", workspace.ID, "team_id", team.ID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}

	s.postSlackTaskAck(ctx, workspace.ID, slackWorkspace.InstallGeneration, creationKey+":confirmation", submission.Source, botToken, workspace.Slug, team.Code, creatorName, story)
	return interactionClearResponse()
}

func (s *Service) handleBlockSuggestion(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if callbackID := strings.TrimSpace(payload.View.CallbackID); callbackID != "" && callbackID != "fortyone_create_task" {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome: "suggestion_skipped_invalid_callback",
			Reason:  "callback_id_not_supported",
		})
		return interactionOptionsResponse(nil)
	}

	source, err := parseSourceFromPrivateMetadata(payload.View.PrivateMetadata)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome: "suggestion_skipped_invalid_metadata",
			Reason:  err.Error(),
		})
		return interactionOptionsResponse(nil)
	}
	if strings.TrimSpace(source.SlackTeamID) == "" {
		source.SlackTeamID = strings.TrimSpace(payload.Team.ID)
	}
	source, err = interactionSourceForPayload(payload, source)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:     "suggestion_skipped_actor_mismatch",
			Reason:      err.Error(),
			SlackTeamID: source.SlackTeamID,
		})
		return interactionOptionsResponse(nil)
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_workspace_not_found",
			Reason:       err.Error(),
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  uuid.Nil,
			ResolvedTeam: uuid.Nil,
		})
		return interactionOptionsResponse(nil)
	}
	actorID, err := s.findLinkedInteractionActor(ctx, slackWorkspace.WorkspaceID, source)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:     "suggestion_skipped_unlinked_actor",
			Reason:      err.Error(),
			SlackTeamID: source.SlackTeamID,
			WorkspaceID: slackWorkspace.WorkspaceID,
		})
		return interactionOptionsResponse(nil)
	}
	rawActionID := suggestionActionID(payload)
	query := suggestionQuery(payload)
	if rawActionID == modalActionTeamSelect {
		if blockID := suggestionBlockID(payload); blockID != "" && blockID != modalBlockTeam {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:     "suggestion_skipped_unknown_action",
				Reason:      "team_action_block_id_not_valid",
				ActionID:    rawActionID,
				SlackTeamID: source.SlackTeamID,
				WorkspaceID: slackWorkspace.WorkspaceID,
			})
			return interactionOptionsResponse(nil)
		}
		teams, teamsErr := s.availableTeamsForSlackSource(ctx, slackWorkspace.WorkspaceID, actorID, source)
		if teamsErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:     "suggestion_search_error_teams",
				Reason:      teamsErr.Error(),
				Query:       query,
				ActionID:    rawActionID,
				SlackTeamID: source.SlackTeamID,
				WorkspaceID: slackWorkspace.WorkspaceID,
			})
			return interactionOptionsResponse(nil)
		}
		options := slackTeamSuggestionOptions(teams, query)
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:     "suggestion_search_teams",
			Query:       query,
			ActionID:    rawActionID,
			SlackTeamID: source.SlackTeamID,
			WorkspaceID: slackWorkspace.WorkspaceID,
			ResultCount: len(options),
		})
		return interactionOptionsResponse(options)
	}

	teamID, err := s.resolveTeamIDForSuggestion(ctx, payload, slackWorkspace.WorkspaceID, actorID, source)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_team_resolution_failed",
			Reason:       err.Error(),
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: uuid.Nil,
		})
		return interactionOptionsResponse(nil)
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, slackWorkspace.WorkspaceID, actorID, teamID, source); err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_team_not_available",
			Reason:       err.Error(),
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}
	actionID, actionMatchesTeam := modalDependentActionBase(rawActionID, teamID)
	expectedBlockID := modalDependentBlockForAction(actionID)
	if !actionMatchesTeam || expectedBlockID == "" || !modalDependentBlockMatches(suggestionBlockID(payload), expectedBlockID, teamID) {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_unknown_action",
			Reason:       "action_or_block_id_not_valid_for_selected_team",
			ActionID:     rawActionID,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}

	if actionID != modalActionStatusSelect && len([]rune(query)) < 2 {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_query_too_short",
			Query:        query,
			ActionID:     rawActionID,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}

	const optionsLimit = 25
	switch actionID {
	case modalActionStatusSelect:
		statuses, statusesErr := s.repo.ListTeamStatuses(ctx, teamID)
		if statusesErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:      "suggestion_search_error_statuses",
				Reason:       statusesErr.Error(),
				Query:        query,
				ActionID:     modalActionStatusSelect,
				SlackTeamID:  source.SlackTeamID,
				WorkspaceID:  slackWorkspace.WorkspaceID,
				ResolvedTeam: teamID,
			})
			return interactionOptionsResponse(nil)
		}
		options := slackStatusSuggestionOptions(statuses, query)
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_search_statuses",
			Query:        query,
			ActionID:     modalActionStatusSelect,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
			ResultCount:  len(options),
		})
		return interactionOptionsResponse(options)
	case modalActionAssigneeSelect:
		members, membersErr := s.repo.SearchTeamMembers(ctx, teamID, query, optionsLimit)
		if membersErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:      "suggestion_search_error_members",
				Reason:       membersErr.Error(),
				Query:        query,
				ActionID:     modalActionAssigneeSelect,
				SlackTeamID:  source.SlackTeamID,
				WorkspaceID:  slackWorkspace.WorkspaceID,
				ResolvedTeam: teamID,
			})
			return interactionOptionsResponse(nil)
		}
		options := make([]map[string]any, 0, len(members))
		for _, member := range members {
			options = append(options, toSlackOption(teamMemberDisplayName(member), member.UserID.String()))
		}
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_search_members",
			Query:        query,
			ActionID:     modalActionAssigneeSelect,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
			ResultCount:  len(options),
		})
		return interactionOptionsResponse(options)
	case modalActionLabelsMultiSelect:
		labels, labelsErr := s.repo.SearchTeamLabels(ctx, slackWorkspace.WorkspaceID, teamID, query, optionsLimit)
		if labelsErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:      "suggestion_search_error_labels",
				Reason:       labelsErr.Error(),
				Query:        query,
				ActionID:     modalActionLabelsMultiSelect,
				SlackTeamID:  source.SlackTeamID,
				WorkspaceID:  slackWorkspace.WorkspaceID,
				ResolvedTeam: teamID,
			})
			return interactionOptionsResponse(nil)
		}
		options := make([]map[string]any, 0, len(labels))
		for _, label := range labels {
			options = append(options, toSlackOption(label.Name, label.ID.String()))
		}
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_search_labels",
			Query:        query,
			ActionID:     modalActionLabelsMultiSelect,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
			ResultCount:  len(options),
		})
		return interactionOptionsResponse(options)
	case modalActionObjectiveSelect:
		objectives, objectivesErr := s.repo.SearchTeamObjectives(ctx, slackWorkspace.WorkspaceID, teamID, query, optionsLimit)
		if objectivesErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:      "suggestion_search_error_objectives",
				Reason:       objectivesErr.Error(),
				Query:        query,
				ActionID:     modalActionObjectiveSelect,
				SlackTeamID:  source.SlackTeamID,
				WorkspaceID:  slackWorkspace.WorkspaceID,
				ResolvedTeam: teamID,
			})
			return interactionOptionsResponse(nil)
		}
		options := make([]map[string]any, 0, len(objectives))
		for _, objective := range objectives {
			options = append(options, toSlackOption(objective.Name, objective.ID.String()))
		}
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_search_objectives",
			Query:        query,
			ActionID:     modalActionObjectiveSelect,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
			ResultCount:  len(options),
		})
		return interactionOptionsResponse(options)
	default:
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_unknown_action",
			Query:        query,
			ActionID:     rawActionID,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}
}

type suggestionDebugInput struct {
	Outcome      string
	Reason       string
	Query        string
	ActionID     string
	SlackTeamID  string
	WorkspaceID  uuid.UUID
	ResolvedTeam uuid.UUID
	ResultCount  int
}

func (s *Service) recordSuggestionDebug(ctx context.Context, payload interactionPayload, input suggestionDebugInput) {
	var workspaceIDPtr *uuid.UUID
	if input.WorkspaceID != uuid.Nil {
		workspaceIDPtr = &input.WorkspaceID
	}
	slackTeamID := optionalString(input.SlackTeamID)
	slackUserID := optionalString(payload.User.ID)
	slackChannelID := optionalString(payload.Channel.ID)
	errorMessage := optionalString(input.Reason)

	if insertErr := s.repo.InsertRequestLog(ctx, slackrepository.SlackRequestLogInsert{
		RequestType:  "suggestion_search",
		Endpoint:     "/integrations/slack/interactivity",
		WorkspaceID:  workspaceIDPtr,
		SlackTeamID:  slackTeamID,
		SlackUserID:  slackUserID,
		SlackChannel: slackChannelID,
		Headers:      []byte("{}"),
		ResponseCode: http.StatusOK,
		Outcome:      truncateForLog(strings.TrimSpace(input.Outcome), 120),
		ErrorMessage: errorMessage,
	}); insertErr != nil {
		s.log.Warn(ctx, "failed writing suggestion diagnostic log", "error", insertErr)
	}
}

func suggestionActionID(payload interactionPayload) string {
	if actionID := strings.TrimSpace(payload.ActionID); actionID != "" {
		return actionID
	}
	if len(payload.Actions) > 0 {
		return strings.TrimSpace(payload.Actions[0].ActionID)
	}
	return ""
}

func suggestionBlockID(payload interactionPayload) string {
	if blockID := strings.TrimSpace(payload.BlockID); blockID != "" {
		return blockID
	}
	if len(payload.Actions) > 0 {
		return strings.TrimSpace(payload.Actions[0].BlockID)
	}
	return ""
}

func suggestionQuery(payload interactionPayload) string {
	if query := strings.TrimSpace(payload.Value); query != "" {
		return query
	}
	if len(payload.Actions) > 0 {
		if query := strings.TrimSpace(payload.Actions[0].Value); query != "" {
			return query
		}
	}
	return ""
}

func (s *Service) AcceptIntegrationRequest(ctx context.Context, request integrationrequests.CoreIntegrationRequest, story stories.CoreSingleStory) error {
	if request.Provider != integrationrequests.ProviderSlack {
		return nil
	}
	channelID := metadataString(request.Metadata, "slack_channel_id")
	threadTS := metadataString(request.Metadata, "slack_thread_ts")
	if threadTS == "" {
		threadTS = metadataString(request.Metadata, "slack_message_ts")
	}
	providerThread, threadErr := s.requests.FindProviderThread(ctx, request.WorkspaceID, request.ID, integrationrequests.ProviderSlack)
	if threadErr == nil {
		channelID = strings.TrimSpace(providerThread.ExternalChannelID)
		threadTS = strings.TrimSpace(providerThread.ExternalThreadID)
	} else if !errors.Is(threadErr, integrationrequests.ErrProviderThreadNotFound) {
		return fmt.Errorf("find canonical Slack request thread: %w", threadErr)
	}
	if channelID == "" || threadTS == "" {
		return nil
	}

	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return nil
		}
		return err
	}
	requestSlackTeamID := metadataString(request.Metadata, "slack_team_id")
	if requestSlackTeamID == "" || requestSlackTeamID != slackWorkspace.SlackTeamID {
		s.log.Warn(ctx, "skipping Slack acceptance update for a replaced installation", "request_id", request.ID, "request_slack_team_id", requestSlackTeamID, "active_slack_team_id", slackWorkspace.SlackTeamID)
		return nil
	}
	if threadErr == nil && (providerThread.InstallationGeneration == nil || *providerThread.InstallationGeneration != slackWorkspace.InstallGeneration || providerThread.ExternalWorkspaceID != slackWorkspace.SlackTeamID) {
		s.log.Warn(ctx, "skipping Slack acceptance update for a stale provider thread", "request_id", request.ID)
		return nil
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return err
	}
	workspaceSlug := metadataString(request.Metadata, "workspace_slug")
	teamCode := strings.TrimSpace(story.TeamCode)
	if teamCode == "" {
		teamCode = metadataString(request.Metadata, "team_code")
	}
	creatorName := "A team member"
	if story.Reporter != nil && *story.Reporter != uuid.Nil {
		creator, creatorErr := s.repo.FindTeamMemberByID(ctx, request.TeamID, *story.Reporter)
		if creatorErr != nil {
			return fmt.Errorf("find accepted Slack request story creator: %w", creatorErr)
		}
		creatorName = storyCreatorDisplayName(creator)
	}
	s.postSlackTaskAck(
		ctx,
		request.WorkspaceID,
		slackWorkspace.InstallGeneration,
		fmt.Sprintf("slack:%s:request:%s:accepted", request.WorkspaceID, request.ID),
		requestSourceContext{
			SlackTeamID:    slackWorkspace.SlackTeamID,
			SlackChannelID: channelID,
			SlackThreadTS:  threadTS,
			SlackUserID:    metadataString(request.Metadata, "slack_user_id"),
		},
		botToken,
		workspaceSlug,
		teamCode,
		creatorName,
		story,
	)
	return nil
}

func (s *Service) openCreateTaskModal(ctx context.Context, triggerID, title, description string, source requestSourceContext, workspaceID, actorID uuid.UUID, botToken string) error {
	if strings.TrimSpace(botToken) == "" {
		return errors.New("missing slack bot token")
	}
	if strings.TrimSpace(triggerID) == "" {
		return errors.New("missing trigger id")
	}
	if workspaceID == uuid.Nil {
		return errors.New("missing workspace id")
	}
	if actorID == uuid.Nil {
		return errors.New("missing actor id")
	}

	view, err := s.buildCreateTaskModalView(ctx, createTaskModalViewInput{
		Title:       title,
		Description: description,
		Source:      source,
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		return err
	}

	payload := map[string]any{
		"trigger_id": triggerID,
		"view":       view,
	}
	return s.callSlackAPI(ctx, botToken, "https://slack.com/api/views.open", payload, nil)
}

type createTaskModalViewInput struct {
	Title       string
	Description string
	Source      requestSourceContext
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	Selection   createTaskModalSelection
}

type createTaskModalSelection struct {
	StatusKind  string
	TeamID      uuid.UUID
	StatusID    *uuid.UUID
	Priority    string
	AssigneeID  *uuid.UUID
	LabelIDs    []uuid.UUID
	ObjectiveID *uuid.UUID
}

func (s *Service) buildCreateTaskModalView(ctx context.Context, input createTaskModalViewInput) (map[string]any, error) {
	if input.ActorID == uuid.Nil {
		return nil, errors.New("actor id is required")
	}
	teams, err := s.availableTeamsForSlackSource(ctx, input.WorkspaceID, input.ActorID, input.Source)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, ErrSlackNoTeamsAvailable
	}

	selectedTeam := selectTeam(teams, input.Selection.TeamID)
	selectedTeamOption := slackTeamOption(selectedTeam)

	statuses, err := s.repo.ListTeamStatuses(ctx, selectedTeam.ID)
	if err != nil {
		return nil, err
	}

	useExternalStatusSelect := len(statuses)+1 > slackSelectMaxOptions
	statusOptions := make([]map[string]any, 0, min(len(statuses)+1, slackSelectMaxOptions))
	requestStatusOption := toSlackOption("Request", slackRequestStatusValue)
	if !useExternalStatusSelect {
		statusOptions = append(statusOptions, requestStatusOption)
	}
	selectedStatusOption := requestStatusOption
	for index, status := range statuses {
		option := toSlackOption(status.Name, status.ID.String())
		if !useExternalStatusSelect {
			statusOptions = append(statusOptions, option)
		}
		if index == 0 && input.Selection.StatusKind == "" && input.Selection.StatusID == nil {
			selectedStatusOption = option
		}
		if input.Selection.StatusKind == slackStatusKindStory && input.Selection.StatusID != nil && *input.Selection.StatusID == status.ID {
			selectedStatusOption = option
		}
	}

	var selectedAssigneeOption map[string]any
	if input.Selection.AssigneeID != nil && *input.Selection.AssigneeID != uuid.Nil {
		member, memberErr := s.repo.FindTeamMemberByID(ctx, selectedTeam.ID, *input.Selection.AssigneeID)
		if memberErr == nil {
			selectedAssigneeOption = toSlackOption(teamMemberDisplayName(member), member.UserID.String())
		}
	}
	selectedLabelOptions := make([]map[string]any, 0, len(input.Selection.LabelIDs))
	for _, labelID := range input.Selection.LabelIDs {
		label, labelErr := s.repo.FindTeamLabelByID(ctx, input.WorkspaceID, selectedTeam.ID, labelID)
		if labelErr != nil {
			continue
		}
		selectedLabelOptions = append(selectedLabelOptions, toSlackOption(label.Name, label.ID.String()))
	}

	var selectedObjectiveOption map[string]any
	if input.Selection.ObjectiveID != nil && *input.Selection.ObjectiveID != uuid.Nil {
		objective, objectiveErr := s.repo.FindTeamObjectiveByID(ctx, input.WorkspaceID, selectedTeam.ID, *input.Selection.ObjectiveID)
		if objectiveErr == nil {
			selectedObjectiveOption = toSlackOption(objective.Name, objective.ID.String())
		}
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "New task"
	}
	metadataSource := input.Source
	metadataSource.SlackText = truncateRunes(metadataSource.SlackText, modalSourceTextMaxRunes)
	metadataPayload, err := json.Marshal(slackModalPrivateMetadata{
		Source:         metadataSource,
		SelectedTeamID: selectedTeam.ID.String(),
	})
	if err != nil {
		return nil, err
	}
	if len(metadataPayload) > modalMetadataMaxBytes {
		metadataSource.SlackText = ""
		metadataPayload, err = json.Marshal(slackModalPrivateMetadata{
			Source:         metadataSource,
			SelectedTeamID: selectedTeam.ID.String(),
		})
		if err != nil {
			return nil, err
		}
	}
	if len(metadataPayload) > modalMetadataMaxBytes {
		return nil, errors.New("Slack modal metadata exceeds the provider limit")
	}

	priorityOption := toSlackOption(normalizeSlackPriority(input.Selection.Priority), normalizeSlackPriority(input.Selection.Priority))
	statusBlockID := modalTeamScopedID(modalBlockStatus, selectedTeam.ID)
	statusActionID := modalTeamScopedID(modalActionStatusSelect, selectedTeam.ID)
	assigneeBlockID := modalTeamScopedID(modalBlockAssignee, selectedTeam.ID)
	assigneeActionID := modalTeamScopedID(modalActionAssigneeSelect, selectedTeam.ID)
	labelsBlockID := modalTeamScopedID(modalBlockLabels, selectedTeam.ID)
	labelsActionID := modalTeamScopedID(modalActionLabelsMultiSelect, selectedTeam.ID)
	objectiveBlockID := modalTeamScopedID(modalBlockObjective, selectedTeam.ID)
	objectiveActionID := modalTeamScopedID(modalActionObjectiveSelect, selectedTeam.ID)
	statusBlock := selectInputBlock(statusBlockID, statusActionID, "Status", statusOptions, selectedStatusOption, true, false)
	if useExternalStatusSelect {
		statusBlock = externalSelectInputBlock(statusBlockID, statusActionID, "Status", selectedStatusOption, true, slackExternalSearchMinRunes, false)
	}

	blocks := []map[string]any{
		externalSelectInputBlock(modalBlockTeam, modalActionTeamSelect, "Team", selectedTeamOption, false, slackExternalSearchMinRunes, true),
		plainInputBlock(modalBlockTitle, modalActionTitleInput, "Title", truncateRunes(title, modalTitleMaxRunes), false, "", false),
		plainInputBlock(modalBlockDescription, modalActionDescriptionInput, "Description", truncateRunes(input.Description, modalDescriptionMaxRunes), true, "", true),
		statusBlock,
		externalSelectInputBlock(assigneeBlockID, assigneeActionID, "Assignee", selectedAssigneeOption, true, 2, false),
		externalMultiSelectInputBlock(labelsBlockID, labelsActionID, "Labels", selectedLabelOptions, true, 2),
		externalSelectInputBlock(objectiveBlockID, objectiveActionID, "Objective", selectedObjectiveOption, true, 2, false),
	}
	blocks = append(blocks, selectInputBlock(modalBlockPriority, modalActionPrioritySelect, "Priority", slackPriorityOptions(), priorityOption, true, false))

	return map[string]any{
		"type":             "modal",
		"callback_id":      "fortyone_create_task",
		"private_metadata": string(metadataPayload),
		"title": map[string]string{
			"type": "plain_text",
			"text": "Create Story",
		},
		"submit": map[string]string{
			"type": "plain_text",
			"text": "Create",
		},
		"close": map[string]string{
			"type": "plain_text",
			"text": "Cancel",
		},
		"blocks": blocks,
	}, nil
}

func selectInputBlock(blockID, actionID, label string, options []map[string]any, initialOption map[string]any, optional, dispatchAction bool) map[string]any {
	element := map[string]any{
		"type":      "static_select",
		"action_id": actionID,
		"options":   limitedSlackOptions(options),
	}
	if initialOption != nil {
		element["initial_option"] = initialOption
	}

	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	if dispatchAction {
		block["dispatch_action"] = true
	}
	return block
}

func multiSelectInputBlock(blockID, actionID, label string, options []map[string]any, initialOptions []map[string]any, optional bool) map[string]any {
	element := map[string]any{
		"type":      "multi_static_select",
		"action_id": actionID,
		"options":   limitedSlackOptions(options),
	}
	if len(initialOptions) > 0 {
		element["initial_options"] = limitedSlackOptions(initialOptions)
	}

	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	return block
}

func externalSelectInputBlock(blockID, actionID, label string, initialOption map[string]any, optional bool, minQueryLength int, dispatchAction bool) map[string]any {
	element := map[string]any{
		"type":      "external_select",
		"action_id": actionID,
	}
	if initialOption != nil {
		element["initial_option"] = initialOption
	}
	if minQueryLength >= 0 {
		element["min_query_length"] = minQueryLength
	}
	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	if dispatchAction {
		block["dispatch_action"] = true
	}
	return block
}

func externalMultiSelectInputBlock(blockID, actionID, label string, initialOptions []map[string]any, optional bool, minQueryLength int) map[string]any {
	element := map[string]any{
		"type":      "multi_external_select",
		"action_id": actionID,
	}
	if len(initialOptions) > 0 {
		element["initial_options"] = limitedSlackOptions(initialOptions)
	}
	if minQueryLength > 0 {
		element["min_query_length"] = minQueryLength
	}
	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	return block
}

func toSlackOption(text, value string) map[string]any {
	trimmedText := strings.TrimSpace(text)
	trimmedValue := strings.TrimSpace(value)
	if trimmedText == "" {
		trimmedText = trimmedValue
	}
	if trimmedValue == "" {
		trimmedValue = trimmedText
	}
	trimmedText = truncateRunes(trimmedText, slackOptionTextMaxRunes)
	trimmedValue = truncateRunes(trimmedValue, slackOptionValueMaxRunes)
	return map[string]any{
		"text": map[string]string{
			"type": "plain_text",
			"text": trimmedText,
		},
		"value": trimmedValue,
	}
}

func slackPriorityOptions() []map[string]any {
	priorities := []string{slackPriorityNoPriority, "Low", "Medium", "High", "Urgent"}
	options := make([]map[string]any, 0, len(priorities))
	for _, priority := range priorities {
		options = append(options, toSlackOption(priority, priority))
	}
	return options
}

func plainInputBlock(blockID, actionID, label, initial string, multiline bool, placeholder string, optional bool) map[string]any {
	element := map[string]any{
		"type":      "plain_text_input",
		"action_id": actionID,
	}
	if multiline {
		element["multiline"] = true
	}
	switch blockID {
	case modalBlockTitle:
		element["max_length"] = modalTitleMaxRunes
	case modalBlockDescription:
		element["max_length"] = modalDescriptionMaxRunes
	}
	if initial != "" {
		element["initial_value"] = initial
	}
	if placeholder != "" {
		element["placeholder"] = map[string]string{
			"type": "plain_text",
			"text": placeholder,
		}
	}
	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	return block
}

func parseCommandTitle(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "New task"
	}
	parts := strings.Fields(trimmed)
	if len(parts) >= 2 && strings.EqualFold(parts[0], "create") && strings.EqualFold(parts[1], "task") {
		parts = parts[2:]
	}
	if len(parts) == 0 {
		return "New task"
	}
	return truncateRunes(strings.TrimSpace(strings.Join(parts, " ")), modalTitleMaxRunes)
}

func parseViewSubmission(payload interactionPayload) (viewSubmissionData, error) {
	state := payload.View.State.Values
	readValue := func(blockID string) string {
		block := state[blockID]
		for _, action := range block {
			if strings.TrimSpace(action.Value) != "" {
				return strings.TrimSpace(action.Value)
			}
		}
		return ""
	}
	readSelectedOption := func(blockID string) string {
		block := state[blockID]
		for _, action := range block {
			if strings.TrimSpace(action.SelectedOption.Value) != "" {
				return strings.TrimSpace(action.SelectedOption.Value)
			}
		}
		return ""
	}

	metadata, err := parseSlackModalPrivateMetadata(payload.View.PrivateMetadata)
	if err != nil {
		return viewSubmissionData{}, err
	}
	source := metadata.Source

	selectedTeamID := readSelectedOption(modalBlockTeam)
	if selectedTeamID == "" {
		selectedTeamID = strings.TrimSpace(metadata.SelectedTeamID)
	}
	if selectedTeamID == "" {
		return viewSubmissionData{}, ErrSlackTeamSelectionRequired
	}
	teamID, err := uuid.Parse(selectedTeamID)
	if err != nil {
		return viewSubmissionData{}, errors.New("invalid selected team")
	}

	var statusID *uuid.UUID
	statusKind := slackStatusKindRequest
	statusState, statusBlockID, _ := modalDependentStateValue(
		state,
		modalBlockStatus,
		modalActionStatusSelect,
		teamID,
	)
	selectedStatusID := strings.TrimSpace(statusState.SelectedOption.Value)
	if selectedStatusID == slackRequestStatusValue || selectedStatusID == "" {
		statusKind = slackStatusKindRequest
	} else {
		parsedStatusID, parseErr := uuid.Parse(selectedStatusID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected status")
		}
		statusKind = slackStatusKindStory
		statusID = &parsedStatusID
	}

	var assigneeID *uuid.UUID
	assigneeState, assigneeBlockID, _ := modalDependentStateValue(
		state,
		modalBlockAssignee,
		modalActionAssigneeSelect,
		teamID,
	)
	selectedAssigneeID := strings.TrimSpace(assigneeState.SelectedOption.Value)
	if selectedAssigneeID != "" {
		parsedAssigneeID, parseErr := uuid.Parse(selectedAssigneeID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected assignee")
		}
		assigneeID = &parsedAssigneeID
	}

	labelsState, labelsBlockID, _ := modalDependentStateValue(
		state,
		modalBlockLabels,
		modalActionLabelsMultiSelect,
		teamID,
	)
	selectedLabelIDs := make([]uuid.UUID, 0)
	for _, selected := range labelsState.SelectedOptions {
		selectedLabelID := strings.TrimSpace(selected.Value)
		if selectedLabelID == "" {
			continue
		}
		parsedLabelID, parseErr := uuid.Parse(selectedLabelID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected label")
		}
		selectedLabelIDs = append(selectedLabelIDs, parsedLabelID)
	}

	var objectiveID *uuid.UUID
	objectiveState, objectiveBlockID, _ := modalDependentStateValue(
		state,
		modalBlockObjective,
		modalActionObjectiveSelect,
		teamID,
	)
	selectedObjectiveID := strings.TrimSpace(objectiveState.SelectedOption.Value)
	if selectedObjectiveID != "" {
		parsedObjectiveID, parseErr := uuid.Parse(selectedObjectiveID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected objective")
		}
		objectiveID = &parsedObjectiveID
	}

	return viewSubmissionData{
		Title:       readValue(modalBlockTitle),
		Description: readValue(modalBlockDescription),
		TeamID:      teamID,
		StatusKind:  statusKind,
		StatusID:    statusID,
		Priority:    readSelectedOption(modalBlockPriority),
		AssigneeID:  assigneeID,
		LabelIDs:    selectedLabelIDs,
		ObjectiveID: objectiveID,
		Source:      source,
		BlockIDs: modalDependentBlockIDs{
			Status:    statusBlockID,
			Assignee:  assigneeBlockID,
			Labels:    labelsBlockID,
			Objective: objectiveBlockID,
		},
	}, nil
}

func parseSourceFromPrivateMetadata(privateMetadata string) (requestSourceContext, error) {
	metadata, err := parseSlackModalPrivateMetadata(privateMetadata)
	if err != nil {
		return requestSourceContext{}, err
	}
	return metadata.Source, nil
}

func selectedTeamIDFromState(values interactionViewStateValues) string {
	block := values[modalBlockTeam]
	for _, action := range block {
		value := strings.TrimSpace(action.SelectedOption.Value)
		if value != "" {
			return value
		}
	}
	return ""
}

func (s *Service) resolveTeamIDForSuggestion(ctx context.Context, payload interactionPayload, workspaceID, actorID uuid.UUID, source requestSourceContext) (uuid.UUID, error) {
	if metadata, err := parseSlackModalPrivateMetadata(payload.View.PrivateMetadata); err == nil {
		if selectedFromMetadata := strings.TrimSpace(metadata.SelectedTeamID); selectedFromMetadata != "" {
			teamID, parseErr := uuid.Parse(selectedFromMetadata)
			if parseErr == nil && teamID != uuid.Nil {
				return teamID, nil
			}
		}
	}

	if selectedFromState := selectedTeamIDFromState(payload.View.State.Values); selectedFromState != "" {
		teamID, err := uuid.Parse(selectedFromState)
		if err == nil && teamID != uuid.Nil {
			return teamID, nil
		}
	}

	for _, block := range payload.View.Blocks {
		if strings.TrimSpace(block.BlockID) != modalBlockTeam {
			continue
		}
		value := strings.TrimSpace(block.Element.InitialOption.Value)
		if value == "" {
			continue
		}
		teamID, err := uuid.Parse(value)
		if err == nil && teamID != uuid.Nil {
			return teamID, nil
		}
	}

	teams, err := s.availableTeamsForSlackSource(ctx, workspaceID, actorID, source)
	if err != nil {
		return uuid.Nil, err
	}
	if len(teams) == 0 {
		return uuid.Nil, ErrSlackNoTeamsAvailable
	}
	return teams[0].ID, nil
}

func (s *Service) postSlackRequestAck(ctx context.Context, workspaceID, installGeneration uuid.UUID, idempotencyKey string, source requestSourceContext, botToken, workspaceSlug string, teamID, requestID, actorID uuid.UUID) string {
	requestURL := buildRequestURL(s.cfg.WebsiteURL, workspaceSlug, teamID.String(), requestID.String())
	text := "📥 Request created in FortyOne."
	if requestURL != "" {
		text = fmt.Sprintf("📥 Request created in FortyOne: <%s|Open request>", requestURL)
	}
	return s.postSlackCreationAckWithPayload(ctx, workspaceID, installGeneration, idempotencyKey, source, botToken, text, SlackProviderPayload{
		Authorization: &SlackDeliveryAuthorization{
			AllowedTeamIDs: []uuid.UUID{teamID},
			ActorUserID:    &actorID,
		},
		RequestThreadBinding: &SlackRequestThreadBinding{
			IntegrationRequestID:    requestID,
			ExternalSourceMessageID: strings.TrimSpace(source.SlackMessageTS),
			SourceURL:               permalinkFromContext(source),
		},
	})
}

func (s *Service) postSlackTaskAck(ctx context.Context, workspaceID, installGeneration uuid.UUID, idempotencyKey string, source requestSourceContext, botToken, workspaceSlug, teamCode, creatorName string, story stories.CoreSingleStory) string {
	storyCode := buildStoryCode(teamCode, story.SequenceID)
	taskURL := buildTaskURL(
		s.cfg.WebsiteURL,
		workspaceSlug,
		buildStoryReference(teamCode, story.SequenceID, story.ID.String()),
	)
	text := buildSlackStoryCreatedText(creatorName, storyCode, taskURL)
	authorization := &SlackDeliveryAuthorization{AllowedTeamIDs: []uuid.UUID{story.Team}}
	if story.Reporter != nil && *story.Reporter != uuid.Nil {
		authorization.ActorUserID = story.Reporter
	} else if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(source.SlackChannelID)), "D") {
		linkedUserID, linkErr := s.repo.FindLinkedUserIDBySlackUser(ctx, workspaceID, strings.TrimSpace(source.SlackTeamID), strings.TrimSpace(source.SlackUserID))
		if linkErr != nil {
			s.log.Warn(ctx, "failed resolving Slack receipt recipient", "error", linkErr, "workspace_id", workspaceID, "story_id", story.ID)
		} else if linkedUserID != nil && *linkedUserID != uuid.Nil {
			authorization.ActorUserID = linkedUserID
		}
	}
	receipt, err := s.buildSlackStoryCreationReceipt(ctx, source, taskURL, creatorName, story)
	if err != nil {
		if s.log != nil {
			s.log.Warn(ctx, "failed building rich Slack story receipt", "error", err, "workspace_id", workspaceID, "story_id", story.ID)
		}
		return s.postSlackCreationAckWithPayload(ctx, workspaceID, installGeneration, idempotencyKey, source, botToken, text, SlackProviderPayload{Authorization: authorization})
	}
	receipt.ProviderPayload.Authorization = authorization
	return s.postSlackCreationAckWithPayload(ctx, workspaceID, installGeneration, idempotencyKey, source, botToken, receipt.Text, receipt.ProviderPayload)
}

func buildSlackStoryCreatedText(creatorName, storyCode, taskURL string) string {
	creatorLabel := slackMrkdwnTextEscaper.Replace(strings.TrimSpace(creatorName))
	if creatorLabel == "" {
		creatorLabel = "A team member"
	}
	storyLabel := strings.TrimSpace(storyCode)
	if storyLabel == "" {
		storyLabel = "a story"
	}
	if taskURL = strings.TrimSpace(taskURL); taskURL != "" {
		storyLabel = fmt.Sprintf("<%s|%s>", taskURL, storyLabel)
	}
	return fmt.Sprintf("%s created %s", creatorLabel, storyLabel)
}

func (s *Service) buildSlackStoryCreationReceipt(
	ctx context.Context,
	source requestSourceContext,
	storyURL, creatorName string,
	story stories.CoreSingleStory,
) (SlackStoryCreationReceipt, error) {
	statusName := ""
	if story.Status != nil {
		statuses, err := s.repo.ListTeamStatuses(ctx, story.Team)
		if err != nil {
			return SlackStoryCreationReceipt{}, err
		}
		for _, status := range statuses {
			if status.ID == *story.Status {
				statusName = status.Name
				break
			}
		}
	}
	assigneeName := ""
	assigneeSlackUserID := ""
	if story.Assignee != nil {
		member, err := s.repo.FindTeamMemberByID(ctx, story.Team, *story.Assignee)
		if err != nil && !slackrepository.IsNotFound(err) {
			return SlackStoryCreationReceipt{}, err
		}
		if err == nil {
			assigneeName = slackMemberDisplayName(member)
		}
		if story.Reporter != nil && *story.Reporter == *story.Assignee {
			assigneeSlackUserID = strings.TrimSpace(source.SlackUserID)
		}
	}
	description := ""
	if story.Description != nil {
		description = strings.TrimSpace(*story.Description)
	}
	return BuildSlackStoryCreationReceipt(creatorName, SlackStoryWorkObjectInput{
		AccessGranted:       true,
		ExternalID:          story.ID.String(),
		StoryURL:            storyURL,
		Title:               story.Title,
		Description:         description,
		Status:              statusName,
		Priority:            story.Priority,
		AssigneeSlackUserID: assigneeSlackUserID,
		AssigneeName:        assigneeName,
		CreatorSlackUserID:  strings.TrimSpace(source.SlackUserID),
		CreatorName:         creatorName,
		DueDate:             story.EndDate,
		CreatedAt:           story.CreatedAt,
		UpdatedAt:           story.UpdatedAt,
	})
}

func slackMemberDisplayName(member slackrepository.TeamMemberRecord) string {
	for _, value := range []string{member.FullName, member.Username, member.Email} {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func (s *Service) postSlackCreationAck(ctx context.Context, workspaceID, installGeneration uuid.UUID, idempotencyKey string, source requestSourceContext, botToken, text string) string {
	return s.postSlackCreationAckWithPayload(ctx, workspaceID, installGeneration, idempotencyKey, source, botToken, text, SlackProviderPayload{})
}

func (s *Service) postSlackCreationAckWithPayload(ctx context.Context, workspaceID, installGeneration uuid.UUID, idempotencyKey string, source requestSourceContext, botToken, text string, providerPayload SlackProviderPayload) string {
	externalWorkspaceID := strings.TrimSpace(source.SlackTeamID)
	channelID := strings.TrimSpace(source.SlackChannelID)
	threadTS := strings.TrimSpace(source.SlackThreadTS)
	if channelID == "" {
		channelID = strings.TrimSpace(source.SlackUserID)
		threadTS = ""
	}
	if workspaceID == uuid.Nil || installGeneration == uuid.Nil || externalWorkspaceID == "" || channelID == "" || strings.TrimSpace(idempotencyKey) == "" {
		return ""
	}

	var encodedProviderPayload []byte
	if !slackProviderPayloadIsEmpty(providerPayload) {
		var err error
		encodedProviderPayload, err = EncodeSlackProviderPayload(providerPayload)
		if err != nil {
			s.log.Warn(ctx, "failed encoding Slack creation acknowledgement payload", "error", err, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
			return ""
		}
	}
	deliveryID := uuid.Nil
	var actorUserID *uuid.UUID
	if providerPayload.Authorization != nil && providerPayload.Authorization.ActorUserID != nil {
		actorID := *providerPayload.Authorization.ActorUserID
		actorUserID = &actorID
	}
	if s.outbound != nil {
		deliveryExpiresAt := s.clock.Now().UTC().Add(time.Hour)
		delivery, shouldSend, err := s.outbound.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
			Provider:                slackProviderMessaging,
			WorkspaceID:             workspaceID,
			UserID:                  actorUserID,
			InstallGeneration:       &installGeneration,
			ExternalWorkspaceID:     externalWorkspaceID,
			ExternalRecipientUserID: strings.TrimSpace(source.SlackUserID),
			IdempotencyKey:          idempotencyKey,
			ExternalChannelID:       channelID,
			ExternalThreadID:        threadTS,
			Content:                 text,
			ProviderPayload:         encodedProviderPayload,
			Purpose:                 "creation_confirmation",
			ExpiresAt:               &deliveryExpiresAt,
		})
		if err != nil {
			s.log.Warn(ctx, "failed claiming Slack creation acknowledgement", "error", err, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
			return ""
		}
		if !shouldSend {
			if delivery.ExternalMessageID != nil {
				return strings.TrimSpace(*delivery.ExternalMessageID)
			}
			return ""
		}
		deliveryID = delivery.ID
		if err := persistSlackOutboundContent(ctx, s.outbound, delivery.ID, text, providerPayload); err != nil {
			if failErr := failOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, truncateError(err)); failErr != nil {
				s.log.Error(ctx, "failed releasing Slack creation acknowledgement after persistence error", "error", failErr, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
			}
			s.log.Warn(ctx, "failed persisting Slack creation acknowledgement", "error", err, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
			return ""
		}
	}
	if err := s.requireCurrentSlackInstallation(ctx, workspaceID, externalWorkspaceID, installGeneration); err != nil {
		if s.outbound != nil && deliveryID != uuid.Nil {
			if cancelErr := cancelOutboundDeliveryDetached(ctx, s.outbound, deliveryID, "Slack installation changed before creation acknowledgement"); cancelErr != nil {
				s.log.Error(ctx, "failed cancelling stale Slack creation acknowledgement", "error", cancelErr, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
			}
		}
		if !errors.Is(err, errSlackInstallationChanged) {
			s.log.Error(ctx, "failed revalidating Slack creation acknowledgement installation", "error", err, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
		}
		return ""
	}
	if current, err := s.slackDeliveryAuthorizationCurrent(ctx, workspaceID, externalWorkspaceID, channelID, source.SlackUserID, providerPayload); err != nil {
		if s.outbound != nil && deliveryID != uuid.Nil {
			_ = failOutboundDeliveryDetached(ctx, s.outbound, deliveryID, truncateError(err))
		}
		s.log.Error(ctx, "failed revalidating Slack creation acknowledgement audience", "error", err, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
		return ""
	} else if !current {
		if s.outbound != nil && deliveryID != uuid.Nil {
			_ = cancelOutboundDeliveryDetached(ctx, s.outbound, deliveryID, "Slack creation acknowledgement actor or channel audience changed")
		}
		return ""
	}

	externalMessageID, err := (&slackAPISender{client: s.slackClient()}).Send(ctx, botToken, SlackOutboundMessage{
		ChannelID:       channelID,
		UserID:          strings.TrimSpace(source.SlackUserID),
		ThreadTS:        threadTS,
		Text:            truncateSlackText(text),
		ClientMessageID: deterministicSlackMessageID(idempotencyKey),
		ProviderPayload: providerPayload,
	})
	if err != nil {
		if s.outbound != nil && deliveryID != uuid.Nil {
			if failErr := failOutboundDeliveryDetached(ctx, s.outbound, deliveryID, truncateError(err)); failErr != nil {
				s.log.Error(ctx, "failed releasing Slack creation acknowledgement after provider error", "error", failErr, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
			}
		}
		s.log.Error(ctx, "failed posting Slack creation acknowledgement", "error", err, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
		return ""
	}
	if err := bindSlackRequestThreadContinuation(ctx, s.requests, workspaceID, installGeneration, externalWorkspaceID, channelID, threadTS, externalMessageID, providerPayload); err != nil {
		if s.outbound != nil && deliveryID != uuid.Nil {
			if failErr := failOutboundDeliveryDetached(ctx, s.outbound, deliveryID, truncateError(err)); failErr != nil {
				s.log.Error(ctx, "failed releasing Slack request acknowledgement after thread binding error", "error", failErr, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
			}
		}
		s.log.Error(ctx, "failed binding Slack request acknowledgement thread", "error", err, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
		return ""
	}
	if s.outbound != nil && deliveryID != uuid.Nil {
		if err := s.outbound.CompleteOutboundDelivery(ctx, deliveryID, externalMessageID); err != nil {
			s.log.Error(ctx, "failed completing Slack creation acknowledgement", "error", err, "workspace_id", workspaceID, "idempotency_key", idempotencyKey)
		}
	}
	return externalMessageID
}

func (s *Service) requireCurrentSlackInstallation(ctx context.Context, workspaceID uuid.UUID, slackTeamID string, generation uuid.UUID) error {
	current, err := s.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return fmt.Errorf("%w: Slack team is no longer connected", errSlackInstallationChanged)
		}
		return err
	}
	if !current.IsActive || current.WorkspaceID != workspaceID || current.InstallGeneration != generation {
		return fmt.Errorf("%w: active generation no longer matches", errSlackInstallationChanged)
	}
	return nil
}

func (s *Service) postMessage(ctx context.Context, botToken, channelID, threadTS, text string) error {
	if strings.TrimSpace(botToken) == "" {
		return errors.New("missing slack bot token")
	}
	payload := map[string]any{
		"channel": channelID,
		"text":    text,
	}
	if strings.TrimSpace(threadTS) != "" {
		payload["thread_ts"] = threadTS
	}
	return s.callSlackAPI(ctx, botToken, "https://slack.com/api/chat.postMessage", payload, nil)
}

func (s *Service) postEphemeralMessage(ctx context.Context, botToken, channelID, userID, text string) error {
	if strings.TrimSpace(botToken) == "" {
		return errors.New("missing slack bot token")
	}
	channelID = strings.TrimSpace(channelID)
	userID = strings.TrimSpace(userID)
	if channelID == "" || userID == "" {
		return nil
	}

	payload := map[string]any{
		"channel": channelID,
		"user":    userID,
		"text":    strings.TrimSpace(text),
	}
	return s.callSlackAPI(ctx, botToken, "https://slack.com/api/chat.postEphemeral", payload, nil)
}

func (s *Service) postCommandResponse(ctx context.Context, responseURL, text string) error {
	responseURL = strings.TrimSpace(responseURL)
	if responseURL == "" {
		return nil
	}
	parsedResponseURL, err := url.Parse(responseURL)
	if err != nil || parsedResponseURL.Scheme != "https" || !strings.EqualFold(parsedResponseURL.Hostname(), "hooks.slack.com") {
		return errors.New("invalid Slack response URL")
	}
	payload := CommandResponse{
		ResponseType: "ephemeral",
		Text:         text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, responseURL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
		}
		return fmt.Errorf("slack command response failed with status %d: %s", resp.StatusCode, string(respBytes))
	}
	return nil
}

func (s *Service) postInteractionFailure(ctx context.Context, payload interactionPayload, text string) error {
	var responseErr error
	if responseURL := strings.TrimSpace(payload.ResponseURL); responseURL != "" {
		if err := s.postCommandResponse(ctx, responseURL, text); err == nil {
			return nil
		} else {
			responseErr = fmt.Errorf("post via response URL: %w", err)
		}
	}

	teamID := strings.TrimSpace(payload.Team.ID)
	channelID := strings.TrimSpace(payload.Channel.ID)
	userID := strings.TrimSpace(payload.User.ID)
	if teamID == "" || channelID == "" || userID == "" {
		if responseErr != nil {
			return responseErr
		}
		return errors.New("Slack interaction has no private feedback destination")
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, teamID)
	if err != nil {
		return errors.Join(responseErr, fmt.Errorf("load Slack installation for private feedback: %w", err))
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return errors.Join(responseErr, fmt.Errorf("load Slack credential for private feedback: %w", err))
	}
	if err := s.postEphemeralMessage(ctx, botToken, channelID, userID, text); err != nil {
		return errors.Join(responseErr, fmt.Errorf("post private Slack feedback: %w", err))
	}
	return nil
}

func interactionFailureMessage(err error) string {
	switch {
	case errors.Is(err, ErrSlackNoWorkspaceLinked):
		return "Slack is no longer connected to this FortyOne workspace."
	case errors.Is(err, ErrSlackUserNotLinked):
		return "Your Slack account is no longer linked to FortyOne. Connect it again and reopen the form."
	case errors.Is(err, ErrSlackTeamNotAvailable):
		return "The selected team is no longer available to you. Reopen the form and choose another team."
	case errors.Is(err, ErrSlackTeamSelectionRequired):
		return "FortyOne could not read the selected team. Reopen the form and try again."
	case errors.Is(err, ErrSlackInteractionActorMismatch):
		return "This FortyOne form is no longer valid for your Slack account. Reopen it and try again."
	default:
		return "FortyOne could not update this form. Please reopen it and try again."
	}
}

func (s *Service) fetchChannels(ctx context.Context, botToken string) ([]slackrepository.SlackChannelPayload, error) {
	cursor := ""
	channels := make([]slackrepository.SlackChannelPayload, 0)

	for {
		endpoint := "https://slack.com/api/conversations.list?limit=200&types=public_channel,private_channel"
		if cursor != "" {
			endpoint += "&cursor=" + url.QueryEscape(cursor)
		}
		var response struct {
			OK       bool   `json:"ok"`
			Error    string `json:"error"`
			Channels []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				IsPrivate  bool   `json:"is_private"`
				IsArchived bool   `json:"is_archived"`
				IsMember   bool   `json:"is_member"`
			} `json:"channels"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := s.callSlackAPI(ctx, botToken, endpoint, nil, &response); err != nil {
			return nil, err
		}
		for _, channel := range response.Channels {
			if strings.TrimSpace(channel.ID) == "" {
				continue
			}
			name := strings.TrimSpace(channel.Name)
			if name == "" {
				name = channel.ID
			}
			channels = append(channels, slackrepository.SlackChannelPayload{
				SlackChannelID: channel.ID,
				Name:           name,
				IsPrivate:      channel.IsPrivate,
				IsArchived:     channel.IsArchived,
				IsMember:       channel.IsMember,
			})
		}
		cursor = strings.TrimSpace(response.ResponseMetadata.NextCursor)
		if cursor == "" {
			break
		}
	}

	return channels, nil
}

func (s *Service) autoLinkWorkspaceMembers(ctx context.Context, slackWorkspace slackrepository.SlackWorkspaceRecord) error {
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return err
	}
	slackUsers, err := s.fetchWorkspaceUsers(ctx, botToken)
	if err != nil {
		return err
	}
	if len(slackUsers) == 0 {
		return nil
	}

	workspaceMembers, err := s.repo.ListWorkspaceMembersForSlackLinking(ctx, slackWorkspace.WorkspaceID)
	if err != nil {
		return err
	}
	if len(workspaceMembers) == 0 {
		return nil
	}

	memberByEmail := make(map[string]uuid.UUID, len(workspaceMembers))
	for _, member := range workspaceMembers {
		normalizedEmail := normalizeEmail(member.Email)
		if normalizedEmail == "" {
			continue
		}
		memberByEmail[normalizedEmail] = member.UserID
	}
	if len(memberByEmail) == 0 {
		return nil
	}

	links := make([]slackrepository.SlackUserLinkUpsert, 0, len(slackUsers))
	for _, slackUser := range slackUsers {
		normalizedEmail := normalizeEmail(slackUser.Email)
		if normalizedEmail == "" {
			continue
		}
		userID, ok := memberByEmail[normalizedEmail]
		if !ok || userID == uuid.Nil {
			continue
		}
		links = append(links, slackrepository.SlackUserLinkUpsert{
			SlackUserID: slackUser.ID,
			UserID:      userID,
			LinkedVia:   "email_match",
		})
	}
	if len(links) == 0 {
		return nil
	}

	return s.repo.UpsertSlackUserLinks(ctx, slackWorkspace.WorkspaceID, slackWorkspace.ID, slackWorkspace.SlackTeamID, links)
}

func (s *Service) fetchWorkspaceUsers(ctx context.Context, botToken string) ([]slackWorkspaceUser, error) {
	cursor := ""
	users := make([]slackWorkspaceUser, 0)

	for {
		endpoint := "https://slack.com/api/users.list?limit=200"
		if cursor != "" {
			endpoint += "&cursor=" + url.QueryEscape(cursor)
		}
		var response struct {
			OK      bool   `json:"ok"`
			Error   string `json:"error"`
			Members []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				RealName string `json:"real_name"`
				Deleted  bool   `json:"deleted"`
				IsBot    bool   `json:"is_bot"`
				Profile  struct {
					Email string `json:"email"`
				} `json:"profile"`
			} `json:"members"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := s.callSlackAPI(ctx, botToken, endpoint, nil, &response); err != nil {
			return nil, err
		}

		for _, member := range response.Members {
			if member.Deleted || member.IsBot {
				continue
			}
			memberID := strings.TrimSpace(member.ID)
			if memberID == "" {
				continue
			}
			users = append(users, slackWorkspaceUser{
				ID:       memberID,
				Username: strings.TrimSpace(member.Name),
				FullName: strings.TrimSpace(member.RealName),
				Email:    strings.TrimSpace(member.Profile.Email),
			})
		}

		cursor = strings.TrimSpace(response.ResponseMetadata.NextCursor)
		if cursor == "" {
			break
		}
	}

	return users, nil
}

func (s *Service) resolveLinkedSlackUser(ctx context.Context, workspaceID uuid.UUID, source requestSourceContext) (uuid.UUID, string, error) {
	slackTeamID := strings.TrimSpace(source.SlackTeamID)
	slackUserID := strings.TrimSpace(source.SlackUserID)
	if slackTeamID != "" && slackUserID != "" {
		mappedUserID, err := s.repo.FindLinkedUserIDBySlackUser(ctx, workspaceID, slackTeamID, slackUserID)
		if err != nil {
			return uuid.Nil, "", err
		}
		if mappedUserID != nil && *mappedUserID != uuid.Nil {
			return *mappedUserID, "", nil
		}
	}

	connectURL, err := s.buildSlackUserLinkURL(ctx, workspaceID, slackTeamID, slackUserID)
	if err != nil {
		return uuid.Nil, "", err
	}
	return uuid.Nil, connectURL, nil
}

func interactionSourceForPayload(payload interactionPayload, source requestSourceContext) (requestSourceContext, error) {
	slackTeamID := strings.TrimSpace(payload.Team.ID)
	slackUserID := strings.TrimSpace(payload.User.ID)
	if slackTeamID == "" || slackUserID == "" {
		return requestSourceContext{}, ErrSlackInteractionActorMismatch
	}
	if sourceTeamID := strings.TrimSpace(source.SlackTeamID); sourceTeamID != "" && sourceTeamID != slackTeamID {
		return requestSourceContext{}, ErrSlackInteractionActorMismatch
	}
	if sourceUserID := strings.TrimSpace(source.SlackUserID); sourceUserID != "" && sourceUserID != slackUserID {
		return requestSourceContext{}, ErrSlackInteractionActorMismatch
	}

	source.SlackTeamID = slackTeamID
	source.SlackUserID = slackUserID
	if username := strings.TrimSpace(payload.User.Username); username != "" {
		source.SlackUsername = username
	} else if username := strings.TrimSpace(payload.User.Name); username != "" {
		source.SlackUsername = username
	}
	return source, nil
}

func (s *Service) findLinkedInteractionActor(ctx context.Context, workspaceID uuid.UUID, source requestSourceContext) (uuid.UUID, error) {
	linkedUserID, err := s.repo.FindLinkedUserIDBySlackUser(
		ctx,
		workspaceID,
		strings.TrimSpace(source.SlackTeamID),
		strings.TrimSpace(source.SlackUserID),
	)
	if err != nil {
		return uuid.Nil, err
	}
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		return uuid.Nil, ErrSlackUserNotLinked
	}
	return *linkedUserID, nil
}

func (s *Service) findTeamForActor(ctx context.Context, workspaceID, actorID, teamID uuid.UUID) (slackrepository.TeamRecord, error) {
	if workspaceID == uuid.Nil || actorID == uuid.Nil || teamID == uuid.Nil {
		return slackrepository.TeamRecord{}, ErrSlackTeamNotAvailable
	}
	teams, err := s.repo.ListWorkspaceTeamsForUser(ctx, workspaceID, actorID)
	if err != nil {
		return slackrepository.TeamRecord{}, err
	}
	for _, team := range teams {
		if team.ID == teamID {
			return team, nil
		}
	}
	return slackrepository.TeamRecord{}, ErrSlackTeamNotAvailable
}

func (s *Service) buildSlackUserLinkURL(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (string, error) {
	if workspaceID == uuid.Nil {
		return "", errors.New("workspace id is required")
	}
	if s.nonces == nil {
		return "", errors.New("slack account-link nonce store is not configured")
	}
	slackTeamID = strings.TrimSpace(slackTeamID)
	slackUserID = strings.TrimSpace(slackUserID)
	if slackTeamID == "" || slackUserID == "" {
		return "", nil
	}

	workspace, err := s.repo.FindWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	token, digest, err := s.newOpaqueNonce()
	if err != nil {
		return "", err
	}
	if err := s.nonces.CreateNonce(ctx, messagingrepository.NonceInput{
		Provider:            slackProviderMessaging,
		Purpose:             slackNoncePurposeAccount,
		NonceHash:           digest,
		WorkspaceID:         workspaceID,
		ExternalWorkspaceID: slackTeamID,
		ExternalUserID:      slackUserID,
		ExpiresAt:           s.clock.Now().UTC().Add(slackAccountLinkNonceTTL),
	}); err != nil {
		return "", fmt.Errorf("create Slack account-link token: %w", err)
	}

	baseLink := s.buildAccountIntegrationURL(workspace.Slug)
	linkURL, err := url.Parse(baseLink)
	if err != nil {
		return "", nil
	}
	query := linkURL.Query()
	query.Set("slack_link_token", token)
	linkURL.RawQuery = query.Encode()
	return linkURL.String(), nil
}

func (s *Service) newOpaqueNonce() (string, []byte, error) {
	if s.random == nil {
		return "", nil, errors.New("secure random source is not configured")
	}
	nonce := make([]byte, slackOpaqueNonceSize)
	if _, err := io.ReadFull(s.random, nonce); err != nil {
		return "", nil, fmt.Errorf("generate Slack nonce: %w", err)
	}
	digest := sha256.Sum256(nonce)
	return base64.RawURLEncoding.EncodeToString(nonce), digest[:], nil
}

func (s *Service) consumeNonce(ctx context.Context, purpose, rawToken string, workspaceID, userID *uuid.UUID) (messagingrepository.NonceRecord, error) {
	if s.nonces == nil {
		return messagingrepository.NonceRecord{}, errors.New("slack nonce store is not configured")
	}
	nonce, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(rawToken))
	if err != nil || len(nonce) != slackOpaqueNonceSize {
		return messagingrepository.NonceRecord{}, errors.New("invalid Slack nonce")
	}
	digest := sha256.Sum256(nonce)
	record, err := s.nonces.ConsumeNonce(ctx, messagingrepository.NonceConsumeInput{
		Provider:    slackProviderMessaging,
		Purpose:     purpose,
		NonceHash:   digest[:],
		WorkspaceID: workspaceID,
		UserID:      userID,
		Now:         s.clock.Now().UTC(),
	})
	if err != nil {
		return messagingrepository.NonceRecord{}, err
	}
	return record, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Service) exchangeOAuthCode(ctx context.Context, code string) (oauthAccessResponse, error) {
	var response oauthAccessResponse
	if err := s.slackClient().oauthV2Access(ctx, s.cfg.ClientID, s.cfg.ClientSecret, s.cfg.RedirectURL, code, &response); err != nil {
		return oauthAccessResponse{}, err
	}
	if strings.TrimSpace(response.AccessToken) == "" {
		return oauthAccessResponse{}, errors.New("slack oauth returned empty access token")
	}
	if strings.TrimSpace(response.Team.ID) == "" {
		return oauthAccessResponse{}, errors.New("slack oauth returned empty team id")
	}
	if strings.TrimSpace(response.Team.Domain) == "" {
		response.Team.Domain = response.Team.ID
	}
	if strings.TrimSpace(response.Team.Name) == "" {
		response.Team.Name = response.Team.Domain
	}
	return response, nil
}

func (s *Service) callSlackAPI(ctx context.Context, botToken, endpoint string, payload any, out any) error {
	method, err := slackAPIMethod(endpoint)
	if err != nil {
		return err
	}
	return s.slackClient().callJSON(ctx, botToken, method, payload, out)
}

func (s *Service) slackClient() *slackWebClient {
	if s.webClient == nil {
		s.webClient = newSlackWebClient(s.client)
		return s.webClient
	}
	if s.client != nil && s.webClient.client != s.client {
		baseURL := s.webClient.baseURL
		s.webClient = newSlackWebClient(s.client)
		if strings.TrimSpace(baseURL) != "" && baseURL != defaultSlackAPIBaseURL {
			s.webClient.baseURL = baseURL
		}
	}
	return s.webClient
}

func slackAPIMethod(endpoint string) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return "", errors.New("slack api endpoint is required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse slack api endpoint: %w", err)
	}
	method := strings.TrimPrefix(parsed.Path, "/api/")
	if method == parsed.Path {
		method = strings.TrimPrefix(parsed.Path, "/")
	}
	if method == "" {
		return "", errors.New("slack api method is required")
	}
	if parsed.RawQuery != "" {
		method += "?" + parsed.RawQuery
	}
	return method, nil
}

func (s *Service) buildWorkspaceIntegrationURL(workspaceSlug string) string {
	link := buildWorkspaceURL(s.cfg.WebsiteURL, workspaceSlug, "settings", "workspace", "integrations", "slack")
	if link == "" {
		return "/"
	}
	return link
}

func (s *Service) buildAccountIntegrationURL(workspaceSlug string) string {
	link := buildWorkspaceURL(s.cfg.WebsiteURL, workspaceSlug, "settings", "integrations", "slack")
	if link == "" {
		return "/"
	}
	return link
}

func (s *Service) canUseOAuth() bool {
	return strings.TrimSpace(s.cfg.ClientID) != "" &&
		strings.TrimSpace(s.cfg.ClientSecret) != "" &&
		strings.TrimSpace(s.cfg.RedirectURL) != "" &&
		strings.TrimSpace(s.cfg.SecretKey) != "" &&
		s.credentials != nil &&
		s.nonces != nil
}

func interactionValidationErrors(errorsByBlock map[string]string) (InteractionResponse, error) {
	body, err := json.Marshal(map[string]any{
		"response_action": "errors",
		"errors":          errorsByBlock,
	})
	if err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
}

func interactionClearResponse() (InteractionResponse, error) {
	body, err := json.Marshal(map[string]string{"response_action": "clear"})
	if err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
}

func interactionOptionsResponse(options []map[string]any) (InteractionResponse, error) {
	if options == nil {
		options = make([]map[string]any, 0)
	}
	options = limitedSlackOptions(options)
	body, err := json.Marshal(map[string]any{"options": options})
	if err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
}

func messageToTitle(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "New task"
	}
	firstLine := strings.Split(trimmed, "\n")[0]
	firstLine = strings.TrimSpace(firstLine)
	if len([]rune(firstLine)) <= 120 {
		return firstLine
	}
	return strings.TrimSpace(truncateRunes(firstLine, 120))
}

func buildConnectSlackAccountMessage(connectURL string) string {
	if strings.TrimSpace(connectURL) == "" {
		return "Connect your FortyOne account before creating tasks from Slack."
	}
	return fmt.Sprintf("Connect your FortyOne account to continue: <%s|Connect FortyOne account>", connectURL)
}

func buildPrefilledDescription(source requestSourceContext) string {
	message := strings.TrimSpace(source.SlackText)
	if message == "" {
		return ""
	}
	identity := strings.TrimSpace(source.SlackUsername)
	if identity == "" {
		identity = strings.TrimSpace(source.SlackUserID)
	}
	if identity == "" {
		return truncateRunes("> "+message, modalDescriptionMaxRunes)
	}
	if strings.TrimSpace(source.SlackUserID) == "" {
		return truncateRunes(fmt.Sprintf("@%s said:\n> %s", identity, message), modalDescriptionMaxRunes)
	}
	return truncateRunes(fmt.Sprintf("@[%s](%s) said:\n> %s", identity, strings.TrimSpace(source.SlackUserID), message), modalDescriptionMaxRunes)
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func buildSlackStoryLinkTitle(source requestSourceContext) string {
	name := strings.TrimSpace(source.SlackUsername)
	if name == "" {
		name = strings.TrimSpace(source.SlackUserID)
	}
	channel := strings.TrimSpace(source.SlackChannel)
	if channel == "" {
		channel = strings.TrimSpace(source.SlackChannelID)
	}
	switch {
	case name != "" && channel != "":
		return fmt.Sprintf("Slack message from %s in #%s", name, channel)
	case name != "":
		return fmt.Sprintf("Slack message from %s", name)
	case channel != "":
		return fmt.Sprintf("Slack message in #%s", channel)
	default:
		return "Slack message"
	}
}

func buildSourceExternalID(source requestSourceContext) string {
	parts := []string{}
	if source.SlackTeamID != "" {
		parts = append(parts, source.SlackTeamID)
	}
	if source.SlackChannelID != "" {
		parts = append(parts, source.SlackChannelID)
	}
	if source.SlackMessageTS != "" {
		parts = append(parts, source.SlackMessageTS)
	}
	if source.SlackThreadTS != "" {
		parts = append(parts, source.SlackThreadTS)
	}
	return strings.Join(parts, ":")
}

func permalinkFromContext(source requestSourceContext) string {
	if source.SlackTeamDomain == "" || source.SlackChannelID == "" || source.SlackMessageTS == "" {
		return ""
	}
	messageTS := strings.ReplaceAll(source.SlackMessageTS, ".", "")
	return fmt.Sprintf("https://%s.slack.com/archives/%s/p%s", source.SlackTeamDomain, source.SlackChannelID, messageTS)
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func interactionErrorMessage(err error) string {
	if err == nil {
		return "Unable to create story. Please try again."
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "Unable to create story. Please try again."
	}
	const maxLength = 180
	if len(message) > maxLength {
		return strings.TrimSpace(message[:maxLength-3]) + "..."
	}
	return message
}

func findStatusByID(statuses []slackrepository.StatusRecord, statusID uuid.UUID) (slackrepository.StatusRecord, bool) {
	for _, status := range statuses {
		if status.ID == statusID {
			return status, true
		}
	}
	return slackrepository.StatusRecord{}, false
}

func normalizeSlackPriority(value string) string {
	switch strings.TrimSpace(value) {
	case "Low", "Medium", "High", "Urgent", slackPriorityNoPriority:
		return strings.TrimSpace(value)
	default:
		return slackPriorityNoPriority
	}
}

func selectTeam(teams []slackrepository.TeamRecord, preferredTeamID uuid.UUID) slackrepository.TeamRecord {
	if preferredTeamID != uuid.Nil {
		for _, team := range teams {
			if team.ID == preferredTeamID {
				return team
			}
		}
	}
	return teams[0]
}

func teamMemberDisplayName(member slackrepository.TeamMemberRecord) string {
	if fullName := strings.TrimSpace(member.FullName); fullName != "" {
		if email := strings.TrimSpace(member.Email); email != "" {
			return fmt.Sprintf("%s (%s)", fullName, email)
		}
		return fullName
	}
	if username := strings.TrimSpace(member.Username); username != "" {
		if email := strings.TrimSpace(member.Email); email != "" {
			return fmt.Sprintf("%s (%s)", username, email)
		}
		return username
	}
	if email := strings.TrimSpace(member.Email); email != "" {
		return email
	}
	return member.UserID.String()
}

func storyCreatorDisplayName(member slackrepository.TeamMemberRecord) string {
	if fullName := strings.TrimSpace(member.FullName); fullName != "" {
		return fullName
	}
	if username := strings.TrimSpace(member.Username); username != "" {
		return username
	}
	if email := strings.TrimSpace(member.Email); email != "" {
		return email
	}
	return "A team member"
}

func teamMemberExists(members []slackrepository.TeamMemberRecord, userID uuid.UUID) bool {
	for _, member := range members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}

func filterValidLabelIDs(labels []slackrepository.LabelRecord, selected []uuid.UUID) []uuid.UUID {
	if len(selected) == 0 {
		return nil
	}
	valid := make(map[uuid.UUID]struct{}, len(labels))
	for _, label := range labels {
		valid[label.ID] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(selected))
	seen := make(map[uuid.UUID]struct{}, len(selected))
	for _, labelID := range selected {
		if _, alreadySeen := seen[labelID]; alreadySeen {
			continue
		}
		seen[labelID] = struct{}{}
		if _, ok := valid[labelID]; ok {
			result = append(result, labelID)
		}
	}
	return result
}

func uuidStrings(values []uuid.UUID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}

type requestLogDetails struct {
	SlackTeamID    string
	SlackUserID    string
	SlackChannelID string
	Command        string
	TriggerID      string
}

func parseRequestLogDetails(requestType string, rawBody []byte) requestLogDetails {
	switch strings.TrimSpace(requestType) {
	case "commands":
		return parseCommandLogDetails(rawBody)
	case "interactivity":
		return parseInteractivityLogDetails(rawBody)
	case "events":
		return parseEventsLogDetails(rawBody)
	default:
		return requestLogDetails{}
	}
}

func parseCommandLogDetails(rawBody []byte) requestLogDetails {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return requestLogDetails{}
	}
	return requestLogDetails{
		SlackTeamID:    strings.TrimSpace(values.Get("team_id")),
		SlackUserID:    strings.TrimSpace(values.Get("user_id")),
		SlackChannelID: strings.TrimSpace(values.Get("channel_id")),
		Command:        strings.TrimSpace(values.Get("command")),
		TriggerID:      strings.TrimSpace(values.Get("trigger_id")),
	}
}

func parseInteractivityLogDetails(rawBody []byte) requestLogDetails {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return requestLogDetails{}
	}
	payloadText := strings.TrimSpace(values.Get("payload"))
	if payloadText == "" {
		return requestLogDetails{}
	}
	var payload interactionPayload
	if err := json.Unmarshal([]byte(payloadText), &payload); err != nil {
		return requestLogDetails{}
	}

	teamID := strings.TrimSpace(payload.Team.ID)
	if teamID == "" {
		if metadata, metadataErr := parseSlackModalPrivateMetadata(payload.View.PrivateMetadata); metadataErr == nil {
			teamID = strings.TrimSpace(metadata.Source.SlackTeamID)
		}
	}
	actionID := suggestionActionID(payload)
	if actionID == "" && len(payload.Actions) > 0 {
		actionID = strings.TrimSpace(payload.Actions[0].ActionID)
	}
	userID := strings.TrimSpace(payload.User.ID)
	if userID == "" {
		userID = strings.TrimSpace(payload.Message.User)
	}

	return requestLogDetails{
		SlackTeamID:    teamID,
		SlackUserID:    userID,
		SlackChannelID: strings.TrimSpace(payload.Channel.ID),
		Command:        actionID,
		TriggerID:      strings.TrimSpace(payload.TriggerID),
	}
}

func parseEventsLogDetails(rawBody []byte) requestLogDetails {
	var payload struct {
		TeamID string `json:"team_id"`
		Event  struct {
			Channel string `json:"channel"`
			User    string `json:"user"`
		} `json:"event"`
	}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return requestLogDetails{}
	}

	return requestLogDetails{
		SlackTeamID:    strings.TrimSpace(payload.TeamID),
		SlackUserID:    strings.TrimSpace(payload.Event.User),
		SlackChannelID: strings.TrimSpace(payload.Event.Channel),
	}
}

func (s *Service) resolveWorkspaceIDFromLog(ctx context.Context, slackTeamID string) *uuid.UUID {
	if strings.TrimSpace(slackTeamID) == "" {
		return nil
	}
	workspace, err := s.repo.GetWorkspaceBySlackTeamID(ctx, slackTeamID)
	if err != nil {
		return nil
	}
	return &workspace.ID
}

func truncateForLog(value string, maxLength int) string {
	trimmed := strings.TrimSpace(value)
	if maxLength <= 0 || len(trimmed) <= maxLength {
		return trimmed
	}
	return trimmed[:maxLength]
}

func buildWorkspaceURL(websiteURL, workspaceSlug string, routeSegments ...string) string {
	baseURL, err := url.Parse(strings.TrimRight(strings.TrimSpace(websiteURL), "/"))
	if err != nil {
		return ""
	}
	if strings.TrimSpace(baseURL.Hostname()) == "" || strings.TrimSpace(workspaceSlug) == "" {
		return ""
	}

	cleanSegments := make([]string, 0, len(routeSegments))
	for _, segment := range routeSegments {
		if trimmed := strings.TrimSpace(segment); trimmed != "" {
			cleanSegments = append(cleanSegments, trimmed)
		}
	}

	host := baseURL.Hostname()
	if isLocalWebsiteHost(host) {
		baseURL.Path = path.Join(append([]string{"/", workspaceSlug}, cleanSegments...)...)
		return baseURL.String()
	}

	baseURL.Path = path.Join(append([]string{"/"}, cleanSegments...)...)
	if !strings.HasPrefix(host, workspaceSlug+".") {
		if port := baseURL.Port(); port != "" {
			baseURL.Host = fmt.Sprintf("%s.%s:%s", workspaceSlug, host, port)
		} else {
			baseURL.Host = fmt.Sprintf("%s.%s", workspaceSlug, host)
		}
	}

	return baseURL.String()
}

func buildStoryReference(teamCode string, sequenceID int, fallbackID string) string {
	if storyCode := buildStoryCode(teamCode, sequenceID); storyCode != "" {
		return storyCode
	}
	return strings.TrimSpace(fallbackID)
}

func buildStoryCode(teamCode string, sequenceID int) string {
	normalizedCode := strings.ToUpper(strings.TrimSpace(teamCode))
	if normalizedCode == "" || sequenceID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s-%d", normalizedCode, sequenceID)
}

func buildTaskURL(websiteURL, workspaceSlug, storyReference string) string {
	if strings.TrimSpace(storyReference) == "" {
		return ""
	}
	return buildWorkspaceURL(websiteURL, workspaceSlug, "work", storyReference)
}

func buildRequestURL(websiteURL, workspaceSlug, teamID, requestID string) string {
	if strings.TrimSpace(teamID) == "" || strings.TrimSpace(requestID) == "" {
		return ""
	}
	return buildWorkspaceURL(websiteURL, workspaceSlug, "teams", teamID, "requests", requestID)
}

func isLocalWebsiteHost(host string) bool {
	return strings.EqualFold(host, "localhost") || strings.EqualFold(host, "0.0.0.0") || net.ParseIP(host) != nil
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func toCoreSlackWorkspace(record slackrepository.SlackWorkspaceRecord) CoreSlackWorkspace {
	return CoreSlackWorkspace{
		ID:                record.ID,
		SlackTeamID:       record.SlackTeamID,
		SlackTeamName:     record.SlackTeamName,
		SlackTeamDomain:   record.SlackTeamDomain,
		BotUserID:         record.BotUserID,
		Scope:             record.Scope,
		IsActive:          record.IsActive,
		InstalledByUserID: record.InstalledByUserID,
		CreatedAt:         record.CreatedAt,
		UpdatedAt:         record.UpdatedAt,
	}
}

func toCoreChannels(records []slackrepository.SlackChannelRecord) []CoreSlackChannel {
	channels := make([]CoreSlackChannel, 0, len(records))
	for _, record := range records {
		channels = append(channels, CoreSlackChannel{
			ID:             record.ID,
			SlackChannelID: record.SlackChannelID,
			Name:           record.Name,
			IsPrivate:      record.IsPrivate,
			IsArchived:     record.IsArchived,
			IsMember:       record.IsMember,
			IsActive:       record.IsActive,
			LastSyncedAt:   record.LastSyncedAt,
			CreatedAt:      record.CreatedAt,
			UpdatedAt:      record.UpdatedAt,
		})
	}
	return channels
}

func toCoreRequestLog(record slackrepository.SlackRequestLogRecord) CoreRequestLog {
	headers := map[string]string{}
	if len(record.Headers) > 0 {
		_ = json.Unmarshal(record.Headers, &headers)
	}
	return CoreRequestLog{
		ID:           record.ID,
		RequestType:  record.RequestType,
		Endpoint:     record.Endpoint,
		WorkspaceID:  record.WorkspaceID,
		SlackTeamID:  record.SlackTeamID,
		SlackUserID:  record.SlackUserID,
		SlackChannel: record.SlackChannel,
		Command:      record.Command,
		TriggerID:    record.TriggerID,
		RequestBody:  record.RequestBody,
		Headers:      headers,
		ResponseCode: record.ResponseCode,
		Outcome:      record.Outcome,
		ErrorMessage: record.ErrorMessage,
		CreatedAt:    record.CreatedAt,
	}
}

type oauthAccessResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	BotUserID    string `json:"bot_user_id"`
	AppID        string `json:"app_id"`
	Scope        string `json:"scope"`
	Team         struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Domain string `json:"domain"`
	} `json:"team"`
	Enterprise struct {
		ID string `json:"id"`
	} `json:"enterprise"`
	AuthedUser struct {
		ID string `json:"id"`
	} `json:"authed_user"`
}

type oauthInstallNoncePayload struct {
	WorkspaceSlug string `json:"workspace_slug"`
}

type slackWorkspaceUser struct {
	ID       string
	Username string
	FullName string
	Email    string
}

type slackModalPrivateMetadata struct {
	Source         requestSourceContext `json:"source"`
	SelectedTeamID string               `json:"selected_team_id,omitempty"`
}

func parseSlackModalPrivateMetadata(raw string) (slackModalPrivateMetadata, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return slackModalPrivateMetadata{}, nil
	}

	var structured slackModalPrivateMetadata
	if err := json.Unmarshal([]byte(trimmed), &structured); err == nil && !isZeroRequestSource(structured.Source) {
		return structured, nil
	}

	var legacy requestSourceContext
	if err := json.Unmarshal([]byte(trimmed), &legacy); err != nil {
		return slackModalPrivateMetadata{}, err
	}
	return slackModalPrivateMetadata{Source: legacy}, nil
}

func isZeroRequestSource(source requestSourceContext) bool {
	return strings.TrimSpace(source.SlackTeamID) == "" &&
		strings.TrimSpace(source.SlackTeamDomain) == "" &&
		strings.TrimSpace(source.SlackChannelID) == "" &&
		strings.TrimSpace(source.SlackChannel) == "" &&
		strings.TrimSpace(source.SlackMessageTS) == "" &&
		strings.TrimSpace(source.SlackThreadTS) == "" &&
		strings.TrimSpace(source.SlackUserID) == "" &&
		strings.TrimSpace(source.SlackUsername) == "" &&
		strings.TrimSpace(source.SlackText) == ""
}

type interactionPayload struct {
	Type        string `json:"type"`
	TriggerID   string `json:"trigger_id"`
	ResponseURL string `json:"response_url"`
	ActionID    string `json:"action_id"`
	BlockID     string `json:"block_id"`
	Value       string `json:"value"`
	Team        struct {
		ID     string `json:"id"`
		Domain string `json:"domain"`
	} `json:"team"`
	User struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"user"`
	Channel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	Message struct {
		Text     string `json:"text"`
		TS       string `json:"ts"`
		ThreadTS string `json:"thread_ts"`
		User     string `json:"user"`
	} `json:"message"`
	View struct {
		ID              string                     `json:"id"`
		Hash            string                     `json:"hash"`
		Type            string                     `json:"type"`
		CallbackID      string                     `json:"callback_id"`
		PrivateMetadata string                     `json:"private_metadata"`
		EntityURL       string                     `json:"entity_url"`
		AppUnfurlURL    string                     `json:"app_unfurl_url"`
		Channel         string                     `json:"channel"`
		MessageTS       string                     `json:"message_ts"`
		ThreadTS        string                     `json:"thread_ts"`
		ExternalRef     SlackWorkObjectExternalRef `json:"external_ref"`
		Blocks          []struct {
			BlockID string `json:"block_id"`
			Element struct {
				Type          string `json:"type"`
				ActionID      string `json:"action_id"`
				InitialOption struct {
					Value string `json:"value"`
				} `json:"initial_option"`
			} `json:"element"`
		} `json:"blocks"`
		State struct {
			Values interactionViewStateValues `json:"values"`
		} `json:"state"`
	} `json:"view"`
	Actions []struct {
		ActionID       string `json:"action_id"`
		BlockID        string `json:"block_id"`
		Type           string `json:"type"`
		Value          string `json:"value"`
		SelectedOption struct {
			Value string `json:"value"`
		} `json:"selected_option"`
		SelectedOptions []struct {
			Value string `json:"value"`
		} `json:"selected_options"`
	} `json:"actions"`
}
