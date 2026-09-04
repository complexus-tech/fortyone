package taskhandlers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type GoogleDriveRevocationProcessor interface {
	DispatchPendingRevocations(context.Context) (int, error)
}

func (h *handlers) HandleGoogleDriveRevocationDispatch(ctx context.Context, _ *asynq.Task) error {
	if h.googleDriveRevocations == nil {
		return fmt.Errorf("Google Drive revocation processor is not configured")
	}
	if _, err := h.googleDriveRevocations.DispatchPendingRevocations(ctx); err != nil {
		return fmt.Errorf("dispatch Google Drive revocations: %w", err)
	}
	return nil
}
