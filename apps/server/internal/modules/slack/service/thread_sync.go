package slack

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/google/uuid"
)

type SlackThreadSync interface {
	IngestInboundProviderComment(ctx context.Context, input InboundProviderCommentInput) (handled bool, err error)
	BindProviderThread(ctx context.Context, input BindProviderThreadInput) (ProviderThread, error)
}

func bindSlackRequestThreadContinuation(
	ctx context.Context,
	binder interface {
		BindProviderThread(context.Context, BindProviderThreadInput) (ProviderThread, error)
	},
	workspaceID uuid.UUID,
	installGeneration uuid.UUID,
	externalWorkspaceID string,
	externalChannelID string,
	externalThreadID string,
	externalMessageID string,
	payload SlackProviderPayload,
) error {
	binding := payload.RequestThreadBinding
	if binding == nil {
		return nil
	}
	if binder == nil {
		return errors.New("slack request thread binder is not configured")
	}
	canonicalThreadID := strings.TrimSpace(externalThreadID)
	if canonicalThreadID == "" {
		canonicalThreadID = strings.TrimSpace(externalMessageID)
	}
	if workspaceID == uuid.Nil || installGeneration == uuid.Nil || strings.TrimSpace(externalWorkspaceID) == "" || strings.TrimSpace(externalChannelID) == "" || canonicalThreadID == "" {
		return errors.New("slack request thread continuation is missing its installation or destination")
	}

	var sourceMessageID *string
	if value := strings.TrimSpace(binding.ExternalSourceMessageID); value != "" {
		sourceMessageID = &value
	}
	var sourceURL *string
	if value := strings.TrimSpace(binding.SourceURL); value != "" {
		sourceURL = &value
	}
	_, err := binder.BindProviderThread(ctx, BindProviderThreadInput{
		WorkspaceID:             workspaceID,
		IntegrationRequestID:    binding.IntegrationRequestID,
		Provider:                ProviderSlack,
		ExternalWorkspaceID:     strings.TrimSpace(externalWorkspaceID),
		InstallationGeneration:  &installGeneration,
		ExternalChannelID:       strings.TrimSpace(externalChannelID),
		ExternalThreadID:        canonicalThreadID,
		ExternalSourceMessageID: sourceMessageID,
		SourceURL:               sourceURL,
	})
	return err
}

func (p *EventProcessor) syncIntegrationRequestThreadReply(
	ctx context.Context,
	installation slackdomain.Installation,
	linkedUserID *uuid.UUID,
	event normalizedSlackEvent,
) (bool, error) {
	threadTS := strings.TrimSpace(event.ThreadTS)
	if threadTS == "" {
		return false, nil
	}
	return p.threadSync.IngestInboundProviderComment(ctx, InboundProviderCommentInput{
		Provider:               ProviderSlack,
		ExternalWorkspaceID:    installation.SlackTeamID,
		InstallationGeneration: installation.InstallGeneration,
		ExternalChannelID:      event.ChannelID,
		ExternalThreadID:       threadTS,
		ExternalMessageID:      event.MessageTS,
		ExternalAuthorID:       event.UserID,
		AuthorUserID:           linkedUserID,
		Body:                   event.Text,
		CreatedAt:              slackMessageTime(event.MessageTS),
	})
}

func slackMessageTime(messageTS string) time.Time {
	secondsText, _, _ := strings.Cut(strings.TrimSpace(messageTS), ".")
	seconds, err := strconv.ParseInt(secondsText, 10, 64)
	if err != nil || seconds <= 0 {
		return time.Time{}
	}
	return time.Unix(seconds, 0).UTC()
}
