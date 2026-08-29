package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxSlackCommandResponseDrainBytes int64 = 64 << 10

type slackCommandResponseStatusError struct {
	StatusCode int
}

func (err *slackCommandResponseStatusError) Error() string {
	return fmt.Sprintf("slack command response returned HTTP status %d", err.StatusCode)
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
		delivery, shouldSend, err := s.outbound.StartOutboundDelivery(ctx, outboundDeliveryInput{
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
		if isSlackRepositoryNotFound(err) {
			return fmt.Errorf("%w: Slack team is no longer connected", errSlackInstallationChanged)
		}
		return err
	}
	if !current.IsActive || current.WorkspaceID != workspaceID || current.InstallGeneration != generation {
		return fmt.Errorf("%w: active generation no longer matches", errSlackInstallationChanged)
	}
	return nil
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
	client := *s.client
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		// A response URL is a provider-issued bearer capability. Its response is
		// not part of the product contract, so bound the work needed for
		// connection reuse and never retain or report the provider body.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxSlackCommandResponseDrainBytes))
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		return &slackCommandResponseStatusError{StatusCode: resp.StatusCode}
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
		return errors.New("slack interaction has no private feedback destination")
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
