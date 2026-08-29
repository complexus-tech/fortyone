package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	messagingAliasRetention    = 24 * time.Hour
	messagingProviderRetention = 30 * 24 * time.Hour
)

// MessagingDataPurger is the worker-owned persistence capability for bounded,
// atomic messaging retention batches.
type MessagingDataPurger interface {
	PurgeMessagingDataBatch(
		context.Context,
		messagingdomain.RetentionCutoffs,
		int,
	) (messagingdomain.RetentionPurgeResult, error)
}

// PurgeMessagingData removes expired nonces, redacts expired mutation
// proposals, and deletes provider message data after its operational retention
// window. Durable Maya email threads and messages are preserved so later
// replies retain the complete conversation; only their expired or revoked
// reply-address aliases are removed here.
func PurgeMessagingData(ctx context.Context, store MessagingDataPurger, log *logger.Logger) error {
	return purgeMessagingDataAt(ctx, store, log, time.Now().UTC())
}

func purgeMessagingDataAt(
	ctx context.Context,
	store MessagingDataPurger,
	log *logger.Logger,
	now time.Time,
) error {
	if ctx == nil {
		return errors.New("messaging retention context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.PurgeMessagingData")
	defer span.End()
	if store == nil {
		return errors.New("messaging retention store is required")
	}
	if log == nil {
		return errors.New("messaging retention logger is required")
	}
	if now.IsZero() {
		return errors.New("messaging retention clock is required")
	}
	now = now.UTC()
	cutoffs := messagingdomain.RetentionCutoffs{
		ExpiredNoncesBefore:    now.Add(-messagingAliasRetention),
		ConfirmationsExpiredAt: now,
		ProviderDataBefore:     now.Add(-messagingProviderRetention),
		ReplyTokensBefore:      now.Add(-messagingAliasRetention),
	}

	log.Info(ctx, "Purging expired messaging data")
	result, err := drainMessagingRetentionBatches(ctx, store, cutoffs, func(batch int, result messagingdomain.RetentionPurgeResult) {
		recordMessagingRetentionEvents(span, batch, result)
	})
	if err != nil {
		span.RecordError(err)
		return err
	}
	log.Info(ctx, "Purged expired messaging data", "rows_affected", result.TotalAffected())
	return nil
}

func drainMessagingRetentionBatches(
	ctx context.Context,
	store MessagingDataPurger,
	cutoffs messagingdomain.RetentionCutoffs,
	onCommitted func(int, messagingdomain.RetentionPurgeResult),
) (messagingdomain.RetentionPurgeResult, error) {
	if ctx == nil {
		return messagingdomain.RetentionPurgeResult{}, errors.New("messaging retention context is required")
	}
	if store == nil {
		return messagingdomain.RetentionPurgeResult{}, errors.New("messaging retention store is required")
	}

	var total messagingdomain.RetentionPurgeResult
	for batch := 0; batch < maintenancePurgeMaxBatches; batch++ {
		if err := ctx.Err(); err != nil {
			return total, fmt.Errorf(
				"purge messaging retention interrupted after %d rows: %w",
				total.TotalAffected(),
				err,
			)
		}

		result, err := store.PurgeMessagingDataBatch(ctx, cutoffs, maintenancePurgeBatchSize)
		if err != nil {
			return total, fmt.Errorf("purge messaging retention batch: %w", err)
		}
		if err := validateMessagingRetentionResult(result, maintenancePurgeBatchSize); err != nil {
			return total, err
		}
		total = addMessagingRetentionResults(total, result)
		if onCommitted != nil {
			onCommitted(batch+1, result)
		}
		if !messagingRetentionBatchFull(result, maintenancePurgeBatchSize) {
			return total, nil
		}
	}

	return total, fmt.Errorf(
		"purge messaging retention after %d rows: %w",
		total.TotalAffected(),
		errMaintenanceBacklogRemaining,
	)
}

type messagingRetentionCount struct {
	kind string
	rows int64
}

func messagingRetentionCounts(result messagingdomain.RetentionPurgeResult) []messagingRetentionCount {
	return []messagingRetentionCount{
		{kind: "expired_nonces", rows: result.NoncesDeleted},
		{kind: "expired_story_mutation_confirmations", rows: result.ConfirmationsRedacted},
		{kind: "old_deliveries", rows: result.OutboundDeliveriesDeleted},
		{kind: "old_inbound_events", rows: result.InboundEventsDeleted},
		{kind: "completed_slack_uninstalls", rows: result.CompletedSlackUninstallsDeleted},
		{kind: "old_messages", rows: result.MessagesDeleted},
		{kind: "expired_email_reply_tokens", rows: result.ReplyTokensDeleted},
		{kind: "empty_conversations", rows: result.ConversationsDeleted},
	}
}

func validateMessagingRetentionResult(result messagingdomain.RetentionPurgeResult, batchSize int) error {
	for _, count := range messagingRetentionCounts(result) {
		if count.rows < 0 || count.rows > int64(batchSize) {
			return fmt.Errorf(
				"purge messaging retention %s: %w: got %d, want 0..%d",
				count.kind,
				errInvalidMaintenancePurgeResult,
				count.rows,
				batchSize,
			)
		}
	}
	return nil
}

func messagingRetentionBatchFull(result messagingdomain.RetentionPurgeResult, batchSize int) bool {
	for _, count := range messagingRetentionCounts(result) {
		if count.rows == int64(batchSize) {
			return true
		}
	}
	return false
}

func addMessagingRetentionResults(
	total messagingdomain.RetentionPurgeResult,
	result messagingdomain.RetentionPurgeResult,
) messagingdomain.RetentionPurgeResult {
	total.NoncesDeleted += result.NoncesDeleted
	total.ConfirmationsRedacted += result.ConfirmationsRedacted
	total.OutboundDeliveriesDeleted += result.OutboundDeliveriesDeleted
	total.InboundEventsDeleted += result.InboundEventsDeleted
	total.CompletedSlackUninstallsDeleted += result.CompletedSlackUninstallsDeleted
	total.MessagesDeleted += result.MessagesDeleted
	total.ReplyTokensDeleted += result.ReplyTokensDeleted
	total.ConversationsDeleted += result.ConversationsDeleted
	return total
}

func recordMessagingRetentionEvents(
	span trace.Span,
	batch int,
	result messagingdomain.RetentionPurgeResult,
) {
	for _, count := range messagingRetentionCounts(result) {
		span.AddEvent("messaging_data_deleted", trace.WithAttributes(
			attribute.String("kind", count.kind),
			attribute.Int("batch", batch),
			attribute.Int64("rows_affected", count.rows),
		))
	}
}
