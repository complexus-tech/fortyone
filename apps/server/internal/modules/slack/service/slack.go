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
	modalBlockObjective   = "objective"

	modalActionTeamSelect        = "team_select"
	modalActionTitleInput        = "title_input"
	modalActionDescriptionInput  = "description_input"
	modalActionStatusSelect      = "status_select"
	modalActionPrioritySelect    = "priority_select"
	modalActionAssigneeSelect    = "assignee_select"
	modalActionLabelsMultiSelect = "labels_multi_select"
	modalActionObjectiveSelect   = "objective_select"
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

func (s *Service) RuntimeSearchTeams(ctx context.Context, actor CoreRuntimeActor, query string) ([]CoreRuntimeOption, error) {
	workspace, actorID, err := s.runtimeWorkspaceAndLinkedActor(ctx, actor)
	if err != nil {
		return nil, err
	}
	teams, err := s.repo.ListWorkspaceTeamsForUser(ctx, workspace.ID, actorID)
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	options := make([]CoreRuntimeOption, 0, len(teams))
	for _, team := range teams {
		label := fmt.Sprintf("%s (%s)", team.Name, team.Code)
		if query != "" && !strings.Contains(strings.ToLower(label), query) {
			continue
		}
		options = append(options, CoreRuntimeOption{
			Label: label,
			Value: team.ID.String(),
		})
	}
	return options, nil
}

func (s *Service) RuntimeSearchStatuses(ctx context.Context, actor CoreRuntimeActor, teamIDRaw, query string) ([]CoreRuntimeOption, error) {
	_, _, team, err := s.findRuntimeTeamForActor(ctx, actor, teamIDRaw)
	if err != nil {
		return nil, err
	}
	statuses, err := s.repo.ListTeamStatuses(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	options := make([]CoreRuntimeOption, 0, len(statuses))
	for _, status := range statuses {
		if query != "" && !strings.Contains(strings.ToLower(status.Name), query) {
			continue
		}
		options = append(options, CoreRuntimeOption{
			Label: status.Name,
			Value: status.ID.String(),
		})
	}
	return options, nil
}

func (s *Service) RuntimeSearchMembers(ctx context.Context, actor CoreRuntimeActor, teamIDRaw, query string) ([]CoreRuntimeOption, error) {
	_, _, team, err := s.findRuntimeTeamForActor(ctx, actor, teamIDRaw)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.SearchTeamMembers(ctx, team.ID, query, 25)
	if err != nil {
		return nil, err
	}
	options := make([]CoreRuntimeOption, 0, len(members))
	for _, member := range members {
		options = append(options, CoreRuntimeOption{
			Label: teamMemberDisplayName(member),
			Value: member.UserID.String(),
		})
	}
	return options, nil
}

func (s *Service) RuntimeSearchObjectives(ctx context.Context, actor CoreRuntimeActor, teamIDRaw, query string) ([]CoreRuntimeOption, error) {
	workspace, _, team, err := s.findRuntimeTeamForActor(ctx, actor, teamIDRaw)
	if err != nil {
		return nil, err
	}
	objectives, err := s.repo.SearchTeamObjectives(ctx, workspace.ID, team.ID, query, 25)
	if err != nil {
		return nil, err
	}
	options := make([]CoreRuntimeOption, 0, len(objectives))
	for _, objective := range objectives {
		options = append(options, CoreRuntimeOption{
			Label: objective.Name,
			Value: objective.ID.String(),
		})
	}
	return options, nil
}

func (s *Service) RuntimeSearchLabels(ctx context.Context, actor CoreRuntimeActor, teamIDRaw, query string) ([]CoreRuntimeOption, error) {
	workspace, _, team, err := s.findRuntimeTeamForActor(ctx, actor, teamIDRaw)
	if err != nil {
		return nil, err
	}
	labels, err := s.repo.SearchTeamLabels(ctx, workspace.ID, team.ID, query, 25)
	if err != nil {
		return nil, err
	}
	options := make([]CoreRuntimeOption, 0, len(labels))
	for _, label := range labels {
		options = append(options, CoreRuntimeOption{
			Label: label.Name,
			Value: label.ID.String(),
		})
	}
	return options, nil
}

func (s *Service) RuntimeCreateStory(ctx context.Context, input CoreRuntimeCreateStoryInput) (CoreRuntimeCreatedStory, error) {
	source := requestSourceContext{
		SlackTeamID:    strings.TrimSpace(input.Source.SlackTeamID),
		SlackChannelID: strings.TrimSpace(input.Source.SlackChannelID),
		SlackChannel:   strings.TrimSpace(input.Source.SlackChannel),
		SlackMessageTS: strings.TrimSpace(input.Source.SlackMessageTS),
		SlackThreadTS:  strings.TrimSpace(input.Source.SlackThreadTS),
		SlackUserID:    strings.TrimSpace(input.Source.SlackUserID),
		SlackUsername:  strings.TrimSpace(input.Source.SlackUserName),
		SlackText:      strings.TrimSpace(input.Source.SlackText),
	}
	if strings.TrimSpace(input.Title) == "" {
		return CoreRuntimeCreatedStory{}, errors.New("title is required")
	}

	actor := CoreRuntimeActor{
		SlackChannel:   source.SlackChannel,
		SlackChannelID: source.SlackChannelID,
		SlackMessageTS: source.SlackMessageTS,
		SlackTeamID:    source.SlackTeamID,
		SlackThreadTS:  source.SlackThreadTS,
		SlackUserID:    source.SlackUserID,
		SlackUserName:  source.SlackUsername,
	}
	workspace, actorID, team, err := s.findRuntimeTeamForActor(ctx, actor, input.TeamID)
	if err != nil {
		return CoreRuntimeCreatedStory{}, err
	}

	description := strings.TrimSpace(input.Description)
	sourceDescription := buildPrefilledDescription(source)
	if description == "" {
		description = sourceDescription
	} else if sourceDescription != "" {
		description = strings.TrimSpace(description + "\n\n" + sourceDescription)
	}
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	statusID, err := parseOptionalUUID(input.StatusID)
	if err != nil {
		return CoreRuntimeCreatedStory{}, fmt.Errorf("invalid status: %w", err)
	}
	if statusID == nil {
		statusID, err = s.repo.FindFirstStatusByCategory(ctx, team.ID, "unstarted")
		if err != nil {
			return CoreRuntimeCreatedStory{}, err
		}
	}
	assigneeID, err := parseOptionalUUID(input.AssigneeID)
	if err != nil {
		return CoreRuntimeCreatedStory{}, fmt.Errorf("invalid assignee: %w", err)
	}
	objectiveID, err := parseOptionalUUID(input.ObjectiveID)
	if err != nil {
		return CoreRuntimeCreatedStory{}, fmt.Errorf("invalid objective: %w", err)
	}
	labelIDs, err := s.validRuntimeLabelIDs(ctx, workspace.ID, team.ID, input.LabelIDs)
	if err != nil {
		return CoreRuntimeCreatedStory{}, err
	}

	story, err := s.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title:       strings.TrimSpace(input.Title),
		Description: descriptionPtr,
		Objective:   objectiveID,
		Status:      statusID,
		Assignee:    assigneeID,
		Reporter:    &actorID,
		Priority:    normalizeSlackPriority(input.Priority),
		Team:        team.ID,
		LabelIDs:    labelIDs,
	}, workspace.ID)
	if err != nil {
		return CoreRuntimeCreatedStory{}, err
	}

	if sourceURL := permalinkFromContext(source); sourceURL != "" {
		if err := s.repo.CreateStoryLink(ctx, story.ID, buildSlackStoryLinkTitle(source), sourceURL); err != nil {
			s.log.Warn(ctx, "failed creating slack source link", "error", err, "story_id", story.ID)
		}
	}

	ref := fmt.Sprintf("%s-%d", team.Code, story.SequenceID)
	return CoreRuntimeCreatedStory{
		ID:    story.ID.String(),
		Ref:   ref,
		Title: story.Title,
		URL:   buildTaskURL(s.cfg.WebsiteURL, workspace.Slug, story.ID.String()),
	}, nil
}

func (s *Service) RuntimeGetInstallation(ctx context.Context, slackTeamID string) (CoreRuntimeSlackInstallation, error) {
	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, strings.TrimSpace(slackTeamID))
	if err != nil {
		return CoreRuntimeSlackInstallation{}, err
	}
	installation := CoreRuntimeSlackInstallation{
		BotToken: strings.TrimSpace(slackWorkspace.BotAccessToken),
		TeamName: strings.TrimSpace(slackWorkspace.SlackTeamName),
	}
	if slackWorkspace.BotUserID != nil {
		installation.BotUserID = strings.TrimSpace(*slackWorkspace.BotUserID)
	}
	if installation.BotToken == "" {
		return CoreRuntimeSlackInstallation{}, errors.New("slack installation is missing bot token")
	}
	return installation, nil
}

func (s *Service) RuntimeResolveIdentity(ctx context.Context, actor CoreRuntimeActor) (CoreRuntimeIdentity, error) {
	slackTeamID := strings.TrimSpace(actor.SlackTeamID)
	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err != nil {
		return CoreRuntimeIdentity{}, err
	}
	workspace, err := s.repo.FindWorkspaceByID(ctx, slackWorkspace.WorkspaceID)
	if err != nil {
		return CoreRuntimeIdentity{}, err
	}

	source := requestSourceContext{
		SlackTeamID:    slackTeamID,
		SlackChannelID: strings.TrimSpace(actor.SlackChannelID),
		SlackChannel:   strings.TrimSpace(actor.SlackChannel),
		SlackMessageTS: strings.TrimSpace(actor.SlackMessageTS),
		SlackThreadTS:  strings.TrimSpace(actor.SlackThreadTS),
		SlackUserID:    strings.TrimSpace(actor.SlackUserID),
		SlackUsername:  strings.TrimSpace(actor.SlackUserName),
	}
	userID, connectURL, err := s.resolveLinkedSlackUser(ctx, workspace.ID, source)
	if err != nil {
		return CoreRuntimeIdentity{}, err
	}

	identity := CoreRuntimeIdentity{
		WorkspaceID:   workspace.ID,
		WorkspaceSlug: workspace.Slug,
		ConnectURL:    connectURL,
	}
	if userID != uuid.Nil {
		identity.UserID = &userID
	}
	return identity, nil
}

func (s *Service) RuntimeRecordLog(ctx context.Context, input CoreRuntimeLogInput) error {
	statusCode := input.ResponseCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	requestType := strings.TrimSpace(input.RequestType)
	if requestType == "" {
		requestType = "bot_runtime"
	}
	outcome := strings.TrimSpace(input.Outcome)
	if outcome == "" {
		outcome = "recorded"
	}

	workspaceID := s.resolveWorkspaceIDFromLog(ctx, input.Actor.SlackTeamID)
	return s.repo.InsertRequestLog(ctx, slackrepository.SlackRequestLogInsert{
		RequestType:  requestType,
		Endpoint:     strings.TrimSpace(input.Endpoint),
		WorkspaceID:  workspaceID,
		SlackTeamID:  optionalString(input.Actor.SlackTeamID),
		SlackUserID:  optionalString(input.Actor.SlackUserID),
		SlackChannel: optionalString(input.Actor.SlackChannelID),
		Headers:      []byte("{}"),
		ResponseCode: statusCode,
		Outcome:      truncateForLog(outcome, 120),
		ErrorMessage: optionalString(truncateForLog(input.ErrorMessage, 1000)),
	})
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
			"users:read",
			"users:read.email",
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

	slackWorkspace, err := s.repo.UpsertSlackWorkspace(ctx, workspaceID, installedByUserID, slackrepository.OAuthInstallPayload{
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

	if autoLinkErr := s.autoLinkWorkspaceMembers(ctx, slackWorkspace); autoLinkErr != nil {
		s.log.Warn(ctx, "slack connect succeeded but user auto-link failed", "workspace_id", workspaceID, "error", autoLinkErr)
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

func (s *Service) LinkSlackAccount(ctx context.Context, workspaceID, userID uuid.UUID, token string) error {
	if workspaceID == uuid.Nil {
		return errors.New("workspace id is required")
	}
	if userID == uuid.Nil {
		return errors.New("user id is required")
	}
	if strings.TrimSpace(s.cfg.SecretKey) == "" {
		return errors.New("slack link token signing secret is not configured")
	}

	values, err := s.verifyState(strings.TrimSpace(token))
	if err != nil {
		return err
	}
	if values["kind"] != "slack_user_link" {
		return errors.New("invalid slack link token")
	}

	expiresAt, err := strconv.ParseInt(values["expires"], 10, 64)
	if err != nil {
		return errors.New("invalid slack link token expiry")
	}
	if !s.clock.Now().Before(time.Unix(expiresAt, 0)) {
		return errors.New("slack link token has expired")
	}

	tokenWorkspaceID, err := uuid.Parse(values["workspace_id"])
	if err != nil {
		return errors.New("invalid slack link token workspace id")
	}
	if tokenWorkspaceID != workspaceID {
		return errors.New("slack link token workspace mismatch")
	}

	slackTeamID := strings.TrimSpace(values["slack_team_id"])
	slackUserID := strings.TrimSpace(values["slack_user_id"])
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

	linkedUserID, connectURL, err := s.resolveLinkedSlackUser(ctx, slackWorkspace.WorkspaceID, source)
	if err != nil {
		return CommandResponse{}, err
	}
	if linkedUserID == uuid.Nil {
		return CommandResponse{
			ResponseType: "ephemeral",
			Text:         buildConnectSlackAccountMessage(connectURL),
		}, nil
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
	case "block_suggestion":
		return s.handleBlockSuggestion(ctx, payload)
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
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return InteractionResponse{StatusCode: http.StatusOK}, nil
		}
		return InteractionResponse{}, err
	}

	linkedUserID, connectURL, err := s.resolveLinkedSlackUser(ctx, slackWorkspace.WorkspaceID, source)
	if err != nil {
		return InteractionResponse{}, err
	}
	if linkedUserID == uuid.Nil {
		message := buildConnectSlackAccountMessage(connectURL)
		if responseURL := strings.TrimSpace(payload.ResponseURL); responseURL != "" {
			if responseErr := s.postCommandResponse(ctx, responseURL, message); responseErr != nil {
				s.log.Error(ctx, "failed posting slack connect prompt via response_url", "error", responseErr)
			}
		} else {
			if responseErr := s.postEphemeralMessage(ctx, slackWorkspace.BotAccessToken, source.SlackChannelID, source.SlackUserID, message); responseErr != nil {
				s.log.Error(ctx, "failed posting slack connect prompt", "error", responseErr)
			}
		}
		return InteractionResponse{StatusCode: http.StatusOK}, nil
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
			StatusKind:  submission.StatusKind,
			TeamID:      submission.TeamID,
			StatusID:    submission.StatusID,
			Priority:    submission.Priority,
			AssigneeID:  submission.AssigneeID,
			LabelIDs:    submission.LabelIDs,
			ObjectiveID: submission.ObjectiveID,
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

	actorID, connectURL, err := s.resolveLinkedSlackUser(ctx, workspace.ID, submission.Source)
	if err != nil {
		s.log.Error(ctx, "failed resolving slack actor user", "error", err, "workspace_id", workspace.ID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	if actorID == uuid.Nil {
		message := buildConnectSlackAccountMessage(connectURL)
		if promptErr := s.postEphemeralMessage(ctx, slackWorkspace.BotAccessToken, submission.Source.SlackChannelID, submission.Source.SlackUserID, message); promptErr != nil {
			s.log.Error(ctx, "failed posting slack connect prompt for view submission", "error", promptErr)
		}
		return interactionValidationErrors(map[string]string{"title": "Connect your FortyOne account first, then submit again."})
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

	var objectiveID *uuid.UUID
	if submission.ObjectiveID != nil {
		if _, objectiveErr := s.repo.FindTeamObjectiveByID(ctx, workspace.ID, team.ID, *submission.ObjectiveID); objectiveErr != nil {
			if !slackrepository.IsNotFound(objectiveErr) {
				s.log.Error(ctx, "failed validating objective for slack submission", "error", objectiveErr, "workspace_id", workspace.ID, "team_id", team.ID)
				return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(objectiveErr)})
			}
		} else {
			objectiveID = submission.ObjectiveID
		}
	}

	priority := normalizeSlackPriority(submission.Priority)
	story, err := s.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title:       submission.Title,
		Description: descriptionPtr,
		Status:      statusID,
		Assignee:    assigneeID,
		Objective:   objectiveID,
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

	if sourceURL != "" {
		linkTitle := buildSlackStoryLinkTitle(submission.Source)
		if linkErr := s.repo.CreateStoryLink(ctx, story.ID, linkTitle, sourceURL); linkErr != nil {
			s.log.Error(ctx, "failed creating slack story link", "error", linkErr, "workspace_id", workspace.ID, "story_id", story.ID)
		}
	}

	s.postSlackTaskAck(ctx, submission.Source, slackWorkspace.BotAccessToken, workspace.Slug, story)
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
	teamID, err := s.resolveTeamIDForSuggestion(ctx, payload, slackWorkspace.WorkspaceID)
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

	query := suggestionQuery(payload)
	if len([]rune(query)) < 2 {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_query_too_short",
			Query:        query,
			ActionID:     suggestionActionID(payload),
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}

	const optionsLimit = 25
	switch suggestionActionID(payload) {
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
			ActionID:     suggestionActionID(payload),
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
	details := map[string]any{
		"type":          payload.Type,
		"action_id":     strings.TrimSpace(input.ActionID),
		"query":         strings.TrimSpace(input.Query),
		"team_id":       strings.TrimSpace(input.SlackTeamID),
		"resolved_team": strings.TrimSpace(input.ResolvedTeam.String()),
		"result_count":  input.ResultCount,
		"reason":        strings.TrimSpace(input.Reason),
		"user_id":       strings.TrimSpace(payload.User.ID),
		"channel_id":    strings.TrimSpace(payload.Channel.ID),
	}
	body, err := json.Marshal(details)
	if err != nil {
		return
	}

	var workspaceIDPtr *uuid.UUID
	if input.WorkspaceID != uuid.Nil {
		workspaceIDPtr = &input.WorkspaceID
	}
	slackTeamID := optionalString(input.SlackTeamID)
	slackUserID := optionalString(payload.User.ID)
	slackChannelID := optionalString(payload.Channel.ID)
	triggerID := optionalString(payload.TriggerID)
	requestBody := optionalString(string(body))
	errorMessage := optionalString(input.Reason)

	if insertErr := s.repo.InsertRequestLog(ctx, slackrepository.SlackRequestLogInsert{
		RequestType:  "suggestion_search",
		Endpoint:     "/integrations/slack/interactivity",
		WorkspaceID:  workspaceIDPtr,
		SlackTeamID:  slackTeamID,
		SlackUserID:  slackUserID,
		SlackChannel: slackChannelID,
		TriggerID:    triggerID,
		RequestBody:  requestBody,
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
	StatusKind  string
	TeamID      uuid.UUID
	StatusID    *uuid.UUID
	Priority    string
	AssigneeID  *uuid.UUID
	LabelIDs    []uuid.UUID
	ObjectiveID *uuid.UUID
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

	statusOptions := make([]map[string]any, 0, len(statuses)+1)

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
	metadataPayload, err := json.Marshal(slackModalPrivateMetadata{
		Source:         input.Source,
		SelectedTeamID: selectedTeam.ID.String(),
	})
	if err != nil {
		return nil, err
	}

	priorityOption := toSlackOption(normalizeSlackPriority(input.Selection.Priority), normalizeSlackPriority(input.Selection.Priority))

	blocks := []map[string]any{
		selectInputBlock(modalBlockTeam, modalActionTeamSelect, "Team", teamOptions, selectedTeamOption, false, true),
		plainInputBlock(modalBlockTitle, modalActionTitleInput, "Title", title, false, "", false),
		plainInputBlock(modalBlockDescription, modalActionDescriptionInput, "Description", input.Description, true, "", true),
		selectInputBlock(modalBlockStatus, modalActionStatusSelect, "Status", statusOptions, selectedStatusOption, true, false),
		externalSelectInputBlock(modalBlockAssignee, modalActionAssigneeSelect, "Assignee", selectedAssigneeOption, true, 2),
		externalMultiSelectInputBlock(modalBlockLabels, modalActionLabelsMultiSelect, "Labels", selectedLabelOptions, true, 2),
		externalSelectInputBlock(modalBlockObjective, modalActionObjectiveSelect, "Objective", selectedObjectiveOption, true, 2),
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

func externalSelectInputBlock(blockID, actionID, label string, initialOption map[string]any, optional bool, minQueryLength int) map[string]any {
	element := map[string]any{
		"type":      "external_select",
		"action_id": actionID,
	}
	if initialOption != nil {
		element["initial_option"] = initialOption
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

func externalMultiSelectInputBlock(blockID, actionID, label string, initialOptions []map[string]any, optional bool, minQueryLength int) map[string]any {
	element := map[string]any{
		"type":      "multi_external_select",
		"action_id": actionID,
	}
	if len(initialOptions) > 0 {
		element["initial_options"] = initialOptions
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

	var objectiveID *uuid.UUID
	selectedObjectiveID := readSelectedOption(modalBlockObjective)
	if selectedObjectiveID != "" {
		parsedObjectiveID, parseErr := uuid.Parse(selectedObjectiveID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected objective")
		}
		objectiveID = &parsedObjectiveID
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
		ObjectiveID: objectiveID,
		Source:      source,
	}, nil
}

func parseSourceFromPrivateMetadata(privateMetadata string) (requestSourceContext, error) {
	metadata, err := parseSlackModalPrivateMetadata(privateMetadata)
	if err != nil {
		return requestSourceContext{}, err
	}
	return metadata.Source, nil
}

func selectedTeamIDFromState(values map[string]map[string]struct {
	Type           string `json:"type"`
	Value          string `json:"value"`
	SelectedOption struct {
		Value string `json:"value"`
	} `json:"selected_option"`
	SelectedOptions []struct {
		Value string `json:"value"`
	} `json:"selected_options"`
}) string {
	block := values[modalBlockTeam]
	for _, action := range block {
		value := strings.TrimSpace(action.SelectedOption.Value)
		if value != "" {
			return value
		}
	}
	return ""
}

func (s *Service) resolveTeamIDForSuggestion(ctx context.Context, payload interactionPayload, workspaceID uuid.UUID) (uuid.UUID, error) {
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

	teams, err := s.repo.ListWorkspaceTeams(ctx, workspaceID)
	if err != nil {
		return uuid.Nil, err
	}
	if len(teams) == 0 {
		return uuid.Nil, ErrSlackNoTeamsAvailable
	}
	return teams[0].ID, nil
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

func (s *Service) autoLinkWorkspaceMembers(ctx context.Context, slackWorkspace slackrepository.SlackWorkspaceRecord) error {
	slackUsers, err := s.fetchWorkspaceUsers(ctx, slackWorkspace.BotAccessToken)
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

func (s *Service) buildSlackUserLinkURL(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (string, error) {
	if workspaceID == uuid.Nil {
		return "", errors.New("workspace id is required")
	}
	if strings.TrimSpace(s.cfg.SecretKey) == "" {
		return "", errors.New("slack link token signing secret is not configured")
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
	token, err := s.signState(map[string]string{
		"kind":          "slack_user_link",
		"workspace_id":  workspaceID.String(),
		"slack_team_id": slackTeamID,
		"slack_user_id": slackUserID,
		"expires":       strconv.FormatInt(s.clock.Now().Add(30*time.Minute).Unix(), 10),
	})
	if err != nil {
		return "", err
	}

	baseLink := s.buildWorkspaceIntegrationURL(workspace.Slug)
	linkURL, err := url.Parse(baseLink)
	if err != nil {
		return "", nil
	}
	query := linkURL.Query()
	query.Set("slack_link_token", token)
	linkURL.RawQuery = query.Encode()
	return linkURL.String(), nil
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

func interactionOptionsResponse(options []map[string]any) (InteractionResponse, error) {
	if options == nil {
		options = make([]map[string]any, 0)
	}
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
	if len(firstLine) <= 120 {
		return firstLine
	}
	return strings.TrimSpace(firstLine[:120])
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
		return "> " + message
	}
	if strings.TrimSpace(source.SlackUserID) == "" {
		return fmt.Sprintf("@%s said:\n> %s", identity, message)
	}
	return fmt.Sprintf("@[%s](%s) said:\n> %s", identity, strings.TrimSpace(source.SlackUserID), message)
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

func parseOptionalUUID(value string) (*uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (s *Service) runtimeWorkspaceAndLinkedActor(ctx context.Context, actor CoreRuntimeActor) (slackrepository.WorkspaceRecord, uuid.UUID, error) {
	slackTeamID := strings.TrimSpace(actor.SlackTeamID)
	workspace, err := s.repo.GetWorkspaceBySlackTeamID(ctx, slackTeamID)
	if err != nil {
		return slackrepository.WorkspaceRecord{}, uuid.Nil, err
	}
	userID, err := s.repo.FindLinkedUserIDBySlackUser(ctx, workspace.ID, slackTeamID, strings.TrimSpace(actor.SlackUserID))
	if err != nil {
		return slackrepository.WorkspaceRecord{}, uuid.Nil, err
	}
	if userID == nil || *userID == uuid.Nil {
		return slackrepository.WorkspaceRecord{}, uuid.Nil, errors.New("slack user is not linked to a fortyone user")
	}
	return workspace, *userID, nil
}

func (s *Service) findRuntimeTeamForActor(ctx context.Context, actor CoreRuntimeActor, teamIDRaw string) (slackrepository.WorkspaceRecord, uuid.UUID, slackrepository.TeamRecord, error) {
	workspace, actorID, err := s.runtimeWorkspaceAndLinkedActor(ctx, actor)
	if err != nil {
		return slackrepository.WorkspaceRecord{}, uuid.Nil, slackrepository.TeamRecord{}, err
	}
	teamID, err := uuid.Parse(strings.TrimSpace(teamIDRaw))
	if err != nil {
		return slackrepository.WorkspaceRecord{}, uuid.Nil, slackrepository.TeamRecord{}, err
	}
	team, err := s.repo.FindTeamByID(ctx, workspace.ID, teamID)
	if err != nil {
		return slackrepository.WorkspaceRecord{}, uuid.Nil, slackrepository.TeamRecord{}, err
	}
	if _, err := s.repo.FindTeamMemberByID(ctx, team.ID, actorID); err != nil {
		return slackrepository.WorkspaceRecord{}, uuid.Nil, slackrepository.TeamRecord{}, errors.New("selected team is not available to the slack user")
	}
	return workspace, actorID, team, nil
}

func (s *Service) validRuntimeLabelIDs(ctx context.Context, workspaceID, teamID uuid.UUID, rawLabelIDs []string) ([]uuid.UUID, error) {
	if len(rawLabelIDs) == 0 {
		return nil, nil
	}
	parsed := make([]uuid.UUID, 0, len(rawLabelIDs))
	for _, rawLabelID := range rawLabelIDs {
		rawLabelID = strings.TrimSpace(rawLabelID)
		if rawLabelID == "" {
			continue
		}
		labelID, err := uuid.Parse(rawLabelID)
		if err != nil {
			return nil, fmt.Errorf("invalid label: %w", err)
		}
		parsed = append(parsed, labelID)
	}
	if len(parsed) == 0 {
		return nil, nil
	}

	labels, err := s.repo.ListTeamLabels(ctx, workspaceID, teamID)
	if err != nil {
		return nil, err
	}
	valid := filterValidLabelIDs(labels, parsed)
	if len(valid) != len(parsed) {
		return nil, errors.New("selected label is no longer available")
	}
	return valid, nil
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
	AccessToken string `json:"access_token"`
	BotUserID   string `json:"bot_user_id"`
	Scope       string `json:"scope"`
	Team        struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Domain string `json:"domain"`
	} `json:"team"`
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
		ID              string `json:"id"`
		Hash            string `json:"hash"`
		CallbackID      string `json:"callback_id"`
		PrivateMetadata string `json:"private_metadata"`
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
		Value          string `json:"value"`
		SelectedOption struct {
			Value string `json:"value"`
		} `json:"selected_option"`
		SelectedOptions []struct {
			Value string `json:"value"`
		} `json:"selected_options"`
	} `json:"actions"`
}
