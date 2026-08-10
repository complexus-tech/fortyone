package slack

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
)

const (
	slackOnboardingPurpose                   = "onboarding"
	slackFirstInteractionGuidePrepareTimeout = 500 * time.Millisecond
	slackFirstInteractionGuideRetryTimeout   = 15 * time.Second
	slackFirstInteractionGuideSendTimeout    = 15 * time.Second
	slackFirstInteractionGuideKeyPrefix      = "slack-onboarding"
)

type preparedSlackFirstInteractionGuide struct {
	installation   slackrepository.SlackWorkspaceRecord
	slackUserID    string
	text           string
	idempotencyKey string
	providerKey    string
	delivery       messagingrepository.OutboundDeliveryRecord
}

func slackEventCountsAsFirstInteraction(kind slackEventKind) bool {
	switch kind {
	case slackEventKindMention,
		slackEventKindDirect,
		slackEventKindChannelThread,
		slackEventKindLinkShared,
		slackEventKindEntityDetails:
		return true
	default:
		return false
	}
}

func (s *Service) dispatchFirstInteractionGuide(
	parent context.Context,
	installation slackrepository.SlackWorkspaceRecord,
	slackUserID string,
) {
	if !s.slackFirstInteractionGuideConfigured() || installation.WorkspaceID == uuid.Nil || strings.TrimSpace(slackUserID) == "" {
		return
	}
	prepareCtx, cancel := context.WithTimeout(parent, slackFirstInteractionGuidePrepareTimeout)
	prepared, shouldSend, err := s.prepareFirstInteractionGuide(prepareCtx, installation, slackUserID)
	cancel()
	if err != nil {
		baseCtx := context.WithoutCancel(parent)
		s.logFirstInteractionGuideError(baseCtx, installation, slackUserID, err)
		s.retryFirstInteractionGuideAsync(baseCtx, installation, slackUserID)
		return
	}
	if shouldSend {
		s.sendPreparedFirstInteractionGuideAsync(parent, prepared)
	}
}

func (s *Service) dispatchFirstInteractionGuideByTeam(parent context.Context, slackTeamID, slackUserID string) {
	if !s.slackFirstInteractionGuideConfigured() || strings.TrimSpace(slackTeamID) == "" || strings.TrimSpace(slackUserID) == "" {
		return
	}
	prepareCtx, cancel := context.WithTimeout(parent, slackFirstInteractionGuidePrepareTimeout)
	installation, err := s.repo.GetSlackWorkspaceByTeamID(prepareCtx, strings.TrimSpace(slackTeamID))
	if err != nil {
		cancel()
		if !slackrepository.IsNotFound(err) {
			baseCtx := context.WithoutCancel(parent)
			s.logFirstInteractionGuideError(baseCtx, installation, slackUserID, err)
			s.retryFirstInteractionGuideByTeamAsync(baseCtx, slackTeamID, slackUserID)
		}
		return
	}
	prepared, shouldSend, err := s.prepareFirstInteractionGuide(prepareCtx, installation, slackUserID)
	cancel()
	if err != nil {
		baseCtx := context.WithoutCancel(parent)
		s.logFirstInteractionGuideError(baseCtx, installation, slackUserID, err)
		s.retryFirstInteractionGuideAsync(baseCtx, installation, slackUserID)
		return
	}
	if shouldSend {
		s.sendPreparedFirstInteractionGuideAsync(parent, prepared)
	}
}

// The bounded synchronous attempt normally persists the outbox claim before
// Slack is acknowledged. If a transient database stall exhausts that budget,
// continue once outside the request lifetime; the same durable idempotency key
// makes an ambiguous first attempt safe to retry.
func (s *Service) retryFirstInteractionGuideAsync(
	parent context.Context,
	installation slackrepository.SlackWorkspaceRecord,
	slackUserID string,
) {
	go func() {
		ctx, cancel := context.WithTimeout(parent, slackFirstInteractionGuideRetryTimeout)
		defer cancel()
		prepared, shouldSend, err := s.prepareFirstInteractionGuide(ctx, installation, slackUserID)
		if err != nil {
			s.logFirstInteractionGuideError(parent, installation, slackUserID, err)
			return
		}
		if shouldSend {
			if err := s.sendPreparedFirstInteractionGuide(ctx, prepared); err != nil {
				s.logFirstInteractionGuideError(parent, installation, slackUserID, err)
			}
		}
	}()
}

func (s *Service) retryFirstInteractionGuideByTeamAsync(parent context.Context, slackTeamID, slackUserID string) {
	go func() {
		ctx, cancel := context.WithTimeout(parent, slackFirstInteractionGuideRetryTimeout)
		defer cancel()
		installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, strings.TrimSpace(slackTeamID))
		if err != nil {
			if !slackrepository.IsNotFound(err) {
				s.logFirstInteractionGuideError(parent, installation, slackUserID, err)
			}
			return
		}
		prepared, shouldSend, err := s.prepareFirstInteractionGuide(ctx, installation, slackUserID)
		if err != nil {
			s.logFirstInteractionGuideError(parent, installation, slackUserID, err)
			return
		}
		if shouldSend {
			if err := s.sendPreparedFirstInteractionGuide(ctx, prepared); err != nil {
				s.logFirstInteractionGuideError(parent, installation, slackUserID, err)
			}
		}
	}()
}

func (s *Service) slackFirstInteractionGuideConfigured() bool {
	return s != nil && s.outbound != nil && s.repo != nil
}

func (s *Service) prepareFirstInteractionGuide(
	ctx context.Context,
	installation slackrepository.SlackWorkspaceRecord,
	slackUserID string,
) (preparedSlackFirstInteractionGuide, bool, error) {
	if s.outbound == nil {
		return preparedSlackFirstInteractionGuide{}, false, nil
	}
	slackTeamID := strings.TrimSpace(installation.SlackTeamID)
	slackUserID = strings.TrimSpace(slackUserID)
	if installation.WorkspaceID == uuid.Nil || installation.InstallGeneration == uuid.Nil || slackTeamID == "" || slackUserID == "" || !installation.IsActive {
		return preparedSlackFirstInteractionGuide{}, false, nil
	}
	delivered, err := s.repo.HasSlackUserOnboardingReceipt(ctx, installation.WorkspaceID, slackTeamID, slackUserID)
	if err != nil || delivered {
		return preparedSlackFirstInteractionGuide{}, false, err
	}

	text := buildSlackFirstInteractionGuide(valueOrEmpty(installation.BotUserID))
	identityKey := slackFirstInteractionGuideIdentityKey(slackTeamID, slackUserID)
	idempotencyKey := fmt.Sprintf(
		"%s:%s:%s",
		slackFirstInteractionGuideKeyPrefix,
		installation.InstallGeneration,
		identityKey,
	)
	providerKey := slackFirstInteractionGuideProviderKey(installation.WorkspaceID, slackTeamID, slackUserID)
	delivery, shouldSend, err := s.outbound.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
		Provider:                slackProviderMessaging,
		WorkspaceID:             installation.WorkspaceID,
		InstallGeneration:       &installation.InstallGeneration,
		ExternalWorkspaceID:     slackTeamID,
		ExternalRecipientUserID: slackUserID,
		IdempotencyKey:          idempotencyKey,
		ExternalChannelID:       slackUserID,
		Content:                 text,
		Purpose:                 slackOnboardingPurpose,
	})
	if errors.Is(err, messagingrepository.ErrLeaseBusy) {
		return preparedSlackFirstInteractionGuide{}, false, nil
	}
	if err != nil || !shouldSend {
		return preparedSlackFirstInteractionGuide{}, false, err
	}
	return preparedSlackFirstInteractionGuide{
		installation:   installation,
		slackUserID:    slackUserID,
		text:           text,
		idempotencyKey: idempotencyKey,
		providerKey:    providerKey,
		delivery:       delivery,
	}, true, nil
}

func slackFirstInteractionGuideIdentityKey(slackTeamID, slackUserID string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(slackTeamID) + "\x1f" + strings.TrimSpace(slackUserID)))
	return hex.EncodeToString(digest[:])
}

func slackFirstInteractionGuideProviderKey(workspaceID uuid.UUID, slackTeamID, slackUserID string) string {
	return fmt.Sprintf(
		"%s:%s:%s",
		slackFirstInteractionGuideKeyPrefix,
		workspaceID,
		slackFirstInteractionGuideIdentityKey(slackTeamID, slackUserID),
	)
}

func (s *Service) sendPreparedFirstInteractionGuideAsync(parent context.Context, prepared preparedSlackFirstInteractionGuide) {
	baseCtx := context.WithoutCancel(parent)
	go func() {
		ctx, cancel := context.WithTimeout(baseCtx, slackFirstInteractionGuideSendTimeout)
		defer cancel()
		if err := s.sendPreparedFirstInteractionGuide(ctx, prepared); err != nil {
			s.logFirstInteractionGuideError(baseCtx, prepared.installation, prepared.slackUserID, err)
		}
	}()
}

func (s *Service) sendPreparedFirstInteractionGuide(ctx context.Context, prepared preparedSlackFirstInteractionGuide) error {
	botToken, err := s.botToken(ctx, prepared.installation)
	if err != nil {
		if failErr := failOutboundDeliveryDetached(ctx, s.outbound, prepared.delivery.ID, truncateError(err)); failErr != nil {
			return errors.Join(err, failErr)
		}
		return err
	}
	text := prepared.text
	delivery := prepared.delivery
	if err := s.outbound.SetOutboundDeliveryContent(ctx, delivery.ID, text); err != nil {
		if failErr := failOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, truncateError(err)); failErr != nil {
			return errors.Join(err, failErr)
		}
		return err
	}
	installation := prepared.installation
	if err := s.requireCurrentSlackInstallation(ctx, installation.WorkspaceID, installation.SlackTeamID, installation.InstallGeneration); err != nil {
		if errors.Is(err, errSlackInstallationChanged) {
			return cancelOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, "Slack installation changed before first-use guide delivery")
		}
		if failErr := failOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, truncateError(err)); failErr != nil {
			return errors.Join(err, failErr)
		}
		return err
	}

	externalMessageID, err := (&slackAPISender{client: s.slackClient()}).Send(ctx, botToken, SlackOutboundMessage{
		ChannelID: prepared.slackUserID,
		UserID:    prepared.slackUserID,
		Text:      truncateSlackText(text),
		// Keep Slack's provider-side idempotency identity stable across install
		// generations. This closes the narrow handoff where an old token accepts
		// the post while a newly installed generation starts its own delivery.
		ClientMessageID: deterministicSlackMessageID(prepared.providerKey),
	})
	if err != nil {
		if failErr := failOutboundDeliveryDetached(ctx, s.outbound, delivery.ID, truncateError(err)); failErr != nil {
			return errors.Join(err, failErr)
		}
		return err
	}
	return s.outbound.CompleteOutboundDelivery(ctx, delivery.ID, externalMessageID)
}

func (s *Service) logFirstInteractionGuideError(
	ctx context.Context,
	installation slackrepository.SlackWorkspaceRecord,
	slackUserID string,
	err error,
) {
	if s.log != nil {
		s.log.Warn(ctx, "failed sending first-use Slack guide",
			"error", err,
			"workspace_id", installation.WorkspaceID,
			"slack_team_id", strings.TrimSpace(installation.SlackTeamID),
			"slack_user_id", strings.TrimSpace(slackUserID),
		)
	}
}

func buildSlackFirstInteractionGuide(botUserID string) string {
	mention := "mention the FortyOne app"
	if botUserID = strings.TrimSpace(botUserID); botUserID != "" {
		mention = "mention <@" + botUserID + ">"
	}
	return fmt.Sprintf(`*Welcome to FortyOne in Slack*

I'm Maya, your FortyOne work assistant. Here's how to get started:

• *Ask about your work* — when Maya is enabled, DM me here or %s in a channel. Try “What am I working on?”, “Find onboarding work”, or “What's the status of WEB-546?”
• *Create a story or request* — use %s, or choose *Create a story* from any Slack message.
• *Make changes with Maya* — when workflow actions are enabled, ask me to create or update a story. I'll show the proposed change and wait for you to confirm before anything is updated.
• *Share work* — paste a FortyOne story or request link for a permission-aware preview.
• *Keep context together* — continue in the thread after I reply. Replies in a Request thread created by FortyOne also sync back to the Request.

I only use FortyOne work you're allowed to access, including team limits configured for the Slack channel. If your Slack account isn't linked yet, I'll ask you to connect it first. Maya availability and workflow actions depend on your plan and workspace settings.`, mention, "`/fortyone [title]`")
}
