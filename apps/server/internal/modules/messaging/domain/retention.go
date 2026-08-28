package messagingdomain

import "time"

// RetentionCutoffs contains the application-clock boundaries used by one
// messaging retention transaction. Keeping the clock outside SQL makes every
// operation in a batch evaluate the same instant and keeps tests deterministic.
type RetentionCutoffs struct {
	ExpiredNoncesBefore    time.Time
	ConfirmationsExpiredAt time.Time
	ProviderDataBefore     time.Time
	ReplyTokensBefore      time.Time
}

// RetentionPurgeResult reports the committed work performed by one atomic
// messaging retention batch. Confirmation rows are redacted rather than
// deleted so their security lifecycle remains auditable.
type RetentionPurgeResult struct {
	NoncesDeleted                   int64
	ConfirmationsRedacted           int64
	OutboundDeliveriesDeleted       int64
	InboundEventsDeleted            int64
	CompletedSlackUninstallsDeleted int64
	MessagesDeleted                 int64
	ReplyTokensDeleted              int64
	ConversationsDeleted            int64
}

// TotalAffected returns the number of rows deleted or redacted by the batch.
func (result RetentionPurgeResult) TotalAffected() int64 {
	return result.NoncesDeleted +
		result.ConfirmationsRedacted +
		result.OutboundDeliveriesDeleted +
		result.InboundEventsDeleted +
		result.CompletedSlackUninstallsDeleted +
		result.MessagesDeleted +
		result.ReplyTokensDeleted +
		result.ConversationsDeleted
}
