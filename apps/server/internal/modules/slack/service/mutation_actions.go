package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func isSlackMutationAction(payload interactionPayload) bool {
	actionID := strings.TrimSpace(payload.ActionID)
	if len(payload.Actions) > 0 {
		actionID = strings.TrimSpace(payload.Actions[0].ActionID)
	}
	return actionID == slackConfirmMutationActionID || actionID == slackCancelMutationActionID
}

func (s *Service) handleMutationAction(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if s.mutationConfirmer == nil {
		return InteractionResponse{}, errors.New("slack mutation confirmer is not configured")
	}
	if len(payload.Actions) == 0 {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	action := payload.Actions[0]
	channelID := strings.TrimSpace(payload.Channel.ID)
	messageTS := strings.TrimSpace(payload.Message.TS)
	if channelID == "" || messageTS == "" {
		return InteractionResponse{}, errors.New("slack mutation action is missing its message destination")
	}
	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, strings.TrimSpace(payload.Team.ID))
	if err != nil {
		if isSlackRepositoryNotFound(err) {
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
		_, cancelErr := s.mutationConfirmer.CancelStoryMutation(ctx, storyMutationScope{
			WorkspaceID: installation.WorkspaceID,
			UserID:      *linkedUserID,
		}, actionValue.Token)
		if cancelErr != nil {
			switch {
			case errors.Is(cancelErr, errStoryMutationApplied):
				if err := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, "This story change was already confirmed."); err != nil {
					return InteractionResponse{}, errors.Join(cancelErr, err)
				}
				return InteractionResponse{StatusCode: http.StatusOK}, nil
			case errors.Is(cancelErr, errStoryMutationExpired),
				errors.Is(cancelErr, errStoryMutationInvalid),
				errors.Is(cancelErr, errStoryMutationCancelled):
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
	result, err := s.mutationConfirmer.ConfirmStoryMutation(ctx, storyMutationScope{
		WorkspaceID:    installation.WorkspaceID,
		UserID:         *linkedUserID,
		AllowedTeamIDs: allowedTeamIDs,
		AllowMutations: true,
	}, actionValue.Token)
	if err != nil {
		if errors.Is(err, errStoryMutationNotAllowed) ||
			errors.Is(err, errStoryMutationInvalid) ||
			errors.Is(err, errStoryMutationExpired) ||
			errors.Is(err, errStoryMutationStale) ||
			errors.Is(err, errStoryMutationTeamRestricted) {
			if updateErr := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, "This story change is no longer valid. Ask Maya to try again."); updateErr != nil {
				return InteractionResponse{}, errors.Join(err, updateErr)
			}
			return InteractionResponse{StatusCode: http.StatusOK}, nil
		}
		if result.Operation == storyMutationCreateBatch {
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
			workspace, workspaceErr := s.repo.FindWorkspaceByID(ctx, installation.WorkspaceID)
			if workspaceErr != nil {
				return InteractionResponse{}, errors.Join(err, workspaceErr)
			}
			text := buildSlackStoryBatchPartialReceiptText(
				creatorName,
				s.cfg.WebsiteURL,
				workspace.Slug,
				result.Items,
			)
			retryPayload, payloadErr := BuildSlackMutationRetryProviderPayload(
				text,
				actionValue.Token,
				payload.User.ID,
			)
			if payloadErr != nil {
				return InteractionResponse{}, errors.Join(err, payloadErr)
			}
			if updateErr := s.updateSlackInteractiveMessageWithProviderPayload(
				ctx,
				botToken,
				channelID,
				messageTS,
				text,
				retryPayload,
			); updateErr != nil {
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
	text := ""
	if result.Operation == storyMutationCreateBatch {
		text = buildSlackStoryBatchMutationReceiptText(creatorName, s.cfg.WebsiteURL, workspace.Slug, result.Items)
	} else {
		storyURL := buildTaskURL(s.cfg.WebsiteURL, workspace.Slug, reference)
		text = buildSlackStoryMutationReceiptText(creatorName, reference, storyURL, result.Operation)
	}
	if err := s.updateSlackInteractiveMessage(ctx, botToken, channelID, messageTS, text); err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK}, nil
}

func buildSlackStoryBatchMutationReceiptText(
	creatorName, websiteURL, workspaceSlug string,
	items []storyMutationItemResult,
) string {
	creatorLabel := slackMrkdwnTextEscaper.Replace(strings.TrimSpace(creatorName))
	if creatorLabel == "" {
		creatorLabel = "A team member"
	}
	if len(items) == 0 {
		return creatorLabel + " created the proposed stories"
	}

	var receipt strings.Builder
	fmt.Fprintf(&receipt, "%s created %d stories:", creatorLabel, len(items))
	appendSlackStoryBatchReceiptItems(&receipt, websiteURL, workspaceSlug, items)
	return receipt.String()
}

func buildSlackStoryBatchPartialReceiptText(
	creatorName, websiteURL, workspaceSlug string,
	items []storyMutationItemResult,
) string {
	creatorLabel := slackMrkdwnTextEscaper.Replace(strings.TrimSpace(creatorName))
	if creatorLabel == "" {
		creatorLabel = "A team member"
	}
	if len(items) == 0 {
		return "FortyOne couldn't create the proposed stories because of an error. No stories were created. Select *Retry remaining* to try again."
	}

	var receipt strings.Builder
	fmt.Fprintf(&receipt, "%s created %d of the proposed stories before FortyOne hit an error:", creatorLabel, len(items))
	appendSlackStoryBatchReceiptItems(&receipt, websiteURL, workspaceSlug, items)
	receipt.WriteString("\nSelect *Retry remaining* to try again. Already-created stories will not be duplicated.")
	return receipt.String()
}

func appendSlackStoryBatchReceiptItems(
	receipt *strings.Builder,
	websiteURL, workspaceSlug string,
	items []storyMutationItemResult,
) {
	for _, item := range items {
		reference := strings.ToUpper(strings.TrimSpace(item.Reference))
		if reference == "" {
			reference = "Story"
		}
		label := reference
		if taskURL := buildTaskURL(websiteURL, workspaceSlug, reference); taskURL != "" {
			label = fmt.Sprintf("<%s|%s>", taskURL, reference)
		}
		title := slackMrkdwnTextEscaper.Replace(strings.TrimSpace(item.Title))
		if title == "" {
			fmt.Fprintf(receipt, "\n• %s", label)
		} else {
			fmt.Fprintf(receipt, "\n• %s — %s", label, title)
		}
	}
}

func buildSlackStoryMutationReceiptText(creatorName, reference, storyURL string, operation storyMutationOperation) string {
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
	if operation == storyMutationCreate {
		verb = "created"
	}
	return fmt.Sprintf("%s %s %s", creatorLabel, verb, storyLabel)
}
