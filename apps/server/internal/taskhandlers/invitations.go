package taskhandlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
)

type InvitationOutboxProcessor interface {
	DispatchReady(context.Context) error
}

func (h *handlers) HandleInvitationOutboxDispatch(ctx context.Context, _ *asynq.Task) error {
	if h.invitationOutbox == nil {
		return errors.New("invitation outbox processor is unavailable")
	}
	if err := h.invitationOutbox.DispatchReady(ctx); err != nil {
		return fmt.Errorf("dispatch invitation outbox events: %w", err)
	}
	return nil
}
