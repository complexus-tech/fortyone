package messagingrepository

import (
	"context"
	"errors"
	"fmt"

	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

var errMessagingMaintenanceUnavailable = errors.New("messaging maintenance repository is not configured")

// PurgeMessagingDataBatch applies one bounded, ordered retention batch inside
// a single transaction. Outbound deliveries are removed before inbound events
// that they reference, and messages are removed before empty conversations.
func (repository *Repository) PurgeMessagingDataBatch(
	ctx context.Context,
	cutoffs messagingdomain.RetentionCutoffs,
	batchSize int,
) (messagingdomain.RetentionPurgeResult, error) {
	if !repository.configured() || repository.runTransaction == nil {
		return messagingdomain.RetentionPurgeResult{}, errMessagingMaintenanceUnavailable
	}
	if err := validateMessagingRetentionCutoffs(cutoffs); err != nil {
		return messagingdomain.RetentionPurgeResult{}, err
	}
	if batchSize <= 0 {
		return messagingdomain.RetentionPurgeResult{}, errors.New("messaging retention batch size must be positive")
	}
	databaseBatchSize, err := safecast.Int32(batchSize)
	if err != nil {
		return messagingdomain.RetentionPurgeResult{}, fmt.Errorf("validate messaging retention batch size: %w", err)
	}
	cutoffs = utcMessagingRetentionCutoffs(cutoffs)

	var result messagingdomain.RetentionPurgeResult
	err = repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		result.NoncesDeleted, err = queries.PurgeExpiredMessagingNonces(
			ctx,
			messagingsql.PurgeExpiredMessagingNoncesParams{
				ExpiredBefore: cutoffs.ExpiredNoncesBefore,
				BatchSize:     databaseBatchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("purge expired messaging nonces: %w", err)
		}

		confirmationExpiry := cutoffs.ConfirmationsExpiredAt
		result.ConfirmationsRedacted, err = queries.ExpireMessagingStoryMutationConfirmations(
			ctx,
			messagingsql.ExpireMessagingStoryMutationConfirmationsParams{
				ExpiredAt: &confirmationExpiry,
				BatchSize: databaseBatchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("expire messaging story mutation confirmations: %w", err)
		}

		result.OutboundDeliveriesDeleted, err = queries.PurgeOldMessagingOutboundDeliveries(
			ctx,
			messagingsql.PurgeOldMessagingOutboundDeliveriesParams{
				CreatedBefore: cutoffs.ProviderDataBefore,
				BatchSize:     databaseBatchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("purge old messaging outbound deliveries: %w", err)
		}

		result.InboundEventsDeleted, err = queries.PurgeOldMessagingInboundEvents(
			ctx,
			messagingsql.PurgeOldMessagingInboundEventsParams{
				ReceivedBefore: cutoffs.ProviderDataBefore,
				BatchSize:      databaseBatchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("purge old messaging inbound events: %w", err)
		}

		completedBefore := cutoffs.ProviderDataBefore
		result.CompletedSlackUninstallsDeleted, err = queries.PurgeCompletedSlackUninstalls(
			ctx,
			messagingsql.PurgeCompletedSlackUninstallsParams{
				CompletedBefore: &completedBefore,
				BatchSize:       databaseBatchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("purge completed Slack uninstalls: %w", err)
		}

		result.MessagesDeleted, err = queries.PurgeOldMessagingMessages(
			ctx,
			messagingsql.PurgeOldMessagingMessagesParams{
				CreatedBefore: cutoffs.ProviderDataBefore,
				BatchSize:     databaseBatchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("purge old messaging messages: %w", err)
		}

		result.ReplyTokensDeleted, err = queries.PurgeExpiredMessagingEmailReplyTokens(
			ctx,
			messagingsql.PurgeExpiredMessagingEmailReplyTokensParams{
				RetainedBefore: cutoffs.ReplyTokensBefore,
				BatchSize:      databaseBatchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("purge expired messaging email reply tokens: %w", err)
		}

		result.ConversationsDeleted, err = queries.PurgeEmptyMessagingConversations(
			ctx,
			messagingsql.PurgeEmptyMessagingConversationsParams{
				UpdatedBefore: cutoffs.ProviderDataBefore,
				BatchSize:     databaseBatchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("purge empty messaging conversations: %w", err)
		}
		return nil
	})
	if err != nil {
		return messagingdomain.RetentionPurgeResult{}, fmt.Errorf("purge messaging retention batch: %w", err)
	}
	return result, nil
}

func validateMessagingRetentionCutoffs(cutoffs messagingdomain.RetentionCutoffs) error {
	if cutoffs.ExpiredNoncesBefore.IsZero() ||
		cutoffs.ConfirmationsExpiredAt.IsZero() ||
		cutoffs.ProviderDataBefore.IsZero() ||
		cutoffs.ReplyTokensBefore.IsZero() {
		return errors.New("messaging retention cutoffs are required")
	}
	if !cutoffs.ExpiredNoncesBefore.Before(cutoffs.ConfirmationsExpiredAt) ||
		!cutoffs.ProviderDataBefore.Before(cutoffs.ConfirmationsExpiredAt) ||
		!cutoffs.ReplyTokensBefore.Before(cutoffs.ConfirmationsExpiredAt) {
		return errors.New("messaging retention cutoffs must precede the expiry time")
	}
	return nil
}

func utcMessagingRetentionCutoffs(cutoffs messagingdomain.RetentionCutoffs) messagingdomain.RetentionCutoffs {
	return messagingdomain.RetentionCutoffs{
		ExpiredNoncesBefore:    cutoffs.ExpiredNoncesBefore.UTC(),
		ConfirmationsExpiredAt: cutoffs.ConfirmationsExpiredAt.UTC(),
		ProviderDataBefore:     cutoffs.ProviderDataBefore.UTC(),
		ReplyTokensBefore:      cutoffs.ReplyTokensBefore.UTC(),
	}
}
