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
)

var (
	ErrSlackSigningSecretNotConfigured = errors.New("slack signing secret is not configured")
	ErrSlackRequestExpired             = errors.New("slack request timestamp is too old")
	ErrSlackInvalidSignature           = errors.New("invalid slack request signature")
	ErrSlackBotTokenNotConfigured      = errors.New("slack bot token is not configured")
	ErrInvalidCommandText              = errors.New("invalid slash command text")
)

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type Service struct {
	log      *logger.Logger
	repo     Repository
	requests RequestStore
	cfg      Config
	client   *http.Client
	clock    Clock
}

func New(log *logger.Logger, repo Repository, requests RequestStore, cfg Config) *Service {
	return &Service{
		log:      log,
		repo:     repo,
		requests: requests,
		cfg:      cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		clock: realClock{},
	}
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
	triggerID := strings.TrimSpace(values.Get("trigger_id"))
	if triggerID == "" {
		return CommandResponse{}, errors.New("missing trigger_id")
	}

	workspaceSlug, teamCode, title, err := parseCommandText(values.Get("text"))
	if err != nil {
		return CommandResponse{
			ResponseType: "ephemeral",
			Text:         "Usage: /fortyone create story <workspace-slug> <team-code> <title>",
		}, nil
	}

	source := requestSourceContext{
		SlackTeamID:     strings.TrimSpace(values.Get("team_id")),
		SlackTeamDomain: strings.TrimSpace(values.Get("team_domain")),
		SlackChannelID:  strings.TrimSpace(values.Get("channel_id")),
		SlackChannel:    strings.TrimSpace(values.Get("channel_name")),
		SlackUserID:     strings.TrimSpace(values.Get("user_id")),
		SlackUsername:   strings.TrimSpace(values.Get("user_name")),
	}
	if err := s.openCreateRequestModal(ctx, triggerID, workspaceSlug, teamCode, title, "", source); err != nil {
		return CommandResponse{}, err
	}

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

	if err := s.openCreateRequestModal(ctx, payload.TriggerID, "", "", title, description, source); err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK}, nil
}

func (s *Service) handleViewSubmission(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if payload.View.CallbackID != "fortyone_create_request" {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	submission, err := parseViewSubmission(payload)
	if err != nil {
		return InteractionResponse{}, err
	}

	errorsByBlock := map[string]string{}
	if submission.WorkspaceSlug == "" {
		errorsByBlock["workspace_slug"] = "Workspace slug is required"
	}
	if submission.TeamCode == "" {
		errorsByBlock["team_code"] = "Team code is required"
	}
	if submission.Title == "" {
		errorsByBlock["title"] = "Title is required"
	}
	if len(errorsByBlock) > 0 {
		body, err := json.Marshal(map[string]any{
			"response_action": "errors",
			"errors":          errorsByBlock,
		})
		if err != nil {
			return InteractionResponse{}, err
		}
		return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
	}

	workspace, err := s.repo.FindWorkspaceBySlug(ctx, submission.WorkspaceSlug)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			body, marshalErr := json.Marshal(map[string]any{
				"response_action": "errors",
				"errors": map[string]string{
					"workspace_slug": "Unknown workspace slug",
				},
			})
			if marshalErr != nil {
				return InteractionResponse{}, marshalErr
			}
			return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
		}
		return InteractionResponse{}, err
	}

	team, err := s.repo.FindTeamByCode(ctx, workspace.ID, submission.TeamCode)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			body, marshalErr := json.Marshal(map[string]any{
				"response_action": "errors",
				"errors": map[string]string{
					"team_code": "Unknown team code for this workspace",
				},
			})
			if marshalErr != nil {
				return InteractionResponse{}, marshalErr
			}
			return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
		}
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

	input := integrationrequests.CoreUpsertRequestInput{
		WorkspaceID:      workspace.ID,
		TeamID:           team.ID,
		Provider:         integrationrequests.ProviderSlack,
		SourceType:       SourceTypeSlackMessage,
		SourceExternalID: sourceExternalID,
		SourceURL:        nil,
		Title:            submission.Title,
		Description:      descriptionPtr,
		Metadata:         metadata,
	}
	if sourceURL != "" {
		input.SourceURL = &sourceURL
	}

	request, err := s.requests.UpsertPending(ctx, input)
	if err != nil {
		return InteractionResponse{}, err
	}

	s.postSlackAck(ctx, submission.Source, workspace.Slug, team.ID.String(), request.ID.String())

	body, err := json.Marshal(map[string]string{"response_action": "clear"})
	if err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
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
	workspaceSlug := metadataString(request.Metadata, "workspace_slug")
	storyURL := buildStoryURL(s.cfg.WebsiteURL, workspaceSlug, story.ID.String())
	text := fmt.Sprintf("✅ Accepted in FortyOne: %s", story.Title)
	if storyURL != "" {
		text = fmt.Sprintf("✅ Accepted in FortyOne: <%s|%s>", storyURL, story.Title)
	}
	if err := s.postMessage(ctx, channelID, threadTS, text); err != nil {
		s.log.Error(ctx, "failed posting acceptance update to slack", "error", err, "request_id", request.ID)
	}
	return nil
}

func (s *Service) openCreateRequestModal(ctx context.Context, triggerID, workspaceSlug, teamCode, title, description string, source requestSourceContext) error {
	if strings.TrimSpace(s.cfg.BotToken) == "" {
		return ErrSlackBotTokenNotConfigured
	}
	if strings.TrimSpace(triggerID) == "" {
		return errors.New("missing trigger id")
	}

	if title == "" {
		title = "New story"
	}
	metadataPayload, err := json.Marshal(source)
	if err != nil {
		return err
	}

	view := map[string]any{
		"type":             "modal",
		"callback_id":      "fortyone_create_request",
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
		"blocks": []map[string]any{
			plainInputBlock("workspace_slug", "value", "Workspace slug", workspaceSlug, false, "e.g. acme"),
			plainInputBlock("team_code", "value", "Team code", teamCode, false, "e.g. ENG"),
			plainInputBlock("title", "value", "Title", title, false, ""),
			plainInputBlock("description", "value", "Description", description, true, ""),
		},
	}

	payload := map[string]any{
		"trigger_id": triggerID,
		"view":       view,
	}
	return s.callSlackAPI(ctx, "https://slack.com/api/views.open", payload, nil)
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

func parseCommandText(text string) (workspaceSlug string, teamCode string, title string, err error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", "", "", ErrInvalidCommandText
	}
	parts := strings.Fields(trimmed)
	if len(parts) < 3 {
		return "", "", "", ErrInvalidCommandText
	}
	if len(parts) >= 5 && strings.EqualFold(parts[0], "create") && strings.EqualFold(parts[1], "story") {
		parts = parts[2:]
	}
	if len(parts) < 3 {
		return "", "", "", ErrInvalidCommandText
	}
	workspaceSlug = strings.TrimSpace(parts[0])
	teamCode = strings.TrimSpace(parts[1])
	title = strings.TrimSpace(strings.Join(parts[2:], " "))
	if workspaceSlug == "" || teamCode == "" || title == "" {
		return "", "", "", ErrInvalidCommandText
	}
	return workspaceSlug, teamCode, title, nil
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
		WorkspaceSlug: read("workspace_slug"),
		TeamCode:      read("team_code"),
		Title:         read("title"),
		Description:   read("description"),
		Source:        source,
	}, nil
}

func (s *Service) postSlackAck(ctx context.Context, source requestSourceContext, workspaceSlug, teamID, requestID string) {
	if source.SlackChannelID == "" {
		return
	}
	threadTS := source.SlackThreadTS
	if threadTS == "" {
		threadTS = source.SlackMessageTS
	}
	requestURL := buildRequestURL(s.cfg.WebsiteURL, workspaceSlug, teamID, requestID)
	text := "📥 Story queued in FortyOne."
	if requestURL != "" {
		text = fmt.Sprintf("📥 Story queued in FortyOne: <%s|Open story intake>", requestURL)
	}
	if err := s.postMessage(ctx, source.SlackChannelID, threadTS, text); err != nil {
		s.log.Error(ctx, "failed posting slack request acknowledgement", "error", err)
	}
}

func (s *Service) postMessage(ctx context.Context, channelID, threadTS, text string) error {
	if strings.TrimSpace(s.cfg.BotToken) == "" {
		return ErrSlackBotTokenNotConfigured
	}
	payload := map[string]any{
		"channel": channelID,
		"text":    text,
	}
	if strings.TrimSpace(threadTS) != "" {
		payload["thread_ts"] = threadTS
	}
	return s.callSlackAPI(ctx, "https://slack.com/api/chat.postMessage", payload, nil)
}

func (s *Service) callSlackAPI(ctx context.Context, endpoint string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.BotToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

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
		OK    bool            `json:"ok"`
		Error string          `json:"error"`
		Raw   json.RawMessage `json:"-"`
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

func messageToTitle(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "New request"
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

func buildStoryURL(websiteURL, workspaceSlug string, storyID string) string {
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
