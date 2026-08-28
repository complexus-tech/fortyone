package emailreply

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const processorStateWriteTimeout = 5 * time.Second

// processorStateContext gives terminal bookkeeping a short opportunity to
// finish after the task context is cancelled. Detaching without a deadline can
// pin a worker indefinitely while the database or network is unavailable.
func processorStateContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(parent), processorStateWriteTimeout)
}

func (processor *Processor) failOutboundDelivery(ctx context.Context, deliveryID uuid.UUID, cause error) error {
	stateCtx, cancel := processorStateContext(ctx)
	defer cancel()
	return processor.store.FailOutboundDelivery(stateCtx, deliveryID, truncateProcessorError(cause))
}
