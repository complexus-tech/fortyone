package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

const legacyWebhookPayloadBatchSize = 100

type legacyWebhookPayloadRepository interface {
	ListLegacySlackWebhookPayloads(
		ctx context.Context,
		afterID uuid.UUID,
		limit int,
	) ([]webhooks.Record, error)
	UpgradeLegacySlackWebhookPayload(
		ctx context.Context,
		record webhooks.Record,
		previousPayload, replacementPayload string,
	) error
}

// BackfillLegacyWebhookPayloads performs the one-way key-domain cutover for
// retryable Slack receipts. Rows are scanned in stable ID pages, authenticated
// against their stored provider identity, resealed with the dedicated key, and
// replaced by an exact compare-and-swap. Normal processing never calls this
// legacy decoder and therefore fails closed until the cutover succeeds.
func (p *EventProcessor) BackfillLegacyWebhookPayloads(ctx context.Context, cutover *LegacyCutover) (int, error) {
	repository, ok := p.repo.(legacyWebhookPayloadRepository)
	if !ok {
		return 0, errors.New("slack repository does not support webhook payload cutover")
	}
	if cutover == nil {
		return 0, errors.New("slack legacy webhook cutover is not configured")
	}
	if p.webhookPayloads == nil {
		return 0, errors.New("slack webhook payload encryption is not configured")
	}

	updated := 0
	afterID := uuid.Nil
	for {
		records, err := repository.ListLegacySlackWebhookPayloads(ctx, afterID, legacyWebhookPayloadBatchSize)
		if err != nil {
			return updated, fmt.Errorf("list legacy Slack webhook payloads: %w", err)
		}
		for _, record := range records {
			if err := validateSlackLegacyCutoverRecord(record, afterID); err != nil {
				return updated, err
			}
			previous := strings.TrimSpace(*record.EncryptedPayload)
			body, err := cutover.openWebhookPayload(record, previous)
			if err != nil {
				return updated, fmt.Errorf("open legacy Slack webhook payload %s: %w", record.ID, err)
			}
			replacement, err := p.webhookPayloads.Seal(ctx, slackWebhookPayloadBinding(record), body)
			clear(body)
			if err != nil {
				return updated, fmt.Errorf("seal legacy Slack webhook payload %s: %w", record.ID, err)
			}
			if err := repository.UpgradeLegacySlackWebhookPayload(ctx, record, previous, replacement); err != nil {
				if !isSlackRepositoryNotFound(err) {
					return updated, fmt.Errorf("upgrade legacy Slack webhook payload %s: %w", record.ID, err)
				}
			} else {
				updated++
			}
			afterID = record.ID
		}
		if len(records) < legacyWebhookPayloadBatchSize {
			return updated, nil
		}
	}
}

func validateSlackLegacyCutoverRecord(record webhooks.Record, afterID uuid.UUID) error {
	if record.ID == uuid.Nil || record.Provider != slackWebhookProvider ||
		record.DeliveryID == "" || record.ExternalAccountID == "" ||
		record.WorkspaceID == uuid.Nil || record.InstallationID == uuid.Nil ||
		record.InstallationGeneration == uuid.Nil || record.EncryptedPayload == nil ||
		strings.TrimSpace(*record.EncryptedPayload) == "" ||
		strings.HasPrefix(strings.TrimSpace(*record.EncryptedPayload), slackWebhookPayloadEnvelopePrefix) {
		return fmt.Errorf("legacy Slack webhook payload %s has invalid durable identity", record.ID)
	}
	if afterID != uuid.Nil && strings.Compare(record.ID.String(), afterID.String()) <= 0 {
		return errors.New("legacy Slack webhook payload page is not in stable ID order")
	}
	return nil
}
