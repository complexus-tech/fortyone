package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrSlackNotConfigured              = errors.New("slack integration is not configured")
	ErrSlackSigningSecretNotConfigured = errors.New("slack signing secret is not configured")
	ErrSlackRequestExpired             = errors.New("slack request timestamp is too old")
	ErrSlackInvalidSignature           = errors.New("invalid slack request signature")
	ErrSlackNoWorkspaceLinked          = errors.New("slack workspace is not connected")
	ErrSlackChannelNotLinked           = errors.New("this slack channel is not linked to a team")
	ErrSlackInvalidCreateMode          = errors.New("invalid slack create mode")
)

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type Service struct {
	log      *logger.Logger
	repo     Repository
	requests RequestStore
	stories  StoryService
	cfg      Config
	client   *http.Client
	clock    Clock
}

func New(log *logger.Logger, repo Repository, requests RequestStore, stories StoryService, cfg Config) *Service {
	return &Service{
		log:      log,
		repo:     repo,
		requests: requests,
		stories:  stories,
		cfg:      cfg,
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
		clock: realClock{},
	}
}

func (s *Service) GetIntegration(ctx context.Context, workspaceID uuid.UUID) (CoreIntegration, error) {
	settingsRecord, err := s.repo.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		return CoreIntegration{}, err
	}

	integration := CoreIntegration{
		Settings:     toCoreWorkspaceSettings(settingsRecord),
		Channels:     make([]CoreSlackChannel, 0),
		ChannelLinks: make([]CoreSlackChannelLink, 0),
	}

	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			links, listErr := s.repo.ListChannelLinks(ctx, workspaceID)
			if listErr == nil {
				integration.ChannelLinks = toCoreChannelLinks(links)
			}
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

	links, err := s.repo.ListChannelLinks(ctx, workspaceID)
	if err != nil && !slackrepository.IsNotFound(err) {
		return CoreIntegration{}, err
	}
	if err == nil {
		integration.ChannelLinks = toCoreChannelLinks(links)
	}

	return integration, nil
}

func (s *Service) CreateInstallSession(ctx context.Context, workspaceID, userID uuid.UUID, workspaceSlug string) (CoreCreateInstallSession, error) {
	if !s.canUseOAuth() {
		return CoreCreateInstallSession{}, ErrSlackNotConfigured
	}
	state, err := s.signState(map[string]string{
		"kind":           "slack_install",
		"workspace_id":   workspaceID.String(),
		"workspace_slug": workspaceSlug,
		"user_id":        userID.String(),
		"expires":        strconv.FormatInt(s.clock.Now().Add(10*time.Minute).Unix(), 10),
	})
	if err != nil {
		return CoreCreateInstallSession{}, err
	}

	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&scope=%s&state=%s&redirect_uri=%s",
		url.QueryEscape(s.cfg.ClientID),
		url.QueryEscape(strings.Join([]string{
			"commands",
			"chat:write",
			"channels:read",
			"groups:read",
			"channels:history",
			"groups:history",
		}, ",")),
		url.QueryEscape(state),
		url.QueryEscape(s.cfg.RedirectURL),
	)

	return CoreCreateInstallSession{InstallURL: authURL}, nil
}

func (s *Service) HandleSetup(ctx context.Context, code, state, slackError string) (string, error) {
	if !s.canUseOAuth() {
		return "", ErrSlackNotConfigured
	}
	if strings.TrimSpace(slackError) != "" {
		return "", fmt.Errorf("slack oauth failed: %s", slackError)
	}
	if strings.TrimSpace(code) == "" {
		return "", errors.New("missing slack oauth code")
	}
	values, err := s.verifyState(state)
	if err != nil {
		return "", err
	}
	if values["kind"] != "slack_install" {
		return "", errors.New("invalid slack install state")
	}
	expiresAt, err := strconv.ParseInt(values["expires"], 10, 64)
	if err != nil {
		return "", errors.New("invalid slack install state expiry")
	}
	if !s.clock.Now().Before(time.Unix(expiresAt, 0)) {
		return "", errors.New("slack install state has expired")
	}
	workspaceID, err := uuid.Parse(values["workspace_id"])
	if err != nil {
		return "", errors.New("invalid slack install workspace id")
	}
	installedByUserID, err := uuid.Parse(values["user_id"])
	if err != nil {
		return "", errors.New("invalid slack install user id")
	}
	workspaceSlug := strings.TrimSpace(values["workspace_slug"])

	oauthResp, err := s.exchangeOAuthCode(ctx, code)
	if err != nil {
		return "", err
	}

	_, err = s.repo.UpsertSlackWorkspace(ctx, workspaceID, installedByUserID, slackrepository.OAuthInstallPayload{
		SlackTeamID:     strings.TrimSpace(oauthResp.Team.ID),
		SlackTeamName:   strings.TrimSpace(oauthResp.Team.Name),
		SlackTeamDomain: strings.TrimSpace(oauthResp.Team.Domain),
		BotUserID:       optionalString(oauthResp.BotUserID),
		BotAccessToken:  strings.TrimSpace(oauthResp.AccessToken),
		Scope:           optionalString(oauthResp.Scope),
	})
	if err != nil {
		return "", err
	}

	if _, err := s.repo.GetWorkspaceSettings(ctx, workspaceID); err != nil {
		return "", err
	}
	if err := s.SyncChannels(ctx, workspaceID); err != nil {
		s.log.Warn(ctx, "slack connect succeeded but initial channel sync failed", "workspace_id", workspaceID, "error", err)
	}

	return s.buildWorkspaceIntegrationURL(workspaceSlug), nil
}

func (s *Service) SyncChannels(ctx context.Context, workspaceID uuid.UUID) error {
	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	channels, err := s.fetchChannels(ctx, slackWorkspace.BotAccessToken)
	if err != nil {
		return err
	}
	return s.repo.UpsertChannels(ctx, workspaceID, slackWorkspace.ID, channels)
}

func (s *Service) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, input CoreUpdateWorkspaceSettingsInput) (CoreWorkspaceSettings, error) {
	settings, err := s.repo.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		return CoreWorkspaceSettings{}, err
	}
	mode := settings.DefaultCreateMode
	if input.DefaultCreateMode != nil {
		mode = strings.TrimSpace(*input.DefaultCreateMode)
	}
	if mode != CreateModeCreateTaskNow && mode != CreateModeSendToRequests {
		return CoreWorkspaceSettings{}, ErrSlackInvalidCreateMode
	}
	updated, err := s.repo.UpdateWorkspaceSettings(ctx, workspaceID, mode)
	if err != nil {
		return CoreWorkspaceSettings{}, err
	}
	return toCoreWorkspaceSettings(updated), nil
}

func (s *Service) CreateChannelLink(ctx context.Context, workspaceID, createdByUserID uuid.UUID, input CoreCreateChannelLinkInput) (CoreSlackChannelLink, error) {
	if strings.TrimSpace(input.SlackChannelID) == "" {
		return CoreSlackChannelLink{}, errors.New("slack channel id is required")
	}
	if input.TeamID == uuid.Nil {
		return CoreSlackChannelLink{}, errors.New("team id is required")
	}
	if _, err := s.repo.FindTeamByID(ctx, workspaceID, input.TeamID); err != nil {
		return CoreSlackChannelLink{}, err
	}
	link, err := s.repo.UpsertChannelLink(ctx, workspaceID, strings.TrimSpace(input.SlackChannelID), input.TeamID, createdByUserID)
	if err != nil {
		return CoreSlackChannelLink{}, err
	}
	return toCoreChannelLink(link), nil
}

func (s *Service) DeleteChannelLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	return s.repo.DeleteChannelLink(ctx, workspaceID, linkID)
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

func (s *Service) HandleEvents(rawBody []byte) (EventResponse, error) {
	var payload struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return EventResponse{}, err
	}
	if payload.Type == "url_verification" {
		return EventResponse{Challenge: payload.Challenge}, nil
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
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return CommandResponse{ResponseType: "ephemeral", Text: "Slack is not connected to this FortyOne workspace."}, nil
		}
		return CommandResponse{}, err
	}

	title := parseCommandTitle(values.Get("text"))
	if err := s.openCreateTaskModal(ctx, triggerID, title, "", source, slackWorkspace.BotAccessToken); err != nil {
		return CommandResponse{}, err
	}

	return CommandResponse{
		ResponseType: "ephemeral",
		Text:         "Opening FortyOne create task form...",
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
		return s.handleMessageAction(ctx, payload)
	case "view_submission":
		return s.handleViewSubmission(ctx, payload)
	default:
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
}

func (s *Service) handleMessageAction(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	title := messageToTitle(payload.Message.Text)
	description := strings.TrimSpace(payload.Message.Text)
	source := requestSourceContext{
		SlackTeamID:     strings.TrimSpace(payload.Team.ID),
		SlackTeamDomain: strings.TrimSpace(payload.Team.Domain),
		SlackChannelID:  strings.TrimSpace(payload.Channel.ID),
		SlackChannel:    strings.TrimSpace(payload.Channel.Name),
		SlackMessageTS:  strings.TrimSpace(payload.Message.TS),
		SlackThreadTS:   strings.TrimSpace(payload.Message.ThreadTS),
		SlackUserID:     strings.TrimSpace(payload.Message.User),
		SlackUsername:   strings.TrimSpace(payload.User.Username),
		SlackText:       strings.TrimSpace(payload.Message.Text),
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return InteractionResponse{StatusCode: http.StatusOK}, nil
		}
		return InteractionResponse{}, err
	}

	if err := s.openCreateTaskModal(ctx, payload.TriggerID, title, description, source, slackWorkspace.BotAccessToken); err != nil {
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
		return InteractionResponse{}, err
	}

	errorsByBlock := map[string]string{}
	if submission.Title == "" {
		errorsByBlock["title"] = "Title is required"
	}
	if len(errorsByBlock) > 0 {
		return interactionValidationErrors(errorsByBlock)
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, submission.Source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return interactionValidationErrors(map[string]string{"title": "Slack workspace is not connected"})
		}
		return InteractionResponse{}, err
	}

	workspace, err := s.repo.FindWorkspaceByID(ctx, slackWorkspace.WorkspaceID)
	if err != nil {
		return InteractionResponse{}, err
	}
	team, err := s.repo.FindTeamByChannel(ctx, workspace.ID, submission.Source.SlackChannelID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return interactionValidationErrors(map[string]string{"title": "This Slack channel is not linked to a team in FortyOne settings."})
		}
		return InteractionResponse{}, err
	}

	settings, err := s.repo.GetWorkspaceSettings(ctx, workspace.ID)
	if err != nil {
		return InteractionResponse{}, err
	}

	sourceURL := permalinkFromContext(submission.Source)
	sourceExternalID := buildSourceExternalID(submission.Source)
	if sourceExternalID == "" {
		sourceExternalID = fmt.Sprintf("slack:%d", s.clock.Now().UnixNano())
	}

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

	mode := normalizeCreateMode(settings.DefaultCreateMode)
	if mode == CreateModeSendToRequests {
		requestInput := integrationrequests.CoreUpsertRequestInput{
			WorkspaceID:      workspace.ID,
			TeamID:           team.ID,
			Provider:         integrationrequests.ProviderSlack,
			SourceType:       SourceTypeSlackMessage,
			SourceExternalID: sourceExternalID,
			Title:            submission.Title,
			Description:      descriptionPtr,
			Metadata:         metadata,
		}
		if sourceURL != "" {
			requestInput.SourceURL = &sourceURL
		}
		request, err := s.requests.UpsertPending(ctx, requestInput)
		if err != nil {
			return InteractionResponse{}, err
		}
		s.postSlackRequestAck(ctx, submission.Source, slackWorkspace.BotAccessToken, workspace.Slug, team.ID.String(), request.ID.String())
		return interactionClearResponse()
	}

	if slackWorkspace.InstalledByUserID == nil || *slackWorkspace.InstalledByUserID == uuid.Nil {
		return interactionValidationErrors(map[string]string{"title": "Slack install user is missing. Reconnect Slack from FortyOne settings."})
	}
	statusID, err := s.repo.FindFirstStatusByCategory(ctx, team.ID, "unstarted")
	if err != nil {
		return InteractionResponse{}, err
	}
	actorID := *slackWorkspace.InstalledByUserID
	story, err := s.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title:       submission.Title,
		Description: descriptionPtr,
		Status:      statusID,
		Team:        team.ID,
		Reporter:    &actorID,
		Priority:    "No Priority",
	}, workspace.ID)
	if err != nil {
		return InteractionResponse{}, err
	}

	s.postSlackTaskAck(ctx, submission.Source, slackWorkspace.BotAccessToken, workspace.Slug, story)
	return interactionClearResponse()
}

func (s *Service) AcceptIntegrationRequest(ctx context.Context, request integrationrequests.CoreIntegrationRequest, story stories.CoreSingleStory) error {
	if request.Provider != integrationrequests.ProviderSlack {
		return nil
	}
	channelID := metadataString(request.Metadata, "slack_channel_id")
	if channelID == "" {
		return nil
	}
	threadTS := metadataString(request.Metadata, "slack_thread_ts")
	if threadTS == "" {
		threadTS = metadataString(request.Metadata, "slack_message_ts")
	}

	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return nil
		}
		return err
	}
	workspaceSlug := metadataString(request.Metadata, "workspace_slug")
	storyURL := buildTaskURL(s.cfg.WebsiteURL, workspaceSlug, story.ID.String())
	text := fmt.Sprintf("✅ Request accepted in FortyOne: %s", story.Title)
	if storyURL != "" {
		text = fmt.Sprintf("✅ Request accepted in FortyOne: <%s|%s>", storyURL, story.Title)
	}
	if err := s.postMessage(ctx, slackWorkspace.BotAccessToken, channelID, threadTS, text); err != nil {
		s.log.Error(ctx, "failed posting acceptance update to slack", "error", err, "request_id", request.ID)
	}
	return nil
}

func (s *Service) openCreateTaskModal(ctx context.Context, triggerID, title, description string, source requestSourceContext, botToken string) error {
	if strings.TrimSpace(botToken) == "" {
		return errors.New("missing slack bot token")
	}
	if strings.TrimSpace(triggerID) == "" {
		return errors.New("missing trigger id")
	}

	if title == "" {
		title = "New task"
	}
	metadataPayload, err := json.Marshal(source)
	if err != nil {
		return err
	}

	view := map[string]any{
		"type":             "modal",
		"callback_id":      "fortyone_create_task",
		"private_metadata": string(metadataPayload),
		"title": map[string]string{
			"type": "plain_text",
			"text": "Create Task",
		},
		"submit": map[string]string{
			"type": "plain_text",
			"text": "Create",
		},
		"close": map[string]string{
			"type": "plain_text",
			"text": "Cancel",
		},
		"blocks": []map[string]any{
			plainInputBlock("title", "value", "Title", title, false, ""),
			plainInputBlock("description", "value", "Description", description, true, ""),
		},
	}

	payload := map[string]any{
		"trigger_id": triggerID,
		"view":       view,
	}
	return s.callSlackAPI(ctx, botToken, "https://slack.com/api/views.open", payload, nil)
}

func plainInputBlock(blockID, actionID, label, initial string, multiline bool, placeholder string) map[string]any {
	element := map[string]any{
		"type":      "plain_text_input",
		"action_id": actionID,
	}
	if multiline {
		element["multiline"] = true
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
	return map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
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
	return strings.TrimSpace(strings.Join(parts, " "))
}

func parseViewSubmission(payload interactionPayload) (viewSubmissionData, error) {
	state := payload.View.State.Values
	read := func(blockID string) string {
		block := state[blockID]
		for _, action := range block {
			if strings.TrimSpace(action.Value) != "" {
				return strings.TrimSpace(action.Value)
			}
		}
		return ""
	}

	var source requestSourceContext
	if strings.TrimSpace(payload.View.PrivateMetadata) != "" {
		if err := json.Unmarshal([]byte(payload.View.PrivateMetadata), &source); err != nil {
			return viewSubmissionData{}, err
		}
	}

	return viewSubmissionData{
		Title:       read("title"),
		Description: read("description"),
		Source:      source,
	}, nil
}

func (s *Service) postSlackRequestAck(ctx context.Context, source requestSourceContext, botToken, workspaceSlug, teamID, requestID string) {
	if source.SlackChannelID == "" {
		return
	}
	threadTS := source.SlackThreadTS
	if threadTS == "" {
		threadTS = source.SlackMessageTS
	}
	requestURL := buildRequestURL(s.cfg.WebsiteURL, workspaceSlug, teamID, requestID)
	text := "📥 Request created in FortyOne."
	if requestURL != "" {
		text = fmt.Sprintf("📥 Request created in FortyOne: <%s|Open request>", requestURL)
	}
	if err := s.postMessage(ctx, botToken, source.SlackChannelID, threadTS, text); err != nil {
		s.log.Error(ctx, "failed posting slack request acknowledgement", "error", err)
	}
}

func (s *Service) postSlackTaskAck(ctx context.Context, source requestSourceContext, botToken, workspaceSlug string, story stories.CoreSingleStory) {
	if source.SlackChannelID == "" {
		return
	}
	threadTS := source.SlackThreadTS
	if threadTS == "" {
		threadTS = source.SlackMessageTS
	}
	taskURL := buildTaskURL(s.cfg.WebsiteURL, workspaceSlug, story.ID.String())
	text := fmt.Sprintf("✅ Task created in FortyOne: %s", story.Title)
	if taskURL != "" {
		text = fmt.Sprintf("✅ Task created in FortyOne: <%s|%s>", taskURL, story.Title)
	}
	if err := s.postMessage(ctx, botToken, source.SlackChannelID, threadTS, text); err != nil {
		s.log.Error(ctx, "failed posting slack task acknowledgement", "error", err)
	}
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

func (s *Service) exchangeOAuthCode(ctx context.Context, code string) (oauthAccessResponse, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", s.cfg.ClientID)
	form.Set("client_secret", s.cfg.ClientSecret)
	form.Set("redirect_uri", s.cfg.RedirectURL)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://slack.com/api/oauth.v2.access",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return oauthAccessResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return oauthAccessResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return oauthAccessResponse{}, err
	}
	if resp.StatusCode >= 300 {
		return oauthAccessResponse{}, fmt.Errorf("slack oauth exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var envelope struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return oauthAccessResponse{}, err
	}
	if !envelope.OK {
		if envelope.Error == "" {
			envelope.Error = "unknown_error"
		}
		return oauthAccessResponse{}, fmt.Errorf("slack oauth exchange failed: %s", envelope.Error)
	}

	var response oauthAccessResponse
	if err := json.Unmarshal(body, &response); err != nil {
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
	var req *http.Request
	var err error
	if payload == nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
	} else {
		body, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return marshalErr
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	if strings.TrimSpace(botToken) != "" {
		req.Header.Set("Authorization", "Bearer "+botToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack api %s failed with status %d: %s", endpoint, resp.StatusCode, string(respBytes))
	}

	var envelope struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &envelope); err != nil {
		return err
	}
	if !envelope.OK {
		if envelope.Error == "" {
			envelope.Error = "unknown_error"
		}
		return fmt.Errorf("slack api %s returned error: %s", endpoint, envelope.Error)
	}
	if out != nil {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) signState(values map[string]string) (string, error) {
	payloadBytes, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

func (s *Service) verifyState(value string) (map[string]string, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid slack install state")
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	_, _ = mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return nil, errors.New("invalid slack install state signature")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var values map[string]string
	if err := json.Unmarshal(payloadBytes, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func (s *Service) buildWorkspaceIntegrationURL(workspaceSlug string) string {
	base := strings.TrimRight(s.cfg.WebsiteURL, "/")
	if strings.TrimSpace(base) == "" || strings.TrimSpace(workspaceSlug) == "" {
		return "/"
	}
	return fmt.Sprintf("%s/%s/settings/workspace/integrations/slack", base, workspaceSlug)
}

func (s *Service) canUseOAuth() bool {
	return strings.TrimSpace(s.cfg.ClientID) != "" &&
		strings.TrimSpace(s.cfg.ClientSecret) != "" &&
		strings.TrimSpace(s.cfg.RedirectURL) != "" &&
		strings.TrimSpace(s.cfg.SecretKey) != ""
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

func messageToTitle(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "New task"
	}
	firstLine := strings.Split(trimmed, "\n")[0]
	firstLine = strings.TrimSpace(firstLine)
	if len(firstLine) <= 120 {
		return firstLine
	}
	return strings.TrimSpace(firstLine[:120])
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

func buildTaskURL(websiteURL, workspaceSlug, storyID string) string {
	if strings.TrimSpace(websiteURL) == "" || strings.TrimSpace(workspaceSlug) == "" || strings.TrimSpace(storyID) == "" {
		return ""
	}
	base := strings.TrimRight(websiteURL, "/")
	return fmt.Sprintf("%s/%s/story/%s", base, workspaceSlug, storyID)
}

func buildRequestURL(websiteURL, workspaceSlug, teamID, requestID string) string {
	if strings.TrimSpace(websiteURL) == "" || strings.TrimSpace(workspaceSlug) == "" || strings.TrimSpace(teamID) == "" || strings.TrimSpace(requestID) == "" {
		return ""
	}
	base := strings.TrimRight(websiteURL, "/")
	return fmt.Sprintf("%s/%s/teams/%s/requests/%s", base, workspaceSlug, teamID, requestID)
}

func normalizeCreateMode(mode string) string {
	trimmed := strings.TrimSpace(mode)
	if trimmed == CreateModeSendToRequests {
		return CreateModeSendToRequests
	}
	return CreateModeCreateTaskNow
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toCoreWorkspaceSettings(record slackrepository.SlackWorkspaceSettingsRecord) CoreWorkspaceSettings {
	return CoreWorkspaceSettings{
		WorkspaceID:       record.WorkspaceID,
		DefaultCreateMode: normalizeCreateMode(record.DefaultCreateMode),
		CreatedAt:         record.CreatedAt,
		UpdatedAt:         record.UpdatedAt,
	}
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

func toCoreChannelLinks(records []slackrepository.SlackChannelLinkRecord) []CoreSlackChannelLink {
	links := make([]CoreSlackChannelLink, 0, len(records))
	for _, record := range records {
		links = append(links, toCoreChannelLink(record))
	}
	return links
}

func toCoreChannelLink(record slackrepository.SlackChannelLinkRecord) CoreSlackChannelLink {
	return CoreSlackChannelLink{
		ID:             record.ID,
		SlackChannelID: record.SlackChannelID,
		TeamID:         record.TeamID,
		TeamCode:       record.TeamCode,
		TeamName:       record.TeamName,
		TeamColor:      record.TeamColor,
		IsActive:       record.IsActive,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}

type oauthAccessResponse struct {
	AccessToken string `json:"access_token"`
	BotUserID   string `json:"bot_user_id"`
	Scope       string `json:"scope"`
	Team        struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Domain string `json:"domain"`
	} `json:"team"`
}

type interactionPayload struct {
	Type      string `json:"type"`
	TriggerID string `json:"trigger_id"`
	Team      struct {
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
		CallbackID      string `json:"callback_id"`
		PrivateMetadata string `json:"private_metadata"`
		State           struct {
			Values map[string]map[string]struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"values"`
		} `json:"state"`
	} `json:"view"`
}
