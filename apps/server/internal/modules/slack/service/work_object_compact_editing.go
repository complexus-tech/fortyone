package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/google/uuid"
)

const (
	slackWorkObjectCompactEditCallbackID = "fortyone_work_object_compact_edit"
	slackWorkObjectCompactStatusField    = "status"
	slackWorkObjectCompactPriorityField  = "priority"
)

var errSlackWorkObjectNoStatuses = errors.New("slack story has no statuses available")

type slackWorkObjectCompactEditMetadata struct {
	Source         requestSourceContext `json:"source"`
	StoryURL       string               `json:"story_url"`
	StoryReference string               `json:"story_reference"`
	Field          string               `json:"field"`
}

func isSlackStoryCompactEditAction(payload interactionPayload) bool {
	if strings.TrimSpace(payload.Type) != "block_actions" || len(payload.Actions) != 1 {
		return false
	}
	actionID := strings.TrimSpace(payload.Actions[0].ActionID)
	return actionID == slackEditStoryStatusActionID || actionID == slackEditStoryPriorityActionID
}

func isSlackStoryCompactEditSubmission(payload interactionPayload) bool {
	return strings.TrimSpace(payload.Type) == "view_submission" &&
		strings.TrimSpace(payload.View.CallbackID) == slackWorkObjectCompactEditCallbackID
}

func slackStoryCompactEditField(actionID string) (string, bool) {
	switch strings.TrimSpace(actionID) {
	case slackEditStoryStatusActionID:
		return slackWorkObjectCompactStatusField, true
	case slackEditStoryPriorityActionID:
		return slackWorkObjectCompactPriorityField, true
	default:
		return "", false
	}
}

func (s *Service) handleSlackStoryCompactEditAction(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if len(payload.Actions) != 1 {
		return InteractionResponse{}, fmt.Errorf("%w: compact edit action is required", errSlackWorkObjectEditMalformed)
	}
	field, ok := slackStoryCompactEditField(payload.Actions[0].ActionID)
	if !ok {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	if strings.TrimSpace(payload.TriggerID) == "" {
		return InteractionResponse{}, fmt.Errorf("%w: compact edit trigger is required", errSlackWorkObjectEditMalformed)
	}

	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, strings.TrimSpace(payload.Team.ID))
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return InteractionResponse{}, ErrSlackNoWorkspaceLinked
		}
		return InteractionResponse{}, err
	}
	if !installation.IsActive || installation.WorkspaceID == uuid.Nil || installation.ID == uuid.Nil {
		return InteractionResponse{}, fmt.Errorf("%w: Slack installation is inactive", errSlackWorkObjectEditDenied)
	}

	workspace, err := s.repo.FindWorkspaceByID(ctx, installation.WorkspaceID)
	if err != nil {
		return InteractionResponse{}, fmt.Errorf("load Slack Work Object workspace: %w", err)
	}
	postedURL := firstNonEmptyString(
		payload.AppUnfurl.AppUnfurlURL,
		payload.Container.AppUnfurlURL,
		payload.View.AppUnfurlURL,
	)
	link, err := ParseFortyOneStoryURL(postedURL)
	if err != nil || !strings.EqualFold(link.WorkspaceSlug, workspace.Slug) {
		return InteractionResponse{}, fmt.Errorf("%w: invalid compact story URL", errSlackWorkObjectEditMalformed)
	}
	if actionValue := strings.TrimSpace(payload.Actions[0].Value); actionValue != "" && !strings.EqualFold(actionValue, link.StoryReference) {
		return InteractionResponse{}, fmt.Errorf("%w: compact action story does not match its URL", errSlackWorkObjectEditMalformed)
	}

	channelID := firstNonEmptyString(payload.Container.ChannelID, payload.Channel.ID)
	messageTS := firstNonEmptyString(payload.Container.MessageTS, payload.Message.TS)
	if channelID == "" || messageTS == "" {
		return InteractionResponse{}, fmt.Errorf("%w: compact story message destination is required", errSlackWorkObjectEditMalformed)
	}
	source := requestSourceContext{
		SlackTeamID:     strings.TrimSpace(payload.Team.ID),
		SlackTeamDomain: strings.TrimSpace(payload.Team.Domain),
		SlackChannelID:  channelID,
		SlackChannel:    strings.TrimSpace(payload.Channel.Name),
		SlackMessageTS:  messageTS,
		SlackUserID:     strings.TrimSpace(payload.User.ID),
		SlackUsername:   firstNonEmptyString(payload.User.Username, payload.User.Name),
		ResponseURL:     strings.TrimSpace(payload.ResponseURL),
	}
	actorID, err := s.findLinkedInteractionActor(ctx, installation.WorkspaceID, source)
	if err != nil {
		return InteractionResponse{}, err
	}

	storyService, ok := s.stories.(slackWorkObjectStoryService)
	if !ok {
		return InteractionResponse{}, errors.New("slack Work Object story editing is not configured")
	}
	story, err := storyService.QueryByRef(ctx, installation.WorkspaceID, actorID, link.StoryReference)
	if err != nil {
		return InteractionResponse{}, fmt.Errorf("load compact Work Object story: %w", err)
	}
	if story.Workspace != uuid.Nil && story.Workspace != installation.WorkspaceID {
		return InteractionResponse{}, fmt.Errorf("%w: story workspace does not match installation", errSlackWorkObjectEditDenied)
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, installation.WorkspaceID, actorID, story.Team, source); err != nil {
		return InteractionResponse{}, fmt.Errorf("%w: story team is not available: %v", errSlackWorkObjectEditDenied, err)
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, installation.WorkspaceID, actorID, source); err != nil {
		return InteractionResponse{}, err
	}
	if err := s.requireCurrentSlackInstallation(ctx, installation.WorkspaceID, source.SlackTeamID, installation.InstallGeneration); err != nil {
		return InteractionResponse{}, err
	}

	statuses := []slackdomain.Status(nil)
	if field == slackWorkObjectCompactStatusField {
		statuses, err = s.repo.ListTeamStatuses(ctx, story.Team)
		if err != nil {
			return InteractionResponse{}, err
		}
	}
	view, err := buildSlackStoryCompactEditModal(slackWorkObjectCompactEditMetadata{
		Source:         source,
		StoryURL:       link.PostedURL,
		StoryReference: link.StoryReference,
		Field:          field,
	}, story, statuses)
	if err != nil {
		return InteractionResponse{}, err
	}
	botToken, err := s.botToken(ctx, installation)
	if err != nil {
		return InteractionResponse{}, err
	}
	if err := s.callSlackAPI(ctx, botToken, "https://slack.com/api/views.open", map[string]any{
		"trigger_id": payload.TriggerID,
		"view":       view,
	}, nil); err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK}, nil
}

func buildSlackStoryCompactEditModal(metadata slackWorkObjectCompactEditMetadata, story Story, statuses []slackdomain.Status) (map[string]any, error) {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	if len(metadataBytes) > modalMetadataMaxBytes {
		return nil, errors.New("slack compact edit metadata exceeds the provider limit")
	}

	fieldLabel := "Priority"
	options := slackPriorityOptions()
	initialOption := toSlackOption(normalizeSlackPriority(story.Priority), normalizeSlackPriority(story.Priority))
	if metadata.Field == slackWorkObjectCompactStatusField {
		fieldLabel = "Status"
		options = make([]map[string]any, 0, len(statuses))
		initialOption = nil
		for _, status := range statuses {
			option := toSlackOption(status.Name, status.ID.String())
			options = append(options, option)
			if story.Status != nil && status.ID == *story.Status {
				initialOption = option
			}
		}
		if len(options) == 0 {
			return nil, errSlackWorkObjectNoStatuses
		}
	} else if metadata.Field != slackWorkObjectCompactPriorityField {
		return nil, fmt.Errorf("%w: unsupported compact edit field", errSlackWorkObjectEditMalformed)
	}

	return map[string]any{
		"type":             "modal",
		"callback_id":      slackWorkObjectCompactEditCallbackID,
		"private_metadata": string(metadataBytes),
		"title": map[string]string{
			"type": "plain_text",
			"text": "Edit story",
		},
		"submit": map[string]string{
			"type": "plain_text",
			"text": "Save",
		},
		"close": map[string]string{
			"type": "plain_text",
			"text": "Cancel",
		},
		"blocks": []map[string]any{
			selectInputBlock(metadata.Field, metadata.Field+".input", fieldLabel, options, initialOption, false, false),
		},
	}, nil
}

func (s *Service) handleSlackStoryCompactEditSubmission(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if !isSlackStoryCompactEditSubmission(payload) {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	var metadata slackWorkObjectCompactEditMetadata
	if err := json.Unmarshal([]byte(payload.View.PrivateMetadata), &metadata); err != nil {
		return InteractionResponse{}, fmt.Errorf("%w: invalid compact edit metadata", errSlackWorkObjectEditMalformed)
	}
	metadata.Field = strings.TrimSpace(metadata.Field)
	if metadata.Field != slackWorkObjectCompactStatusField && metadata.Field != slackWorkObjectCompactPriorityField {
		return InteractionResponse{}, fmt.Errorf("%w: invalid compact edit field", errSlackWorkObjectEditMalformed)
	}
	link, err := ParseFortyOneStoryURL(metadata.StoryURL)
	if err != nil || !strings.EqualFold(link.StoryReference, strings.TrimSpace(metadata.StoryReference)) {
		return InteractionResponse{}, fmt.Errorf("%w: invalid compact story metadata", errSlackWorkObjectEditMalformed)
	}
	metadata.Source, err = interactionSourceForPayload(payload, metadata.Source)
	if err != nil {
		return InteractionResponse{}, err
	}

	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, metadata.Source.SlackTeamID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return InteractionResponse{}, ErrSlackNoWorkspaceLinked
		}
		return InteractionResponse{}, err
	}
	if !installation.IsActive || installation.WorkspaceID == uuid.Nil || installation.ID == uuid.Nil {
		return InteractionResponse{}, fmt.Errorf("%w: Slack installation is inactive", errSlackWorkObjectEditDenied)
	}
	workspace, err := s.repo.FindWorkspaceByID(ctx, installation.WorkspaceID)
	if err != nil {
		return InteractionResponse{}, err
	}
	if !strings.EqualFold(link.WorkspaceSlug, workspace.Slug) {
		return InteractionResponse{}, fmt.Errorf("%w: story workspace does not match installation", errSlackWorkObjectEditDenied)
	}
	actorID, err := s.findLinkedInteractionActor(ctx, installation.WorkspaceID, metadata.Source)
	if err != nil {
		return InteractionResponse{}, err
	}
	storyService, ok := s.stories.(slackWorkObjectStoryService)
	if !ok {
		return InteractionResponse{}, errors.New("slack Work Object story editing is not configured")
	}
	story, err := storyService.QueryByRef(ctx, installation.WorkspaceID, actorID, link.StoryReference)
	if err != nil {
		return InteractionResponse{}, err
	}
	if story.Workspace != uuid.Nil && story.Workspace != installation.WorkspaceID {
		return InteractionResponse{}, fmt.Errorf("%w: story workspace does not match installation", errSlackWorkObjectEditDenied)
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, installation.WorkspaceID, actorID, story.Team, metadata.Source); err != nil {
		return InteractionResponse{}, fmt.Errorf("%w: story team is not available: %v", errSlackWorkObjectEditDenied, err)
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, installation.WorkspaceID, actorID, metadata.Source); err != nil {
		return InteractionResponse{}, err
	}
	if err := s.requireCurrentSlackInstallation(ctx, installation.WorkspaceID, metadata.Source.SlackTeamID, installation.InstallGeneration); err != nil {
		return InteractionResponse{}, err
	}

	value, present, err := slackWorkObjectStateValue(payload.View.State.Values, metadata.Field, "static_select")
	if err != nil || !present {
		if err != nil {
			return InteractionResponse{}, err
		}
		return InteractionResponse{}, fmt.Errorf("%w: compact edit value is required", errSlackWorkObjectEditMalformed)
	}
	updates := map[string]any{}
	switch metadata.Field {
	case slackWorkObjectCompactStatusField:
		statusID, parseErr := uuid.Parse(strings.TrimSpace(value.SelectedOption.Value))
		if parseErr != nil || statusID == uuid.Nil {
			return InteractionResponse{}, fmt.Errorf("%w: invalid status", errSlackWorkObjectEditMalformed)
		}
		statuses, listErr := s.repo.ListTeamStatuses(ctx, story.Team)
		if listErr != nil {
			return InteractionResponse{}, listErr
		}
		if _, found := findStatusByID(statuses, statusID); !found {
			return InteractionResponse{}, fmt.Errorf("%w: status does not belong to the story team", errSlackWorkObjectEditDenied)
		}
		updates["status_id"] = statusID
	case slackWorkObjectCompactPriorityField:
		priority := strings.TrimSpace(value.SelectedOption.Value)
		if !isSlackStoryPriority(priority) {
			return InteractionResponse{}, fmt.Errorf("%w: invalid priority", errSlackWorkObjectEditMalformed)
		}
		updates["priority"] = priority
	}
	if err := storyService.UpdateExternalUserActionIfUnchanged(ctx, actorID, story.ID, installation.WorkspaceID, story.UpdatedAt, updates); err != nil {
		return InteractionResponse{}, fmt.Errorf("apply compact Slack Work Object edit: %w", err)
	}

	refreshed, err := storyService.QueryByRef(ctx, installation.WorkspaceID, actorID, link.StoryReference)
	if err != nil {
		return InteractionResponse{}, err
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, installation.WorkspaceID, actorID, refreshed.Team, metadata.Source); err != nil {
		return InteractionResponse{}, err
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, installation.WorkspaceID, actorID, metadata.Source); err != nil {
		return InteractionResponse{}, err
	}
	if err := s.requireCurrentSlackInstallation(ctx, installation.WorkspaceID, metadata.Source.SlackTeamID, installation.InstallGeneration); err != nil {
		return InteractionResponse{}, err
	}
	repository, ok := s.repo.(slackWorkObjectRepository)
	if !ok {
		return InteractionResponse{}, errors.New("slack Work Object repository is not configured")
	}
	input, err := buildSlackStoryWorkObjectInput(ctx, repository, actorID, metadata.Source.SlackUserID, link.PostedURL, refreshed, false)
	if err != nil {
		return InteractionResponse{}, err
	}
	request, err := BuildSlackStoryUnfurlRequest(metadata.Source.SlackChannelID, metadata.Source.SlackMessageTS, input)
	if err != nil {
		return InteractionResponse{}, err
	}
	botToken, err := s.botToken(ctx, installation)
	if err != nil {
		return InteractionResponse{}, err
	}
	if err := newSlackWorkObjectPublisher(s.webClient).Unfurl(ctx, botToken, request); err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK}, nil
}
