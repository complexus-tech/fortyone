package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	slackWorkObjectEditCallbackID = "work-object-edit"
	slackWorkObjectEntityViewType = "entity_detail"
)

var (
	errSlackWorkObjectEditDenied    = errors.New("slack Work Object edit is not authorized")
	errSlackWorkObjectEditMalformed = errors.New("slack Work Object edit is malformed")
)

type slackWorkObjectStoryService interface {
	QueryByRef(ctx context.Context, workspaceID uuid.UUID, storyRef string) (Story, error)
	UpdateExternalUserActionIfUnchanged(
		ctx context.Context,
		actorID, storyID, workspaceID uuid.UUID,
		expectedUpdatedAt time.Time,
		updates map[string]any,
	) error
}

func isSlackWorkObjectEditSubmission(payload interactionPayload) bool {
	if strings.TrimSpace(payload.Type) != "view_submission" || strings.TrimSpace(payload.View.Type) != slackWorkObjectEntityViewType {
		return false
	}
	callbackID := strings.TrimSpace(payload.View.CallbackID)
	return callbackID == "" || callbackID == slackWorkObjectEditCallbackID
}

// dispatchSlackWorkObjectEdit acknowledges Slack before doing any network or
// database work. Slack requires view submissions to be acknowledged within
// three seconds; the submission trigger is then used to refresh the flexpane.
func (s *Service) dispatchSlackWorkObjectEdit(parent context.Context, payload interactionPayload) {
	baseCtx := context.WithoutCancel(parent)
	ctx, cancel := s.newSlackWorkObjectTriggerContext(baseCtx)
	go func() {
		defer cancel()

		if err := s.processSlackWorkObjectEdit(ctx, payload); err != nil {
			if s.log != nil {
				s.log.Error(baseCtx, "failed processing Slack Work Object edit", "error", err, "view_id", payload.View.ID, "slack_team_id", payload.Team.ID, "slack_user_id", payload.User.ID)
			}
			if slackEntityDetailsPresentationWasAttempted(err) || isSlackEntityDetailsTerminalError(err) || ctx.Err() != nil {
				return
			}
			if presentErr := s.presentSlackWorkObjectEditError(ctx, payload, slackWorkObjectEditErrorMessage(err)); presentErr != nil && s.log != nil {
				s.log.Error(baseCtx, "failed presenting Slack Work Object edit error", "error", presentErr, "view_id", payload.View.ID, "slack_team_id", payload.Team.ID)
			}
		}
	}()
}

func (s *Service) processSlackWorkObjectEdit(ctx context.Context, payload interactionPayload) error {
	if !isSlackWorkObjectEditSubmission(payload) {
		return fmt.Errorf("%w: unsupported view", errSlackWorkObjectEditMalformed)
	}
	triggerID := strings.TrimSpace(payload.TriggerID)
	slackTeamID := strings.TrimSpace(payload.Team.ID)
	slackUserID := strings.TrimSpace(payload.User.ID)
	if triggerID == "" || slackTeamID == "" || slackUserID == "" {
		return fmt.Errorf("%w: trigger, team, and user are required", errSlackWorkObjectEditMalformed)
	}

	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err != nil {
		return fmt.Errorf("%w: load Slack installation: %v", errSlackWorkObjectEditDenied, err)
	}
	if !installation.IsActive || installation.WorkspaceID == uuid.Nil || installation.ID == uuid.Nil {
		return fmt.Errorf("%w: Slack installation is inactive", errSlackWorkObjectEditDenied)
	}
	workspace, err := s.repo.FindWorkspaceByID(ctx, installation.WorkspaceID)
	if err != nil {
		return fmt.Errorf("load Work Object workspace: %w", err)
	}

	entityURL := strings.TrimSpace(payload.View.EntityURL)
	if entityURL == "" {
		entityURL = strings.TrimSpace(payload.View.AppUnfurlURL)
	}
	link, err := ParseFortyOneStoryURL(entityURL)
	if err != nil || !strings.EqualFold(link.WorkspaceSlug, workspace.Slug) {
		return fmt.Errorf("%w: invalid entity URL", errSlackWorkObjectEditMalformed)
	}
	if refType := strings.TrimSpace(payload.View.ExternalRef.Type); refType != "" && !strings.EqualFold(refType, slackStoryExternalRefType) {
		return fmt.Errorf("%w: invalid external reference type", errSlackWorkObjectEditMalformed)
	}
	if strings.TrimSpace(payload.View.ExternalRef.ID) == "" {
		return fmt.Errorf("%w: external reference is required", errSlackWorkObjectEditMalformed)
	}

	source := requestSourceContext{
		SlackTeamID:    slackTeamID,
		SlackChannelID: strings.TrimSpace(payload.View.Channel),
		SlackMessageTS: strings.TrimSpace(payload.View.MessageTS),
		SlackThreadTS:  strings.TrimSpace(payload.View.ThreadTS),
		SlackUserID:    slackUserID,
		SlackUsername:  firstNonEmptyString(payload.User.Username, payload.User.Name),
	}
	actorID, err := s.findLinkedInteractionActor(ctx, installation.WorkspaceID, source)
	if err != nil {
		return fmt.Errorf("%w: resolve linked actor: %v", errSlackWorkObjectEditDenied, err)
	}

	storyService, ok := s.stories.(slackWorkObjectStoryService)
	if !ok {
		return errors.New("slack Work Object story editing is not configured")
	}
	story, err := storyService.QueryByRef(ctx, installation.WorkspaceID, link.StoryReference)
	if err != nil {
		return fmt.Errorf("load Work Object story: %w", err)
	}
	if !slackStoryExternalRefMatches(link, story.ID.String(), payload.View.ExternalRef.ID) {
		return fmt.Errorf("%w: external reference does not match story", errSlackWorkObjectEditMalformed)
	}
	if story.Workspace != uuid.Nil && story.Workspace != installation.WorkspaceID {
		return fmt.Errorf("%w: story workspace does not match installation", errSlackWorkObjectEditDenied)
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, installation.WorkspaceID, actorID, story.Team, source); err != nil {
		return fmt.Errorf("%w: story team is no longer available: %v", errSlackWorkObjectEditDenied, err)
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, installation.WorkspaceID, actorID, source); err != nil {
		return fmt.Errorf("%w: Slack actor is no longer linked: %v", errSlackWorkObjectEditDenied, err)
	}
	if err := s.requireCurrentSlackInstallation(ctx, installation.WorkspaceID, slackTeamID, installation.InstallGeneration); err != nil {
		return fmt.Errorf("%w: Slack installation changed: %v", errSlackWorkObjectEditDenied, err)
	}

	updates, err := s.authorizedSlackWorkObjectUpdates(ctx, story, payload.View.State.Values)
	if err != nil {
		return err
	}
	if len(updates) > 0 {
		if err := storyService.UpdateExternalUserActionIfUnchanged(
			ctx,
			actorID,
			story.ID,
			installation.WorkspaceID,
			story.UpdatedAt,
			updates,
		); err != nil {
			return fmt.Errorf("apply Slack Work Object edit: %w", err)
		}
	}

	refreshed, err := storyService.QueryByRef(ctx, installation.WorkspaceID, link.StoryReference)
	if err != nil {
		return fmt.Errorf("reload Work Object story: %w", err)
	}
	// Recheck access after the write so a concurrent audience revocation cannot
	// disclose refreshed metadata in the flexpane.
	if err := s.ensureTeamAvailableForSlackSource(ctx, installation.WorkspaceID, actorID, refreshed.Team, source); err != nil {
		return fmt.Errorf("%w: refreshed story team is no longer available: %v", errSlackWorkObjectEditDenied, err)
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, installation.WorkspaceID, actorID, source); err != nil {
		return fmt.Errorf("%w: Slack actor is no longer linked: %v", errSlackWorkObjectEditDenied, err)
	}
	if err := s.requireCurrentSlackInstallation(ctx, installation.WorkspaceID, slackTeamID, installation.InstallGeneration); err != nil {
		return fmt.Errorf("%w: Slack installation changed: %v", errSlackWorkObjectEditDenied, err)
	}
	repository, ok := s.repo.(slackWorkObjectRepository)
	if !ok {
		return errors.New("slack Work Object repository is not configured")
	}
	input, err := buildSlackStoryWorkObjectInput(ctx, repository, actorID, slackUserID, link.CanonicalURL, refreshed, true)
	if err != nil {
		return fmt.Errorf("build refreshed Work Object: %w", err)
	}
	request, err := BuildSlackStoryEntityDetailsRequest(triggerID, input)
	if err != nil {
		return err
	}
	botToken, err := s.botToken(ctx, installation)
	if err != nil {
		return err
	}
	return newSlackWorkObjectPublisher(s.webClient).PresentDetails(ctx, botToken, request)
}

func (s *Service) authorizedSlackWorkObjectUpdates(
	ctx context.Context,
	story Story,
	state interactionViewStateValues,
) (map[string]any, error) {
	updates := make(map[string]any, 7)

	if value, present, err := slackWorkObjectStateValue(state, "title", "plain_text_input"); err != nil {
		return nil, err
	} else if present {
		title := strings.TrimSpace(value.Value)
		if title == "" || len([]rune(title)) > modalTitleMaxRunes {
			return nil, fmt.Errorf("%w: title must contain 1-%d characters", errSlackWorkObjectEditMalformed, modalTitleMaxRunes)
		}
		updates["title"] = title
	}

	if value, present, err := slackWorkObjectStateValue(state, "description", "plain_text_input"); err != nil {
		return nil, err
	} else if present {
		description := strings.TrimSpace(value.Value)
		if len([]rune(description)) > modalDescriptionMaxRunes {
			return nil, fmt.Errorf("%w: description must be %d characters or fewer", errSlackWorkObjectEditMalformed, modalDescriptionMaxRunes)
		}
		if description == "" {
			updates["description"] = nil
		} else {
			updates["description"] = description
		}
		// Slack edits raw Markdown in a plain input. Clear any stale rich-editor
		// snapshot so the canonical plain description is visible in FortyOne.
		updates["description_html"] = nil
	}

	if value, present, err := slackWorkObjectStateValue(state, "status", "static_select"); err != nil {
		return nil, err
	} else if present {
		statusID, parseErr := uuid.Parse(strings.TrimSpace(value.SelectedOption.Value))
		if parseErr != nil || statusID == uuid.Nil {
			return nil, fmt.Errorf("%w: invalid status", errSlackWorkObjectEditMalformed)
		}
		statuses, listErr := s.repo.ListTeamStatuses(ctx, story.Team)
		if listErr != nil {
			return nil, listErr
		}
		if _, found := findStatusByID(statuses, statusID); !found {
			return nil, fmt.Errorf("%w: status does not belong to the story team", errSlackWorkObjectEditDenied)
		}
		updates["status_id"] = statusID
	}

	if value, present, err := slackWorkObjectStateValue(state, "priority", "static_select"); err != nil {
		return nil, err
	} else if present {
		priority := strings.TrimSpace(value.SelectedOption.Value)
		if !isSlackStoryPriority(priority) {
			return nil, fmt.Errorf("%w: invalid priority", errSlackWorkObjectEditMalformed)
		}
		updates["priority"] = priority
	}

	if value, present, err := slackWorkObjectStateValue(state, "assignee", "static_select"); err != nil {
		return nil, err
	} else if present {
		rawAssigneeID := strings.TrimSpace(value.SelectedOption.Value)
		if rawAssigneeID == "" {
			updates["assignee_id"] = nil
		} else {
			assigneeID, parseErr := uuid.Parse(rawAssigneeID)
			if parseErr != nil || assigneeID == uuid.Nil {
				return nil, fmt.Errorf("%w: invalid assignee", errSlackWorkObjectEditMalformed)
			}
			if _, memberErr := s.repo.FindTeamMemberByID(ctx, story.Team, assigneeID); memberErr != nil {
				return nil, fmt.Errorf("%w: assignee does not belong to the story team", errSlackWorkObjectEditDenied)
			}
			updates["assignee_id"] = assigneeID
		}
	}

	if value, present, err := slackWorkObjectStateValue(state, "due_date", "datepicker"); err != nil {
		return nil, err
	} else if present {
		rawDate := strings.TrimSpace(value.SelectedDate)
		if rawDate == "" {
			updates["end_date"] = nil
		} else {
			dueDate, parseErr := time.Parse(time.DateOnly, rawDate)
			if parseErr != nil {
				return nil, fmt.Errorf("%w: due date must use YYYY-MM-DD", errSlackWorkObjectEditMalformed)
			}
			updates["end_date"] = dueDate.UTC()
		}
	}

	return updates, nil
}

func slackWorkObjectStateValue(
	state interactionViewStateValues,
	field, expectedType string,
) (interactionViewStateValue, bool, error) {
	block, present := state[field]
	if !present {
		return interactionViewStateValue{}, false, nil
	}
	if len(block) != 1 {
		return interactionViewStateValue{}, false, fmt.Errorf("%w: %s has invalid actions", errSlackWorkObjectEditMalformed, field)
	}
	value, ok := block[field+".input"]
	if !ok {
		// Slack Work Object submissions have historically used the documented
		// <field>.input action ID, but some payloads only preserve the single
		// action under a provider-generated key. Because this block is required
		// to contain exactly one action, accepting that sole action cannot widen
		// the field being updated.
		for _, candidate := range block {
			value = candidate
			ok = true
		}
	}
	if !ok || (strings.TrimSpace(value.Type) != "" && strings.TrimSpace(value.Type) != expectedType) {
		return interactionViewStateValue{}, false, fmt.Errorf("%w: %s has an invalid input", errSlackWorkObjectEditMalformed, field)
	}
	return value, true, nil
}

func isSlackStoryPriority(value string) bool {
	switch strings.TrimSpace(value) {
	case slackPriorityNoPriority, "Low", "Medium", "High", "Urgent":
		return true
	default:
		return false
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func (s *Service) presentSlackWorkObjectEditError(ctx context.Context, payload interactionPayload, message string) error {
	triggerID := strings.TrimSpace(payload.TriggerID)
	slackTeamID := strings.TrimSpace(payload.Team.ID)
	if triggerID == "" || slackTeamID == "" {
		return nil
	}
	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err != nil || !installation.IsActive {
		return err
	}
	botToken, err := s.botToken(ctx, installation)
	if err != nil {
		return err
	}
	request, err := BuildSlackStoryEntityDetailsErrorRequest(triggerID, message)
	if err != nil {
		return err
	}
	return newSlackWorkObjectPublisher(s.webClient).PresentDetails(ctx, botToken, request)
}

func slackWorkObjectEditErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrStoryChanged):
		return "This task changed while you were editing it. Refresh the task and try again."
	case errors.Is(err, errSlackWorkObjectEditDenied), errors.Is(err, ErrSlackUserNotLinked), errors.Is(err, ErrSlackTeamNotAvailable):
		return "This task is no longer available to you in this Slack context."
	case errors.Is(err, errSlackWorkObjectEditMalformed):
		return "FortyOne could not validate these changes. Refresh the task and try again."
	default:
		return "FortyOne could not save these changes. Refresh the task and try again."
	}
}
