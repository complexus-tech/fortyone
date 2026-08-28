package slack

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

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
		s.dispatchFirstInteractionGuideByTeam(ctx, payload.Team.ID, payload.User.ID)
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	case "view_submission":
		if isSlackWorkObjectEditSubmission(payload) {
			s.dispatchSlackWorkObjectEdit(ctx, payload)
			s.dispatchFirstInteractionGuideByTeam(ctx, payload.Team.ID, payload.User.ID)
			return InteractionResponse{StatusCode: http.StatusOK}, nil
		}
		if isSlackStoryCompactEditSubmission(payload) {
			s.dispatchInteraction(ctx, payload.Type, payload, s.handleSlackStoryCompactEditSubmission)
			s.dispatchFirstInteractionGuideByTeam(ctx, payload.Team.ID, payload.User.ID)
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
		s.dispatchFirstInteractionGuideByTeam(ctx, payload.Team.ID, payload.User.ID)
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	case "block_suggestion":
		return s.handleBlockSuggestion(ctx, payload)
	default:
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
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
