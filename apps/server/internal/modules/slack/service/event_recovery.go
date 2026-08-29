package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
)

// RecoverPendingEvents republishes encrypted inbox payloads when the original
// database-to-queue handoff failed or a worker task disappeared.
func (p *EventProcessor) RecoverPendingEvents(ctx context.Context) (int, error) {
	store, ok := p.store.(recoverableSlackDeliveryStore)
	if !ok {
		return 0, errors.New("slack event store does not support outbound recovery")
	}
	if p.webhookRecovery == nil {
		return 0, errors.New("slack webhook recovery gateway is not configured")
	}
	policy := webhooks.DefaultRecoveryPolicy()
	policy.ClaimLimit = 500
	policy.ProcessingLease = slackWebhookProcessingLease
	report, webhookErr := p.webhookRecovery.Recover(ctx, slackWebhookProvider, policy)
	recovered := report.Dispatched
	var recoveryErrors []error
	if webhookErr != nil {
		recoveryErrors = append(recoveryErrors, webhookErr)
	}
	outboundRecovered, outboundErr := p.recoverPendingOutboundDeliveries(ctx, store)
	if outboundErr != nil {
		recoveryErrors = append(recoveryErrors, outboundErr)
	}
	uninstallRecovered, uninstallErr := p.recoverSlackUninstalls(ctx)
	if uninstallErr != nil {
		recoveryErrors = append(recoveryErrors, uninstallErr)
	}
	return recovered + outboundRecovered + uninstallRecovered, errors.Join(recoveryErrors...)
}

func (p *EventProcessor) recoverSlackUninstalls(ctx context.Context) (int, error) {
	records, err := p.repo.ClaimRecoverableSlackUninstalls(ctx, 100)
	if err != nil {
		return 0, err
	}
	recovered := 0
	var recoveryErrors []error
	for _, record := range records {
		attemptCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		completed, uninstallErr := executeSlackUninstall(
			attemptCtx,
			p.repo,
			p.repo,
			p.webClient,
			p.codec,
			p.clientID,
			p.clientSecret,
			p.clock.Now(),
			record,
		)
		cancel()
		if uninstallErr != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recover Slack uninstall %s: %w", record.ID, uninstallErr))
			continue
		}
		if completed {
			recovered++
		}
	}
	return recovered, errors.Join(recoveryErrors...)
}

func (p *EventProcessor) recoverPendingOutboundDeliveries(ctx context.Context, store recoverableSlackDeliveryStore) (int, error) {
	records, err := store.ListRecoverableOutboundDeliveries(ctx, "slack", 500)
	if err != nil {
		return 0, err
	}
	recovered := 0
	var recoveryErrors []error
	for _, record := range records {
		externalWorkspaceID := strings.TrimSpace(record.ExternalWorkspaceID)
		if externalWorkspaceID == "" {
			err := errors.New("slack outbound delivery has no external workspace binding")
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, record.ID, err.Error()); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel malformed Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recover Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		threadID := ""
		if record.ExternalThreadID != nil {
			threadID = strings.TrimSpace(*record.ExternalThreadID)
		}
		delivery, claimed, err := p.store.StartOutboundDelivery(ctx, outboundDeliveryInput{
			Provider:                "slack",
			WorkspaceID:             record.WorkspaceID,
			UserID:                  record.UserID,
			InstallGeneration:       record.InstallGeneration,
			ExternalWorkspaceID:     externalWorkspaceID,
			ExternalRecipientUserID: valueOrEmpty(record.ExternalRecipientUserID),
			InboundEventID:          record.InboundEventID,
			IdempotencyKey:          record.IdempotencyKey,
			ExternalChannelID:       record.ExternalChannelID,
			ExternalThreadID:        threadID,
			Content:                 valueOrEmpty(record.Content),
			ProviderPayload:         append([]byte(nil), record.ProviderPayload...),
			Purpose:                 record.Purpose,
			ExpiresAt:               record.ExpiresAt,
		})
		if err != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf("claim Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if !claimed {
			continue
		}
		if delivery.ExpiresAt != nil && !p.clock.Now().UTC().Before(delivery.ExpiresAt.UTC()) {
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack delivery expired before recovery"); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel expired Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
			}
			continue
		}
		content := ""
		if delivery.Content != nil {
			content = strings.TrimSpace(*delivery.Content)
		}
		if content == "" && record.Content != nil {
			content = strings.TrimSpace(*record.Content)
		}
		if content == "" {
			err := errors.New("slack outbound delivery has no content")
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, err.Error()); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel empty Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recover Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		providerPayload, payloadErr := DecodeSlackProviderPayload(delivery.ProviderPayload)
		if payloadErr != nil {
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack outbound delivery has an invalid provider payload"); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel invalid Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recover Slack delivery %s provider payload: %w", record.IdempotencyKey, payloadErr))
			continue
		}
		installation, err := p.repo.GetSlackWorkspaceByTeamID(ctx, externalWorkspaceID)
		if err != nil {
			if isSlackRepositoryNotFound(err) {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack installation no longer exists"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel disconnected Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after installation lookup: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("load Slack installation for delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if err := validateOutboundInstallation(record, installation); err != nil {
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack installation binding changed before recovery"); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel Slack delivery %s after installation change: %w", record.IdempotencyKey, cancelErr))
			}
			continue
		}
		if delivery.Purpose == "assistant" || providerPayload.Authorization != nil {
			if delivery.UserID == nil {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack delivery is missing its actor binding"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel unbound Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
			if providerPayload.Authorization != nil && providerPayload.Authorization.ActorUserID != nil && *providerPayload.Authorization.ActorUserID != *delivery.UserID {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack delivery actor binding does not match its authorization payload"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel mismatched Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
			requireLinkedActor := delivery.Purpose == "assistant"
			if requireLinkedActor {
				if delivery.ExternalRecipientUserID == nil || strings.TrimSpace(*delivery.ExternalRecipientUserID) == "" {
					if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack delivery is missing its recipient binding"); cancelErr != nil {
						recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel unbound Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
					}
					continue
				}
				linkedUserID, linkErr := p.repo.FindLinkedUserIDBySlackUser(
					ctx,
					installation.WorkspaceID,
					externalWorkspaceID,
					strings.TrimSpace(*delivery.ExternalRecipientUserID),
				)
				if linkErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("revalidate Slack delivery %s actor: %w", record.IdempotencyKey, linkErr))
					if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(linkErr)); failErr != nil {
						recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after actor lookup: %w", record.IdempotencyKey, failErr))
					}
					continue
				}
				if linkedUserID == nil || *linkedUserID != *delivery.UserID {
					if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack delivery actor is no longer linked or active"); cancelErr != nil {
						recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel unauthorized Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
					}
					continue
				}
			}
			current, audienceErr := p.slackChannelDeliveryAuthorizationCurrent(
				ctx,
				installation.WorkspaceID,
				installation,
				*delivery.UserID,
				delivery.ExternalChannelID,
				valueOrEmpty(delivery.ExternalRecipientUserID),
				providerPayload,
			)
			if audienceErr != nil {
				if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(audienceErr)); failErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after audience lookup: %w", record.IdempotencyKey, failErr))
				}
				recoveryErrors = append(recoveryErrors, fmt.Errorf("revalidate Slack delivery %s channel audience: %w", record.IdempotencyKey, audienceErr))
				continue
			}
			if !current {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack channel audience narrowed before delivery recovery"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel narrowed Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
		}
		if delivery.Purpose == "assistant" {
			currentSettings, settingsErr := p.agentSettings(ctx, installation.WorkspaceID)
			if settingsErr != nil {
				if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(settingsErr)); failErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after settings lookup: %w", record.IdempotencyKey, failErr))
				}
				recoveryErrors = append(recoveryErrors, fmt.Errorf("revalidate Slack delivery %s agent settings: %w", record.IdempotencyKey, settingsErr))
				continue
			}
			if !assistantSettingsAllowDelivery(currentSettings, providerPayload) {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack agent settings changed before assistant delivery recovery"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel disabled Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
		}
		botToken, err := p.botToken(ctx, installation)
		if err != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after credential lookup: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("load Slack credential for delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if err := p.requireCurrentInstallation(ctx, record.WorkspaceID, externalWorkspaceID, *record.InstallGeneration); err != nil {
			if errors.Is(err, errSlackInstallationChanged) {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack installation changed before recovered delivery"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel stale Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after installation recheck: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recheck Slack installation for delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		providerIdempotencyKey := record.IdempotencyKey
		if delivery.Purpose == slackOnboardingPurpose {
			providerIdempotencyKey = slackFirstInteractionGuideProviderKey(
				record.WorkspaceID,
				externalWorkspaceID,
				valueOrEmpty(delivery.ExternalRecipientUserID),
			)
		}
		externalMessageID, err := p.sender.Send(ctx, botToken, SlackOutboundMessage{
			ChannelID:        record.ExternalChannelID,
			ThreadTS:         threadID,
			Text:             content,
			ClientMessageID:  deterministicSlackMessageID(providerIdempotencyKey),
			StandardMarkdown: delivery.Purpose == "assistant" && len(providerPayload.Blocks) == 0,
			ProviderPayload:  providerPayload,
		})
		if err != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after send failure: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("send Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if err := bindSlackRequestThreadContinuation(ctx, p.threadSync, record.WorkspaceID, installation.InstallGeneration, externalWorkspaceID, record.ExternalChannelID, threadID, externalMessageID, providerPayload); err != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after thread binding: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("bind Slack request thread for delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if err := p.store.CompleteOutboundDelivery(ctx, delivery.ID, externalMessageID); err != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf("complete Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		recovered++
	}
	return recovered, errors.Join(recoveryErrors...)
}
