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
	"net"
	"net/http"
	"net/url"
	"path"
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
	ErrSlackNoTeamsAvailable           = errors.New("no teams are available in this workspace")
	ErrSlackTeamSelectionRequired      = errors.New("team selection is required")
)

const (
	slackPriorityNoPriority = "No Priority"
	slackRequestStatusValue = "__fortyone_request__"
	slackStatusKindRequest  = "request"
	slackStatusKindStory    = "story"

	modalBlockTeam        = "team"
	modalBlockTitle       = "title"
	modalBlockDescription = "description"
	modalBlockStatus      = "status"
	modalBlockPriority    = "priority"
	modalBlockAssignee    = "assignee"
	modalBlockLabels      = "labels"

	modalActionTeamSelect        = "team_select"
	modalActionTitleInput        = "title_input"
	modalActionDescriptionInput  = "description_input"
	modalActionStatusSelect      = "status_select"
	modalActionPrioritySelect    = "priority_select"
	modalActionAssigneeSelect    = "assignee_select"
	modalActionLabelsMultiSelect = "labels_multi_select"
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

func (s *Service) DisconnectWorkspace(ctx context.Context, workspaceID uuid.UUID) error {
	if workspaceID == uuid.Nil {
		return errors.New("workspace id is required")
	}
	return s.repo.DisconnectSlackWorkspace(ctx, workspaceID)
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
	responseURL := strings.TrimSpace(values.Get("response_url"))
	s.openCreateTaskModalForCommand(triggerID, title, source, slackWorkspace.WorkspaceID, slackWorkspace.BotAccessToken, responseURL)

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
	case "block_actions":
		return s.handleBlockActions(ctx, payload)
	default:
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
}

func (s *Service) RecordRequestLog(ctx context.Context, input CoreRequestLogInput) {
	statusCode := input.ResponseCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	requestDetails := parseRequestLogDetails(input.RequestType, input.RawBody)
	workspaceID := s.resolveWorkspaceIDFromLog(ctx, requestDetails.SlackTeamID)
	bodyText := truncateForLog(string(input.RawBody), 8000)
	headersJSON, err := json.Marshal(input.Headers)
	if err != nil {
		headersJSON = []byte("{}")
	}
	if len(headersJSON) == 0 {
		headersJSON = []byte("{}")
	}

	entry := slackrepository.SlackRequestLogInsert{
		RequestType:  input.RequestType,
		Endpoint:     strings.TrimSpace(input.Endpoint),
		WorkspaceID:  workspaceID,
		SlackTeamID:  optionalString(requestDetails.SlackTeamID),
		SlackUserID:  optionalString(requestDetails.SlackUserID),
		SlackChannel: optionalString(requestDetails.SlackChannelID),
		Command:      optionalString(requestDetails.Command),
		TriggerID:    optionalString(requestDetails.TriggerID),
		RequestBody:  optionalString(bodyText),
		Headers:      headersJSON,
		ResponseCode: statusCode,
		Outcome:      truncateForLog(strings.TrimSpace(input.Outcome), 120),
		ErrorMessage: optionalString(truncateForLog(input.ErrorMessage, 1000)),
	}

	if err := s.repo.InsertRequestLog(ctx, entry); err != nil {
		s.log.Warn(ctx, "failed to insert slack request log", "error", err, "request_type", input.RequestType)
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

	if err := s.openCreateTaskModal(ctx, payload.TriggerID, title, description, source, slackWorkspace.WorkspaceID, slackWorkspace.BotAccessToken); err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK}, nil
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
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, submission.Source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return InteractionResponse{StatusCode: http.StatusOK}, nil
		}
		return InteractionResponse{}, err
	}

	view, err := s.buildCreateTaskModalView(ctx, createTaskModalViewInput{
		Title:       submission.Title,
		Description: submission.Description,
		Source:      submission.Source,
		WorkspaceID: slackWorkspace.WorkspaceID,
		Selection: createTaskModalSelection{
			StatusKind: submission.StatusKind,
			TeamID:     submission.TeamID,
			StatusID:   submission.StatusID,
			Priority:   submission.Priority,
			AssigneeID: submission.AssigneeID,
			LabelIDs:   submission.LabelIDs,
		},
	})
	if err != nil {
		return InteractionResponse{}, err
	}

	updatePayload := map[string]any{
		"view_id": payload.View.ID,
		"hash":    payload.View.Hash,
		"view":    view,
	}
	if err := s.callSlackAPI(ctx, slackWorkspace.BotAccessToken, "https://slack.com/api/views.update", updatePayload, nil); err != nil {
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
	if len(errorsByBlock) > 0 {
		return interactionValidationErrors(errorsByBlock)
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, submission.Source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return interactionValidationErrors(map[string]string{"title": "Slack workspace is not connected"})
		}
		s.log.Error(ctx, "failed loading slack workspace by team id", "error", err, "slack_team_id", submission.Source.SlackTeamID)
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
	team, err := s.repo.FindTeamByID(ctx, workspace.ID, submission.TeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return interactionValidationErrors(map[string]string{modalBlockTeam: "Selected team is no longer available"})
		}
		s.log.Error(ctx, "failed finding selected team for slack submission", "error", err, "workspace_id", workspace.ID, "team_id", submission.TeamID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
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
			submission.StatusID = nil
		}
	}
	if submission.StatusKind == slackStatusKindStory && submission.StatusID == nil {
		return interactionValidationErrors(map[string]string{modalBlockStatus: "Selected status is no longer available"})
	}

	if sendToRequests {
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
			s.log.Error(ctx, "failed creating slack integration request", "error", err, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
		}
		s.postSlackRequestAck(ctx, submission.Source, slackWorkspace.BotAccessToken, workspace.Slug, team.ID.String(), request.ID.String())
		return interactionClearResponse()
	}

	if slackWorkspace.InstalledByUserID == nil || *slackWorkspace.InstalledByUserID == uuid.Nil {
		return interactionValidationErrors(map[string]string{"title": "Slack install user is missing. Reconnect Slack from FortyOne settings."})
	}
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

	var assigneeID *uuid.UUID
	if submission.AssigneeID != nil {
		members, membersErr := s.repo.ListTeamMembers(ctx, team.ID)
		if membersErr != nil {
			s.log.Error(ctx, "failed loading team members for slack submission", "error", membersErr, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(membersErr)})
		}
		if teamMemberExists(members, *submission.AssigneeID) {
			assigneeID = submission.AssigneeID
		}
	}

	priority := normalizeSlackPriority(submission.Priority)
	actorID := *slackWorkspace.InstalledByUserID
	story, err := s.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title:       submission.Title,
		Description: descriptionPtr,
		Status:      statusID,
		Assignee:    assigneeID,
		Team:        team.ID,
		Reporter:    &actorID,
		Priority:    priority,
	}, workspace.ID)
	if err != nil {
		s.log.Error(ctx, "failed creating story from slack submission", "error", err, "workspace_id", workspace.ID, "team_id", team.ID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}

	if len(submission.LabelIDs) > 0 {
		labels, labelsErr := s.repo.ListTeamLabels(ctx, workspace.ID, team.ID)
		if labelsErr != nil {
			s.log.Error(ctx, "failed loading team labels for slack submission", "error", labelsErr, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(labelsErr)})
		}
		filteredLabelIDs := filterValidLabelIDs(labels, submission.LabelIDs)
		if len(filteredLabelIDs) > 0 {
			if updateErr := s.stories.UpdateLabels(ctx, story.ID, workspace.ID, filteredLabelIDs); updateErr != nil {
				s.log.Error(ctx, "failed applying labels on slack story", "error", updateErr, "workspace_id", workspace.ID, "story_id", story.ID)
				return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(updateErr)})
			}
		}
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

func (s *Service) openCreateTaskModal(ctx context.Context, triggerID, title, description string, source requestSourceContext, workspaceID uuid.UUID, botToken string) error {
	if strings.TrimSpace(botToken) == "" {
		return errors.New("missing slack bot token")
	}
	if strings.TrimSpace(triggerID) == "" {
		return errors.New("missing trigger id")
	}
	if workspaceID == uuid.Nil {
		return errors.New("missing workspace id")
	}

	view, err := s.buildCreateTaskModalView(ctx, createTaskModalViewInput{
		Title:       title,
		Description: description,
		Source:      source,
		WorkspaceID: workspaceID,
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

func (s *Service) openCreateTaskModalForCommand(triggerID, title string, source requestSourceContext, workspaceID uuid.UUID, botToken, responseURL string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := s.openCreateTaskModal(ctx, triggerID, title, "", source, workspaceID, botToken); err != nil {
			s.log.Error(ctx, "failed opening slack create task modal from command", "error", err, "workspace_id", workspaceID, "slack_team_id", source.SlackTeamID)
			if notifyErr := s.postCommandResponse(ctx, responseURL, "Unable to open the FortyOne create task form. Please try again."); notifyErr != nil {
				s.log.Error(ctx, "failed posting slack command failure response", "error", notifyErr, "workspace_id", workspaceID, "slack_team_id", source.SlackTeamID)
			}
		}
	}()
}

type createTaskModalViewInput struct {
	Title       string
	Description string
	Source      requestSourceContext
	WorkspaceID uuid.UUID
	Selection   createTaskModalSelection
}

type createTaskModalSelection struct {
	StatusKind string
	TeamID     uuid.UUID
	StatusID   *uuid.UUID
	Priority   string
	AssigneeID *uuid.UUID
	LabelIDs   []uuid.UUID
}

func (s *Service) buildCreateTaskModalView(ctx context.Context, input createTaskModalViewInput) (map[string]any, error) {
	teams, err := s.repo.ListWorkspaceTeams(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, ErrSlackNoTeamsAvailable
	}

	selectedTeam := selectTeam(teams, input.Selection.TeamID)
	teamOptions := make([]map[string]any, 0, len(teams))
	var selectedTeamOption map[string]any
	for _, team := range teams {
		option := toSlackOption(fmt.Sprintf("%s (%s)", team.Name, team.Code), team.ID.String())
		teamOptions = append(teamOptions, option)
		if team.ID == selectedTeam.ID {
			selectedTeamOption = option
		}
	}

	statuses, err := s.repo.ListTeamStatuses(ctx, selectedTeam.ID)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.ListTeamMembers(ctx, selectedTeam.ID)
	if err != nil {
		return nil, err
	}
	labels, err := s.repo.ListTeamLabels(ctx, input.WorkspaceID, selectedTeam.ID)
	if err != nil {
		return nil, err
	}

	statusOptions := make([]map[string]any, 0, len(statuses)+1)
	assigneeOptions := make([]map[string]any, 0, len(members))
	labelOptions := make([]map[string]any, 0, len(labels))

	requestStatusOption := toSlackOption("Request", slackRequestStatusValue)
	statusOptions = append(statusOptions, requestStatusOption)
	selectedStatusOption := requestStatusOption
	for _, status := range statuses {
		option := toSlackOption(status.Name, status.ID.String())
		statusOptions = append(statusOptions, option)
		if input.Selection.StatusKind == slackStatusKindStory && input.Selection.StatusID != nil && *input.Selection.StatusID == status.ID {
			selectedStatusOption = option
		}
	}

	var selectedAssigneeOption map[string]any
	for _, member := range members {
		displayName := teamMemberDisplayName(member)
		option := toSlackOption(displayName, member.UserID.String())
		assigneeOptions = append(assigneeOptions, option)
		if input.Selection.AssigneeID != nil && *input.Selection.AssigneeID == member.UserID {
			selectedAssigneeOption = option
		}
	}

	labelOptionByID := make(map[uuid.UUID]map[string]any, len(labels))
	for _, label := range labels {
		option := toSlackOption(label.Name, label.ID.String())
		labelOptions = append(labelOptions, option)
		labelOptionByID[label.ID] = option
	}
	selectedLabelOptions := make([]map[string]any, 0, len(input.Selection.LabelIDs))
	for _, labelID := range input.Selection.LabelIDs {
		option, ok := labelOptionByID[labelID]
		if !ok {
			continue
		}
		selectedLabelOptions = append(selectedLabelOptions, option)
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "New task"
	}
	metadataPayload, err := json.Marshal(input.Source)
	if err != nil {
		return nil, err
	}

	priorityOption := toSlackOption(normalizeSlackPriority(input.Selection.Priority), normalizeSlackPriority(input.Selection.Priority))

	blocks := []map[string]any{
		selectInputBlock(modalBlockTeam, modalActionTeamSelect, "Team", teamOptions, selectedTeamOption, false, true),
		plainInputBlock(modalBlockTitle, modalActionTitleInput, "Title", title, false, ""),
		plainInputBlock(modalBlockDescription, modalActionDescriptionInput, "Description", input.Description, true, ""),
		selectInputBlock(modalBlockStatus, modalActionStatusSelect, "Status", statusOptions, selectedStatusOption, true, false),
	}
	if len(assigneeOptions) > 0 {
		blocks = append(blocks, selectInputBlock(modalBlockAssignee, modalActionAssigneeSelect, "Assignee", assigneeOptions, selectedAssigneeOption, true, false))
	}
	if len(labelOptions) > 0 {
		blocks = append(blocks, multiSelectInputBlock(modalBlockLabels, modalActionLabelsMultiSelect, "Labels", labelOptions, selectedLabelOptions, true))
	}
	blocks = append(blocks, selectInputBlock(modalBlockPriority, modalActionPrioritySelect, "Priority", slackPriorityOptions(), priorityOption, true, false))

	return map[string]any{
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
		"blocks": blocks,
	}, nil
}

func selectInputBlock(blockID, actionID, label string, options []map[string]any, initialOption map[string]any, optional, dispatchAction bool) map[string]any {
	element := map[string]any{
		"type":      "static_select",
		"action_id": actionID,
		"options":   options,
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
		"options":   options,
	}
	if len(initialOptions) > 0 {
		element["initial_options"] = initialOptions
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
	readSelectedOption := func(blockID string) string {
		block := state[blockID]
		for _, action := range block {
			if strings.TrimSpace(action.SelectedOption.Value) != "" {
				return strings.TrimSpace(action.SelectedOption.Value)
			}
		}
		return ""
	}
	readSelectedOptions := func(blockID string) []string {
		block := state[blockID]
		for _, action := range block {
			if len(action.SelectedOptions) == 0 {
				continue
			}
			values := make([]string, 0, len(action.SelectedOptions))
			for _, selected := range action.SelectedOptions {
				value := strings.TrimSpace(selected.Value)
				if value == "" {
					continue
				}
				values = append(values, value)
			}
			return values
		}
		return nil
	}

	var source requestSourceContext
	if strings.TrimSpace(payload.View.PrivateMetadata) != "" {
		if err := json.Unmarshal([]byte(payload.View.PrivateMetadata), &source); err != nil {
			return viewSubmissionData{}, err
		}
	}

	selectedTeamID := readSelectedOption(modalBlockTeam)
	if selectedTeamID == "" {
		return viewSubmissionData{}, ErrSlackTeamSelectionRequired
	}
	teamID, err := uuid.Parse(selectedTeamID)
	if err != nil {
		return viewSubmissionData{}, errors.New("invalid selected team")
	}

	var statusID *uuid.UUID
	statusKind := slackStatusKindRequest
	selectedStatusID := readSelectedOption(modalBlockStatus)
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
	selectedAssigneeID := readSelectedOption(modalBlockAssignee)
	if selectedAssigneeID != "" {
		parsedAssigneeID, parseErr := uuid.Parse(selectedAssigneeID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected assignee")
		}
		assigneeID = &parsedAssigneeID
	}

	selectedLabelIDs := make([]uuid.UUID, 0)
	for _, selectedLabelID := range readSelectedOptions(modalBlockLabels) {
		parsedLabelID, parseErr := uuid.Parse(selectedLabelID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected label")
		}
		selectedLabelIDs = append(selectedLabelIDs, parsedLabelID)
	}

	return viewSubmissionData{
		Title:       read(modalBlockTitle),
		Description: read(modalBlockDescription),
		TeamID:      teamID,
		StatusKind:  statusKind,
		StatusID:    statusID,
		Priority:    readSelectedOption(modalBlockPriority),
		AssigneeID:  assigneeID,
		LabelIDs:    selectedLabelIDs,
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

func (s *Service) postCommandResponse(ctx context.Context, responseURL, text string) error {
	if strings.TrimSpace(responseURL) == "" {
		return nil
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
	link := buildWorkspaceURL(s.cfg.WebsiteURL, workspaceSlug, "settings", "workspace", "integrations", "slack")
	if link == "" {
		return "/"
	}
	return link
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

func interactionErrorMessage(err error) string {
	if err == nil {
		return "Unable to create task. Please try again."
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "Unable to create task. Please try again."
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

	return requestLogDetails{
		SlackTeamID:    strings.TrimSpace(payload.Team.ID),
		SlackUserID:    strings.TrimSpace(payload.User.ID),
		SlackChannelID: strings.TrimSpace(payload.Channel.ID),
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

func buildTaskURL(websiteURL, workspaceSlug, storyID string) string {
	if strings.TrimSpace(storyID) == "" {
		return ""
	}
	return buildWorkspaceURL(websiteURL, workspaceSlug, "story", storyID)
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
		ID              string `json:"id"`
		Hash            string `json:"hash"`
		CallbackID      string `json:"callback_id"`
		PrivateMetadata string `json:"private_metadata"`
		State           struct {
			Values map[string]map[string]struct {
				Type           string `json:"type"`
				Value          string `json:"value"`
				SelectedOption struct {
					Value string `json:"value"`
				} `json:"selected_option"`
				SelectedOptions []struct {
					Value string `json:"value"`
				} `json:"selected_options"`
			} `json:"values"`
		} `json:"state"`
	} `json:"view"`
	Actions []struct {
		ActionID       string `json:"action_id"`
		BlockID        string `json:"block_id"`
		Type           string `json:"type"`
		SelectedOption struct {
			Value string `json:"value"`
		} `json:"selected_option"`
		SelectedOptions []struct {
			Value string `json:"value"`
		} `json:"selected_options"`
	} `json:"actions"`
}
