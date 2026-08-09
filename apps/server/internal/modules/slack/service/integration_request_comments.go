package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
)

func (s *Service) PrepareIntegrationRequestComment(
	_ context.Context,
	request integrationrequests.CoreIntegrationRequest,
	thread integrationrequests.CoreProviderThread,
	input integrationrequests.CoreCreateCommentInput,
) (integrationrequests.CorePreparedProviderComment, error) {
	if thread.Provider != integrationrequests.ProviderSlack || request.Provider != integrationrequests.ProviderSlack {
		return integrationrequests.CorePreparedProviderComment{}, fmt.Errorf("unsupported Slack comment provider binding")
	}
	if request.WorkspaceID != thread.WorkspaceID || request.ID != thread.IntegrationRequestID || request.TeamID != thread.TeamID {
		return integrationrequests.CorePreparedProviderComment{}, errors.New("Slack comment request and thread binding do not match")
	}
	if input.AuthorID == uuid.Nil {
		return integrationrequests.CorePreparedProviderComment{}, errors.New("Slack comment author is required")
	}

	externalRecipientUserID := ""
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(thread.ExternalChannelID)), "D") {
		externalRecipientUserID = metadataString(request.Metadata, "slack_user_id")
		if externalRecipientUserID == "" {
			return integrationrequests.CorePreparedProviderComment{}, errors.New("Slack direct-message request has no bound recipient")
		}
	}
	payload, err := EncodeSlackProviderPayload(SlackProviderPayload{
		Authorization: &SlackDeliveryAuthorization{
			AllowedTeamIDs: []uuid.UUID{request.TeamID},
			ActorUserID:    &input.AuthorID,
		},
	})
	if err != nil {
		return integrationrequests.CorePreparedProviderComment{}, err
	}
	return integrationrequests.CorePreparedProviderComment{
		ExternalRecipientUserID: externalRecipientUserID,
		ProviderPayload:         payload,
	}, nil
}

func (s *Service) DeliverIntegrationRequestComment(
	ctx context.Context,
	request integrationrequests.CoreIntegrationRequest,
	thread integrationrequests.CoreProviderThread,
	comment integrationrequests.CoreIntegrationRequestComment,
	prepared integrationrequests.CorePreparedProviderComment,
) error {
	if thread.Provider != integrationrequests.ProviderSlack || request.Provider != integrationrequests.ProviderSlack {
		return nil
	}
	if s.outbound == nil {
		return errors.New("Slack outbound delivery store is not configured")
	}
	if comment.OutboundIdempotencyKey == nil || strings.TrimSpace(*comment.OutboundIdempotencyKey) == "" {
		return errors.New("Slack integration request comment has no idempotency key")
	}
	if comment.AuthorUserID == nil || *comment.AuthorUserID == uuid.Nil {
		return errors.New("Slack integration request comment has no internal author")
	}
	providerPayload, err := DecodeSlackProviderPayload(prepared.ProviderPayload)
	if err != nil {
		return err
	}
	if err := validateIntegrationRequestCommentAuthorization(providerPayload, request.TeamID, *comment.AuthorUserID); err != nil {
		return err
	}
	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, thread.ExternalWorkspaceID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return errors.New("Slack installation is no longer active")
		}
		return err
	}
	if installation.WorkspaceID != request.WorkspaceID || thread.InstallationGeneration == nil || *thread.InstallationGeneration != installation.InstallGeneration {
		return errSlackInstallationChanged
	}
	botToken, err := s.botToken(ctx, installation)
	if err != nil {
		return err
	}
	idempotencyKey := strings.TrimSpace(*comment.OutboundIdempotencyKey)
	delivery, shouldSend, err := s.outbound.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
		Provider:                integrationrequests.ProviderSlack,
		WorkspaceID:             request.WorkspaceID,
		UserID:                  comment.AuthorUserID,
		InstallGeneration:       &installation.InstallGeneration,
		ExternalWorkspaceID:     installation.SlackTeamID,
		ExternalRecipientUserID: strings.TrimSpace(prepared.ExternalRecipientUserID),
		IdempotencyKey:          idempotencyKey,
		ExternalChannelID:       thread.ExternalChannelID,
		ExternalThreadID:        thread.ExternalThreadID,
		Content:                 comment.Body,
		ProviderPayload:         append([]byte(nil), prepared.ProviderPayload...),
		Purpose:                 "provider_message",
	})
	if err != nil {
		return err
	}
	if !shouldSend {
		return nil
	}
	providerPayload, err = DecodeSlackProviderPayload(delivery.ProviderPayload)
	if err != nil {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, "Slack comment delivery has an invalid durable provider payload"); cancelErr != nil {
			return errors.Join(err, cancelErr)
		}
		return err
	}
	if err := validateIntegrationRequestCommentAuthorization(providerPayload, request.TeamID, *comment.AuthorUserID); err != nil {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, "Slack comment delivery authorization no longer matches its request"); cancelErr != nil {
			return errors.Join(err, cancelErr)
		}
		return err
	}
	deliveryChannelID := strings.TrimSpace(delivery.ExternalChannelID)
	deliveryThreadID := strings.TrimSpace(valueOrEmpty(delivery.ExternalThreadID))
	deliveryRecipientUserID := strings.TrimSpace(valueOrEmpty(delivery.ExternalRecipientUserID))
	deliveryContent := strings.TrimSpace(valueOrEmpty(delivery.Content))
	if deliveryChannelID == "" || deliveryContent == "" {
		err := errors.New("Slack comment delivery is missing its durable destination or content")
		if cancelErr := cancelOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, err.Error()); cancelErr != nil {
			return errors.Join(err, cancelErr)
		}
		return err
	}
	current, err := s.slackDeliveryAuthorizationCurrent(
		ctx,
		request.WorkspaceID,
		delivery.ExternalWorkspaceID,
		deliveryChannelID,
		deliveryRecipientUserID,
		providerPayload,
	)
	if err != nil {
		if failErr := failOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, truncateError(err)); failErr != nil {
			return errors.Join(err, fmt.Errorf("release Slack comment delivery after audience lookup: %w", failErr))
		}
		return err
	}
	if !current {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, "Slack comment actor or channel audience changed"); cancelErr != nil {
			return cancelErr
		}
		return nil
	}
	if err := s.requireCurrentSlackInstallation(ctx, request.WorkspaceID, installation.SlackTeamID, installation.InstallGeneration); err != nil {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, "Slack installation changed before comment delivery"); cancelErr != nil {
			return errors.Join(err, cancelErr)
		}
		return err
	}
	externalMessageID, err := (&slackAPISender{client: s.slackClient()}).Send(ctx, botToken, SlackOutboundMessage{
		ChannelID:       deliveryChannelID,
		UserID:          deliveryRecipientUserID,
		ThreadTS:        deliveryThreadID,
		Text:            truncateSlackText(deliveryContent),
		ClientMessageID: deterministicSlackMessageID(idempotencyKey),
		ProviderPayload: providerPayload,
	})
	if err != nil {
		if failErr := failOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, truncateError(err)); failErr != nil {
			return errors.Join(err, fmt.Errorf("release Slack comment delivery: %w", failErr))
		}
		return err
	}
	if err := s.outbound.CompleteOutboundDelivery(ctx, delivery.ID, externalMessageID); err != nil {
		return err
	}
	return nil
}

func validateIntegrationRequestCommentAuthorization(payload SlackProviderPayload, teamID, authorID uuid.UUID) error {
	authorization := payload.Authorization
	if authorization == nil || authorization.ActorUserID == nil || *authorization.ActorUserID != authorID || len(authorization.AllowedTeamIDs) != 1 || authorization.AllowedTeamIDs[0] != teamID {
		return errors.New("Slack integration request comment authorization does not match its request")
	}
	return nil
}

var _ integrationrequests.ProviderCommenter = (*Service)(nil)
